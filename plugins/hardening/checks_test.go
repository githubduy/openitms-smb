package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// AC P2-05: phát hiện ≥ finding chuẩn (mật khẩu default, quyền file); fix quyền hoạt động;
// không finding giả trên host đã harden.

func writeConfig(t *testing.T, dir, content string, mode os.FileMode) scanConfig {
	t.Helper()
	os.MkdirAll(filepath.Join(dir, "config"), 0o755)
	os.MkdirAll(filepath.Join(dir, "certs"), 0o750)
	cfg := filepath.Join(dir, "config", "config.json")
	if err := os.WriteFile(cfg, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	return scanConfig{prefix: dir, configPath: cfg, certsDir: filepath.Join(dir, "certs"),
		dbPassFile: filepath.Join(dir, ".db-pass")}
}

func find(fs []Finding, id string) Finding {
	for _, f := range fs {
		if f.ID == id {
			return f
		}
	}
	return Finding{ID: "MISSING"}
}

func TestScan_DetectsDefaultPassword(t *testing.T) {
	dir := t.TempDir()
	c := writeConfig(t, dir, `{"port":"3000","admin_pass":"quickwin123"}`, 0o600)
	f := find(runChecks(c), "default-admin-password")
	if f.Passed {
		t.Fatal("phải phát hiện mật khẩu default 'quickwin123'")
	}
	if f.Severity != SevHigh {
		t.Fatalf("severity phải high, got %s", f.Severity)
	}
}

func TestScan_CleanHostNoFalsePositive(t *testing.T) {
	dir := t.TempDir()
	c := writeConfig(t, dir, `{"port":"3000","tls":{"cert":"x"}}`, 0o600)
	for _, f := range runChecks(c) {
		// trên host sạch (config 0600, có tls, không default pass) → không finding fail
		// (trừ Windows nơi perm bỏ qua = pass)
		if !f.Passed {
			t.Fatalf("false positive trên host sạch: %s — %s", f.ID, f.Detail)
		}
	}
}

func TestScan_DetectsLoosePerms(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("perm check bỏ qua trên Windows")
	}
	dir := t.TempDir()
	c := writeConfig(t, dir, `{"port":"3000","tls":true}`, 0o644) // config quá mở
	f := find(runChecks(c), "config-perms")
	if f.Passed {
		t.Fatal("config 0644 phải bị báo (nên 0600)")
	}
	if !f.Fixable {
		t.Fatal("config-perms phải fixable")
	}
	// fix → chmod 0600 → scan lại pass
	if err := applyFix(c, "config-perms"); err != nil {
		t.Fatalf("fix fail: %v", err)
	}
	if f2 := find(runChecks(c), "config-perms"); !f2.Passed {
		t.Fatalf("sau fix vẫn fail: %s", f2.Detail)
	}
}

func TestFix_UnknownCheck(t *testing.T) {
	c := writeConfig(t, t.TempDir(), "{}", 0o600)
	if err := applyFix(c, "khong-co"); err == nil {
		t.Fatal("fix check không tồn tại phải lỗi")
	}
}
