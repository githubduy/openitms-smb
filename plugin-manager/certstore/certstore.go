// Package certstore — thư mục ./certs của OpenITMS-SMB (spec docs/L2-specs/certs-directory.md):
// người dùng ném file .pem/.pfx vào là hệ thống nhận NÓNG (không restart).
// Dùng bởi: plugin winrs-cert (certificate auth xuống Windows 11) và inventory WinRM.
// Watcher dùng polling (không dep fsnotify) — đơn giản, chạy mọi OS/filesystem.
package certstore

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Entry — 1 file certificate đã nạp.
type Entry struct {
	Name     string // tên file (vd win11-lab.pem)
	Path     string
	Kind     string // "pem" | "pfx"
	Raw      []byte
	SHA256   string
	LoadedAt time.Time
	// PEM đã parse (nil với pfx — consumer tự decode bằng password khi dùng)
	Certificates []*x509.Certificate
	HasKey       bool // PEM chứa PRIVATE KEY block
}

// Store — watch 1 thư mục certs, nạp nóng .pem/.pfx.
type Store struct {
	dir      string
	interval time.Duration
	log      *slog.Logger

	mu      sync.RWMutex
	entries map[string]*Entry // key = file name
	onLoad  []func(*Entry)

	stop chan struct{}
	wg   sync.WaitGroup
}

// New tạo Store. interval <= 0 → 5s.
func New(dir string, interval time.Duration, logger *slog.Logger) *Store {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Store{
		dir: dir, interval: interval, log: logger,
		entries: map[string]*Entry{}, stop: make(chan struct{}),
	}
}

// Start quét ngay 1 lần rồi chạy watcher nền. Thư mục chưa tồn tại → vẫn watch
// (người dùng có thể tạo sau).
func (s *Store) Start() {
	s.scan()
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		t := time.NewTicker(s.interval)
		defer t.Stop()
		for {
			select {
			case <-s.stop:
				return
			case <-t.C:
				s.scan()
			}
		}
	}()
}

func (s *Store) Stop() { close(s.stop); s.wg.Wait() }

// OnLoad đăng ký callback khi có cert mới/đổi (gọi ngoài lock).
func (s *Store) OnLoad(f func(*Entry)) { s.onLoad = append(s.onLoad, f) }

// Get trả entry theo tên file, nil nếu không có.
func (s *Store) Get(name string) *Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entries[name]
}

// List trả mọi entry, sort theo tên.
func (s *Store) List() []*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Store) scan() {
	files, err := os.ReadDir(s.dir)
	if err != nil {
		if !os.IsNotExist(err) {
			s.log.Warn("certstore: đọc thư mục fail", "dir", s.dir, "err", err)
		}
		return
	}
	seen := map[string]bool{}
	var loaded []*Entry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".pem" && ext != ".pfx" && ext != ".p12" && ext != ".crt" && ext != ".key" {
			continue
		}
		seen[name] = true
		path := filepath.Join(s.dir, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			s.log.Warn("certstore: đọc file fail", "file", name, "err", err)
			continue
		}
		sum := sha256.Sum256(raw)
		hexSum := hex.EncodeToString(sum[:])

		s.mu.RLock()
		old := s.entries[name]
		s.mu.RUnlock()
		if old != nil && old.SHA256 == hexSum {
			continue // không đổi
		}

		e, err := parseEntry(name, path, raw, hexSum)
		if err != nil {
			// file rác không làm crash — bỏ qua + warning (spec certs-directory)
			s.log.Warn("certstore: file không hợp lệ, bỏ qua", "file", name, "err", err)
			continue
		}
		s.mu.Lock()
		s.entries[name] = e
		s.mu.Unlock()
		s.log.Info("certstore: nạp certificate", "file", name, "kind", e.Kind,
			"certs", len(e.Certificates), "hasKey", e.HasKey)
		loaded = append(loaded, e)
	}
	// file bị xóa → gỡ khỏi store
	s.mu.Lock()
	for name := range s.entries {
		if !seen[name] {
			delete(s.entries, name)
			s.log.Info("certstore: certificate bị gỡ", "file", name)
		}
	}
	s.mu.Unlock()

	for _, e := range loaded {
		for _, f := range s.onLoad {
			f(e)
		}
	}
}

func parseEntry(name, path string, raw []byte, sum string) (*Entry, error) {
	e := &Entry{Name: name, Path: path, Raw: raw, SHA256: sum, LoadedAt: time.Now()}
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".pfx" || ext == ".p12" {
		// PKCS#12 thường có password — giữ raw, consumer decode lúc dùng
		e.Kind = "pfx"
		if len(raw) < 4 {
			return nil, fmt.Errorf("file quá ngắn, không phải PKCS#12")
		}
		return e, nil
	}
	e.Kind = "pem"
	rest := raw
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		switch {
		case block.Type == "CERTIFICATE":
			c, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("CERTIFICATE block hỏng: %w", err)
			}
			e.Certificates = append(e.Certificates, c)
		case strings.Contains(block.Type, "PRIVATE KEY"):
			e.HasKey = true
		}
	}
	if len(e.Certificates) == 0 && !e.HasKey {
		return nil, fmt.Errorf("không có PEM block hợp lệ")
	}
	return e, nil
}
