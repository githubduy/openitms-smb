package pluginmanager

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// verifyChecksums kiểm sha256 từng file khai trong manifest.checksum.
// Manifest KHÔNG khai checksum → trả về danh sách cảnh báo (dev mode);
// registry (Phase 3) sẽ bắt buộc checksum khi cài từ registry.
func verifyChecksums(m *Manifest) (warnings []string, err error) {
	if len(m.Checksum) == 0 {
		return []string{fmt.Sprintf("plugin %q không khai checksum — chỉ chấp nhận cho plugin dev local", m.Name)}, nil
	}
	for file, want := range m.Checksum {
		want = strings.TrimPrefix(want, "sha256:")
		got, herr := sha256File(filepath.Join(m.Dir, file))
		if herr != nil {
			return nil, fmt.Errorf("checksum %s: %w", file, herr)
		}
		if !strings.EqualFold(got, want) {
			return nil, fmt.Errorf("checksum LỆCH ở %s: manifest=%s thực tế=%s", file, want, got)
		}
	}
	return nil, nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
