package registry

import (
	"crypto/ed25519"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Source — 1 registry (local hoặc public). OpenITMS-SMB cấu hình mặc định 2 source:
// local (thư mục cài, air-gapped) + public (GitHub Pages).
type Source struct {
	Name string // "local" | "public"
	URL  string // "file:///opt/openitms/registry" hoặc "https://<user>.github.io/openitms-registry"
}

// Fetcher đọc bytes theo path tương đối với source (tách ra để test không cần mạng/đĩa).
type Fetcher interface {
	Fetch(rel string) ([]byte, error)
}

// Client verify chữ ký + checksum khi cài artifact từ nhiều source.
type Client struct {
	trusted []ed25519.PublicKey
	http    *http.Client
}

func NewClient(trusted ...ed25519.PublicKey) *Client {
	return &Client{trusted: trusted, http: &http.Client{Timeout: 30 * time.Second}}
}

// fetcherFor chọn fetcher theo scheme của source URL.
func (c *Client) fetcherFor(s Source) (Fetcher, error) {
	switch {
	case strings.HasPrefix(s.URL, "file://"):
		return &fileFetcher{root: strings.TrimPrefix(s.URL, "file://")}, nil
	case strings.HasPrefix(s.URL, "http://"), strings.HasPrefix(s.URL, "https://"):
		return &httpFetcher{base: strings.TrimRight(s.URL, "/"), c: c.http}, nil
	default:
		return nil, fmt.Errorf("scheme không hỗ trợ: %s", s.URL)
	}
}

// LoadIndex tải + verify chữ ký index.json của 1 source. trusted rỗng → BỎ verify
// (chỉ dùng cho local dev; production luôn có trusted key).
func (c *Client) LoadIndex(s Source) (*Index, error) {
	f, err := c.fetcherFor(s)
	if err != nil {
		return nil, err
	}
	return c.loadIndexFrom(f)
}

func (c *Client) loadIndexFrom(f Fetcher) (*Index, error) {
	raw, err := f.Fetch("index.json")
	if err != nil {
		return nil, fmt.Errorf("tải index.json: %w", err)
	}
	ix, err := ParseIndex(raw)
	if err != nil {
		return nil, err
	}
	if len(c.trusted) > 0 {
		sig, err := f.Fetch("index.json.sig")
		if err != nil {
			return nil, fmt.Errorf("tải chữ ký: %w", err)
		}
		if err := VerifyIndex(ix, strings.TrimSpace(string(sig)), c.trusted...); err != nil {
			return nil, err
		}
	}
	return ix, nil
}

// Search gộp artifact khớp query (substring name) từ mọi source, sort theo name+version.
func (c *Client) Search(sources []Source, query string, typ ArtifactType) ([]Artifact, error) {
	var out []Artifact
	for _, s := range sources {
		ix, err := c.LoadIndex(s)
		if err != nil {
			return nil, fmt.Errorf("[%s] %w", s.Name, err)
		}
		for _, a := range ix.Artifacts {
			if typ != "" && a.Type != typ {
				continue
			}
			if query == "" || strings.Contains(a.Name, query) {
				out = append(out, a)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Version < out[j].Version
	})
	return out, nil
}

// Install tải artifact (name[@version]) từ source, verify chữ ký index + checksum tarball,
// trả bytes tarball đã verify. version rỗng → chọn version cao nhất.
func (c *Client) Install(s Source, typ ArtifactType, name, version string) (Artifact, []byte, error) {
	f, err := c.fetcherFor(s)
	if err != nil {
		return Artifact{}, nil, err
	}
	ix, err := c.loadIndexFrom(f)
	if err != nil {
		return Artifact{}, nil, err
	}
	a, ok := pickArtifact(ix.Artifacts, typ, name, version)
	if !ok {
		return Artifact{}, nil, fmt.Errorf("không tìm thấy %s/%s%s", typ, name, verSuffix(version))
	}
	tarball, err := f.Fetch(a.Path)
	if err != nil {
		return Artifact{}, nil, fmt.Errorf("tải %s: %w", a.Path, err)
	}
	if err := VerifyArtifactChecksum(a, tarball); err != nil {
		return Artifact{}, nil, err
	}
	return a, tarball, nil
}

func pickArtifact(arts []Artifact, typ ArtifactType, name, version string) (Artifact, bool) {
	var best Artifact
	found := false
	for _, a := range arts {
		if a.Type != typ || a.Name != name {
			continue
		}
		if version != "" {
			if a.Version == version {
				return a, true
			}
			continue
		}
		if !found || semverLess(best.Version, a.Version) {
			best, found = a, true
		}
	}
	return best, found
}

func verSuffix(v string) string {
	if v == "" {
		return ""
	}
	return "@" + v
}

// --- fetchers ---

type fileFetcher struct{ root string }

func (ff *fileFetcher) Fetch(rel string) ([]byte, error) {
	return os.ReadFile(filepath.Join(ff.root, filepath.FromSlash(rel)))
}

type httpFetcher struct {
	base string
	c    *http.Client
}

func (hf *httpFetcher) Fetch(rel string) ([]byte, error) {
	resp, err := hf.c.Get(hf.base + "/" + path.Clean(rel))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d cho %s", resp.StatusCode, rel)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 100<<20)) // 100MB
}

// semverLess: a < b theo semver rút gọn (major.minor.patch).
func semverLess(a, b string) bool {
	pa, pb := semverParts(a), semverParts(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] < pb[i]
		}
	}
	return false
}

func semverParts(v string) [3]int {
	var out [3]int
	fmt.Sscanf(v, "%d.%d.%d", &out[0], &out[1], &out[2])
	return out
}
