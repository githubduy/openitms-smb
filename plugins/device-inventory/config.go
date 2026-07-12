// config.go — cấu hình thu định kỳ (auto-collect). Lưu trong di_config (key/value).
package main

import (
	"database/sql"
	"encoding/json"
	"strconv"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

// Config — cấu hình scheduler.
type Config struct {
	Enabled     bool `json:"enabled"`
	IntervalMin int  `json:"interval_min"` // phút; <=0 = tắt
}

func loadConfig(db *sql.DB) Config {
	c := Config{Enabled: false, IntervalMin: 240} // mặc định 4h, tắt
	rows, err := db.Query(`SELECT k, v FROM di_config WHERE k IN ('auto_enabled','interval_min')`)
	if err != nil {
		return c
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		if rows.Scan(&k, &v) != nil {
			continue
		}
		switch k {
		case "auto_enabled":
			c.Enabled = v == "1"
		case "interval_min":
			if n, e := strconv.Atoi(v); e == nil && n > 0 {
				c.IntervalMin = n
			}
		}
	}
	return c
}

// handleGetConfig GET /config → cấu hình scheduler.
func (p *plugin) handleGetConfig() (*pluginv1.HttpResponse, error) {
	return jsonResp(200, loadConfig(p.db)), nil
}

// handleSetConfig POST /config {enabled, interval_min}.
func (p *plugin) handleSetConfig(req *pluginv1.HttpRequest) (*pluginv1.HttpResponse, error) {
	var c Config
	if err := json.Unmarshal(req.GetBody(), &c); err != nil {
		return jsonResp(400, map[string]string{"error": "body JSON không hợp lệ"}), nil
	}
	if err := saveConfig(p.db, c); err != nil {
		return jsonResp(500, map[string]string{"error": err.Error()}), nil
	}
	return jsonResp(200, map[string]any{"ok": true}), nil
}

func saveConfig(db *sql.DB, c Config) error {
	en := "0"
	if c.Enabled {
		en = "1"
	}
	iv := c.IntervalMin
	if iv <= 0 {
		iv = 240
	}
	for _, kv := range []struct{ k, v string }{
		{"auto_enabled", en},
		{"interval_min", strconv.Itoa(iv)},
	} {
		if _, err := db.Exec(`INSERT INTO di_config (k,v) VALUES (?,?)
			ON DUPLICATE KEY UPDATE v=VALUES(v)`, kv.k, kv.v); err != nil {
			return err
		}
	}
	return nil
}
