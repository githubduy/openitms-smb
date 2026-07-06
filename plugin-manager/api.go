package pluginmanager

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

// CallerFunc — core (patch 0001) cung cấp: rút thông tin user đã authenticate từ
// request. Trả nil → 401. Test/standalone dùng caller giả.
type CallerFunc func(*http.Request) *pluginv1.Caller

// Handler trả http.Handler phục vụ API động: /api/plugins/<name>/<route...>.
// Mount vào router của core với prefix "/api/plugins/".
func (m *Manager) Handler(caller CallerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// chịu được webPath prefix của core (vd /web/api/plugins/...)
		const marker = "/api/plugins/"
		idx := strings.Index(r.URL.Path, marker)
		if idx < 0 {
			http.Error(w, `{"error":"bad plugin path"}`, http.StatusNotFound)
			return
		}
		rest := r.URL.Path[idx+len(marker):]
		name, routePath, ok := strings.Cut(rest, "/")
		if !ok || name == "" {
			http.Error(w, `{"error":"missing plugin name or route"}`, http.StatusNotFound)
			return
		}

		m.mu.RLock()
		in := m.instances[name]
		m.mu.RUnlock()
		if in == nil {
			http.Error(w, `{"error":"unknown plugin"}`, http.StatusNotFound)
			return
		}

		route := matchRoute(in.manifest.Routes, r.Method, routePath)
		if route == nil {
			// route không khai trong manifest = không tồn tại (spec: 404)
			http.Error(w, `{"error":"unknown route"}`, http.StatusNotFound)
			return
		}

		c := caller(r)
		if c == nil {
			http.Error(w, `{"error":"unauthenticated"}`, http.StatusUnauthorized)
			return
		}
		if route.RequireAdmin && !c.GetIsAdmin() {
			http.Error(w, `{"error":"admin required"}`, http.StatusForbidden)
			return
		}

		in.mu.Lock()
		rpc, state := in.rpc, in.state
		in.mu.Unlock()
		if state != "running" || rpc == nil {
			http.Error(w, `{"error":"plugin not running"}`, http.StatusServiceUnavailable)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20)) // 10MB limit
		if err != nil {
			http.Error(w, `{"error":"read body"}`, http.StatusBadRequest)
			return
		}
		headers := map[string]string{}
		for k := range r.Header {
			headers[k] = r.Header.Get(k)
		}
		query := map[string]string{}
		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				query[k] = v[0]
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		resp, err := rpc.HandleRequest(ctx, &pluginv1.HttpRequest{
			Method:  r.Method,
			Path:    routePath,
			Headers: headers,
			Query:   query,
			Body:    body,
			Caller:  c,
		})
		if err != nil {
			m.log.Error("HandleRequest fail", "plugin", name, "route", routePath, "err", err)
			http.Error(w, `{"error":"plugin error"}`, http.StatusBadGateway)
			return
		}
		for k, v := range resp.GetHeaders() {
			w.Header().Set(k, v)
		}
		status := int(resp.GetStatus())
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		_, _ = w.Write(resp.GetBody())
	})
}

// matchRoute so path request với pattern manifest ("exec", "status/{id}") theo segment.
func matchRoute(routes []RouteDef, method, path string) *RouteDef {
	segs := strings.Split(strings.Trim(path, "/"), "/")
	for i := range routes {
		r := &routes[i]
		if !strings.EqualFold(r.Method, method) {
			continue
		}
		pat := strings.Split(strings.Trim(r.Path, "/"), "/")
		if len(pat) != len(segs) {
			continue
		}
		match := true
		for j := range pat {
			if strings.HasPrefix(pat[j], "{") && strings.HasSuffix(pat[j], "}") {
				continue // param segment — khớp mọi giá trị
			}
			if pat[j] != segs[j] {
				match = false
				break
			}
		}
		if match {
			return r
		}
	}
	return nil
}
