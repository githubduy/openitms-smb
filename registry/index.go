// Package registry — index tĩnh có ký số cho artifact của OpenITMS-SMB
// (plugin / template / endpoint-script). ADR-0003: registry = static index + chữ ký,
// host trên GitHub Pages (public) hoặc thư mục local (air-gapped). Không server động.
//
// Mô hình tin cậy: index.json được KÝ (ed25519) bằng private key của maintainer.
// Client verify chữ ký index → tin các sha256 trong index → verify sha256 mỗi artifact
// tải về. Một chữ ký phủ toàn bộ danh mục.
package registry

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// SchemaVersion — bump khi đổi cấu trúc Index (backward-compat: client cũ đọc được index mới).
const SchemaVersion = 1

// ArtifactType — loại artifact phân phối qua registry.
type ArtifactType string

const (
	TypePlugin         ArtifactType = "plugin"
	TypeTemplate       ArtifactType = "template"
	TypeEndpointScript ArtifactType = "endpoint-script"
)

func (t ArtifactType) valid() bool {
	switch t {
	case TypePlugin, TypeTemplate, TypeEndpointScript:
		return true
	}
	return false
}

// Artifact — 1 mục trong registry (1 phiên bản của 1 artifact).
type Artifact struct {
	Name           string       `json:"name"`
	Version        string       `json:"version"` // semver
	Type           ArtifactType `json:"type"`
	License        string       `json:"license"` // bắt buộc — registry từ chối nếu thiếu
	Description    string       `json:"description,omitempty"`
	Path           string       `json:"path"`             // đường dẫn tarball tương đối với index
	SHA256         string       `json:"sha256"`           // "sha256:<hex>" của tarball
	Size           int64        `json:"size,omitempty"`   // byte
	MinCoreVersion string       `json:"min_core_version,omitempty"`
}

func (a Artifact) key() string { return string(a.Type) + "/" + a.Name + "@" + a.Version }

func (a Artifact) validate() error {
	switch {
	case a.Name == "":
		return fmt.Errorf("artifact thiếu name")
	case a.Version == "":
		return fmt.Errorf("%s: thiếu version", a.Name)
	case !a.Type.valid():
		return fmt.Errorf("%s: type không hợp lệ %q", a.Name, a.Type)
	case a.License == "":
		return fmt.Errorf("%s: thiếu license (bắt buộc)", a.Name)
	case a.Path == "":
		return fmt.Errorf("%s: thiếu path", a.Name)
	case len(a.SHA256) != 71 || a.SHA256[:7] != "sha256:": // "sha256:" + 64 hex
		return fmt.Errorf("%s: sha256 sai định dạng", a.Name)
	}
	return nil
}

// Index — nội dung index.json (phần được KÝ). Không chứa chữ ký (chữ ký ở file .sig riêng).
type Index struct {
	Schema    int        `json:"schema"`
	Name      string     `json:"name"`       // tên registry (vd "local", "public")
	Generated string     `json:"generated"`  // RFC3339 — stamp ngoài (Date bị cấm trong workflow)
	Artifacts []Artifact `json:"artifacts"`
}

// NewIndex tạo index rỗng. now = timestamp (truyền vào để test tất định).
func NewIndex(name string, now time.Time) *Index {
	return &Index{Schema: SchemaVersion, Name: name, Generated: now.UTC().Format(time.RFC3339), Artifacts: []Artifact{}}
}

// Add thêm/thay artifact (theo type+name+version). Validate trước.
func (ix *Index) Add(a Artifact) error {
	if err := a.validate(); err != nil {
		return err
	}
	for i := range ix.Artifacts {
		if ix.Artifacts[i].key() == a.key() {
			ix.Artifacts[i] = a
			return nil
		}
	}
	ix.Artifacts = append(ix.Artifacts, a)
	return nil
}

// Validate kiểm toàn bộ index.
func (ix *Index) Validate() error {
	if ix.Schema > SchemaVersion {
		return fmt.Errorf("index schema %d mới hơn client hỗ trợ (%d)", ix.Schema, SchemaVersion)
	}
	seen := map[string]bool{}
	for _, a := range ix.Artifacts {
		if err := a.validate(); err != nil {
			return err
		}
		if seen[a.key()] {
			return fmt.Errorf("artifact trùng: %s", a.key())
		}
		seen[a.key()] = true
	}
	return nil
}

// Canonical trả JSON tất định (sort artifact) — dùng để KÝ và verify (byte phải khớp tuyệt đối).
func (ix *Index) Canonical() ([]byte, error) {
	cp := *ix
	cp.Artifacts = append([]Artifact(nil), ix.Artifacts...)
	sort.Slice(cp.Artifacts, func(i, j int) bool { return cp.Artifacts[i].key() < cp.Artifacts[j].key() })
	return json.MarshalIndent(&cp, "", "  ")
}

// ParseIndex đọc + validate index.json.
func ParseIndex(data []byte) (*Index, error) {
	var ix Index
	if err := json.Unmarshal(data, &ix); err != nil {
		return nil, fmt.Errorf("index.json không parse được: %w", err)
	}
	if err := ix.Validate(); err != nil {
		return nil, err
	}
	return &ix, nil
}
