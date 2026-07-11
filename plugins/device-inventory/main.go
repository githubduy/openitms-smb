// Plugin device-inventory — inventory device/asset AGENTLESS: thu qua osquery chạy trên máy đích
// (đẩy qua WinRS), lưu thành CMDB gọn trong MariaDB (database riêng openitms_devices).
// Thay cho inventory tự-viết trong core (0025-0028): device data do plugin + tool chuyên quản.
//
// API động:
//   GET  /api/plugins/device-inventory/devices        — danh sách device
//   GET  /api/plugins/device-inventory/device?id=<id> — chi tiết (software/services/patches)
//   GET  /api/plugins/device-inventory/changes?id=<id>— lịch sử thay đổi
//   POST /api/plugins/device-inventory/collect        — thu inventory 1 host (osquery/WinRS)
//   GET  /api/plugins/device-inventory/export?format= — export toàn fleet
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
	sdk "quickwin.dev/sdk"
)

const version = "0.1.0"

type plugin struct {
	db *sql.DB
}

func (p *plugin) Metadata(_ context.Context) (*pluginv1.Metadata, error) {
	return &pluginv1.Metadata{
		Name:    "device-inventory",
		Version: version,
		Routes: []*pluginv1.Route{
			{Method: "GET", Path: "devices", Description: "List inventoried devices"},
			{Method: "GET", Path: "device", Description: "Device detail (facts) by id"},
			{Method: "GET", Path: "changes", Description: "Change history for a device by id"},
			{Method: "POST", Path: "collect", Description: "Collect inventory from a host via osquery over WinRS", RequireAdmin: true},
			{Method: "POST", Path: "collect-switch", Description: "Collect a network switch via SNMP (v2c/v3)", RequireAdmin: true},
			{Method: "GET", Path: "export", Description: "Export the whole fleet inventory (csv|json)"},
		},
		Permissions: []string{"certs:read", "network:outbound", "inventory:read"},
	}, nil
}

func (p *plugin) HandleRequest(ctx context.Context, req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	if p.db == nil {
		return jsonResp(503, map[string]string{"error": "device-inventory: chưa kết nối được MariaDB (xem log plugin)"}), nil
	}
	switch req.GetPath() {
	case "devices":
		return p.handleDevices()
	case "device":
		return p.handleDevice(req)
	case "changes":
		return p.handleChanges(req)
	case "collect":
		return p.handleCollect(ctx, req)
	case "collect-switch":
		return p.handleCollectSwitch(req)
	case "export":
		return p.handleExport(req)
	default:
		return jsonResp(404, map[string]string{"error": "unknown route"}), nil
	}
}

func (p *plugin) handleDevices() (*pluginv1.HttpResponse, error) {
	devs, err := listDevices(p.db)
	if err != nil {
		return jsonResp(500, map[string]string{"error": err.Error()}), nil
	}
	return jsonResp(200, map[string]any{"devices": devs}), nil
}

func (p *plugin) handleDevice(req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	id, ok := queryID(req)
	if !ok {
		return jsonResp(400, map[string]string{"error": "thiếu ?id"}), nil
	}
	det, err := getDevice(p.db, id)
	if err == sql.ErrNoRows {
		return jsonResp(404, map[string]string{"error": "không có device"}), nil
	}
	if err != nil {
		return jsonResp(500, map[string]string{"error": err.Error()}), nil
	}
	return jsonResp(200, det), nil
}

func (p *plugin) handleChanges(req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	id, ok := queryID(req)
	if !ok {
		return jsonResp(400, map[string]string{"error": "thiếu ?id"}), nil
	}
	ch, err := listChanges(p.db, id)
	if err != nil {
		return jsonResp(500, map[string]string{"error": err.Error()}), nil
	}
	return jsonResp(200, map[string]any{"changes": ch}), nil
}

func (p *plugin) RunTask(_ context.Context, _ *pluginv1.TaskSpec, _ sdk.TaskEmitter) (*pluginv1.TaskResult, error) {
	return &pluginv1.TaskResult{Status: pluginv1.TaskStatus_TASK_STATUS_FAILED, Message: "device-inventory chưa hỗ trợ task runner"}, nil
}

func (p *plugin) Health(_ context.Context) (*pluginv1.Health, error) {
	if p.db == nil {
		return &pluginv1.Health{Status: pluginv1.Health_STATUS_DEGRADED, Message: "MariaDB chưa kết nối"}, nil
	}
	if err := p.db.Ping(); err != nil {
		return &pluginv1.Health{Status: pluginv1.Health_STATUS_DEGRADED, Message: "DB: " + err.Error()}, nil
	}
	return &pluginv1.Health{Status: pluginv1.Health_STATUS_HEALTHY}, nil
}

func queryID(req *pluginv1.HttpRequest) (int64, bool) {
	s := req.GetQuery()["id"]
	if s == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func jsonResp(status int32, body any) *pluginv1.HttpResponse {
	b, _ := json.Marshal(body)
	return &pluginv1.HttpResponse{
		Status:  status,
		Headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
		Body:    b,
	}
}

func main() {
	db, err := openDB()
	if err != nil {
		// Không mở được DB → vẫn Serve nhưng Health báo lỗi (core sẽ log). Panic sẽ làm core coi là crash.
		fmt.Fprintln(os.Stderr, "device-inventory: openDB fail:", err)
		sdk.Serve(&plugin{db: nil})
		return
	}
	sdk.Serve(&plugin{db: db})
}
