package servicemanager

import (
	"context"
	"net"
	"runtime"
	"testing"
)

func TestList_StatusUpDown(t *testing.T) {
	// service "up": mở 1 listener thật; "down": port không ai nghe.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	upAddr := ln.Addr().String()

	m := New([]Service{
		{Name: "mariadb", Label: "MariaDB", Addr: upAddr, Unit: "openitms-db", CanRestart: true},
		{Name: "gitea", Label: "Gitea", Addr: "127.0.0.1:1", Unit: "openitms-gitea", CanRestart: true}, // down
		{Name: "app", Label: "core", Addr: upAddr, CanRestart: false},
	})
	st := m.List(context.Background())
	if len(st) != 3 {
		t.Fatalf("mong 3 service, got %d", len(st))
	}
	byName := map[string]Status{}
	for _, s := range st {
		byName[s.Name] = s
	}
	if !byName["mariadb"].Up {
		t.Error("mariadb (có listener) phải Up")
	}
	if byName["gitea"].Up {
		t.Error("gitea (không listener) phải Down")
	}
	if byName["app"].CanRestart {
		t.Error("app không được CanRestart")
	}
}

func TestRestart_Errors(t *testing.T) {
	m := Default("127.0.0.1:3306", "127.0.0.1:3080", "127.0.0.1:3000")
	// service không tồn tại
	if err := m.Restart(context.Background(), "khongco"); err == nil {
		t.Error("service lạ phải lỗi")
	}
	// app không restart được
	if err := m.Restart(context.Background(), "app"); err == nil {
		t.Error("app không được phép restart")
	}
	// mariadb: trên non-Linux phải báo không hỗ trợ (không thực thi systemctl)
	if runtime.GOOS != "linux" {
		if err := m.Restart(context.Background(), "mariadb"); err == nil {
			t.Error("restart trên non-Linux phải báo không hỗ trợ")
		}
	}
}

func TestDefault_OmitsEmpty(t *testing.T) {
	m := Default("", "127.0.0.1:3080", "") // chỉ gitea
	st := m.List(context.Background())
	if len(st) != 1 || st[0].Name != "gitea" {
		t.Fatalf("chỉ nên có gitea, got %+v", st)
	}
}
