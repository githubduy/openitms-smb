package pluginmanager

import (
	"os"
	"strings"
	"testing"
)

// AC P1-01: schema validate manifest hợp lệ + bắt đủ 6 case lỗi chuẩn
// (danh sách trong docs/L2-specs/plugin-manifest.md).

func TestManifest_ValidMinimal(t *testing.T) {
	raw, err := os.ReadFile("schema/examples/valid-minimal.yaml")
	if err != nil {
		t.Fatal(err)
	}
	m, err := parseManifest(raw, "x")
	if err != nil {
		t.Fatalf("manifest hợp lệ bị từ chối: %v", err)
	}
	if m.Name != "hello" || m.Version != "0.1.0" || m.License != "MIT" {
		t.Fatalf("parse sai: %+v", m)
	}
}

func TestManifest_InvalidCases(t *testing.T) {
	base := `
name: %s
version: %s
license: %s
protocol_version: 1
entrypoint:
  linux-amd64: bin
`
	cases := []struct {
		name string
		yaml string
	}{
		{"1-thiếu license", "name: a\nversion: 1.0.0\nprotocol_version: 1\nentrypoint: {linux-amd64: bin}\n"},
		{"2-version không semver", strings.ReplaceAll(base, "%s", "a") /* version "a" */},
		{"3-name viết hoa/underscore", "name: WinRS_Cert\nversion: 1.0.0\nlicense: MIT\nprotocol_version: 1\nentrypoint: {linux-amd64: bin}\n"},
		{"4-method lạ", "name: a\nversion: 1.0.0\nlicense: MIT\nprotocol_version: 1\nentrypoint: {linux-amd64: bin}\nroutes:\n  - {method: PATCH, path: x}\n"},
		{"5-checksum sai format", "name: a\nversion: 1.0.0\nlicense: MIT\nprotocol_version: 1\nentrypoint: {linux-amd64: bin}\nchecksum: {bin: deadbeef}\n"},
		{"6-field lạ", "name: a\nversion: 1.0.0\nlicense: MIT\nprotocol_version: 1\nentrypoint: {linux-amd64: bin}\nsudo: true\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseManifest([]byte(tc.yaml), "x"); err == nil {
				t.Fatalf("manifest lỗi (%s) nhưng validator CHO QUA", tc.name)
			}
		})
	}
}

func TestMatchRoute(t *testing.T) {
	routes := []RouteDef{
		{Method: "POST", Path: "exec"},
		{Method: "GET", Path: "status/{id}"},
	}
	if matchRoute(routes, "POST", "exec") == nil {
		t.Fatal("exec phải match")
	}
	if matchRoute(routes, "GET", "status/42") == nil {
		t.Fatal("status/{id} phải match status/42")
	}
	if matchRoute(routes, "GET", "exec") != nil {
		t.Fatal("sai method không được match")
	}
	if matchRoute(routes, "POST", "exec/extra") != nil {
		t.Fatal("thừa segment không được match")
	}
}
