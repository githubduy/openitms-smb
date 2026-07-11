// snmp.go — thu inventory network switch/router qua SNMP (gosnmp). Hỗ trợ v2c (community) + v3
// (user/auth/priv). Thu: system, hardware (ENTITY-MIB), interfaces (IF-MIB), LLDP neighbors,
// MAC/FDB table (BRIDGE-MIB) — "full + topology".
package main

import (
	"fmt"
	"strings"
	"time"

	g "github.com/gosnmp/gosnmp"
)

// SNMPConfig — tham số kết nối 1 thiết bị.
type SNMPConfig struct {
	Host      string `json:"host"`
	Port      uint16 `json:"port"`     // 0 → 161
	Version   string `json:"version"`  // "v2c" | "v3"
	Community string `json:"community"` // v2c
	// v3:
	User     string `json:"user"`
	AuthProto string `json:"auth_proto"` // MD5 | SHA | "" (noAuth)
	AuthPass  string `json:"auth_pass"`
	PrivProto string `json:"priv_proto"` // DES | AES | "" (noPriv)
	PrivPass  string `json:"priv_pass"`
}

// SwitchIface — 1 cổng/interface.
type SwitchIface struct {
	Index  int    `json:"index"`
	Name   string `json:"name"`
	Alias  string `json:"alias"`
	Type   int    `json:"type"`
	Speed  uint64 `json:"speed_mbps"`
	Oper   string `json:"oper"`
	MAC    string `json:"mac"`
}

// SwitchNeighbor — láng giềng LLDP (topology).
type SwitchNeighbor struct {
	LocalPort  string `json:"local_port"`
	RemoteName string `json:"remote_name"`
	RemotePort string `json:"remote_port"`
	RemoteMAC  string `json:"remote_chassis"`
}

// SwitchFDB — 1 dòng bảng MAC (endpoint nào ở port nào).
type SwitchFDB struct {
	MAC  string `json:"mac"`
	Port int    `json:"port"`
}

// SwitchInventory — kết quả thu 1 switch.
type SwitchInventory struct {
	Host      string           `json:"host"`
	SysName   string           `json:"sysname"`
	SysDescr  string           `json:"sysdescr"`
	Location  string           `json:"location"`
	Contact   string           `json:"contact"`
	Vendor    string           `json:"vendor"`
	Model     string           `json:"model"`
	Serial    string           `json:"serial"`
	Firmware  string           `json:"firmware"`
	Uptime    string           `json:"uptime"`
	Ifaces    []SwitchIface    `json:"interfaces"`
	Neighbors []SwitchNeighbor `json:"neighbors"`
	FDB       []SwitchFDB      `json:"fdb"`
}

// oid chuẩn.
const (
	oidSysDescr    = "1.3.6.1.2.1.1.1.0"
	oidSysObjectID = "1.3.6.1.2.1.1.2.0"
	oidSysUpTime   = "1.3.6.1.2.1.1.3.0"
	oidSysContact  = "1.3.6.1.2.1.1.4.0"
	oidSysName     = "1.3.6.1.2.1.1.5.0"
	oidSysLocation = "1.3.6.1.2.1.1.6.0"

	oidIfDescr       = "1.3.6.1.2.1.2.2.1.2"
	oidIfType        = "1.3.6.1.2.1.2.2.1.3"
	oidIfPhysAddress = "1.3.6.1.2.1.2.2.1.6"
	oidIfOperStatus  = "1.3.6.1.2.1.2.2.1.8"
	oidIfName        = "1.3.6.1.2.1.31.1.1.1.1"
	oidIfHighSpeed   = "1.3.6.1.2.1.31.1.1.1.15"
	oidIfAlias       = "1.3.6.1.2.1.31.1.1.1.18"

	oidEntModel    = "1.3.6.1.2.1.47.1.1.1.1.13"
	oidEntSerial   = "1.3.6.1.2.1.47.1.1.1.1.11"
	oidEntFirmware = "1.3.6.1.2.1.47.1.1.1.1.9"

	oidLldpRemChassisID = "1.0.8802.1.1.2.1.4.1.1.5"
	oidLldpRemPortID    = "1.0.8802.1.1.2.1.4.1.1.7"
	oidLldpRemSysName   = "1.0.8802.1.1.2.1.4.1.1.9"

	oidFdbAddress = "1.3.6.1.2.1.17.4.3.1.1"
	oidFdbPort    = "1.3.6.1.2.1.17.4.3.1.2"
)

