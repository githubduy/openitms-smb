//go:build python_integration

// Chạy: cần python3 + grpcio + stub sinh sẵn + PYTHONPATH (xem CI job python-plugin).
//   go test -tags python_integration -run TestPython ./...
// Chứng minh: plugin Python được Plugin Manager launch Y HỆT plugin Go (cùng proto contract).
package pluginmanager

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

func TestPython_HelloPluginThroughManager(t *testing.T) {
	root, _ := filepath.Abs("..")
	pluginsRoot := t.TempDir()
	dst := filepath.Join(pluginsRoot, "hello-py")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	// copy main.py + plugin.yaml của plugin hello-py
	for _, f := range []string{"main.py", "plugin.yaml"} {
		src := filepath.Join(root, "plugins", "hello-py", f)
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, f), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	m := New(Options{Dir: pluginsRoot, Logger: slog.New(slog.NewTextHandler(os.Stderr, nil))})
	if err := m.Scan(); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Start (python cần grpcio + PYTHONPATH): %v", err)
	}
	t.Cleanup(m.Stop)

	if s := m.States(); len(s) != 1 || s[0].Name != "hello-py" || s[0].State != "running" {
		t.Fatalf("python plugin không running: %+v", s)
	}

	srv := httptest.NewServer(m.Handler(func(*http.Request) *pluginv1.Caller {
		return &pluginv1.Caller{UserId: 1, Username: "tester", IsAdmin: true}
	}))
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/api/plugins/hello-py/echo", "text/plain", strings.NewReader("xin chào python"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("echo status %d", resp.StatusCode)
	}
	var body [512]byte
	n, _ := resp.Body.Read(body[:])
	s := string(body[:n])
	if !strings.Contains(s, "xin chào python") || !strings.Contains(s, `"lang":"python"`) {
		t.Fatalf("python echo sai: %s", s)
	}
}
