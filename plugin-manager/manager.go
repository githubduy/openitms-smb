package pluginmanager

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
	sdk "quickwin.dev/sdk"
)

// Options — cấu hình Manager. Zero value dùng được (default hợp lý).
type Options struct {
	Dir             string          // thư mục /plugins (bắt buộc)
	CoreVersion     string          // để so min_core_version; "" = bỏ qua check
	HealthInterval  time.Duration   // default 30s
	HealthFailLimit int             // default 3 lần fail liên tiếp → restart
	RestartBackoff  []time.Duration // default 1s,5s,25s (lặp phần tử cuối)
	Logger          *slog.Logger    // default slog.Default()
}

func (o *Options) fill() {
	if o.HealthInterval <= 0 {
		o.HealthInterval = 30 * time.Second
	}
	if o.HealthFailLimit <= 0 {
		o.HealthFailLimit = 3
	}
	if len(o.RestartBackoff) == 0 {
		o.RestartBackoff = []time.Duration{time.Second, 5 * time.Second, 25 * time.Second}
	}
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
}

// InstanceState — trạng thái 1 plugin cho UI/API/test.
type InstanceState struct {
	Name     string
	Version  string
	State    string // running | failed | stopped
	Pid      int
	Restarts int
	LastErr  string
}

type instance struct {
	manifest *Manifest

	mu       sync.Mutex
	client   *goplugin.Client
	rpc      pluginv1.PluginClient
	logFile  *os.File
	state    string
	pid      int
	restarts int
	lastErr  string
}

// kill dừng process + đóng log file (Windows giữ lock nếu quên đóng). Gọi khi giữ in.mu.
func (in *instance) killLocked() {
	if in.client != nil {
		in.client.Kill()
		in.client = nil
	}
	if in.logFile != nil {
		_ = in.logFile.Close()
		in.logFile = nil
	}
	in.rpc = nil
}

// Manager — quét, chạy và trông coi vòng đời mọi plugin.
type Manager struct {
	opts      Options
	log       *slog.Logger
	mu        sync.RWMutex
	instances map[string]*instance
	stop      chan struct{}
	wg        sync.WaitGroup
}

func New(opts Options) *Manager {
	opts.fill()
	return &Manager{
		opts:      opts,
		log:       opts.Logger,
		instances: map[string]*instance{},
		stop:      make(chan struct{}),
	}
}

// Scan tìm mọi thư mục con của opts.Dir có plugin.yaml, validate manifest + checksum.
// Plugin hỏng bị BỎ QUA (log error) — không chặn plugin khác; core không có plugin
// nào vẫn chạy bình thường (degrade gracefully).
func (m *Manager) Scan() error {
	entries, err := os.ReadDir(m.opts.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			m.log.Info("thư mục plugins không tồn tại — chạy không plugin", "dir", m.opts.Dir)
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(m.opts.Dir, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "plugin.yaml")); err != nil {
			continue // không phải thư mục plugin
		}
		mf := m.evaluatePlugin(dir) // nil = bị bỏ qua (đã log lý do)
		if mf == nil {
			continue
		}
		m.mu.Lock()
		m.instances[mf.Name] = &instance{manifest: mf, state: "stopped"}
		m.mu.Unlock()
		m.log.Info("phát hiện plugin", "plugin", mf.Name, "version", mf.Version)
	}
	return nil
}

// evaluatePlugin validate 1 thư mục plugin. Trả manifest nếu hợp lệ + được nhận;
// nil (kèm log lý do) nếu bỏ qua — plugin hỏng KHÔNG chặn plugin khác.
func (m *Manager) evaluatePlugin(dir string) *Manifest {
	mf, err := LoadManifest(dir)
	if err != nil {
		m.log.Error("bỏ qua plugin: manifest lỗi", "dir", dir, "err", err)
		return nil
	}
	if mf.ProtocolVersion != sdk.ProtocolVersion {
		m.log.Error("bỏ qua plugin: protocol_version không khớp", "plugin", mf.Name,
			"plugin_protocol", mf.ProtocolVersion, "core_protocol", sdk.ProtocolVersion)
		return nil
	}
	if m.opts.CoreVersion != "" && mf.MinCoreVersion != "" &&
		semverLess(m.opts.CoreVersion, mf.MinCoreVersion) {
		m.log.Error("bỏ qua plugin: cần core mới hơn", "plugin", mf.Name,
			"min_core", mf.MinCoreVersion, "core", m.opts.CoreVersion)
		return nil
	}
	warns, err := verifyChecksums(mf)
	if err != nil {
		m.log.Error("bỏ qua plugin: checksum fail", "plugin", mf.Name, "err", err)
		return nil
	}
	for _, w := range warns {
		m.log.Warn(w)
	}
	m.mu.RLock()
	_, dup := m.instances[mf.Name]
	m.mu.RUnlock()
	if dup {
		m.log.Error("bỏ qua plugin: trùng tên", "plugin", mf.Name, "dir", dir)
		return nil
	}
	return mf
}