func newSNMP(cfg SNMPConfig) (*g.GoSNMP, error) {
	port := cfg.Port
	if port == 0 {
		port = 161
	}
	client := &g.GoSNMP{
		Target:  cfg.Host,
		Port:    port,
		Timeout: 5 * time.Second,
		Retries: 1,
		MaxOids: 60,
	}
	switch cfg.Version {
	case "v3":
		client.Version = g.Version3
		client.SecurityModel = g.UserSecurityModel
		usm := &g.UsmSecurityParameters{UserName: cfg.User}
		flags := g.NoAuthNoPriv
		if cfg.AuthPass != "" {
			flags = g.AuthNoPriv
			usm.AuthenticationProtocol = authProto(cfg.AuthProto)
			usm.AuthenticationPassphrase = cfg.AuthPass
			if cfg.PrivPass != "" {
				flags = g.AuthPriv
				usm.PrivacyProtocol = privProto(cfg.PrivProto)
				usm.PrivacyPassphrase = cfg.PrivPass
			}
		}
		client.MsgFlags = flags
		client.SecurityParameters = usm
	default: // v2c
		client.Version = g.Version2c
		community := cfg.Community
		if community == "" {
			community = "public"
		}
		client.Community = community
	}
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("SNMP connect %s: %w", cfg.Host, err)
	}
	return client, nil
}

func authProto(s string) g.SnmpV3AuthProtocol {
	switch strings.ToUpper(s) {
	case "SHA", "SHA1":
		return g.SHA
	case "SHA256":
		return g.SHA256
	default:
		return g.MD5
	}
}

func privProto(s string) g.SnmpV3PrivProtocol {
	switch strings.ToUpper(s) {
	case "AES", "AES128":
		return g.AES
	case "AES256":
		return g.AES256
	default:
		return g.DES
	}
}

// collectSwitch thu toàn bộ inventory 1 switch qua SNMP.
func collectSwitch(cfg SNMPConfig) (*SwitchInventory, error) {
	client, err := newSNMP(cfg)
	if err != nil {
		return nil, err
	}
	defer client.Conn.Close()

	inv := &SwitchInventory{Host: cfg.Host}

	// 1) system scalars
	sys, err := client.Get([]string{oidSysDescr, oidSysName, oidSysLocation, oidSysContact, oidSysUpTime})
	if err != nil {
		return nil, fmt.Errorf("SNMP get system %s: %w", cfg.Host, err)
	}
	for _, v := range sys.Variables {
		switch {
		case strings.HasPrefix(v.Name[1:], oidSysDescr):
			inv.SysDescr = toStr(v)
		case strings.HasPrefix(v.Name[1:], oidSysName):
			inv.SysName = toStr(v)
		case strings.HasPrefix(v.Name[1:], oidSysLocation):
			inv.Location = toStr(v)
		case strings.HasPrefix(v.Name[1:], oidSysContact):
			inv.Contact = toStr(v)
		case strings.HasPrefix(v.Name[1:], oidSysUpTime):
			inv.Uptime = fmtUptime(g.ToBigInt(v.Value).Int64())
		}
	}
	inv.Vendor = guessVendor(inv.SysDescr)

	// 2) hardware (lấy giá trị đầu tiên khác rỗng của ENTITY-MIB)
	inv.Model = firstNonEmptyWalk(client, oidEntModel)
	inv.Serial = firstNonEmptyWalk(client, oidEntSerial)
	inv.Firmware = firstNonEmptyWalk(client, oidEntFirmware)

	// 3) interfaces
	inv.Ifaces = collectIfaces(client)
	// 4) LLDP neighbors
	inv.Neighbors = collectNeighbors(client)
	// 5) FDB (MAC table)
	inv.FDB = collectFDB(client)

	return inv, nil
}

