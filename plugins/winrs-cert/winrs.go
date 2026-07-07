package main

import (
	"context"
	"fmt"
	"time"

	"quickwin.dev/pluginmanager/certstore"
	"quickwin.dev/winrsexec"
)

type winrsParams struct {
	Host    string
	Port    int
	CertPEM []byte
	KeyPEM  []byte
	Command string
	Timeout time.Duration
}

type winrsResult = winrsexec.Result

// runWinRSCert chạy 1 lệnh qua WinRM cert-auth (delegate winrsexec — dùng chung với core WinRSApp).
func runWinRSCert(ctx context.Context, p winrsParams) (*winrsResult, error) {
	return winrsexec.Run(ctx, winrsexec.Params{
		Host: p.Host, Port: p.Port, CertPEM: p.CertPEM, KeyPEM: p.KeyPEM,
		Command: p.Command, Timeout: p.Timeout,
	})
}

// resolveCertKey lấy PEM cert + key từ certstore entry.
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
	return e.Raw, e.Raw, nil
}

// classifyError delegate winrsexec.Classify.
func classifyError(err error) string { return winrsexec.Classify(err) }
