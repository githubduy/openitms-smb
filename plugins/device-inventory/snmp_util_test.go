package main

import (
	"testing"

	g "github.com/gosnmp/gosnmp"
)

func TestMacFromBytes(t *testing.T) {
	pdu := g.SnmpPDU{Value: []byte{0xaa, 0xbb, 0x0c, 0xdd, 0xee, 0xff}}
	if got := macFromBytes(pdu); got != "aa:bb:0c:dd:ee:ff" {
		t.Fatalf("macFromBytes = %q", got)
	}
	if got := macFromBytes(g.SnmpPDU{Value: []byte{}}); got != "" {
		t.Fatalf("empty MAC phải rỗng, got %q", got)
	}
}

func TestLastIndex(t *testing.T) {
	if got := lastIndex(".1.3.6.1.2.1.2.2.1.2.10"); got != 10 {
		t.Fatalf("lastIndex = %d", got)
	}
	if got := lastIndex("bad"); got != -1 {
		t.Fatalf("lastIndex bad = %d", got)
	}
}

func TestIndexSuffix(t *testing.T) {
	got := indexSuffix(".1.3.6.1.2.1.17.4.3.1.1.0.1.2.3.4.5", oidFdbAddress)
	if got != "0.1.2.3.4.5" {
		t.Fatalf("indexSuffix = %q", got)
	}
	if indexSuffix(".9.9.9", oidFdbAddress) != "" {
		t.Fatal("prefix không khớp phải rỗng")
	}
}

func TestLocalPortFromLldpIndex(t *testing.T) {
	if got := localPortFromLldpIndex("0.5.12"); got != "5" {
		t.Fatalf("localPort = %q", got)
	}
}

func TestOperStatus(t *testing.T) {
	if operStatus(1) != "up" || operStatus(2) != "down" || operStatus(99) != "unknown" {
		t.Fatal("operStatus map sai")
	}
}

func TestFmtUptime(t *testing.T) {
	// 1 ngày 2 giờ 3 phút = (86400+7200+180) giây * 100 ticks
	ticks := int64((86400+7200+180) * 100)
	if got := fmtUptime(ticks); got != "1d 2h 3m" {
		t.Fatalf("fmtUptime = %q", got)
	}
}

func TestGuessVendor(t *testing.T) {
	cases := map[string]string{
		"Cisco IOS Software, C2960":              "cisco",
		"Juniper Networks, Inc. ex2200":          "juniper",
		"HPE Comware Platform":                   "hpe",
		"Hewlett-Packard ProCurve":               "hpe",
		"Linux mikrotik 5.6.3":                   "mikrotik",
		"Some Unknown Vendor Switch":             "",
	}
	for descr, want := range cases {
		if got := guessVendor(descr); got != want {
			t.Errorf("guessVendor(%q) = %q, muốn %q", descr, got, want)
		}
	}
}

func TestToStr(t *testing.T) {
	if toStr(g.SnmpPDU{Value: []byte("hello\x00")}) != "hello" {
		t.Fatal("toStr trim NUL sai")
	}
	if toStr(g.SnmpPDU{Value: nil}) != "" {
		t.Fatal("toStr nil phải rỗng")
	}
}
