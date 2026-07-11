<#
  win-baseline-apps - cai phan mem nen qua winget. Chay TREN may Windows dich (WinRS cert auth).
  Params inject boi OpenITMS task runner tu template.yaml inputs. Idempotent (da co -> nang cap).
#>
param(
  [string]$Packages = "7zip.7zip,Google.Chrome,Notepad++.Notepad++"
)
$ErrorActionPreference = "Stop"

if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
  throw "winget khong co tren may (can App Installer / Windows 10 1809+)."
}

$ids = $Packages.Split(",") | ForEach-Object { $_.Trim() } | Where-Object { $_ }
$fail = 0
foreach ($id in $ids) {
  Write-Host "==> $id"
  # da cai -> winget tu upgrade; chua -> install. Khong tuong tac.
  winget install --id $id --exact --silent --accept-source-agreements `
    --accept-package-agreements --disable-interactivity 2>&1 | Out-Host
  # exit code 0 = OK; -1978335189 = "no applicable upgrade / da moi nhat" cung coi la OK.
  if ($LASTEXITCODE -ne 0 -and $LASTEXITCODE -ne -1978335189 -and $LASTEXITCODE -ne -1978335135) {
    Write-Warning "  $id : winget exit $LASTEXITCODE"
    $fail++
  }
}
Write-Host "OK: xu ly $($ids.Count) goi, $fail loi."
if ($fail -gt 0) { exit 1 }
