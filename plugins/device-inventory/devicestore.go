// devicestore.go — quản lý device như thực thể trung tâm: lưu THÔNG TIN KẾT NỐI cùng device
// (conn_type + cert/community/port) để Collect dùng lại, không phải nhập mỗi lần. Unify model.
package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// DeviceConn — cách kết nối để thu 1 device.
type DeviceConn struct {
	ID            int64  `json:"id"`
	Host          string `json:"host"`
	ConnType      string `json:"conn_type"` // local | winrs | snmp
	ConnCert      string `json:"conn_cert"`
	ConnPort      int    `json:"conn_port"`
	SNMPVersion   string `json:"snmp_version"`
	SNMPCommunity string `json:"snmp_community"`
}

// kindForConn suy ra kind từ conn_type (snmp → switch, còn lại → host).
func kindForConn(t string) string {
	if t == "snmp" {
		return "switch"
	}
	return "host"
}

// upsertDeviceConn tạo/cập nhật device với thông tin kết nối (chưa thu dữ liệu). host UNIQUE.
func upsertDeviceConn(db *sql.DB, c DeviceConn) (int64, error) {
	host := strings.TrimSpace(c.Host)
	if c.ConnType == "local" && host == "" {
		host = "localhost"
	}
	if host == "" {
		return 0, fmt.Errorf("thiếu host")
	}
	kind := kindForConn(c.ConnType)
	now := time.Now()
	_, err := db.Exec(`INSERT INTO di_device
		(host, kind, conn_type, conn_cert, conn_port, snmp_version, snmp_community, first_seen, last_status)
		VALUES (?,?,?,?,?,?,?,?, 'new')
		ON DUPLICATE KEY UPDATE kind=VALUES(kind), conn_type=VALUES(conn_type), conn_cert=VALUES(conn_cert),
			conn_port=VALUES(conn_port), snmp_version=VALUES(snmp_version), snmp_community=VALUES(snmp_community)`,
		host, kind, c.ConnType, c.ConnCert, c.ConnPort, c.SNMPVersion, c.SNMPCommunity, now)
	if err != nil {
		return 0, err
	}
	var id int64
	err = db.QueryRow(`SELECT id FROM di_device WHERE host=?`, host).Scan(&id)
	return id, err
}

// loadDeviceConn nạp thông tin kết nối của 1 device (để Collect dispatch).
func loadDeviceConn(db *sql.DB, id int64) (*DeviceConn, error) {
	var c DeviceConn
	err := db.QueryRow(`SELECT id, host, IFNULL(conn_type,''), IFNULL(conn_cert,''), IFNULL(conn_port,0),
		IFNULL(snmp_version,''), IFNULL(snmp_community,'') FROM di_device WHERE id=?`, id).
		Scan(&c.ID, &c.Host, &c.ConnType, &c.ConnCert, &c.ConnPort, &c.SNMPVersion, &c.SNMPCommunity)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// deleteDevice xóa 1 device (cascade facts).
func deleteDevice(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM di_device WHERE id=?`, id)
	return err
}
