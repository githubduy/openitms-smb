package winrsexec

// E2E test THẬT xuống 1 host Windows có WinRM HTTPS + certificate auth.
// CHỈ chạy khi có đủ env (mặc định SKIP) — CI/dev không có lab Windows sẽ bỏ qua.
//
// Cách chạy (PowerShell, sau khi đã chạy e2e/setup-winrm-cert.ps1 để bật WinRM+cert):
//   $env:WINRS_E2E_HOST = "127.0.0.1"
//   $env:WINRS_E2E_CERT = "D:\open-source\quickwin\run-local\certs\host.pem"   # cert+key PEM
//   $env:WINRS_E2E_PORT = "5986"                                              # optional
//   go test ./winrs-exec/ -run TestE2E -v -count=1
//
// Cert PEM phải chứa CẢ certificate LẪN private key (winrsexec.Run dùng CertPEM=KeyPEM=file này
// nếu chỉ truyền 1 file — xem WinRSApp). Test này đọc 1 file cho cả cert lẫn key.

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_WinRMCertRun(t *testing.T) {
	host := os.Getenv("WINRS_E2E_HOST")
	certPath := os.Getenv("WINRS_E2E_CERT")
	if host == "" || certPath == "" {
		t.Skip("SKIP E2E: cần WINRS_E2E_HOST + WINRS_E2E_CERT (xem e2e/setup-winrm-cert.ps1)")
	}

	pem, err := os.ReadFile(certPath)
	require.NoError(t, err, "đọc cert PEM")

	port := 5986
	if p := os.Getenv("WINRS_E2E_PORT"); p != "" {
		if n, cerr := strconv.Atoi(p); cerr == nil {
			port = n
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Lệnh đơn giản: in hostname. Kỳ vọng exit 0 + stdout không rỗng.
	res, err := Run(ctx, Params{
		Host:    host,
		Port:    port,
		CertPEM: pem,
		KeyPEM:  pem,
		Command: "hostname",
		Timeout: 30 * time.Second,
	})
	require.NoError(t, err, "winrsexec.Run lỗi: %v", err)
	assert.Equal(t, 0, res.ExitCode, "exit code, stderr=%s", res.Stderr)
	assert.NotEmpty(t, strings.TrimSpace(res.Stdout), "hostname phải in ra tên máy")
	t.Logf("E2E OK — host=%s stdout=%q", host, strings.TrimSpace(res.Stdout))
}
