// snmp_util.go — helper thuần (parse PDU, OID index, MAC, uptime, vendor). Tách để unit-test.
package main

import (
	"fmt"
	"strconv"
	"strings"

	g "github.com/gosnmp/gosnmp"
)

// toStr đổi PDU (OctetString/khác) thành string in được.
func toStr(v g.SnmpPDU) string {
	switch b := v.Value.(type) {
	case []byte:
		return strings.TrimRight(string(b), "\x00")
	case string:
		return b
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v.Value)
	}
}

// macFromBytes format []byte (6 octet) thành aa:bb:cc:dd:ee:ff.
func macFromBytes(v g.SnmpPDU) string {
	b, ok := v.Value.([]byte)
	if !ok || len(b) == 0 {
		return ""
	}
	parts := make([]string, len(b))
	for i, x := range b {
		parts[i] = fmt.Sprintf("%02x", x)
	}
	return strings.Join(parts, ":")
}

// lastIndex lấy số cuối trong OID (ifIndex): ".1.3...2.2.1.2.10" → 10.
func lastIndex(oid string) int {
	oid = strings.TrimLeft(oid, ".")
	i := strings.LastIndex(oid, ".")
	if i < 0 {
		return -1
	}
	n, err := strconv.Atoi(oid[i+1:])
	if err != nil {
		return -1
	}
	return n
}

// indexSuffix trả phần index của 1 OID sau khi bỏ base (giữ nguyên chuỗi, dùng làm khóa gộp cột).
func indexSuffix(oid, base string) string {
	oid = strings.TrimLeft(oid, ".")
	base = strings.TrimLeft(base, ".")
	if !strings.HasPrefix(oid, base) {
		return ""
	}
	return strings.TrimPrefix(strings.TrimPrefix(oid, base), ".")
}

// localPortFromLldpIndex: index LLDP dạng "timeMark.localPortNum.remoteIndex" → localPortNum.
func localPortFromLldpIndex(idx string) string {
	parts := strings.Split(idx, ".")
	if len(parts) >= 2 {
		return parts[1]
	}
	return idx
}

// operStatus đổi IF-MIB ifOperStatus số → chữ.
func operStatus(n int) string {
	switch n {
	case 1:
		return "up"
	case 2:
		return "down"
	case 3:
		return "testing"
	case 5:
		return "dormant"
	case 6:
		return "notPresent"
	case 7:
		return "lowerLayerDown"
	default:
		return "unknown"
	}
}

// fmtUptime đổi sysUpTime (timeticks, 1/100 giây) → "Xd Yh Zm".
func fmtUptime(ticks int64) string {
	sec := ticks / 100
	d := sec / 86400
	h := (sec % 86400) / 3600
	m := (sec % 3600) / 60
	return fmt.Sprintf("%dd %dh %dm", d, h, m)
}

// guessVendor đoán hãng từ sysDescr (heuristic đơn giản).
func guessVendor(descr string) string {
	l := strings.ToLower(descr)
	for _, v := range []string{"cisco", "juniper", "aruba", "hpe", "hewlett", "huawei", "mikrotik",
		"ubiquiti", "dell", "netgear", "tp-link", "fortinet", "extreme", "brocade", "zyxel", "ruijie"} {
		if strings.Contains(l, v) {
			if v == "hewlett" {
				return "hpe"
			}
			return v
		}
	}
	return ""
}

// firstNonEmptyWalk walk 1 cột (ENTITY-MIB) và trả giá trị string khác rỗng đầu tiên.
func firstNonEmptyWalk(client *g.GoSNMP, oid string) string {
	var found string
	_ = client.BulkWalk(oid, func(v g.SnmpPDU) error {
		if found != "" {
			return nil
		}
		if s := strings.TrimSpace(toStr(v)); s != "" {
			found = s
		}
		return nil
	})
	return found
}