// Start launch mọi plugin đã Scan + chạy health loop. Trả error tổng hợp nhưng
// plugin launch fail không chặn plugin khác.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.RLock()
	insts := make([]*instance, 0, len(m.instances))
	for _, in := range m.instances {
		insts = append(insts, in)
	}
	m.mu.RUnlock()

	var firstErr error
	for _, in := range insts {
		if err := m.launch(ctx, in); err != nil {
			m.log.Error("launch fail", "plugin", in.manifest.Name, "err", err)
			in.mu.Lock()
			in.state, in.lastErr = "failed", err.Error()
			in.mu.Unlock()
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		m.wg.Add(1)
		go m.healthLoop(in)
	}
	return firstErr
}

// Stop kill mọi plugin + dừng health loop.
func (m *Manager) Stop() {
	close(m.stop)
	m.mu.RLock()
	for _, in := range m.instances {
		in.mu.Lock()
		in.killLocked()
		in.state = "stopped"
		in.mu.Unlock()
	}
	m.mu.RUnlock()
	m.wg.Wait()
}

// States — snapshot trạng thái mọi plugin (cho UI + test).
func (m *Manager) States() []InstanceState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]InstanceState, 0, len(m.instances))
	for _, in := range m.instances {
		in.mu.Lock()
		out = append(out, InstanceState{
			Name: in.manifest.Name, Version: in.manifest.Version,
			State: in.state, Pid: in.pid, Restarts: in.restarts, LastErr: in.lastErr,
		})
		in.mu.Unlock()
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (m *Manager) launch(ctx context.Context, in *instance) error {
	mf := in.manifest
	cmd, err := entrypointCmd(mf)
	if err != nil {
		return err
	}

	logFile, err := os.OpenFile(filepath.Join(mf.Dir, "plugin.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("mở plugin.log: %w", err)
	}

	client := goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig:  sdk.Handshake,
		Plugins:          map[string]goplugin.Plugin{sdk.PluginName: &sdk.GRPCPlugin{}},
		Cmd:              cmd,
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
		Logger: hclog.New(&hclog.LoggerOptions{
			Name: mf.Name, Output: logFile, Level: hclog.Info,
		}),
	})
	// mọi lỗi từ đây phải Kill + đóng logFile (Windows giữ lock file mở)
	fail := func(format string, a ...any) error {
		client.Kill()
		_ = logFile.Close()
		return fmt.Errorf(format, a...)
	}

	rpcClient, err := client.Client()
	if err != nil {
		return fail("handshake fail: %w", err)
	}
	raw, err := rpcClient.Dispense(sdk.PluginName)
	if err != nil {
		return fail("dispense fail: %w", err)
	}
	rpc, ok := raw.(pluginv1.PluginClient)
	if !ok {
		return fail("plugin trả type không mong đợi %T", raw)
	}

	mctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	meta, err := rpc.GetMetadata(mctx, &pluginv1.GetMetadataRequest{})
	if err != nil {
		return fail("GetMetadata fail: %w", err)
	}
	// Chống "manifest nói dối": metadata gRPC phải khớp manifest (spec plugin-manifest.md)
	if meta.GetName() != mf.Name || meta.GetVersion() != mf.Version {
		return fail("metadata lệch manifest: gRPC=%s@%s manifest=%s@%s",
			meta.GetName(), meta.GetVersion(), mf.Name, mf.Version)
	}
	if err := routesMatch(mf.Routes, meta.GetRoutes()); err != nil {
		return fail("routes lệch manifest: %w", err)
	}

	pid := 0
	if rc := client.ReattachConfig(); rc != nil {
		pid = rc.Pid
	}
	in.mu.Lock()
	in.killLocked() // dọn tài nguyên lần chạy trước (nếu có)
	in.client, in.rpc, in.logFile = client, rpc, logFile
	in.state, in.pid, in.lastErr = "running", pid, ""
	in.mu.Unlock()
	m.log.Info("plugin chạy", "plugin", mf.Name, "version", mf.Version, "pid", pid)
	return nil
}

