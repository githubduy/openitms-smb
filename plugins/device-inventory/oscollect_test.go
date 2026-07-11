package main

import "testing"

const sampleOsqueryOut = "@@system\r\n" +
	`[{"hostname":"PC01","hardware_vendor":"Dell Inc. ","hardware_model":"OptiPlex 7090"}]` + "\r\n" +
	"@@os\r\n" +
	`[{"name":"Microsoft Windows 11 Pro","version":"10.0.22631","build":"22631"}]` + "\r\n" +
	"@@software\r\n" +
	`[{"name":"7-Zip 22.01","version":"22.01"},{"name":"Google Chrome","version":"120.0"}]` + "\r\n" +
	"@@services\r\n" +
	`[{"name":"Spooler","status":"RUNNING","start_type":"AUTO_START"}]` + "\r\n" +
	"@@patches\r\n" +
	`[{"hotfix_id":"KB5029921"},{"hotfix_id":"KB5027397"}]` + "\r\n" +
	"@@end\r\n"

func TestParseOsquery(t *testing.T) {
	inv, err := parseOsquery(sampleOsqueryOut)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Hostname != "PC01" || inv.Vendor != "Dell Inc." || inv.Model != "OptiPlex 7090" {
		t.Fatalf("system sai: %+v", inv)
	}
	if inv.OS != "Microsoft Windows 11 Pro" || inv.OSBuild != "22631" {
		t.Fatalf("os sai: %+v", inv)
	}
	if len(inv.Software) != 2 || inv.Software[0].Name != "7-Zip 22.01" {
		t.Fatalf("software sai: %+v", inv.Software)
	}
	if len(inv.Services) != 1 || inv.Services[0].State != "RUNNING" || inv.Services[0].Start != "AUTO_START" {
		t.Fatalf("services sai: %+v", inv.Services)
	}
	if len(inv.Patches) != 2 || inv.Patches[1].KB != "KB5027397" {
		t.Fatalf("patches sai: %+v", inv.Patches)
	}
}

func TestSplitSectionsIgnoresErrorAndEnd(t *testing.T) {
	s := splitSections("@@ERROR osqueryi not found\r\n@@end\r\n")
	if len(s) != 0 {
		t.Fatalf("ERROR/end không được tạo section, got %v", s)
	}
}

func TestDiffNames(t *testing.T) {
	old := map[string]bool{"A": true, "B": true}
	cur := map[string]bool{"B": true, "C": true}
	added, removed := diffNames(old, cur)
	if len(added) != 1 || added[0] != "C" || len(removed) != 1 || removed[0] != "A" {
		t.Fatalf("diff sai: +%v -%v", added, removed)
	}
}
