<#
  win-firewall-baseline - bat Windows Firewall + baseline. Chay TREN may Windows dich (WinRS cert auth).
  Params inject boi OpenITMS task runner. Idempotent.
#>
param(
  [string]$AllowPing = "true"
)
$ErrorActionPreference = "Stop"

Write-Host "==> Bat firewall ca 3 profile (Block inbound mac dinh, Allow outbound)"
Set-NetFirewallProfile -Profile Domain, Public, Private -Enabled True `
  -DefaultInboundAction Block -DefaultOutboundAction Allow

$rule = "OpenITMS Allow ICMPv4 Echo"
Get-NetFirewallRule -DisplayName $rule -ErrorAction SilentlyContinue | Remove-NetFirewallRule -ErrorAction SilentlyContinue
if ($AllowPing -eq "true") {
  Write-Host "==> Cho phep ICMPv4 echo (ping)"
  New-NetFirewallRule -DisplayName $rule -Direction Inbound -Protocol ICMPv4 `
    -IcmpType 8 -Action Allow | Out-Null
}
Write-Host "OK: firewall bat tren $env:COMPUTERNAME (ping=$AllowPing)."
