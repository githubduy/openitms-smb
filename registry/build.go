package registry

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Build 1 registry tĩnh từ thư mục nguồn: đóng gói mỗi artifact thành tarball,
// sinh index.json + ký (nếu có priv). Dùng bởi installer (local registry) và CI (public).

// SourceArtifact — mô tả 1 artifact cần đóng gói.
type SourceArtifact struct {
	Type    ArtifactType
	Name    string
	Version string
	License string
	Dir     string // thư mục chứa nội dung artifact (plugin.yaml + binary, hoặc template.yaml + playbook)
	Meta    map[string]string
}

// Built — kết quả đóng gói 1 artifact (tarball + entry index).
type Built struct {
	Artifact Artifact
	Tarball  []byte
}

// PackArtifact tar.gz toàn bộ Dir (đệ quy), trả Built với sha256 + path chuẩn.
func PackArtifact(sa SourceArtifact, now time.Time) (*Built, error) {
	if sa.License == "" {
		return nil, fmt.Errorf("%s: thiếu license — registry từ chối", sa.Name)
	}
	tarball, err := tarGz(sa.Dir, now)
	if err != nil {
		return nil, err
	}
	relPath := fmt.Sprintf("%s/%s/%s-%s.tar.gz", sa.Type, sa.Name, sa.Name, sa.Version)
	a := Artifact{
		Name: sa.Name, Version: sa.Version, Type: sa.Type, License: sa.License,
		Description: sa.Meta["description"], Path: relPath,
		SHA256: SHA256Hex(tarball), Size: int64(len(tarball)),
		MinCoreVersion: sa.Meta["min_core_version"],
	}
	return &Built{Artifact: a, Tarball: tarball}, nil
}

// BuildIndex đóng gói tất cả + sinh index (chưa ghi ra đĩa). priv != nil → cũng ký.
func BuildIndex(name string, arts []SourceArtifact, now time.Time, priv ed25519.PrivateKey) (*Index, []*Built, string, error) {
	ix := NewIndex(name, now)
	var built []*Built
	for _, sa := range arts {
		b, err := PackArtifact(sa, now)
		if err != nil {
			return nil, nil, "", err
		}
		if err := ix.Add(b.Artifact); err != nil {
			return nil, nil, "", err
		}
		built = append(built, b)
	}
	sig := ""
	if priv != nil {
		s, err := SignIndex(ix, priv)
		if err != nil {
			return nil, nil, "", err
		}
		sig = s
	}
	return ix, built, sig, nil
}

// WriteRegistry ghi index.json (+ .sig) + tarball ra outDir theo layout registry.
func WriteRegistry(outDir string, ix *Index, built []*Built, sig string) error {
	canon, err := ix.Canonical()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "index.json"), canon, 0o644); err != nil {
		return err
	}
	if sig != "" {
		if err := os.WriteFile(filepath.Join(outDir, "index.json.sig"), []byte(sig+"\n"), 0o644); err != nil {
			return err
		}
	}
	for _, b := range built {
		p := filepath.Join(outDir, filepath.FromSlash(b.Artifact.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, b.Tarball, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// tarGz đóng gói thư mục (đệ quy) thành tar.gz tất định (mtime = now, sort path).
func tarGz(dir string, now time.Time) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	var files []string
	err := filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// sort để tarball tất định
	for i := 1; i < len(files); i++ {
		for j := i; j > 0 && files[j-1] > files[j]; j-- {
			files[j-1], files[j] = files[j], files[j-1]
		}
	}
	for _, p := range files {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		rel, _ := filepath.Rel(dir, p)
		hdr := &tar.Header{
			Name: strings.ReplaceAll(rel, "\\", "/"),
			Mode: 0o644, Size: int64(len(data)), ModTime: now.UTC(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write(data); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unpack giải nén tarball artifact vào destDir (dùng khi cài). Chống path traversal.
func Unpack(tarball []byte, destDir string) error {
	gz, err := gzip.NewReader(bytes.NewReader(tarball))
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		clean := filepath.Clean(hdr.Name)
		if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
			return fmt.Errorf("tarball chứa path nguy hiểm: %s", hdr.Name)
		}
		target := filepath.Join(destDir, clean)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		data, err := io.ReadAll(io.LimitReader(tr, 100<<20))
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, data, os.FileMode(hdr.Mode)&0o777); err != nil {
			return err
		}
	}
	return nil
}
