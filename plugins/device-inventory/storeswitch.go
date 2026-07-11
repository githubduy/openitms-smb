// storeswitch.go — lưu inventory switch (SNMP) vào CMDB + đọc lại facts switch.
package main

import (
	"database/sql"
	"fmt"
	"time"
)

// storeSwitch upsert 1 switch (kind=switch) + thay toàn bộ facts (iface/neighbor/fdb).
// Sinh di_device_change nếu số cổng/láng giềng đổi so với lần trước (diff gọn).
func storeSwitch(db *sql.DB, inv *SwitchInventory) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now()
	// upsert device theo host (UNIQUE).
	_, err = tx.Exec(`INSERT INTO di_device
		(host, kind, hostname, vendor, model, serial, firmware, location, descr, uptime, first_seen, last_seen, last_status)
		VALUES (?, 'switch', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'ok')
		ON DUPLICATE KEY UPDATE kind='switch', hostname=VALUES(hostname), vendor=VALUES(vendor),
			model=VALUES(model), serial=VALUES(serial), firmware=VALUES(firmware),
			location=VALUES(location), descr=VALUES(descr), uptime=VALUES(uptime),
			last_seen=VALUES(last_seen), last_status='ok'`,
		inv.Host, inv.SysName, inv.Vendor, inv.Model, inv.Serial, inv.Firmware, inv.Location,
		inv.SysDescr, inv.Uptime, now, now)
	if err != nil {
		return 0, err
	}
	var id int64
	if err := tx.QueryRow(`SELECT id FROM di_device WHERE host=?`, inv.Host).Scan(&id); err != nil {
		return 0, err
	}

	// đếm cũ để diff.
	oldIf := countRows(tx, "di_switch_iface", id)
	oldNb := countRows(tx, "di_switch_neighbor", id)

	// thay facts: xóa rồi chèn.
	for _, t := range []string{"di_switch_iface", "di_switch_neighbor", "di_switch_fdb"} {
		if _, err := tx.Exec("DELETE FROM "+t+" WHERE device_id=?", id); err != nil {
			return 0, err
		}
	}
	for _, f := range inv.Ifaces {
		if _, err := tx.Exec(`INSERT INTO di_switch_iface
			(device_id, if_index, name, alias, iftype, speed_mbps, oper, mac) VALUES (?,?,?,?,?,?,?,?)`,
			id, f.Index, f.Name, f.Alias, f.Type, f.Speed, f.Oper, f.MAC); err != nil {
			return 0, err
		}
	}
	for _, n := range inv.Neighbors {
		if _, err := tx.Exec(`INSERT INTO di_switch_neighbor
			(device_id, local_port, remote_name, remote_port, remote_mac) VALUES (?,?,?,?,?)`,
			id, n.LocalPort, n.RemoteName, n.RemotePort, n.RemoteMAC); err != nil {
			return 0, err
		}
	}
	for _, f := range inv.FDB {
		if _, err := tx.Exec(`INSERT INTO di_switch_fdb (device_id, mac, port) VALUES (?,?,?)`,
			id, f.MAC, f.Port); err != nil {
			return 0, err
		}
	}

	// change-tracking gọn.
	if oldIf >= 0 && oldIf != len(inv.Ifaces) {
		recordChange(tx, id, now, "interfaces", fmt.Sprintf("%d → %d cổng", oldIf, len(inv.Ifaces)))
	}
	if oldNb >= 0 && oldNb != len(inv.Neighbors) {
		recordChange(tx, id, now, "neighbors", fmt.Sprintf("%d → %d láng giềng LLDP", oldNb, len(inv.Neighbors)))
	}
	return id, tx.Commit()
}

func countRows(tx *sql.Tx, table string, deviceID int64) int {
	var n int
	if err := tx.QueryRow("SELECT COUNT(*) FROM "+table+" WHERE device_id=?", deviceID).Scan(&n); err != nil {
		return -1
	}
	return n
}

func recordChange(tx *sql.Tx, deviceID int64, ts time.Time, kind, detail string) {
	_, _ = tx.Exec(`INSERT INTO di_device_change (device_id, ts, kind, detail) VALUES (?,?,?,?)`,
		deviceID, ts, kind, detail)
}

// loadSwitchFacts đọc iface/neighbor/fdb cho device switch.
func loadSwitchFacts(db *sql.DB, id int64, det *DeviceDetail) error {
	det.Ifaces = []SwitchIface{}
	det.Neighbors = []SwitchNeighbor{}
	det.FDB = []SwitchFDB{}

	ir, err := db.Query(`SELECT if_index, IFNULL(name,''), IFNULL(alias,''), IFNULL(iftype,0),
		IFNULL(speed_mbps,0), IFNULL(oper,''), IFNULL(mac,'') FROM di_switch_iface
		WHERE device_id=? ORDER BY if_index`, id)
	if err != nil {
		return err
	}
	defer ir.Close()
	for ir.Next() {
		var f SwitchIface
		if err := ir.Scan(&f.Index, &f.Name, &f.Alias, &f.Type, &f.Speed, &f.Oper, &f.MAC); err != nil {
			return err
		}
		det.Ifaces = append(det.Ifaces, f)
	}

	nr, err := db.Query(`SELECT IFNULL(local_port,''), IFNULL(remote_name,''), IFNULL(remote_port,''),
		IFNULL(remote_mac,'') FROM di_switch_neighbor WHERE device_id=? ORDER BY local_port`, id)
	if err != nil {
		return err
	}
	defer nr.Close()
	for nr.Next() {
		var n SwitchNeighbor
		if err := nr.Scan(&n.LocalPort, &n.RemoteName, &n.RemotePort, &n.RemoteMAC); err != nil {
			return err
		}
		det.Neighbors = append(det.Neighbors, n)
	}

	fr, err := db.Query(`SELECT mac, IFNULL(port,0) FROM di_switch_fdb WHERE device_id=? ORDER BY mac`, id)
	if err != nil {
		return err
	}
	defer fr.Close()
	for fr.Next() {
		var f SwitchFDB
		if err := fr.Scan(&f.MAC, &f.Port); err != nil {
			return err
		}
		det.FDB = append(det.FDB, f)
	}
	return nil
}
