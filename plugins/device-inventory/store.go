// store.go — truy vấn CMDB gọn: liệt kê device, chi tiết (facts), lịch sử thay đổi.
package main

import (
	"database/sql"
	"time"
)

type Device struct {
	ID         int64      `json:"id"`
	Host       string     `json:"host"`
	Kind       string     `json:"kind"` // host | switch
	Hostname   string     `json:"hostname"`
	OS         string     `json:"os"`
	OSVersion  string     `json:"os_version"`
	OSBuild    string     `json:"os_build"`
	Domain     string     `json:"domain"`
	Vendor     string     `json:"vendor,omitempty"`
	Model      string     `json:"model,omitempty"`
	Serial     string     `json:"serial,omitempty"`
	Firmware   string     `json:"firmware,omitempty"`
	Location   string     `json:"location,omitempty"`
	Uptime     string     `json:"uptime,omitempty"`
	FirstSeen  *time.Time `json:"first_seen,omitempty"`
	LastSeen   *time.Time `json:"last_seen,omitempty"`
	LastStatus string     `json:"last_status"`
}

// cột SELECT chung cho di_device (thứ tự khớp scanDevice).
const deviceCols = `id, host, IFNULL(kind,'host'), IFNULL(hostname,''), IFNULL(os,''),
	IFNULL(os_version,''), IFNULL(os_build,''), IFNULL(domain,''), IFNULL(vendor,''),
	IFNULL(model,''), IFNULL(serial,''), IFNULL(firmware,''), IFNULL(location,''),
	IFNULL(uptime,''), first_seen, last_seen, IFNULL(last_status,'')`

func scanDevice(s interface {
	Scan(...any) error
}, d *Device) error {
	return s.Scan(&d.ID, &d.Host, &d.Kind, &d.Hostname, &d.OS, &d.OSVersion, &d.OSBuild, &d.Domain,
		&d.Vendor, &d.Model, &d.Serial, &d.Firmware, &d.Location, &d.Uptime,
		&d.FirstSeen, &d.LastSeen, &d.LastStatus)
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
type DeviceFact struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Detail   string `json:"detail"`
}
type Change struct {
	TS     time.Time `json:"ts"`
	Kind   string    `json:"kind"`
	Detail string    `json:"detail"`
}

type DeviceDetail struct {
	Device
	Software  []Software       `json:"software"`
	Services  []Service        `json:"services"`
	Patches   []Patch          `json:"patches"`
	Facts     []DeviceFact     `json:"facts,omitempty"`
	Ifaces    []SwitchIface    `json:"interfaces,omitempty"`
	Neighbors []SwitchNeighbor `json:"neighbors,omitempty"`
	FDB       []SwitchFDB      `json:"fdb,omitempty"`
}

func listDevices(db *sql.DB) ([]Device, error) {
	rows, err := db.Query(`SELECT ` + deviceCols + ` FROM di_device ORDER BY host`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Device{}
	for rows.Next() {
		var d Device
		if err := scanDevice(rows, &d); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func getDevice(db *sql.DB, id int64) (*DeviceDetail, error) {
	var d Device
	if err := scanDevice(db.QueryRow(`SELECT `+deviceCols+` FROM di_device WHERE id=?`, id), &d); err != nil {
		return nil, err
	}
	det := &DeviceDetail{Device: d, Software: []Software{}, Services: []Service{}, Patches: []Patch{}}
	if d.Kind == "switch" {
		if err := loadSwitchFacts(db, id, det); err != nil {
			return nil, err
		}
		return det, nil
	}

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

	det.Facts = []DeviceFact{}
	fRows, err := db.Query(`SELECT category, name, IFNULL(detail,'') FROM di_device_fact
		WHERE device_id=? ORDER BY category, name`, id)
	if err != nil {
		return nil, err
	}
	defer fRows.Close()
	for fRows.Next() {
		var f DeviceFact
		if err := fRows.Scan(&f.Category, &f.Name, &f.Detail); err != nil {
			return nil, err
		}
		det.Facts = append(det.Facts, f)
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
