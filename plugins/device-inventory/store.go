// store.go — truy vấn CMDB gọn: liệt kê device, chi tiết (facts), lịch sử thay đổi.
package main

import (
	"database/sql"
	"time"
)

type Device struct {
	ID         int64      `json:"id"`
	Host       string     `json:"host"`
	Hostname   string     `json:"hostname"`
	OS         string     `json:"os"`
	OSVersion  string     `json:"os_version"`
	OSBuild    string     `json:"os_build"`
	Domain     string     `json:"domain"`
	FirstSeen  *time.Time `json:"first_seen,omitempty"`
	LastSeen   *time.Time `json:"last_seen,omitempty"`
	LastStatus string     `json:"last_status"`
}

type Software struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type Service struct {
	Name  string `json:"name"`
	State string `json:"state"`
	Start string `json:"start"`
}
type Patch struct {
	KB        string `json:"kb"`
	Installed string `json:"installed"`
}
type Change struct {
	TS     time.Time `json:"ts"`
	Kind   string    `json:"kind"`
	Detail string    `json:"detail"`
}

type DeviceDetail struct {
	Device
	Software []Software `json:"software"`
	Services []Service  `json:"services"`
	Patches  []Patch    `json:"patches"`
}

func listDevices(db *sql.DB) ([]Device, error) {
	rows, err := db.Query(`SELECT id, host, IFNULL(hostname,''), IFNULL(os,''), IFNULL(os_version,''),
		IFNULL(os_build,''), IFNULL(domain,''), first_seen, last_seen, IFNULL(last_status,'')
		FROM di_device ORDER BY host`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Device{}
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.Host, &d.Hostname, &d.OS, &d.OSVersion, &d.OSBuild,
			&d.Domain, &d.FirstSeen, &d.LastSeen, &d.LastStatus); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func getDevice(db *sql.DB, id int64) (*DeviceDetail, error) {
	var d Device
	err := db.QueryRow(`SELECT id, host, IFNULL(hostname,''), IFNULL(os,''), IFNULL(os_version,''),
		IFNULL(os_build,''), IFNULL(domain,''), first_seen, last_seen, IFNULL(last_status,'')
		FROM di_device WHERE id=?`, id).Scan(&d.ID, &d.Host, &d.Hostname, &d.OS, &d.OSVersion,
		&d.OSBuild, &d.Domain, &d.FirstSeen, &d.LastSeen, &d.LastStatus)
	if err != nil {
		return nil, err
	}
	det := &DeviceDetail{Device: d, Software: []Software{}, Services: []Service{}, Patches: []Patch{}}

	swRows, err := db.Query(`SELECT name, IFNULL(version,'') FROM di_device_software WHERE device_id=? ORDER BY name`, id)
	if err != nil {
		return nil, err
	}
	defer swRows.Close()
	for swRows.Next() {
		var s Software
		if err := swRows.Scan(&s.Name, &s.Version); err != nil {
			return nil, err
		}
		det.Software = append(det.Software, s)
	}

	svRows, err := db.Query(`SELECT name, IFNULL(state,''), IFNULL(start,'') FROM di_device_service WHERE device_id=? ORDER BY name`, id)
	if err != nil {
		return nil, err
	}
	defer svRows.Close()
	for svRows.Next() {
		var s Service
		if err := svRows.Scan(&s.Name, &s.State, &s.Start); err != nil {
			return nil, err
		}
		det.Services = append(det.Services, s)
	}

	pRows, err := db.Query(`SELECT kb, IFNULL(installed,'') FROM di_device_patch WHERE device_id=? ORDER BY kb`, id)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()
	for pRows.Next() {
		var p Patch
		if err := pRows.Scan(&p.KB, &p.Installed); err != nil {
			return nil, err
		}
		det.Patches = append(det.Patches, p)
	}
	return det, nil
}

func listChanges(db *sql.DB, id int64) ([]Change, error) {
	rows, err := db.Query(`SELECT ts, kind, IFNULL(detail,'') FROM di_device_change
		WHERE device_id=? ORDER BY ts DESC LIMIT 500`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Change{}
	for rows.Next() {
		var c Change
		if err := rows.Scan(&c.TS, &c.Kind, &c.Detail); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
