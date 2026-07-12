// storehost.go — lưu inventory HOST (osquery) vào CMDB + diff phần mềm/dịch vụ (change-tracking).
package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func storeHost(db *sql.DB, inv *HostInventory) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.Exec(`INSERT INTO di_device
		(host, kind, hostname, os, os_version, os_build, vendor, model, first_seen, last_seen, last_status)
		VALUES (?, 'host', ?, ?, ?, ?, ?, ?, ?, ?, 'ok')
		ON DUPLICATE KEY UPDATE kind='host', hostname=VALUES(hostname), os=VALUES(os),
			os_version=VALUES(os_version), os_build=VALUES(os_build), vendor=VALUES(vendor),
			model=VALUES(model), last_seen=VALUES(last_seen), last_status='ok'`,
		inv.Host, inv.Hostname, inv.OS, inv.OSVersion, inv.OSBuild, inv.Vendor, inv.Model, now, now)
	if err != nil {
		return 0, err
	}
	var id int64
	if err := tx.QueryRow(`SELECT id FROM di_device WHERE host=?`, inv.Host).Scan(&id); err != nil {
		return 0, err
	}

	// diff phần mềm: đọc bộ cũ trước khi xóa.
	oldSW := loadSoftwareSet(tx, id)
	oldSvc := countRows(tx, "di_device_software", id) // dùng chung helper countRows
	_ = oldSvc

	for _, t := range []string{"di_device_software", "di_device_service", "di_device_patch", "di_device_fact"} {
		if _, err := tx.Exec("DELETE FROM "+t+" WHERE device_id=?", id); err != nil {
			return 0, err
		}
	}
	for _, f := range inv.Facts {
		if _, err := tx.Exec(`INSERT INTO di_device_fact (device_id, category, name, detail) VALUES (?,?,?,?)`,
			id, f.Category, f.Name, f.Detail); err != nil {
			return 0, err
		}
	}
	for _, s := range inv.Software {
		if _, err := tx.Exec(`INSERT INTO di_device_software (device_id, name, version) VALUES (?,?,?)`,
			id, s.Name, s.Version); err != nil {
			return 0, err
		}
	}
	for _, s := range inv.Services {
		if _, err := tx.Exec(`INSERT INTO di_device_service (device_id, name, state, start) VALUES (?,?,?,?)`,
			id, s.Name, s.State, s.Start); err != nil {
			return 0, err
		}
	}
	for _, p := range inv.Patches {
		if _, err := tx.Exec(`INSERT INTO di_device_patch (device_id, kb, installed) VALUES (?,?,?)`,
			id, p.KB, p.Installed); err != nil {
			return 0, err
		}
	}

	// change-tracking: software mới cài / gỡ (chỉ khi đã có bộ cũ).
	if len(oldSW) > 0 {
		added, removed := diffNames(oldSW, softwareNames(inv.Software))
		if len(added) > 0 {
			recordChange(tx, id, now, "software+", fmt.Sprintf("cài: %s", joinCap(added, 20)))
		}
		if len(removed) > 0 {
			recordChange(tx, id, now, "software-", fmt.Sprintf("gỡ: %s", joinCap(removed, 20)))
		}
	}
	return id, tx.Commit()
}

func loadSoftwareSet(tx *sql.Tx, id int64) map[string]bool {
	set := map[string]bool{}
	rows, err := tx.Query(`SELECT name FROM di_device_software WHERE device_id=?`, id)
	if err != nil {
		return set
	}
	defer rows.Close()
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			set[n] = true
		}
	}
	return set
}

func softwareNames(sw []Software) map[string]bool {
	set := map[string]bool{}
	for _, s := range sw {
		set[s.Name] = true
	}
	return set
}

func diffNames(old, cur map[string]bool) (added, removed []string) {
	for n := range cur {
		if !old[n] {
			added = append(added, n)
		}
	}
	for n := range old {
		if !cur[n] {
			removed = append(removed, n)
		}
	}
	return added, removed
}

func joinCap(names []string, cap int) string {
	if len(names) > cap {
		return strings.Join(names[:cap], ", ") + fmt.Sprintf(" … (+%d)", len(names)-cap)
	}
	return strings.Join(names, ", ")
}
