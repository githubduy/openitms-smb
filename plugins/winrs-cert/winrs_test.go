package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
	"quickwin.dev/pluginmanager/certstore"
)

// Test không cần Windows thật: logic cert/lỗi/route. E2E cert thật → cần Win11 lab (README).

func TestMetadataMatchesManifest(t *testing.T) {
	p := &plugin{}
	m, _ := p.Metadata(context.Background())
	if m.Name != "winrs-cert" {
		t.Fatalf("name: %s", m.Name)
	}
	// routes trong metadata phải trùng plugin.yaml (Plugin Manager enforce)
	want := map[string]bool{"POST exec": true, "GET certs": true}
	for _, r := range m.Routes {
		delete(want, r.Method+" "+r.Path)
	}
	if len(want) != 0 {
		t.Fatalf("routes thiếu: %v", want)
	}
}

func TestClassifyError(t *testing.T) {
	cases := map[string]string{
		"x509: certificate signed by unknown authority": "LỖI CHỨNG CHỈ",
		"dial tcp 10.0.0.5:5986: connection refused":     "LỖI MẠNG",
		"received 401 Unauthorized":                      "LỖI XÁC THỰC",
		"something weird":                                "LỖI WINRS",
	}
	for in, want := range cases {
		if got := classifyError(errString(in)); !strings.Contains(got, want) {
			t.Errorf("classify(%q) = %q, muốn chứa %q", in, got, want)
		}
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func TestResolveCertKey(t *testing.T) {
	dir := t.TempDir()
	// cert+key trong 1 PEM
	os.WriteFile(filepath.Join(dir, "combo.pem"), selfSignedCertKeyPEM(t), 0o600)
	// pfx
	os.WriteFile(filepath.Join(dir, "win.pfx"), []byte{0x30, 0x82, 0x01, 0x02, 0xAA}, 0o600)

	s := certstore.New(dir, 30*time.Millisecond, nil)
	s.Start()
	defer s.Stop()
	waitFor(t, 3*time.Second, func() bool { return s.Get("combo.pem") != nil && s.Get("win.pfx") != nil })

	// combo.pem: cert+key OK
	c, k, err := resolveCertKey(s, s.Get("combo.pem"), "")
	if err != nil || len(c) == 0 || len(k) == 0 {
		t.Fatalf("combo.pem phải OK: %v", err)
	}
	// pfx: v1 chưa hỗ trợ
	if _, _, err := resolveCertKey(s, s.Get("win.pfx"), ""); err == nil {
		t.Fatal("pfx phải báo lỗi chưa hỗ trợ")
	}
}

func TestHandleExecValidation(t *testing.T) {
	p := &plugin{certs: certstore.New(t.TempDir(), time.Hour, nil)}
	p.certs.Start()
	defer p.certs.Stop()

	// thiếu field bắt buộc → 400
	body, _ := json.Marshal(execRequest{Host: "h"}) // thiếu command, cert
	resp, _ := p.HandleRequest(context.Background(), &pluginv1.HttpRequest{Path: "exec", Body: body})
	if resp.Status != 400 {
		t.Fatalf("thiếu field phải 400, got %d", resp.Status)
	}

	// cert không tồn tại → 404
	body2, _ := json.Marshal(execRequest{Host: "h", Command: "hostname", Cert: "khong-co.pem"})
	resp2, _ := p.HandleRequest(context.Background(), &pluginv1.HttpRequest{Path: "exec", Body: body2})
	if resp2.Status != 404 {
		t.Fatalf("cert không có phải 404, got %d", resp2.Status)
	}
}

func TestHandleListCerts(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.pem"), selfSignedCertKeyPEM(t), 0o600)
	p := &plugin{certs: certstore.New(dir, 30*time.Millisecond, nil)}
	p.certs.Start()
	defer p.certs.Stop()
	waitFor(t, 3*time.Second, func() bool { return p.certs.Get("a.pem") != nil })

	resp, _ := p.HandleRequest(context.Background(), &pluginv1.HttpRequest{Path: "certs", Method: "GET"})
	if resp.Status != 200 {
		t.Fatalf("list certs status %d", resp.Status)
	}
	if !strings.Contains(string(resp.Body), "a.pem") {
		t.Fatalf("list thiếu a.pem: %s", resp.Body)
	}
}
