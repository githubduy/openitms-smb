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
// resolved — kết quả định tuyến + phân quyền 1 request tới plugin.
type resolved struct {
	rpc  pluginv1.PluginClient
	path string
	c    *pluginv1.Caller
}

// resolve tách toàn bộ nhánh 404/401/403/503 ra khỏi Handler (giảm cognitive complexity).
// Trả (nil, code, msg) nếu request bị từ chối; (r, 0, "") nếu OK.
func (m *Manager) resolve(r *http.Request, caller CallerFunc) (*resolved, int, string) {
	const marker = "/api/plugins/"
	idx := strings.Index(r.URL.Path, marker)
	if idx < 0 {
		return nil, http.StatusNotFound, "bad plugin path"
	}
	name, routePath, ok := strings.Cut(r.URL.Path[idx+len(marker):], "/")
	if !ok || name == "" {
		return nil, http.StatusNotFound, "missing plugin name or route"
	}

	m.mu.RLock()
	in := m.instances[name]
	m.mu.RUnlock()
	if in == nil {
		return nil, http.StatusNotFound, "unknown plugin"
	}

	route := matchRoute(in.manifest.Routes, r.Method, routePath)
	if route == nil {
		return nil, http.StatusNotFound, "unknown route" // route ngoài manifest = không tồn tại
	}

	c := caller(r)
	if c == nil {
		return nil, http.StatusUnauthorized, "unauthenticated"
	}
	if route.RequireAdmin && !c.GetIsAdmin() {
		return nil, http.StatusForbidden, "admin required"
	}

	in.mu.Lock()
	rpc, state := in.rpc, in.state
	in.mu.Unlock()
	if state != "running" || rpc == nil {
		return nil, http.StatusServiceUnavailable, "plugin not running"
	}
	return &resolved{rpc: rpc, path: routePath, c: c}, 0, ""
}

func (m *Manager) Handler(caller CallerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, code, msg := m.resolve(r, caller)
		if res == nil {
			http.Error(w, `{"error":"`+msg+`"}`, code)
			return
		}
		preq, err := buildPluginRequest(r, res)
		if err != nil {
			http.Error(w, `{"error":"read body"}`, http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		resp, err := res.rpc.HandleRequest(ctx, preq)
		if err != nil {
			m.log.Error("HandleRequest fail", "route", res.path, "err", err)
			http.Error(w, `{"error":"plugin error"}`, http.StatusBadGateway)
			return
		}
		writePluginResponse(w, resp)
	})
}

// buildPluginRequest dựng HttpRequest gRPC từ http.Request (đọc body + copy headers/query).
func buildPluginRequest(r *http.Request, res *resolved) (*pluginv1.HttpRequest, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20)) // 10MB limit
	if err != nil {
		return nil, err
	}
	headers := make(map[string]string, len(r.Header))
	for k := range r.Header {
		headers[k] = r.Header.Get(k)
	}
	query := map[string]string{}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}
	return &pluginv1.HttpRequest{
		Method: r.Method, Path: res.path, Headers: headers, Query: query, Body: body, Caller: res.c,
	}, nil
}

// writePluginResponse ghi HttpResponse của plugin ra http.ResponseWriter.
func writePluginResponse(w http.ResponseWriter, resp *pluginv1.HttpResponse) {
	for k, v := range resp.GetHeaders() {
		w.Header().Set(k, v)
	}
	status := int(resp.GetStatus())
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write(resp.GetBody())
}

// matchRoute so path request với pattern manifest ("exec", "status/{id}") theo segment.
func matchRoute(routes []RouteDef, method, path string) *RouteDef {
	segs := strings.Split(strings.Trim(path, "/"), "/")
	for i := range routes {
		r := &routes[i]
		if strings.EqualFold(r.Method, method) && segmentsMatch(strings.Split(strings.Trim(r.Path, "/"), "/"), segs) {
			return r
		}
	}
	return nil
}

// segmentsMatch: pattern khớp path theo từng segment; "{param}" khớp mọi giá trị.
func segmentsMatch(pat, segs []string) bool {
	if len(pat) != len(segs) {
		return false
	}
	for j := range pat {
		isParam := strings.HasPrefix(pat[j], "{") && strings.HasSuffix(pat[j], "}")
		if !isParam && pat[j] != segs[j] {
			return false
		}
	}
	return true
}
