package winrsexec

import (
	"strconv"
	"strings"
)

// Host — 1 endpoint Windows từ inventory WinRS.
type Host struct {
	Addr string // host hoặc IP
	Port int    // 0 → 5986
	Cert string // tên file cert trong certs dir ("" → dùng default)
	Key  string // tên file key riêng (optional)
}

// ParseHosts phân tích nội dung inventory WinRS. Mỗi dòng 1 host:
//
//	10.0.0.5
//	win11.lab:5986 cert=win11.pem
//	host.example cert=a.pem key=a.key
//
// Dòng trống / bắt đầu bằng # bị bỏ qua. defaultCert áp cho host không khai cert.
func ParseHosts(inventory, defaultCert string) []Host {
	var hosts []Host
	for _, line := range strings.Split(inventory, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		h := Host{Cert: defaultCert}
		addr := fields[0]
		if i := strings.LastIndex(addr, ":"); i > 0 {
			if p, err := strconv.Atoi(addr[i+1:]); err == nil {
				h.Port = p
				addr = addr[:i]
			}
		}
		h.Addr = addr
		for _, f := range fields[1:] {
			k, v, ok := strings.Cut(f, "=")
			if !ok {
				continue
			}
			switch k {
			case "cert":
				h.Cert = v
			case "key":
				h.Key = v
			case "port":
				if p, err := strconv.Atoi(v); err == nil {
					h.Port = p
				}
			}
		}
		hosts = append(hosts, h)
	}
	return hosts
}
