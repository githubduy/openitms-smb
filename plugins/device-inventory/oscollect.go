// oscollect.go — thu inventory HOST (Windows) qua osquery chạy trên máy đích (đẩy lệnh qua WinRS).
// 1 lệnh PowerShell chạy nhiều query osqueryi --json, phân section bằng marker "@@<name>".
// osqueryi phải có sẵn trên máy đích (PATH hoặc Program Files/ProgramData) — bundle/đẩy: Phase 5.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	winrsexec "quickwin.dev/winrsexec"
)

// osqueryPS — script PowerShell chạy osqueryi cho từng bảng, in kèm marker để plugin tách.
const osqueryPS = `$ErrorActionPreference='SilentlyContinue'
$oq=(Get-Command osqueryi.exe -EA SilentlyContinue).Source
if(-not $oq){foreach($p in @('C:\Program Files\osquery\osqueryi.exe','C:\ProgramData\osquery\osqueryi.exe')){if(Test-Path $p){$oq=$p;break}}}
if(-not $oq){Write-Output '@@ERROR osqueryi not found';exit}
function Q($s){& $oq --json $s}
Write-Output '@@system';Q 'SELECT hostname,cpu_brand,hardware_vendor,hardware_model FROM system_info'
Write-Output '@@os';Q 'SELECT name,version,build FROM os_version'
Write-Output '@@software';Q 'SELECT name,version FROM programs'
Write-Output '@@services';Q 'SELECT name,status,start_type FROM services'
Write-Output '@@patches';Q 'SELECT hotfix_id FROM patches'
Write-Output '@@end'`

type HostInventory struct {
	Host      string
	Hostname  string
	OS        string
	OSVersion string
	OSBuild   string
	Vendor    string
	Model     string
	Software  []Software
	Services  []Service
	Patches   []Patch
}

// collectHost chạy osquery trên host qua WinRS (cert auth) rồi parse.
func collectHost(cfg HostCollectConfig) (*HostInventory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()
	res, err := winrsexec.Run(ctx, winrsexec.Params{
		Host:    cfg.Host,
		Port:    cfg.Port,
		CertPEM: cfg.CertPEM,
		KeyPEM:  cfg.KeyPEM,
		Command: osqueryPS,
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %s", cfg.Host, winrsexec.Classify(err))
	}
	if strings.Contains(res.Stdout, "@@ERROR osqueryi not found") {
		return nil, fmt.Errorf("osqueryi chưa có trên %s — cần deploy osquery (Phase 5)", cfg.Host)
	}
	inv, err := parseOsquery(res.Stdout)
	if err != nil {
		return nil, err
	}
	inv.Host = cfg.Host
	return inv, nil
}

// HostCollectConfig — tham số thu 1 host.
type HostCollectConfig struct {
	Host    string
	Port    int
	CertPEM []byte
	KeyPEM  []byte
	Timeout int
}

// parseOsquery tách stdout theo marker "@@<name>" và unmarshal từng section.
func parseOsquery(stdout string) (*HostInventory, error) {
	sections := splitSections(stdout)
	inv := &HostInventory{Software: []Software{}, Services: []Service{}, Patches: []Patch{}}

	var sys []struct {
		Hostname string `json:"hostname"`
		Vendor   string `json:"hardware_vendor"`
		Model    string `json:"hardware_model"`
	}
	_ = json.Unmarshal([]byte(sections["system"]), &sys)
	if len(sys) > 0 {
		inv.Hostname = sys[0].Hostname
		inv.Vendor = strings.TrimSpace(sys[0].Vendor)
		inv.Model = strings.TrimSpace(sys[0].Model)
	}

	var os []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Build   string `json:"build"`
	}
	_ = json.Unmarshal([]byte(sections["os"]), &os)
	if len(os) > 0 {
		inv.OS = os[0].Name
		inv.OSVersion = os[0].Version
		inv.OSBuild = os[0].Build
	}

	var sw []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	_ = json.Unmarshal([]byte(sections["software"]), &sw)
	for _, s := range sw {
		if s.Name != "" {
			inv.Software = append(inv.Software, Software{Name: s.Name, Version: s.Version})
		}
	}

	var sv []struct {
		Name  string `json:"name"`
		State string `json:"status"`
		Start string `json:"start_type"`
	}
	_ = json.Unmarshal([]byte(sections["services"]), &sv)
	for _, s := range sv {
		if s.Name != "" {
			inv.Services = append(inv.Services, Service{Name: s.Name, State: s.State, Start: s.Start})
		}
	}

	var pt []struct {
		KB string `json:"hotfix_id"`
	}
	_ = json.Unmarshal([]byte(sections["patches"]), &pt)
	for _, p := range pt {
		if p.KB != "" {
			inv.Patches = append(inv.Patches, Patch{KB: p.KB})
		}
	}
	return inv, nil
}

// splitSections trả map[name]=jsonBlob theo marker "@@name" (mỗi marker 1 dòng riêng).
func splitSections(stdout string) map[string]string {
	out := map[string]string{}
	var cur string
	var buf strings.Builder
	flush := func() {
		if cur != "" {
			out[cur] = strings.TrimSpace(buf.String())
		}
		buf.Reset()
	}
	for _, line := range strings.Split(stdout, "\n") {
		t := strings.TrimRight(line, "\r")
		if strings.HasPrefix(t, "@@") {
			flush()
			cur = strings.TrimSpace(strings.TrimPrefix(t, "@@"))
			if cur == "end" || strings.HasPrefix(cur, "ERROR") {
				cur = ""
			}
			continue
		}
		buf.WriteString(t)
		buf.WriteString("\n")
	}
	flush()
	return out
}
