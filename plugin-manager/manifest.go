// Package pluginmanager — quét /plugins, chạy plugin qua HashiCorp go-plugin,
// sinh API động, health-check + restart. Spec: docs/L2-specs/plugin-manager.md.
package pluginmanager

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

//go:embed schema/plugin.schema.json
var manifestSchemaJSON []byte

// Manifest — ánh xạ plugin.yaml. Schema máy-đọc: schema/plugin.schema.json (nguồn chân lý).
type Manifest struct {
	Name            string            `yaml:"name" json:"name"`
	Version         string            `yaml:"version" json:"version"`
	License         string            `yaml:"license" json:"license"`
	Description     string            `yaml:"description,omitempty" json:"description,omitempty"`
	ProtocolVersion int               `yaml:"protocol_version" json:"protocol_version"`
	Entrypoint      map[string]string `yaml:"entrypoint" json:"entrypoint"`
	Checksum        map[string]string `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	Routes          []RouteDef        `yaml:"routes,omitempty" json:"routes,omitempty"`
	Permissions     []string          `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	MinCoreVersion  string            `yaml:"min_core_version,omitempty" json:"min_core_version,omitempty"`
	UI              *UIDef            `yaml:"ui,omitempty" json:"ui,omitempty"`

	Dir string `yaml:"-" json:"-"` // thư mục plugin, gán lúc load
}

type RouteDef struct {
	Method       string `yaml:"method" json:"method"`
	Path         string `yaml:"path" json:"path"`
	Description  string `yaml:"description,omitempty" json:"description,omitempty"`
	RequireAdmin bool   `yaml:"require_admin,omitempty" json:"require_admin,omitempty"`
}

type UIDef struct {
	MenuTitle string `yaml:"menu_title,omitempty" json:"menu_title,omitempty"`
	AssetsDir string `yaml:"assets_dir,omitempty" json:"assets_dir,omitempty"`
}

var manifestSchema = mustCompileSchema()

func mustCompileSchema() *jsonschema.Schema {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(manifestSchemaJSON))
	if err != nil {
		panic(fmt.Sprintf("plugin.schema.json không parse được: %v", err))
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource("plugin.schema.json", doc); err != nil {
		panic(err)
	}
	return c.MustCompile("plugin.schema.json")
}

// LoadManifest đọc + validate <dir>/plugin.yaml theo JSON Schema.
func LoadManifest(dir string) (*Manifest, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "plugin.yaml"))
	if err != nil {
		return nil, fmt.Errorf("đọc plugin.yaml: %w", err)
	}
	return parseManifest(raw, dir)
}

func parseManifest(raw []byte, dir string) (*Manifest, error) {
	// YAML → generic → JSON roundtrip để jsonschema nhận đúng kiểu số/bool
	var generic any
	if err := yaml.Unmarshal(raw, &generic); err != nil {
		return nil, fmt.Errorf("plugin.yaml không phải YAML hợp lệ: %w", err)
	}
	jsonBytes, err := json.Marshal(generic)
	if err != nil {
		return nil, fmt.Errorf("manifest không chuyển được sang JSON: %w", err)
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	if err := manifestSchema.Validate(inst); err != nil {
		return nil, fmt.Errorf("manifest không đạt schema: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return nil, err
	}
	m.Dir = dir
	return &m, nil
}