func (m *Manager) healthLoop(in *instance) {
	defer m.wg.Done()
	fails := 0
	t := time.NewTicker(m.opts.HealthInterval)
	defer t.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-t.C:
		}
		in.mu.Lock()
		rpc, state := in.rpc, in.state
		in.mu.Unlock()
		if state != "running" || rpc == nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		h, err := rpc.HealthCheck(ctx, &pluginv1.HealthCheckRequest{})
		cancel()
		switch {
		case err != nil || h.GetStatus() == pluginv1.Health_STATUS_UNHEALTHY:
			fails++
			m.log.Warn("health fail", "plugin", in.manifest.Name, "fails", fails, "err", err)
		case h.GetStatus() == pluginv1.Health_STATUS_DEGRADED:
			fails = 0
			m.log.Warn("plugin degraded", "plugin", in.manifest.Name, "msg", h.GetMessage())
		default:
			fails = 0
		}
		if fails >= m.opts.HealthFailLimit {
			fails = 0
			m.restart(in)
		}
	}
}

func (m *Manager) restart(in *instance) {
	in.mu.Lock()
	n := in.restarts
	in.restarts++
	in.killLocked()
	in.state = "failed"
	in.mu.Unlock()

	bo := m.opts.RestartBackoff
	d := bo[min(n, len(bo)-1)]
	m.log.Warn("restart plugin", "plugin", in.manifest.Name, "attempt", n+1, "backoff", d)
	select {
	case <-m.stop:
		return
	case <-time.After(d):
	}
	if err := m.launch(context.Background(), in); err != nil {
		m.log.Error("restart fail", "plugin", in.manifest.Name, "err", err)
		in.mu.Lock()
		in.state, in.lastErr = "failed", err.Error()
		in.mu.Unlock()
	}
}

// entrypointCmd chọn binary theo platform; fallback entrypoint "python" (cần python3 trên PATH).
func entrypointCmd(mf *Manifest) (*exec.Cmd, error) {
	key := runtime.GOOS + "-" + runtime.GOARCH
	if bin, ok := mf.Entrypoint[key]; ok {
		return exec.Command(filepath.Join(mf.Dir, bin)), nil
	}
	if script, ok := mf.Entrypoint["python"]; ok {
		py, err := exec.LookPath("python3")
		if err != nil {
			py, err = exec.LookPath("python")
		}
		if err != nil {
			return nil, fmt.Errorf("plugin %s cần python3 nhưng không tìm thấy trên PATH", mf.Name)
		}
		return exec.Command(py, filepath.Join(mf.Dir, script)), nil
	}
	return nil, fmt.Errorf("plugin %s không có entrypoint cho platform %s", mf.Name, key)
}

func routesMatch(manifest []RouteDef, meta []*pluginv1.Route) error {
	set := func(method, path string) string { return method + " " + path }
	want := map[string]bool{}
	for _, r := range manifest {
		want[set(r.Method, r.Path)] = true
	}
	got := map[string]bool{}
	for _, r := range meta {
		got[set(r.GetMethod(), r.GetPath())] = true
	}
	if len(want) != len(got) {
		return fmt.Errorf("số route khác nhau: manifest=%d gRPC=%d", len(want), len(got))
	}
	for k := range want {
		if !got[k] {
			return fmt.Errorf("route %q có trong manifest nhưng plugin không phục vụ", k)
		}
	}
	return nil
}

// semverLess: a < b theo semver rút gọn (major.minor.patch, bỏ pre-release).
func semverLess(a, b string) bool {
	pa, pb := semverParts(a), semverParts(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] < pb[i]
		}
	}
	return false
}

func semverParts(v string) [3]int {
	var out [3]int
	fmt.Sscanf(v, "%d.%d.%d", &out[0], &out[1], &out[2])
	return out
}
