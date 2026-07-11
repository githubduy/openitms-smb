// collect.go — thu inventory: osquery/WinRS (host, Phase 2) + SNMP (switch). export (Phase 3).
package main

import (
	"context"
	"encoding/json"
	"fmt"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

type collectHostReq struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Cert    string `json:"cert"` // tên file .pem (cert+key) trong ./certs
	Key     string `json:"key"`  // (optional) file key riêng
	Timeout int    `json:"timeout"`
}

// handleCollect thu 1 host qua osquery/WinRS rồi lưu CMDB.
func (p *plugin) handleCollect(_ context.Context, req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	var r collectHostReq
	if err := json.Unmarshal(req.GetBody(), &r); err != nil {
		return jsonResp(400, map[string]string{"error": "body JSON không hợp lệ: " + err.Error()}), nil
	}
	if r.Host == "" || r.Cert == "" {
		return jsonResp(400, map[string]string{"error": "cần host và cert"}), nil
	}
	if r.Port == 0 {
		r.Port = 5986
	}
	if r.Timeout <= 0 {
		r.Timeout = 120
	}
	certPEM, keyPEM, err := p.resolveCert(r.Cert, r.Key)
	if err != nil {
		return jsonResp(400, map[string]string{"error": err.Error()}), nil
	}
	inv, err := collectHost(HostCollectConfig{
		Host: r.Host, Port: r.Port, CertPEM: certPEM, KeyPEM: keyPEM, Timeout: r.Timeout,
	})
	if err != nil {
		return jsonResp(502, map[string]string{"error": err.Error()}), nil
	}
	id, err := storeHost(p.db, inv)
	if err != nil {
		return jsonResp(500, map[string]string{"error": err.Error()}), nil
	}
	return jsonResp(200, map[string]any{
		"ok": true, "device_id": id, "hostname": inv.Hostname, "os": inv.OS,
		"software": len(inv.Software), "services": len(inv.Services), "patches": len(inv.Patches),
	}), nil
}

// resolveCert lấy cert+key PEM từ certstore (.pem chứa cả cert lẫn key, hoặc key file riêng).
func (p *plugin) resolveCert(cert, key string) (certPEM, keyPEM []byte, err error) {
	e := p.certs.Get(cert)
	if e == nil {
		return nil, nil, fmt.Errorf("cert %q không có trong ./certs", cert)
	}
	if e.Kind == "pfx" {
		return nil, nil, fmt.Errorf("cert %q là .pfx — dùng .pem (cert+key)", cert)
	}
	if key != "" {
		ke := p.certs.Get(key)
		if ke == nil {
			return nil, nil, fmt.Errorf("key %q không có trong ./certs", key)
		}
		return e.Raw, ke.Raw, nil
	}
	if !e.HasKey {
		return nil, nil, fmt.Errorf("cert %q không chứa private key — cung cấp thêm 'key'", cert)
	}
	return e.Raw, e.Raw, nil
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
