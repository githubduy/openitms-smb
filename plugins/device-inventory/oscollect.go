// oscollect.go — thu inventory HOST (Windows) qua osquery chạy trên máy đích (đẩy lệnh qua WinRS).
// 1 lệnh PowerShell chạy nhiều query osqueryi --json, phân section bằng marker "@@<name>".
// osqueryi phải có sẵn trên máy đích (PATH hoặc Program Files/ProgramData) — bundle/đẩy: Phase 5.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	winrsexec "quickwin.dev/winrsexec"
)

// defaultOsqueryMSI — nguồn cài osquery cho Windows (đổi qua env QUICKWIN_OSQUERY_MSI nếu air-gapped/mirror).
const defaultOsqueryMSI = "https://pkg.osquery.io/windows/osquery-5.12.1.msi"

// buildOsqueryPS sinh script PowerShell chạy osqueryi cho từng bảng (marker @@<name>).
// autoDeploy=true: nếu máy đích chưa có osqueryi thì tải + cài MSI (msiexec /qn) rồi chạy.
func buildOsqueryPS(autoDeploy bool, msiURL string) string {
	deploy := ""
	if autoDeploy {
		deploy = `if(-not $oq){
  try{
    $m="$env:TEMP\osquery-di.msi"
    [Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri '` + msiURL + `' -OutFile $m -UseBasicParsing
    Start-Process msiexec -ArgumentList '/i',"$m",'/qn','/norestart' -Wait
    $c='C:\Program Files\osquery\osqueryi.exe'; if(Test-Path $c){$oq=$c}
  }catch{}
}
`
	}
	return `$ErrorActionPreference='SilentlyContinue'
$oq=$env:QUICKWIN_OSQUERY_BIN; if($oq -and -not (Test-Path $oq)){$oq=$null}
if(-not $oq){$oq=(Get-Command osqueryi.exe -EA SilentlyContinue).Source}
if(-not $oq){foreach($p in @('C:\Program Files\osquery\osqueryi.exe','C:\ProgramData\osquery\osqueryi.exe')){if(Test-Path $p){$oq=$p;break}}}
` + deploy + `if(-not $oq){Write-Output '@@ERROR osqueryi not found';exit}
function Q($s){& $oq --json $s}
function J($a){$j=@($a)|ConvertTo-Json -Compress -Depth 4;if(-not $j){'[]'}elseif($j[0] -ne '['){"[$j]"}else{$j}}
Write-Output '@@system';Q 'SELECT hostname,cpu_brand,hardware_vendor,hardware_model FROM system_info'
Write-Output '@@os';Q 'SELECT name,version,build FROM os_version'
Write-Output '@@software';Q 'SELECT name,version FROM programs'
Write-Output '@@services';Q 'SELECT name,status,start_type FROM services'
Write-Output '@@patches';Q 'SELECT hotfix_id FROM patches'
Write-Output '@@network';Q 'SELECT interface,address,mask FROM interface_addresses'
Write-Output '@@routes';Q 'SELECT destination,gateway,interface,metric FROM routes'
Write-Output '@@users';Q 'SELECT username,description FROM users'
Write-Output '@@groups';Q 'SELECT groupname FROM groups'
Write-Output '@@dns';J (Get-DnsClientServerAddress -AddressFamily IPv4 -EA SilentlyContinue | Select-Object -Expand ServerAddresses -Unique | ForEach-Object {@{name="$_"}})
Write-Output '@@env';J ([Environment]::GetEnvironmentVariables('Machine').GetEnumerator() | ForEach-Object {@{name="$($_.Key)";detail="$($_.Value)"}})
Write-Output '@@ntp';J ((((Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Services\W32Time\Parameters' -EA SilentlyContinue).NtpServer) -split '[ ,]' | Where-Object {$_ -and $_ -notmatch '^0x'} | ForEach-Object {@{name="$_"}}))
Write-Output '@@domain';J (@{name="$((Get-CimInstance Win32_ComputerSystem -EA SilentlyContinue).Domain)"})
Write-Output '@@profiles';J (Get-CimInstance Win32_UserProfile -EA SilentlyContinue | Where-Object {-not $_.Special} | ForEach-Object {@{name="$(Split-Path $_.LocalPath -Leaf)";detail="$($_.LocalPath)"}})
Write-Output '@@end'`
}

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
	Facts     []DeviceFact
}

