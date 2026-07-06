package sdk

import (
	"context"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

// Plugin — interface Go-friendly mà plugin implement (SDK lo phần gRPC/go-plugin).
type Plugin interface {
	// Metadata trả mô tả plugin — PHẢI khớp plugin.yaml (core đối chiếu, lệch → từ chối load).
	Metadata(ctx context.Context) (*pluginv1.Metadata, error)

	// HandleRequest xử lý 1 request từ API động /api/plugins/<name>/<route>.
	HandleRequest(ctx context.Context, req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error)

	// RunTask chạy tác vụ dài; phát log/status qua emit; kết quả trả về được SDK
	// tự gửi thành TaskEvent cuối (result). err != nil → TASK_STATUS_FAILED.
	RunTask(ctx context.Context, spec *pluginv1.TaskSpec, emit TaskEmitter) (*pluginv1.TaskResult, error)

	// Health — core gọi định kỳ; 3 lần fail liên tiếp core sẽ restart plugin.
	Health(ctx context.Context) (*pluginv1.Health, error)
}

// TaskEmitter — plugin phát sự kiện realtime về UI trong lúc RunTask.
type TaskEmitter interface {
	Log(line string)
	Status(s pluginv1.TaskStatus)
}
