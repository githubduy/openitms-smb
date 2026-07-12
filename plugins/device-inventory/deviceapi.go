// deviceapi.go — thêm/xóa device (thực thể trung tâm) với thông tin kết nối. Unify model:
// device = identity + cách kết nối + tài sản thu thập.
package main

import (
	"encoding/json"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

// handleAddDevice tạo/cập nhật 1 device kèm thông tin kết nối (chưa thu). POST /device.
func (p *plugin) handleAddDevice(req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	var c DeviceConn
	if err := json.Unmarshal(req.GetBody(), &c); err != nil {
		return jsonResp(400, map[string]string{"error": "body JSON không hợp lệ: " + err.Error()}), nil
	}
	switch c.ConnType {
	case "local", "winrs", "snmp":
	default:
		return jsonResp(400, map[string]string{"error": "conn_type phải là local|winrs|snmp"}), nil
	}
	id, err := upsertDeviceConn(p.db, c)
	if err != nil {
		return jsonResp(500, map[string]string{"error": err.Error()}), nil
	}
	return jsonResp(200, map[string]any{"ok": true, "id": id}), nil
}

// handleDeleteDevice xóa 1 device. DELETE /device?id=<id>.
func (p *plugin) handleDeleteDevice(req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	id, ok := queryID(req)
	if !ok {
		return jsonResp(400, map[string]string{"error": "thiếu ?id"}), nil
	}
	if err := deleteDevice(p.db, id); err != nil {
		return jsonResp(500, map[string]string{"error": err.Error()}), nil
	}
	return jsonResp(200, map[string]any{"ok": true}), nil
}
