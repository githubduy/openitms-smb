// Package servicemanager — theo dõi status + restart các service hạ tầng của OpenITMS-SMB
// (MariaDB, Gitea, core app). Trang Admin /openitms tab Services gọi qua đây.
//
// Status: dial TCP (platform-agnostic — chạy cả Linux lẫn Windows dev).
// Restart: systemctl (Linux — bản cài thật); nền tảng khác trả "không hỗ trợ" gracefully.
package servicemanager

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"sort"
	"time"
)

// Service — 1 service hạ tầng.
type Service struct {
	Name       string `json:"name"`        // "mariadb" | "gitea" | "app"
	Label      string `json:"label"`       // hiển thị UI
	Addr       string `json:"-"`           // host:port để check status (TCP)
	Unit       string `json:"-"`           // systemd unit (Linux restart)
	CanRestart bool   `json:"can_restart"` // false cho "app" (không tự restart chính mình)
}

// Status — trạng thái runtime của 1 service.
type Status struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	Up         bool   `json:"up"`
	Addr       string `json:"addr"`
	CanRestart bool   `json:"can_restart"`
}

// Manager giữ danh sách service + biết cách check/restart.
type Manager struct {
	services []Service
}

func New(services []Service) *Manager { return &Manager{services: services} }

// Default dựng danh sách service chuẩn từ address (env do core truyền vào).
// selfAddr = địa chỉ app tự nó (chỉ status, không restart).
func Default(mariadbAddr, giteaAddr, selfAddr string) *Manager {
	var s []Service
	if mariadbAddr != "" {
		s = append(s, Service{Name: "mariadb", Label: "MariaDB (database)", Addr: mariadbAddr, Unit: "openitms-db", CanRestart: true})
	}
	if giteaAddr != "" {
		s = append(s, Service{Name: "gitea", Label: "Gitea (git server)", Addr: giteaAddr, Unit: "openitms-gitea", CanRestart: true})
	}
	if selfAddr != "" {
		s = append(s, Service{Name: "app", Label: "OpenITMS-SMB (core)", Addr: selfAddr, Unit: "openitms", CanRestart: false})
	}
	return New(s)
}

// List trả status mọi service (dial TCP song song).
func (m *Manager) List(ctx context.Context) []Status {
	out := make([]Status, len(m.services))
	for i, svc := range m.services {
		out[i] = Status{
			Name: svc.Name, Label: svc.Label, Addr: svc.Addr,
			CanRestart: svc.CanRestart, Up: dialUp(ctx, svc.Addr),
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Restart restart 1 service qua systemctl (Linux). Trả lỗi rõ nếu không hỗ trợ.
func (m *Manager) Restart(ctx context.Context, name string) error {
	var svc *Service
	for i := range m.services {
		if m.services[i].Name == name {
			svc = &m.services[i]
			break
		}
	}
	if svc == nil {
		return fmt.Errorf("service %q không tồn tại", name)
	}
	if !svc.CanRestart {
		return fmt.Errorf("service %q không thể tự restart (vd core app)", name)
	}
	if runtime.GOOS != "linux" {
		return fmt.Errorf("restart chỉ hỗ trợ trên bản cài Linux (systemd); nền tảng hiện tại: %s", runtime.GOOS)
	}
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cctx, "systemctl", "restart", svc.Unit).CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl restart %s: %w: %s", svc.Unit, err, out)
	}
	return nil
}

// dialUp: service sống nếu dial TCP thành công trong 2s.
func dialUp(ctx context.Context, addr string) bool {
	if addr == "" {
		return false
	}
	d := net.Dialer{Timeout: 2 * time.Second}
	dctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	conn, err := d.DialContext(dctx, "tcp", addr)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
