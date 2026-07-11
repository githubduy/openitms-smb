// collect.go — thu inventory: osquery/WinRS (host, Phase 2) + SNMP (switch). export (Phase 3).
package main

import (
	"context"
	"encoding/json"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

func (p *plugin) handleCollect(_ context.Context, _ *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	return jsonResp(501, map[string]string{"error": "collect (host/osquery) chưa hiện thực (Phase 2)"}), nil
}

// handleCollectSwitch thu 1 switch qua SNMP rồi lưu vào CMDB.
// POST {host, version, community | user/auth_*/priv_*, port?} → {ok, device_id, summary}.
func (p *plugin) handleCollectSwitch(req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	var cfg SNMPConfig
	if err := json.Unmarshal(req.GetBody(), &cfg); err != nil {
		return jsonResp(400, map[string]string{"error": "body JSON không hợp lệ: " + err.Error()}), nil
	}
	if cfg.Host == "" {
		return jsonResp(400, map[string]string{"error": "thiếu host"}), nil
	}
	inv, err := collectSwitch(cfg)
	if err != nil {
		return jsonResp(502, map[string]string{"error": err.Error()}), nil
	}
	id, err := storeSwitch(p.db, inv)
	if err != nil {
		return jsonResp(500, map[string]string{"error": err.Error()}), nil
	}
	return jsonResp(200, map[string]any{
		"ok":         true,
		"device_id":  id,
		"sysname":    inv.SysName,
		"vendor":     inv.Vendor,
		"model":      inv.Model,
		"serial":     inv.Serial,
		"interfaces": len(inv.Ifaces),
		"neighbors":  len(inv.Neighbors),
		"fdb":        len(inv.FDB),
	}), nil
}

func (p *plugin) handleExport(_ *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	return jsonResp(501, map[string]string{"error": "export chưa hiện thực (Phase 3)"}), nil
}
