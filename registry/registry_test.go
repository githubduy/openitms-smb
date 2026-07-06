package registry

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var testTime = time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)

func makeSourceArtifact(t *testing.T, typ ArtifactType, name, ver, license string) SourceArtifact {
	t.Helper()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte("name: "+name+"\nversion: "+ver+"\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "bin"), []byte("BINARY-"+name), 0o600)
	return SourceArtifact{Type: typ, Name: name, Version: ver, License: license, Dir: dir,
		Meta: map[string]string{"description": "demo " + name}}
}

// AC P3-01/02/03: build → sign → verify → install round-trip; chống giả mạo.
func TestRoundTrip_BuildSignInstall(t *testing.T) {
	pub, priv := GenerateKeyFromSeed([]byte("openitms-test-seed"))
	arts := []SourceArtifact{
		makeSourceArtifact(t, TypePlugin, "hello", "1.0.0", "MIT"),
		makeSourceArtifact(t, TypePlugin, "hello", "1.1.0", "MIT"),
		makeSourceArtifact(t, TypeTemplate, "odoo", "0.1.0", "MIT"),
	}
	ix, built, sig, err := BuildIndex("local", arts, testTime, priv)
	if err != nil {
		t.Fatal(err)
	}
	outDir := t.TempDir()
	if err := WriteRegistry(outDir, ix, built, sig); err != nil {
		t.Fatal(err)
	}

	// client với trusted key đúng
	client := NewClient(pub)
	src := Source{Name: "local", URL: "file://" + filepath.ToSlash(outDir)}

	// search
	res, err := client.Search([]Source{src}, "hello", TypePlugin)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("mong 2 version hello, got %d", len(res))
	}

	// install version cụ thể + verify checksum
	a, tarball, err := client.Install(src, TypePlugin, "hello", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if a.Version != "1.0.0" {
		t.Fatalf("version sai: %s", a.Version)
	}
	// unpack + kiểm nội dung
	dst := t.TempDir()
	if err := Unpack(tarball, dst); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(dst, "bin")); string(b) != "BINARY-hello" {
		t.Fatalf("nội dung tarball sai: %s", b)
	}

	// install không chỉ version → chọn cao nhất (1.1.0)
	a2, _, err := client.Install(src, TypePlugin, "hello", "")
	if err != nil {
		t.Fatal(err)
	}
	if a2.Version != "1.1.0" {
		t.Fatalf("mong version cao nhất 1.1.0, got %s", a2.Version)
	}
}

func TestReject_TamperedIndex(t *testing.T) {
	pub, priv := GenerateKeyFromSeed([]byte("seed-a"))
	arts := []SourceArtifact{makeSourceArtifact(t, TypePlugin, "hello", "1.0.0", "MIT")}
	ix, built, sig, _ := BuildIndex("local", arts, testTime, priv)
	outDir := t.TempDir()
	WriteRegistry(outDir, ix, built, sig)

	// sửa index.json sau khi ký (đổi checksum artifact) — verify phải fail
	idxPath := filepath.Join(outDir, "index.json")
	raw, _ := os.ReadFile(idxPath)
	tampered := []byte(string(raw[:len(raw)-50]) + "sha256:0000000000000000000000000000000000000000000000000000000000000000\"}]}")
	os.WriteFile(idxPath, tampered, 0o644)

	client := NewClient(pub)
	src := Source{Name: "local", URL: "file://" + filepath.ToSlash(outDir)}
	if _, err := client.LoadIndex(src); err == nil {
		t.Fatal("index bị sửa nhưng verify CHO QUA")
	}
}

func TestReject_WrongKey(t *testing.T) {
	_, priv := GenerateKeyFromSeed([]byte("real-key"))
	wrongPub, _ := GenerateKeyFromSeed([]byte("attacker-key"))
	arts := []SourceArtifact{makeSourceArtifact(t, TypePlugin, "hello", "1.0.0", "MIT")}
	ix, built, sig, _ := BuildIndex("local", arts, testTime, priv)
	outDir := t.TempDir()
	WriteRegistry(outDir, ix, built, sig)

	client := NewClient(wrongPub) // trusted key SAI
	src := Source{Name: "local", URL: "file://" + filepath.ToSlash(outDir)}
	if _, err := client.LoadIndex(src); err == nil {
		t.Fatal("ký bằng key khác nhưng verify CHO QUA")
	}
}

func TestReject_TamperedTarball(t *testing.T) {
	pub, priv := GenerateKeyFromSeed([]byte("seed-b"))
	arts := []SourceArtifact{makeSourceArtifact(t, TypePlugin, "hello", "1.0.0", "MIT")}
	ix, built, sig, _ := BuildIndex("local", arts, testTime, priv)
	outDir := t.TempDir()
	WriteRegistry(outDir, ix, built, sig)

	// sửa tarball (không sửa index) → checksum lệch
	tarPath := filepath.Join(outDir, built[0].Artifact.Path)
	os.WriteFile(tarPath, []byte("CORRUPTED"), 0o644)

	client := NewClient(pub)
	src := Source{Name: "local", URL: "file://" + filepath.ToSlash(outDir)}
	if _, _, err := client.Install(src, TypePlugin, "hello", "1.0.0"); err == nil {
		t.Fatal("tarball bị sửa nhưng checksum CHO QUA")
	}
}

func TestReject_MissingLicense(t *testing.T) {
	sa := makeSourceArtifact(t, TypePlugin, "nolicense", "1.0.0", "")
	if _, err := PackArtifact(sa, testTime); err == nil {
		t.Fatal("artifact thiếu license nhưng được đóng gói")
	}
}

func TestSemverPick(t *testing.T) {
	arts := []Artifact{
		{Type: TypePlugin, Name: "x", Version: "0.9.0"},
		{Type: TypePlugin, Name: "x", Version: "0.10.0"},
		{Type: TypePlugin, Name: "x", Version: "0.2.0"},
	}
	a, ok := pickArtifact(arts, TypePlugin, "x", "")
	if !ok || a.Version != "0.10.0" {
		t.Fatalf("mong 0.10.0 (semver, không phải string sort), got %q ok=%v", a.Version, ok)
	}
}
