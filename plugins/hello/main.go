// Plugin hello — mẫu tối thiểu chứng minh hợp đồng SDK/proto.
// Được dùng làm fixture cho integration test của plugin-manager.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
	sdk "quickwin.dev/sdk"
)

const version = "0.1.0"

type helloPlugin struct{}

func (h *helloPlugin) Metadata(_ context.Context) (*pluginv1.Metadata, error) {
	return &pluginv1.Metadata{
		Name:    "hello",
		Version: version,
		Routes: []*pluginv1.Route{
			{Method: "POST", Path: "echo", Description: "Echo request body back"},
			{Method: "GET", Path: "info", Description: "Plugin info"},
		},
	}, nil
}

func (h *helloPlugin) HandleRequest(_ context.Context, req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	switch req.GetPath() {
	case "echo":
		body, _ := json.Marshal(map[string]any{
			"echo":   string(req.GetBody()),
			"caller": req.GetCaller().GetUsername(),
		})
		return jsonResp(200, body), nil
	case "info":
		body, _ := json.Marshal(map[string]string{"name": "hello", "version": version})
		return jsonResp(200, body), nil
	default:
		return jsonResp(404, []byte(`{"error":"unknown route"}`)), nil
	}
}

func (h *helloPlugin) RunTask(ctx context.Context, spec *pluginv1.TaskSpec, emit sdk.TaskEmitter) (*pluginv1.TaskResult, error) {
	emit.Status(pluginv1.TaskStatus_TASK_STATUS_RUNNING)
	n := 3
	for i := 1; i <= n; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
		emit.Log(fmt.Sprintf("hello task %s: step %d/%d", spec.GetTaskId(), i, n))
	}
	return &pluginv1.TaskResult{Status: pluginv1.TaskStatus_TASK_STATUS_SUCCESS, Message: "done"}, nil
}

func (h *helloPlugin) Health(_ context.Context) (*pluginv1.Health, error) {
	return &pluginv1.Health{Status: pluginv1.Health_STATUS_HEALTHY}, nil
}

func jsonResp(status int32, body []byte) *pluginv1.HttpResponse {
	return &pluginv1.HttpResponse{
		Status:  status,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    body,
	}
}

func main() { sdk.Serve(&helloPlugin{}) }
