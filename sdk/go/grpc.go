package sdk

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

// GRPCPlugin — implement plugin.GRPCPlugin của hashicorp/go-plugin.
// Server side (plugin process): Impl != nil. Client side (core): Impl = nil,
// GRPCClient trả về pluginv1.PluginClient thô cho Plugin Manager dùng trực tiếp.
type GRPCPlugin struct {
	plugin.Plugin
	Impl Plugin
}

func (p *GRPCPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	pluginv1.RegisterPluginServer(s, &grpcServer{impl: p.Impl})
	return nil
}

func (p *GRPCPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return pluginv1.NewPluginClient(c), nil
}

// grpcServer — adapter: pluginv1.PluginServer → interface Plugin của SDK.
type grpcServer struct {
	pluginv1.UnimplementedPluginServer
	impl Plugin
}

func (s *grpcServer) GetMetadata(ctx context.Context, _ *pluginv1.GetMetadataRequest) (*pluginv1.Metadata, error) {
	return s.impl.Metadata(ctx)
}

func (s *grpcServer) HandleRequest(ctx context.Context, req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	return s.impl.HandleRequest(ctx, req)
}

func (s *grpcServer) HealthCheck(ctx context.Context, _ *pluginv1.HealthCheckRequest) (*pluginv1.Health, error) {
	return s.impl.Health(ctx)
}

func (s *grpcServer) RunTask(spec *pluginv1.TaskSpec, stream pluginv1.Plugin_RunTaskServer) error {
	em := &streamEmitter{taskID: spec.GetTaskId(), stream: stream}
	result, err := s.impl.RunTask(stream.Context(), spec, em)
	if err != nil {
		result = &pluginv1.TaskResult{Status: pluginv1.TaskStatus_TASK_STATUS_FAILED, Message: err.Error()}
	}
	if result == nil {
		result = &pluginv1.TaskResult{Status: pluginv1.TaskStatus_TASK_STATUS_SUCCESS}
	}
	// TaskEvent cuối BẮT BUỘC là result (spec proto-contract.md)
	return stream.Send(&pluginv1.TaskEvent{
		TaskId: spec.GetTaskId(),
		Event:  &pluginv1.TaskEvent_Result{Result: result},
	})
}

type streamEmitter struct {
	taskID string
	stream pluginv1.Plugin_RunTaskServer
}

func (e *streamEmitter) Log(line string) {
	_ = e.stream.Send(&pluginv1.TaskEvent{TaskId: e.taskID, Event: &pluginv1.TaskEvent_LogLine{LogLine: line}})
}

func (e *streamEmitter) Status(st pluginv1.TaskStatus) {
	_ = e.stream.Send(&pluginv1.TaskEvent{TaskId: e.taskID, Event: &pluginv1.TaskEvent_Status{Status: st}})
}
