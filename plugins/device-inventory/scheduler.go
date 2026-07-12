// scheduler.go — thu định kỳ tự động: lặp mọi device có kết nối đã lưu + thu lại (không nhập creds).
package main

import (
	"fmt"
	"os"
	"time"
)

// doCollect thu 1 device theo THÔNG TIN KẾT NỐI đã lưu (dispatch theo conn_type), lưu CMDB.
// Dùng chung cho HTTP (collectByID) và scheduler.
func (p *plugin) doCollect(c *DeviceConn, autoDeploy bool) (string, error) {
	switch c.ConnType {
	case "local":
		inv, err := collectHostLocal(autoDeploy, os.Getenv("QUICKWIN_OSQUERY_MSI"), 300)
		if err != nil {
			return "", err
		}
		inv.Host = c.Host
		_, err = storeHost(p.db, inv)
		return "host", err
	case "winrs":
		certPEM, keyPEM, err := p.resolveCert(c.ConnCert, "")
		if err != nil {
			return "", err
		}
		port := c.ConnPort
		if port == 0 {
			port = 5986
		}
		inv, err := collectHost(HostCollectConfig{
			Host: c.Host, Port: port, CertPEM: certPEM, KeyPEM: keyPEM, Timeout: 300,
			AutoDeploy: autoDeploy, MSIURL: os.Getenv("QUICKWIN_OSQUERY_MSI"),
		})
		if err != nil {
			return "", err
		}
		_, err = storeHost(p.db, inv)
		return "host", err
	case "snmp":
		inv, err := collectSwitch(SNMPConfig{
			Host: c.Host, Port: uint16(c.ConnPort), Version: c.SNMPVersion, Community: c.SNMPCommunity,
		})
		if err != nil {
			return "", err
		}
		_, err = storeSwitch(p.db, inv)
		return "switch", err
	default:
		return "", fmt.Errorf("device chưa có thông tin kết nối (conn_type)")
	}
}

// runScheduler vòng lặp nền: mỗi phút kiểm tra, tới hạn thì thu tất cả device có kết nối.
func (p *plugin) runScheduler() {
	var last time.Time
	for {
		time.Sleep(60 * time.Second)
		if p.db == nil {
			continue
		}
		cfg := loadConfig(p.db)
		if !cfg.Enabled || cfg.IntervalMin <= 0 {
			continue
		}
		if !last.IsZero() && time.Since(last) < time.Duration(cfg.IntervalMin)*time.Minute {
			continue
		}
		last = time.Now()
		p.collectAllDue()
	}
}

// collectAllDue thu mọi device có conn_type (bỏ qua lỗi từng máy, chỉ log).
func (p *plugin) collectAllDue() {
	rows, err := p.db.Query(`SELECT id FROM di_device WHERE conn_type IS NOT NULL AND conn_type <> ''`)
	if err != nil {
		return
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		c, err := loadDeviceConn(p.db, id)
		if err != nil {
			continue
		}
		if _, e := p.doCollect(c, true); e != nil {
			fmt.Fprintf(os.Stderr, "device-inventory scheduler: device %d (%s): %v\n", id, c.Host, e)
		}
	}
}