// collectHost chạy osquery trên host qua WinRS (cert auth) rồi parse.
func collectHost(cfg HostCollectConfig) (*HostInventory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()
	msi := cfg.MSIURL
	if msi == "" {
		msi = defaultOsqueryMSI
	}
	res, err := winrsexec.Run(ctx, winrsexec.Params{
		Host:    cfg.Host,
		Port:    cfg.Port,
		CertPEM: cfg.CertPEM,
		KeyPEM:  cfg.KeyPEM,
		Command: buildOsqueryPS(cfg.AutoDeploy, msi),
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %s", cfg.Host, winrsexec.Classify(err))
	}
	if strings.Contains(res.Stdout, "@@ERROR osqueryi not found") {
		if cfg.AutoDeploy {
			return nil, fmt.Errorf("%s: tự cài osquery thất bại (máy đích cần internet tới pkg.osquery.io)", cfg.Host)
		}
		return nil, fmt.Errorf("osqueryi chưa có trên %s — bật auto-deploy hoặc cài osquery thủ công", cfg.Host)
	}
	inv, err := parseOsquery(res.Stdout)
	if err != nil {
		return nil, err
	}
	inv.Host = cfg.Host
	return inv, nil
}

// collectHostLocal chạy osquery NGAY trên máy OpenITMS (không qua WinRS) — server tự kiểm kê chính mình.
// Phần mềm chạy local nên đi vòng WinRS tới 127.0.0.1 là thừa; đây là kết nối trực tiếp.
func collectHostLocal(autoDeploy bool, msiURL string, timeout int) (*HostInventory, error) {
	if msiURL == "" {
		msiURL = defaultOsqueryMSI
	}
	if timeout <= 0 {
		timeout = 300
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	ps := buildOsqueryPS(autoDeploy, msiURL)
	out, _ := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", ps).Output()
	stdout := string(out)
	if strings.Contains(stdout, "@@ERROR osqueryi not found") {
		return nil, fmt.Errorf("osquery chưa có trên server OpenITMS + tự cài thất bại (cần internet tới pkg.osquery.io)")
	}
	inv, err := parseOsquery(stdout)
	if err != nil {
		return nil, err
	}
	if inv.Hostname != "" {
		inv.Host = inv.Hostname
	} else {
		inv.Host = "localhost"
	}
	return inv, nil
}

// HostCollectConfig — tham số thu 1 host.
type HostCollectConfig struct {
	Host       string
	Port       int
	CertPEM    []byte
	KeyPEM     []byte
	Timeout    int
	AutoDeploy bool
	MSIURL     string
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

	inv.Facts = []DeviceFact{}
	addFact := func(cat, name, detail string) {
		if strings.TrimSpace(name) != "" {
			inv.Facts = append(inv.Facts, DeviceFact{Category: cat, Name: strings.TrimSpace(name), Detail: strings.TrimSpace(detail)})
		}
	}

	var net []struct{ Interface, Address, Mask string }
	_ = json.Unmarshal([]byte(sections["network"]), &net)
	for _, n := range net {
		if n.Address == "" || strings.HasPrefix(n.Address, "127.") || strings.HasPrefix(n.Address, "::1") {
			continue
		}
		d := n.Address
		if n.Mask != "" {
			d += " / " + n.Mask
		}
		addFact("network", n.Interface, d)
	}

	var rt []struct{ Destination, Gateway, Interface, Metric string }
	_ = json.Unmarshal([]byte(sections["routes"]), &rt)
	for _, r := range rt {
		addFact("route", r.Destination, "via "+r.Gateway+" dev "+r.Interface)
	}

	var usr []struct{ Username, Description string }
	_ = json.Unmarshal([]byte(sections["users"]), &usr)
	for _, u := range usr {
		addFact("user", u.Username, u.Description)
	}

	var grp []struct {
		Groupname string `json:"groupname"`
	}
	_ = json.Unmarshal([]byte(sections["groups"]), &grp)
	for _, g := range grp {
		addFact("group", g.Groupname, "")
	}

	// dns/env/ntp/domain/profile — PowerShell emit {name, detail?}.
	for _, cat := range []string{"dns", "env", "ntp", "domain", "profile"} {
		key := cat
		if cat == "profile" {
			key = "profiles"
		}
		var fs []struct{ Name, Detail string }
		_ = json.Unmarshal([]byte(sections[key]), &fs)
		for _, f := range fs {
			addFact(cat, f.Name, f.Detail)
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
