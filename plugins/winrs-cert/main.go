// Plugin winrs-cert — chạy lệnh xuống Windows qua WinRS (WinRM) dùng CHỨNG CHỈ
// (certificate auth qua HTTPS), không dùng mật khẩu. Cert lấy từ thư mục ./certs
// của OpenITMS-SMB (certstore). Spec: docs/L2-specs/plugins-winrs-cert.md.
//
// API động:
//   POST /api/plugins/winrs-cert/exec   {host,port?,cert,command,timeout?}
//   GET  /api/plugins/winrs-cert/certs  — liệt kê cert khả dụng trong ./certs
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
	sdk "quickwin.dev/sdk"
	"quickwin.dev/pluginmanager/certstore"
)

const version = "0.1.0"

type plugin struct {
	certs *certstore.Store
}

func (p *plugin) Metadata(_ context.Context) (*pluginv1.Metadata, error) {
	return &pluginv1.Metadata{
		Name:    "winrs-cert",
		Version: version,
		Routes: []*pluginv1.Route{
			{Method: "POST", Path: "exec", Description: "Run a command on Windows via WinRS (certificate auth)", RequireAdmin: true},
			{Method: "GET", Path: "certs", Description: "List certificates available in ./certs"},
		},
		Permissions: []string{"certs:read", "network:outbound"},
	}, nil
}

type execRequest struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Cert    string `json:"cert"`     // tên file cert trong ./certs (.pem chứa cert+key, hoặc .pfx)
	Key     string `json:"key"`      // (optional) file key riêng nếu cert/key tách
	Command string `json:"command"`
	Timeout int    `json:"timeout"`  // giây; 0 = 60
}

type execResponse struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Error    string `json:"error,omitempty"`
}

func (p *plugin) HandleRequest(ctx context.Context, req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	switch req.GetPath() {
	case "certs":
		return p.handleListCerts()
	case "exec":
		return p.handleExec(ctx, req)
	default:
		return jsonResp(404, map[string]string{"error": "unknown route"}), nil
	}
}

func (p *plugin) handleListCerts() (*pluginv1.HttpResponse, error) {
	type certInfo struct {
		Name   string `json:"name"`
		Kind   string `json:"kind"`
		HasKey bool   `json:"has_key"`
		CN     string `json:"cn,omitempty"`
	}
	var out []certInfo
	for _, e := range p.certs.List() {
		ci := certInfo{Name: e.Name, Kind: e.Kind, HasKey: e.HasKey}
		if len(e.Certificates) > 0 {
			ci.CN = e.Certificates[0].Subject.CommonName
		}
		out = append(out, ci)
	}
	return jsonResp(200, map[string]any{"certs": out}), nil
}

func (p *plugin) handleExec(ctx context.Context, req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	var er execRequest
	if err := json.Unmarshal(req.GetBody(), &er); err != nil {
		return jsonResp(400, execResponse{Error: "body JSON không hợp lệ: " + err.Error()}), nil
	}
	if er.Host == "" || er.Command == "" || er.Cert == "" {
		return jsonResp(400, execResponse{Error: "cần host, command, cert"}), nil
	}
	if er.Port == 0 {
		er.Port = 5986 // WinRM HTTPS
	}
	if er.Timeout <= 0 {
		er.Timeout = 60
	}

	certEntry := p.certs.Get(er.Cert)
	if certEntry == nil {
		return jsonResp(404, execResponse{Error: fmt.Sprintf("cert %q không có trong ./certs", er.Cert)}), nil
	}
	pemCert, pemKey, err := resolveCertKey(p.certs, certEntry, er.Key)
	if err != nil {
		return jsonResp(400, execResponse{Error: err.Error()}), nil
	}

	cctx, cancel := context.WithTimeout(ctx, time.Duration(er.Timeout)*time.Second)
	defer cancel()

	res, err := runWinRSCert(cctx, winrsParams{
		Host: er.Host, Port: er.Port, CertPEM: pemCert, KeyPEM: pemKey, Command: er.Command,
		Timeout: time.Duration(er.Timeout) * time.Second,
	})
	if err != nil {
		// phân loại lỗi: cert / mạng / winrm-config (spec) — kèm hint
		return jsonResp(502, execResponse{Error: classifyError(err)}), nil
	}
	return jsonResp(200, execResponse{ExitCode: res.ExitCode, Stdout: res.Stdout, Stderr: res.Stderr}), nil
}

func (p *plugin) RunTask(ctx context.Context, spec *pluginv1.TaskSpec, emit sdk.TaskEmitter) (*pluginv1.TaskResult, error) {
	emit.Status(pluginv1.TaskStatus_TASK_STATUS_RUNNING)
	host := spec.Params["host"]
	cmd := spec.Params["command"]
	cert := spec.Params["cert"]
	if host == "" || cmd == "" || cert == "" {
		return &pluginv1.TaskResult{Status: pluginv1.TaskStatus_TASK_STATUS_FAILED, Message: "cần params host, command, cert"}, nil
	}
	emit.Log(fmt.Sprintf("winrs-cert: %s@%s (cert %s)", cmd, host, cert))
	ce := p.certs.Get(cert)
	if ce == nil {
		return &pluginv1.TaskResult{Status: pluginv1.TaskStatus_TASK_STATUS_FAILED, Message: "cert không có trong ./certs"}, nil
	}
	pemCert, pemKey, err := resolveCertKey(p.certs, ce, spec.Params["key"])
	if err != nil {
		return &pluginv1.TaskResult{Status: pluginv1.TaskStatus_TASK_STATUS_FAILED, Message: err.Error()}, nil
	}
	res, err := runWinRSCert(ctx, winrsParams{Host: host, Port: 5986, CertPEM: pemCert, KeyPEM: pemKey, Command: cmd, Timeout: 60 * time.Second})
	if err != nil {
		return &pluginv1.TaskResult{Status: pluginv1.TaskStatus_TASK_STATUS_FAILED, Message: classifyError(err)}, nil
	}
	emit.Log(res.Stdout)
	data, _ := json.Marshal(res)
	st := pluginv1.TaskStatus_TASK_STATUS_SUCCESS
	if res.ExitCode != 0 {
		st = pluginv1.TaskStatus_TASK_STATUS_FAILED
	}
	return &pluginv1.TaskResult{Status: st, Message: fmt.Sprintf("exit=%d", res.ExitCode), Data: data}, nil
}

func (p *plugin) Health(_ context.Context) (*pluginv1.Health, error) {
	return &pluginv1.Health{Status: pluginv1.Health_STATUS_HEALTHY}, nil
}

func jsonResp(status int32, v any) *pluginv1.HttpResponse {
	body, _ := json.Marshal(v)
	return &pluginv1.HttpResponse{Status: status, Headers: map[string]string{"Content-Type": "application/json"}, Body: body}
}

func main() {
	dir := os.Getenv("QUICKWIN_CERTS_DIR")
	if dir == "" {
		dir = "certs"
	}
	store := certstore.New(dir, 5*time.Second, nil)
	store.Start()
	defer store.Stop()
	sdk.Serve(&plugin{certs: store})
}
