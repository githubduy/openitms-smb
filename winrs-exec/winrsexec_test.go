package winrsexec

import (
	"errors"
	"testing"
)

func TestParseHosts(t *testing.T) {
	inv := `
# comment
10.0.0.5
win11.lab:5986 cert=win11.pem
host.example cert=a.pem key=a.key
1.2.3.4 port=5985
`
	hosts := ParseHosts(inv, "default.pem")
	if len(hosts) != 4 {
		t.Fatalf("mong 4 host, got %d: %+v", len(hosts), hosts)
	}
	if hosts[0].Addr != "10.0.0.5" || hosts[0].Cert != "default.pem" || hosts[0].Port != 0 {
		t.Errorf("host0 sai: %+v", hosts[0])
	}
	if hosts[1].Addr != "win11.lab" || hosts[1].Port != 5986 || hosts[1].Cert != "win11.pem" {
		t.Errorf("host1 sai: %+v", hosts[1])
	}
	if hosts[2].Cert != "a.pem" || hosts[2].Key != "a.key" {
		t.Errorf("host2 sai: %+v", hosts[2])
	}
	if hosts[3].Port != 5985 {
		t.Errorf("host3 port sai: %+v", hosts[3])
	}
}

func TestClassify(t *testing.T) {
	cases := map[string]string{
		"x509: bad cert":            "LỖI CHỨNG CHỈ",
		"dial tcp: connection refused": "LỖI MẠNG",
		"received 401":              "LỖI XÁC THỰC",
		"weird":                     "LỖI WINRS",
	}
	for in, want := range cases {
		if got := Classify(errors.New(in)); !contains(got, want) {
			t.Errorf("Classify(%q)=%q muốn chứa %q", in, got, want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
