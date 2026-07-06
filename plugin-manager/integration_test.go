package pluginmanager

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

// Integration test P1-03/P1-04: build plugin hello THẬT, launch qua go-plugin
// handshake thật, gọi API động, RunTask stream, và health-restart.

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// buildHelloPlugin build plugins/hello vào thư mục plugins tạm + copy manifest.
func buildHelloPlugin(t *testing.T, pluginsRoot string) {
	t.Helper()
	dir := filepath.Join(pluginsRoot, "hello")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go"+exeSuffix())
	out := filepath.Join(dir, "hello"+exeSuffix())
	cmd := exec.Command(goBin, "build", "-o", out, ".")
	cmd.Dir = filepath.Join("..", "plugins", "hello")
	cmd.Env = os.Environ()
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build hello fail: %v\n%s", err, b)
	}
	src, err := os.ReadFile(filepath.Join("..", "plugins", "hello", "plugin.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.yaml"), src, 0o644); err != nil {
		t.Fatal(err)
	}
}

func testCaller(*http.Request) *pluginv1.Caller {
	return &pluginv1.Caller{UserId: 1, Username: "tester", IsAdmin: true}
}

func newTestManager(t *testing.T, opts Options) *Manager {
	t.Helper()
	pluginsRoot := t.TempDir()
	buildHelloPlugin(t, pluginsRoot)
	opts.Dir = pluginsRoot
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	m := New(opts)
	if err := m.Scan(); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(m.Stop)
	return m
}

func TestIntegration_LifecycleAndDynamicAPI(t *testing.T) {
	m := newTestManager(t, Options{})

	states := m.States()
	if len(states) != 1 || states[0].Name != "hello" || states[0].State != "running" {
		t.Fatalf("trạng thái không mong đợi: %+v", states)
	}

	srv := httptest.NewServer(m.Handler(testCaller))
	defer srv.Close()

	// route hợp lệ
	resp, err := http.Post(srv.URL+"/api/plugins/hello/echo", "text/plain", strings.NewReader("xin chào"))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("echo status=%d body=%s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "xin chào") || !strings.Contains(string(body), "tester") {
		t.Fatalf("echo body sai: %s", body)
	}

	// route không khai trong manifest → 404 (spec)
	resp2, err := http.Get(srv.URL + "/api/plugins/hello/not-a-route")
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 404 {
		t.Fatalf("route lạ phải 404, got %d", resp2.StatusCode)
	}

	// plugin không tồn tại → 404
	resp3, err := http.Get(srv.URL + "/api/plugins/nope/x")
	if err != nil {
		t.Fatal(err)
	}
	resp3.Body.Close()
	if resp3.StatusCode != 404 {
		t.Fatalf("plugin lạ phải 404, got %d", resp3.StatusCode)
	}
}

func TestIntegration_RunTaskStream(t *testing.T) {
	m := newTestManager(t, Options{})
	in := m.instances["hello"]

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	stream, err := in.rpc.RunTask(ctx, &pluginv1.TaskSpec{TaskId: "t1", Action: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	logs := 0
	var result *pluginv1.TaskResult
	for {
		ev, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("stream: %v", err)
		}
		switch e := ev.Event.(type) {
		case *pluginv1.TaskEvent_LogLine:
			logs++
		case *pluginv1.TaskEvent_Result:
			result = e.Result
		}
	}
	if logs < 3 {
		t.Fatalf("mong ≥3 log line, got %d", logs)
	}
	if result.GetStatus() != pluginv1.TaskStatus_TASK_STATUS_SUCCESS {
		t.Fatalf("result: %v", result)
	}
}

func TestIntegration_HealthRestart(t *testing.T) {
	m := newTestManager(t, Options{
		HealthInterval:  150 * time.Millisecond,
		HealthFailLimit: 1,
		RestartBackoff:  []time.Duration{50 * time.Millisecond},
	})
	oldPid := m.States()[0].Pid
	if oldPid == 0 {
		t.Fatal("không lấy được pid")
	}
	p, err := os.FindProcess(oldPid)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Kill(); err != nil {
		t.Fatal(err)
	}

	deadline := time.After(10 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("plugin không tự restart: %+v", m.States())
		case <-time.After(100 * time.Millisecond):
		}
		s := m.States()[0]
		if s.State == "running" && s.Pid != 0 && s.Pid != oldPid && s.Restarts >= 1 {
			return // restart thành công với pid mới
		}
	}
}

var _ = fmt.Sprintf // giữ import khi chỉnh test
