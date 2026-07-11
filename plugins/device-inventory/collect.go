// collect.go — thu inventory qua osquery/WinRS + export. (Phase 2/3: sẽ hiện thực đầy đủ.)
package main

import (
	"context"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

func (p *plugin) handleCollect(_ context.Context, _ *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	return jsonResp(501, map[string]string{"error": "collect chưa hiện thực (Phase 2: osquery/WinRS)"}), nil
}

func (p *plugin) handleExport(_ *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	return jsonResp(501, map[string]string{"error": "export chưa hiện thực (Phase 3)"}), nil
}
