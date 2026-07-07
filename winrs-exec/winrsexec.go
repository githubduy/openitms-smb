// Package winrsexec — chạy lệnh xuống Windows qua WinRM HTTPS + client certificate auth.
// Dùng chung bởi plugin winrs-cert (API /exec) và core WinRSApp (task runner, inventory WinRS).
// Bọc timeout để KHÔNG treo khi host chết (masterzen/winrm không luôn tôn trọng ctx cho dial).
package winrsexec

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/masterzen/winrm"
)

// Params — 1 lần exec tới 1 host.
type Params struct {
	Host    string
	Port    int // 0 → 5986 (WinRM HTTPS)
	CertPEM []byte
	KeyPEM  []byte
	Command string
	Timeout time.Duration // 0 → 60s
}

// Result — kết quả exec.
type Result struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// Run chạy 1 lệnh qua WinRM cert-auth. Windows đích cần: WinRM HTTPS + Certificate auth +
// map cert→user (JEA/local). insecure=true chấp nhận server cert self-signed (lab).
func Run(ctx context.Context, p Params) (*Result, error) {
	port := p.Port
	if port == 0 {
		port = 5986
	}
	dialTimeout := p.Timeout
	if dialTimeout <= 0 {
		dialTimeout = 60 * time.Second
	}
	endpoint := winrm.NewEndpoint(p.Host, port, true, true, nil, p.CertPEM, p.KeyPEM, dialTimeout)

	params := winrm.DefaultParameters
	params.TransportDecorator = func() winrm.Transporter { return &winrm.ClientAuthRequest{} }

	client, err := winrm.NewClientWithParameters(endpoint, "", "", params)
	if err != nil {
		return nil, fmt.Errorf("tạo WinRM client: %w", err)
	}

	type outcome struct {
		res *Result
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
		done <- outcome{res: &Result{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String()}}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout: không kết nối được %s trong thời gian cho phép (%v)", p.Host, dialTimeout)
	case o := <-done:
		return o.res, o.err
	}
}

// Classify phân loại lỗi để người dùng biết sửa đâu (cert / mạng / auth / winrs).
func Classify(err error) string {
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
