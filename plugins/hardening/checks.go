// Plugin hardening — quét cấu hình bảo mật của chính host OpenITMS-SMB, báo cáo finding
// và fix được các mục fix được (quyền file). Spec: docs/L2-specs/plugins-hardening.md.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Severity của finding.
type Severity string

const (
	SevHigh   Severity = "high"
	SevMedium Severity = "medium"
	SevLow    Severity = "low"
)

// Finding — 1 phát hiện khi quét.
type Finding struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Severity Severity `json:"severity"`
	Detail   string   `json:"detail"`
	Fixable  bool     `json:"fixable"`  // true → có thể gọi /fix
	Passed   bool     `json:"passed"`   // true → mục này ĐẠT (không phải vấn đề)
}

// scanConfig — đường dẫn để quét (inject từ env, test override được).
type scanConfig struct {
	prefix     string // OPENITMS_PREFIX (thư mục cài)
	configPath string // config/config.json
	certsDir   string // certs/
	dbPassFile string // .db-pass
}

func configFromEnv() scanConfig {
	prefix := getenv("OPENITMS_PREFIX", ".")
	return scanConfig{
		prefix:     prefix,
		configPath: getenv("QUICKWIN_CONFIG", filepath.Join(prefix, "config", "config.json")),
		certsDir:   getenv("QUICKWIN_CERTS_DIR", filepath.Join(prefix, "certs")),
		dbPassFile: filepath.Join(prefix, ".db-pass"),
	}
}

// runChecks chạy toàn bộ check, trả findings (cả pass lẫn fail).
func runChecks(c scanConfig) []Finding {
	var f []Finding
	f = append(f, checkDefaultPassword(c))
	f = append(f, checkConfigPerms(c))
	f = append(f, checkCertsDirPerms(c))
	f = append(f, checkConfigTLS(c))
	f = append(f, checkDBPassFile(c))
	return f
}

// checkDefaultPassword — cảnh báo nếu config.json còn dấu vết mật khẩu admin mặc định.
// Không đọc được hash password (đã hash), nhưng nếu file config chứa "quickwin123" (một số
// bản dev để trong config) hoặc marker "default_admin" → cảnh báo. Chủ yếu nhắc qua banner
// (patch 0003); check này bổ sung ở mức host.
func checkDefaultPassword(c scanConfig) Finding {
	f := Finding{ID: "default-admin-password", Title: "Mật khẩu admin mặc định",
		Severity: SevHigh, Fixable: false}
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		f.Detail = "không đọc được config để kiểm tra"
		f.Passed = true // không kết luận được → không báo động giả
		return f
	}
	if strings.Contains(string(data), "quickwin123") {
		f.Detail = "config còn chứa mật khẩu mặc định 'quickwin123' — đổi ngay qua Settings/UI"
		return f
	}
	f.Passed = true
	f.Detail = "không phát hiện mật khẩu mặc định trong config"
	return f
}

// checkConfigPerms — config.json phải 0600 (chỉ owner đọc; chứa DB pass + keys).
func checkConfigPerms(c scanConfig) Finding {
	return checkFileMode("config-perms", "Quyền file config.json", c.configPath, 0o600, SevHigh)
}

// checkCertsDirPerms — thư mục certs phải ≤ 0750 (chứa private key).
func checkCertsDirPerms(c scanConfig) Finding {
	f := Finding{ID: "certs-dir-perms", Title: "Quyền thư mục certs", Severity: SevHigh, Fixable: true}
	if runtime.GOOS == "windows" {
		f.Passed, f.Detail = true, "bỏ qua trên Windows (ACL khác Unix mode)"
		return f
	}
	fi, err := os.Stat(c.certsDir)
	if err != nil {
		f.Passed, f.Detail = true, "thư mục certs chưa tồn tại"
		return f
	}
	mode := fi.Mode().Perm()
	if mode&0o027 != 0 { // group-write hoặc other bất kỳ
		f.Detail = fmt.Sprintf("certs mode %o quá mở (nên ≤ 0750) — chứa private key", mode)
		return f
	}
	f.Passed, f.Detail = true, fmt.Sprintf("certs mode %o OK", mode)
	return f
}

// checkConfigTLS — UI nên bật TLS (config có 'tls'/'ssl' hoặc chạy sau reverse proxy).
func checkConfigTLS(c scanConfig) Finding {
	f := Finding{ID: "ui-tls", Title: "TLS cho Web UI", Severity: SevMedium, Fixable: false}
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		f.Passed, f.Detail = true, "không đọc được config"
		return f
	}
	s := strings.ToLower(string(data))
	if strings.Contains(s, "\"tls\"") || strings.Contains(s, "\"ssl\"") || strings.Contains(s, "https") {
		f.Passed, f.Detail = true, "cấu hình có dấu hiệu TLS"
		return f
	}
	f.Detail = "chưa thấy cấu hình TLS — dùng reverse proxy HTTPS hoặc bật TLS cho UI công khai"
	return f
}

// checkDBPassFile — .db-pass (nếu có) phải 0600.
func checkDBPassFile(c scanConfig) Finding {
	if _, err := os.Stat(c.dbPassFile); err != nil {
		return Finding{ID: "db-pass-perms", Title: "Quyền file mật khẩu DB", Severity: SevMedium,
			Passed: true, Detail: "không có .db-pass (DB ngoài / dev)"}
	}
	return checkFileMode("db-pass-perms", "Quyền file mật khẩu DB", c.dbPassFile, 0o600, SevMedium)
}

// checkFileMode helper — file phải đúng mode mong đợi (Unix); Windows bỏ qua.
func checkFileMode(id, title, path string, want os.FileMode, sev Severity) Finding {
	f := Finding{ID: id, Title: title, Severity: sev, Fixable: true}
	if runtime.GOOS == "windows" {
		f.Passed, f.Detail = true, "bỏ qua trên Windows (ACL khác Unix mode)"
		return f
	}
	fi, err := os.Stat(path)
	if err != nil {
		f.Passed, f.Detail = true, "file chưa tồn tại"
		return f
	}
	mode := fi.Mode().Perm()
	if mode&^want != 0 { // có bit ngoài want
		f.Detail = fmt.Sprintf("mode %o quá mở (nên %o)", mode, want)
		return f
	}
	f.Passed, f.Detail = true, fmt.Sprintf("mode %o OK", mode)
	return f
}

// applyFix chmod về mode an toàn cho các check fixable. Trả nil nếu OK.
func applyFix(c scanConfig, checkID string) error {
	switch checkID {
	case "config-perms":
		return chmodIfExists(c.configPath, 0o600)
	case "db-pass-perms":
		return chmodIfExists(c.dbPassFile, 0o600)
	case "certs-dir-perms":
		return chmodIfExists(c.certsDir, 0o750)
	default:
		return fmt.Errorf("check %q không fix tự động được", checkID)
	}
}

func chmodIfExists(path string, mode os.FileMode) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("fix quyền file không áp dụng trên Windows")
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("không tìm thấy %s", path)
	}
	return os.Chmod(path, mode)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustJSON(v any) []byte { b, _ := json.Marshal(v); return b }
