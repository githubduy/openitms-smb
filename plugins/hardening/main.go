// Plugin hardening — API động:
//   GET  /api/plugins/hardening/scan        → danh sách finding (pass + fail)
//   POST /api/plugins/hardening/fix (admin)  {check_id} → chmod về mode an toàn
package main

import (
	"context"
	"encoding/json"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
	sdk "quickwin.dev/sdk"
)

const version = "0.1.0"

type plugin struct{ cfg scanConfig }

func (p *plugin) Metadata(_ context.Context) (*pluginv1.Metadata, error) {
	return &pluginv1.Metadata{
		Name:    "hardening",
		Version: version,
		Routes: []*pluginv1.Route{
			{Method: "GET", Path: "scan", Description: "Scan host security configuration"},
			{Method: "POST", Path: "fix", Description: "Fix a fixable finding (file permissions)", RequireAdmin: true},
		},
		Permissions: []string{"settings:read"},
	}, nil
}

func (p *plugin) HandleRequest(_ context.Context, req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	switch req.GetPath() {
	case "scan":
		findings := runChecks(p.cfg)
		fails := 0
		for _, f := range findings {
			if !f.Passed {
				fails++
			}
		}
		return resp(200, map[string]any{"findings": findings, "issues": fails, "total": len(findings)}), nil
	case "fix":
		var body struct {
			CheckID string `json:"check_id"`
		}
		if err := json.Unmarshal(req.GetBody(), &body); err != nil {
			return resp(400, map[string]string{"error": "body JSON không hợp lệ"}), nil
		}
		if err := applyFix(p.cfg, body.CheckID); err != nil {
			return resp(400, map[string]string{"error": err.Error()}), nil
		}
		return resp(200, map[string]string{"fixed": body.CheckID}), nil
	default:
		return resp(404, map[string]string{"error": "unknown route"}), nil
	}
}

func (p *plugin) RunTask(_ context.Context, spec *pluginv1.TaskSpec, emit sdk.TaskEmitter) (*pluginv1.TaskResult, error) {
	emit.Status(pluginv1.TaskStatus_TASK_STATUS_RUNNING)
	findings := runChecks(p.cfg)
	fails := 0
	for _, f := range findings {
		if !f.Passed {
			fails++
			emit.Log(string(f.Severity) + ": " + f.Title + " — " + f.Detail)
		}
	}
	st := pluginv1.TaskStatus_TASK_STATUS_SUCCESS
	if fails > 0 {
		st = pluginv1.TaskStatus_TASK_STATUS_FAILED // có vấn đề → task "đỏ" để chú ý
	}
	return &pluginv1.TaskResult{Status: st, Message: itoa(fails) + " vấn đề bảo mật", Data: mustJSON(findings)}, nil
}

func (p *plugin) Health(_ context.Context) (*pluginv1.Health, error) {
	return &pluginv1.Health{Status: pluginv1.Health_STATUS_HEALTHY}, nil
}

func resp(status int32, v any) *pluginv1.HttpResponse {
	return &pluginv1.HttpResponse{Status: status,
		Headers: map[string]string{"Content-Type": "application/json"}, Body: mustJSON(v)}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func main() { sdk.Serve(&plugin{cfg: configFromEnv()}) }
