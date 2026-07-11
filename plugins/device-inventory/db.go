// db.go — kết nối MariaDB dùng chung với OpenITMS (đọc DSN từ config.json). Dùng CHÍNH database
// của app + prefix bảng "di_" (user app không có quyền CREATE DATABASE trên bản cài mặc định).
// Namespace rõ ràng, backup chung, không cần quyền root. Tự migrate schema lúc start.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// tablePrefix — mọi bảng của plugin mang tiền tố này để tách khỏi bảng core trong cùng database.
const tablePrefix = "di_"

// mysqlConfig — phần MySQL trong config.json của OpenITMS.
type mysqlConfig struct {
	Host string `json:"host"` // "127.0.0.1:3306"
	User string `json:"user"`
	Pass string `json:"pass"`
	Name string `json:"name"`
}

// loadMySQLConfig đọc config.json (đường dẫn qua QUICKWIN_CONFIG) lấy khối mysql.
func loadMySQLConfig() (mysqlConfig, error) {
	path := os.Getenv("QUICKWIN_CONFIG")
	if path == "" {
		return mysqlConfig{}, fmt.Errorf("QUICKWIN_CONFIG chưa set — không xác định được DB")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return mysqlConfig{}, fmt.Errorf("đọc config.json: %w", err)
	}
	var cfg struct {
		Dialect string      `json:"dialect"`
		MySQL   mysqlConfig `json:"mysql"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return mysqlConfig{}, fmt.Errorf("parse config.json: %w", err)
	}
	if cfg.Dialect != "mysql" || cfg.MySQL.Host == "" {
		return mysqlConfig{}, fmt.Errorf("device-inventory cần OpenITMS chạy MariaDB/MySQL (dialect=%q)", cfg.Dialect)
	}
	return cfg.MySQL, nil
}

// openDB kết nối vào chính database của app (name từ config.json) + migrate bảng di_*.
func openDB() (*sql.DB, error) {
	c, err := loadMySQLConfig()
	if err != nil {
		return nil, err
	}
	if c.Name == "" {
		return nil, fmt.Errorf("config.json thiếu mysql.name")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true", c.User, c.Pass, c.Host, c.Name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(5)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// migrate tạo schema (idempotent). Schema CMDB gọn: 1 device + các bảng fact + lịch sử thay đổi.
func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS di_device (
			id           BIGINT PRIMARY KEY AUTO_INCREMENT,
			host         VARCHAR(255) NOT NULL UNIQUE,
			hostname     VARCHAR(255),
			os           VARCHAR(255),
			os_version   VARCHAR(128),
			os_build     VARCHAR(64),
			domain       VARCHAR(255),
			first_seen   DATETIME,
			last_seen    DATETIME,
			last_status  VARCHAR(32)
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS di_device_software (
			device_id BIGINT NOT NULL,
			name      VARCHAR(512) NOT NULL,
			version   VARCHAR(128),
			INDEX (device_id),
			FOREIGN KEY (device_id) REFERENCES di_device(id) ON DELETE CASCADE
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS di_device_service (
			device_id BIGINT NOT NULL,
			name      VARCHAR(255) NOT NULL,
			state     VARCHAR(64),
			start     VARCHAR(64),
			INDEX (device_id),
			FOREIGN KEY (device_id) REFERENCES di_device(id) ON DELETE CASCADE
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS di_device_patch (
			device_id  BIGINT NOT NULL,
			kb         VARCHAR(64) NOT NULL,
			installed  VARCHAR(32),
			INDEX (device_id),
			FOREIGN KEY (device_id) REFERENCES di_device(id) ON DELETE CASCADE
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS di_device_change (
			id        BIGINT PRIMARY KEY AUTO_INCREMENT,
			device_id BIGINT NOT NULL,
			ts        DATETIME NOT NULL,
			kind      VARCHAR(64) NOT NULL,
			detail    TEXT,
			INDEX (device_id),
			FOREIGN KEY (device_id) REFERENCES di_device(id) ON DELETE CASCADE
		) ENGINE=InnoDB`,
		// Network switch (SNMP): cột bổ sung trên di_device + bảng facts riêng.
		`ALTER TABLE di_device
			ADD COLUMN IF NOT EXISTS kind     VARCHAR(32) NOT NULL DEFAULT 'host',
			ADD COLUMN IF NOT EXISTS vendor   VARCHAR(64),
			ADD COLUMN IF NOT EXISTS model    VARCHAR(255),
			ADD COLUMN IF NOT EXISTS serial   VARCHAR(255),
			ADD COLUMN IF NOT EXISTS firmware VARCHAR(255),
			ADD COLUMN IF NOT EXISTS location VARCHAR(255),
			ADD COLUMN IF NOT EXISTS descr    TEXT,
			ADD COLUMN IF NOT EXISTS uptime   VARCHAR(64)`,
		`CREATE TABLE IF NOT EXISTS di_switch_iface (
			device_id  BIGINT NOT NULL,
			if_index   INT NOT NULL,
			name       VARCHAR(255),
			alias      VARCHAR(255),
			iftype     INT,
			speed_mbps BIGINT,
			oper       VARCHAR(32),
			mac        VARCHAR(32),
			INDEX (device_id),
			FOREIGN KEY (device_id) REFERENCES di_device(id) ON DELETE CASCADE
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS di_switch_neighbor (
			device_id   BIGINT NOT NULL,
			local_port  VARCHAR(128),
			remote_name VARCHAR(255),
			remote_port VARCHAR(255),
			remote_mac  VARCHAR(32),
			INDEX (device_id),
			FOREIGN KEY (device_id) REFERENCES di_device(id) ON DELETE CASCADE
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS di_switch_fdb (
			device_id BIGINT NOT NULL,
			mac       VARCHAR(32) NOT NULL,
			port      INT,
			INDEX (device_id),
			FOREIGN KEY (device_id) REFERENCES di_device(id) ON DELETE CASCADE
		) ENGINE=InnoDB`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