func collectIfaces(client *g.GoSNMP) []SwitchIface {
	byIdx := map[int]*SwitchIface{}
	walkInto := func(oid string, fn func(idx int, v g.SnmpPDU)) {
		_ = client.BulkWalk(oid, func(v g.SnmpPDU) error {
			idx := lastIndex(v.Name)
			if idx < 0 {
				return nil
			}
			if byIdx[idx] == nil {
				byIdx[idx] = &SwitchIface{Index: idx}
			}
			fn(idx, v)
			return nil
		})
	}
	walkInto(oidIfDescr, func(i int, v g.SnmpPDU) { byIdx[i].Name = toStr(v) })
	walkInto(oidIfName, func(i int, v g.SnmpPDU) {
		if s := toStr(v); s != "" {
			byIdx[i].Name = s
		}
	})
	walkInto(oidIfAlias, func(i int, v g.SnmpPDU) { byIdx[i].Alias = toStr(v) })
	walkInto(oidIfType, func(i int, v g.SnmpPDU) { byIdx[i].Type = int(g.ToBigInt(v.Value).Int64()) })
	walkInto(oidIfHighSpeed, func(i int, v g.SnmpPDU) { byIdx[i].Speed = g.ToBigInt(v.Value).Uint64() })
	walkInto(oidIfOperStatus, func(i int, v g.SnmpPDU) { byIdx[i].Oper = operStatus(int(g.ToBigInt(v.Value).Int64())) })
	walkInto(oidIfPhysAddress, func(i int, v g.SnmpPDU) { byIdx[i].MAC = macFromBytes(v) })

	out := make([]SwitchIface, 0, len(byIdx))
	for _, iface := range byIdx {
		out = append(out, *iface)
	}
	return out
}

func collectNeighbors(client *g.GoSNMP) []SwitchNeighbor {
	byKey := map[string]*SwitchNeighbor{}
	walk := func(oid string, fn func(key string, v g.SnmpPDU)) {
		_ = client.BulkWalk(oid, func(v g.SnmpPDU) error {
			key := indexSuffix(v.Name, oid)
			if key == "" {
				return nil
			}
			if byKey[key] == nil {
				byKey[key] = &SwitchNeighbor{LocalPort: localPortFromLldpIndex(key)}
			}
			fn(key, v)
			return nil
		})
	}
	walk(oidLldpRemSysName, func(k string, v g.SnmpPDU) { byKey[k].RemoteName = toStr(v) })
	walk(oidLldpRemPortID, func(k string, v g.SnmpPDU) { byKey[k].RemotePort = toStr(v) })
	walk(oidLldpRemChassisID, func(k string, v g.SnmpPDU) { byKey[k].RemoteMAC = macFromBytes(v) })

	out := make([]SwitchNeighbor, 0, len(byKey))
	for _, n := range byKey {
		out = append(out, *n)
	}
	return out
}

func collectFDB(client *g.GoSNMP) []SwitchFDB {
	byKey := map[string]*SwitchFDB{}
	_ = client.BulkWalk(oidFdbAddress, func(v g.SnmpPDU) error {
		key := indexSuffix(v.Name, oidFdbAddress)
		byKey[key] = &SwitchFDB{MAC: macFromBytes(v)}
		return nil
	})
	_ = client.BulkWalk(oidFdbPort, func(v g.SnmpPDU) error {
		key := indexSuffix(v.Name, oidFdbPort)
		if byKey[key] != nil {
			byKey[key].Port = int(g.ToBigInt(v.Value).Int64())
		}
		return nil
	})
	out := make([]SwitchFDB, 0, len(byKey))
	for _, f := range byKey {
		if f.MAC != "" {
			out = append(out, *f)
		}
	}
	return out
}
