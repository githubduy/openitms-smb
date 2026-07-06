package certstore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// AC P2-04: cert mới nạp nóng không restart; file rác bị bỏ qua; xóa file → gỡ entry.

func selfSignedPEM(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "openitms-test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, _ := x509.MarshalECPrivateKey(key)
	return append(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})...,
	)
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("điều kiện không đạt trong timeout")
}

func TestHotLoadAndRemove(t *testing.T) {
	dir := t.TempDir()
	s := New(dir, 50*time.Millisecond, nil)
	loaded := make(chan string, 10)
	s.OnLoad(func(e *Entry) { loaded <- e.Name })
	s.Start()
	defer s.Stop()

	if got := s.List(); len(got) != 0 {
		t.Fatalf("store mới phải rỗng, got %d", len(got))
	}

	// ném cert vào — phải nạp nóng
	if err := os.WriteFile(filepath.Join(dir, "lab.pem"), selfSignedPEM(t), 0o600); err != nil {
		t.Fatal(err)
	}
	waitFor(t, 3*time.Second, func() bool { return s.Get("lab.pem") != nil })
	e := s.Get("lab.pem")
	if e.Kind != "pem" || len(e.Certificates) != 1 || !e.HasKey {
		t.Fatalf("parse sai: %+v", e)
	}
	select {
	case n := <-loaded:
		if n != "lab.pem" {
			t.Fatalf("callback sai file: %s", n)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("OnLoad không được gọi")
	}

	// pfx: giữ raw
	if err := os.WriteFile(filepath.Join(dir, "win.pfx"), []byte{0x30, 0x82, 0x01, 0x02, 0xAA}, 0o600); err != nil {
		t.Fatal(err)
	}
	waitFor(t, 3*time.Second, func() bool { return s.Get("win.pfx") != nil })
	if s.Get("win.pfx").Kind != "pfx" {
		t.Fatal("pfx kind sai")
	}

	// xóa file → entry bị gỡ
	if err := os.Remove(filepath.Join(dir, "lab.pem")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, 3*time.Second, func() bool { return s.Get("lab.pem") == nil })
}

func TestGarbageIgnoredNoCrash(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "trash.pem"), []byte("not a pem at all"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("ignore me"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := New(dir, 50*time.Millisecond, nil)
	s.Start()
	defer s.Stop()
	time.Sleep(200 * time.Millisecond)
	if got := s.List(); len(got) != 0 {
		t.Fatalf("file rác không được vào store: %+v", got)
	}
}

func TestDirNotExistYet(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "certs-chua-tao")
	s := New(dir, 50*time.Millisecond, nil)
	s.Start() // không panic
	defer s.Stop()

	// tạo dir + cert sau khi Start → vẫn nhận
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "late.pem"), selfSignedPEM(t), 0o600); err != nil {
		t.Fatal(err)
	}
	waitFor(t, 3*time.Second, func() bool { return s.Get("late.pem") != nil })
}
