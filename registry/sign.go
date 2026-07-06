package registry

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// Ký số index bằng ed25519 (thuần Go stdlib — không cần minisign/cosign binary).
// - Private key: maintainer giữ (env/file ngoài repo, KHÔNG commit — publishing-policy).
// - Public key: công khai, nhúng vào core làm trusted key để verify.

// GenerateKey sinh cặp khoá ed25519 (dùng khi bootstrap; seed truyền vào để test tất định).
func GenerateKeyFromSeed(seed []byte) (pub, priv []byte) {
	if len(seed) != ed25519.SeedSize {
		s := sha256.Sum256(seed)
		seed = s[:]
	}
	p := ed25519.NewKeyFromSeed(seed)
	return p.Public().(ed25519.PublicKey), p
}

// EncodePublicKey / EncodePrivateKey — base64 để lưu file/env.
func EncodePublicKey(pub []byte) string  { return base64.StdEncoding.EncodeToString(pub) }
func EncodePrivateKey(priv []byte) string { return base64.StdEncoding.EncodeToString(priv) }

func DecodePublicKey(s string) (ed25519.PublicKey, error) {
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, err
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("public key sai độ dài %d", len(b))
	}
	return ed25519.PublicKey(b), nil
}

func DecodePrivateKey(s string) (ed25519.PrivateKey, error) {
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, err
	}
	if len(b) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("private key sai độ dài %d", len(b))
	}
	return ed25519.PrivateKey(b), nil
}

// SignIndex ký canonical bytes của index; trả chữ ký "ed25519:<base64>" (ghi ra index.json.sig).
func SignIndex(ix *Index, priv ed25519.PrivateKey) (string, error) {
	data, err := ix.Canonical()
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, data)
	return "ed25519:" + base64.StdEncoding.EncodeToString(sig), nil
}

// VerifyIndex verify chữ ký index với 1 trong các trusted public key.
func VerifyIndex(ix *Index, signature string, trusted ...ed25519.PublicKey) error {
	raw, ok := strings.CutPrefix(signature, "ed25519:")
	if !ok {
		return fmt.Errorf("chữ ký không phải ed25519")
	}
	sig, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return fmt.Errorf("chữ ký base64 hỏng: %w", err)
	}
	data, err := ix.Canonical()
	if err != nil {
		return err
	}
	for _, pub := range trusted {
		if ed25519.Verify(pub, data, sig) {
			return nil
		}
	}
	return fmt.Errorf("chữ ký KHÔNG khớp trusted key nào — index có thể đã bị sửa")
}

// SHA256Hex trả "sha256:<hex>" của data.
func SHA256Hex(data []byte) string {
	s := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(s[:])
}

// VerifyArtifactChecksum so sha256 tarball với giá trị khai trong index.
func VerifyArtifactChecksum(a Artifact, tarball []byte) error {
	got := SHA256Hex(tarball)
	if !strings.EqualFold(got, a.SHA256) {
		return fmt.Errorf("%s: checksum LỆCH (index=%s thực tế=%s)", a.Name, a.SHA256, got)
	}
	return nil
}
