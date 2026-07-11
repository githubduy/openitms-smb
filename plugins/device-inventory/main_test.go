package main

import (
	"encoding/json"
	"testing"

	pluginv1 "quickwin.dev/proto/quickwin/plugin/v1"
)

func TestQueryID(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"5", 5, true},
		{"", 0, false},
		{"abc", 0, false},
		{"0", 0, false},
		{"-3", 0, false},
	}
	for _, c := range cases {
		req := &pluginv1.HttpRequest{Query: map[string]string{"id": c.in}}
		got, ok := queryID(req)
		if got != c.want || ok != c.ok {
			t.Errorf("queryID(%q) = (%d,%v), muốn (%d,%v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestJSONResp(t *testing.T) {
	r := jsonResp(404, map[string]string{"error": "x"})
	if r.Status != 404 {
		t.Fatalf("status = %d, muốn 404", r.Status)
	}
	if r.Headers["Content-Type"] != "application/json; charset=utf-8" {
		t.Fatalf("content-type sai: %q", r.Headers["Content-Type"])
	}
	var m map[string]string
	if err := json.Unmarshal(r.Body, &m); err != nil || m["error"] != "x" {
		t.Fatalf("body sai: %s err=%v", r.Body, err)
	}
}

func TestMetadataMatchesManifest(t *testing.T) {
	// Metadata.routes PHẢI khớp plugin.yaml (core từ chối load nếu lệch).
	md, err := (&plugin{}).Metadata(nil)
	if err != nil {
		t.Fatal(err)
	}
	if md.Name != "device-inventory" || md.Version != version {
		t.Fatalf("name/version sai: %s %s", md.Name, md.Version)
	}
	want := map[string]bool{"devices": true, "device": true, "changes": true,
		"collect": true, "collect-switch": true, "export": true}
	if len(md.Routes) != len(want) {
		t.Fatalf("số route = %d, muốn %d", len(md.Routes), len(want))
	}
	for _, r := range md.Routes {
		if !want[r.Path] {
			t.Errorf("route lạ: %s", r.Path)
		}
	}
}
