package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/masterzen/winrm"
	"quickwin.dev/pluginmanager/certstore"
)

type winrsParams struct {
	Host    string
	Port    int
	CertPEM []byte
	KeyPEM  []byte
	Command string
	Timeout time.Duration // hard timeout cho dial+HTTP (0 → 60s) — tránh treo khi host không tồn tại
}

type winrsResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// runWinRSCert chạy 1 lệnh qua WinRM HTTPS dùng client certificate auth.
// Windows đích cần: WinRM listener HTTPS + bật Certificate auth + map cert→user (JEA/local).
func runWinRSCert(ctx context.Context, p winrsParams) (*winrsResult, error) {
	// insecure=true: chấp nhận server cert self-signed (lab). Verify CA là tuỳ chọn v2
	// (spec plugins-winrs-cert.md) — khi có, truyền CA vào tham số thứ 5 + insecure=false.
	dialTimeout := p.Timeout
	if dialTimeout <= 0 {
		dialTimeout = 60 * time.Second
	}
	endpoint := winrm.NewEndpoint(
		p.Host, p.Port,
		true, // https
		true, // insecure (xem chú thích trên)
		nil,  // CA cert (v2)
		p.CertPEM,
		p.KeyPEM,
		dialTimeout, // hard timeout — không để 0 (treo vô hạn khi host chết)
	)

	params := winrm.DefaultParameters
	params.TransportDecorator = func() winrm.Transporter {
		return &winrm.ClientAuthRequest{}
	}

	client, err := winrm.NewClientWithParameters(endpoint, "", "", params)
	if err != nil {
		return nil, fmt.Errorf("tạo WinRM client: %w", err)
	}

	// masterzen/winrm không luôn tôn trọng ctx cho TCP dial → bọc goroutine + select
	// để handler LUÔN trả về trong timeout, không treo API dù library dial chậm.
	type outcome struct {
		res *winrsResult
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		var stdout, stderr strings.Builder
		code, rerr := client.RunWithContext(ctx, p.Command, &stdout, &stderr)
		if rerr != nil {
			done <- outcome{err: rerr}
			return
		}
		done <- outcome{res: &winrsResult{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String()}}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout: không kết nối được %s trong thời gian cho phép (%v)", p.Host, dialTimeout)
	case o := <-done:
		return o.res, o.err
	}
}

// resolveCertKey lấy PEM cert + key từ certstore entry.
// - PEM có cả cert + PRIVATE KEY → dùng chung.
// - keyFile != "" → lấy key từ file .pem khác trong ./certs.
// - .pfx → chưa hỗ trợ ở v1 (cần password decode) → báo lỗi rõ.
func resolveCertKey(store *certstore.Store, e *certstore.Entry, keyFile string) (certPEM, keyPEM []byte, err error) {
	if e.Kind == "pfx" {
		return nil, nil, fmt.Errorf("cert %q là PKCS#12 (.pfx) — v1 chưa hỗ trợ, dùng .pem (cert+key)", e.Name)
	}
	if keyFile != "" {
		ke := store.Get(keyFile)
		if ke == nil {
			return nil, nil, fmt.Errorf("key file %q không có trong ./certs", keyFile)
		}
		return e.Raw, ke.Raw, nil
	}
	if !e.HasKey {
		return nil, nil, fmt.Errorf("cert %q không chứa private key — cung cấp thêm 'key' (file .pem)", e.Name)
	}
	// PEM chứa cả cert + key: masterzen/winrm nhận cùng buffer cho cả 2
	return e.Raw, e.Raw, nil
}

// classifyError phân loại lỗi để người dùng biết sửa đâu (spec: cert vs mạng vs winrm-config).
func classifyError(err error) string {
	s := strings.ToLower(err.Error())
	switch {
	case strings.Contains(s, "certificate") || strings.Contains(s, "tls") || strings.Contains(s, "x509"):
		return "LỖI CHỨNG CHỈ: " + err.Error() + " — kiểm tra cert/key hợp lệ + Windows đã map cert→user"
	case strings.Contains(s, "connection refused") || strings.Contains(s, "no route") ||
		strings.Contains(s, "timeout") || strings.Contains(s, "i/o timeout") || strings.Contains(s, "dial"):
		return "LỖI MẠNG: " + err.Error() + " — kiểm tra host/port 5986 + firewall + WinRM HTTPS listener"
	case strings.Contains(s, "401") || strings.Contains(s, "403") || strings.Contains(s, "unauthorized"):
		return "LỖI XÁC THỰC: " + err.Error() + " — WinRM chưa bật Certificate auth hoặc chưa map cert→user"
	default:
		return "LỖI WINRS: " + err.Error()
	}
}

var _ = tls.Certificate{} // giữ import tls cho mở rộng verify CA (spec)
