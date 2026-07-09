# core-patches CHANGELOG

Mỗi patch thêm/sửa/xóa phải có 1 entry ở đây (mới nhất lên đầu).
Format: `## <ngày> — <patch-file>` + WHY (vì sao cần) + WHAT (đổi gì, mức cao).

## 2026-07-09 — 0024-enroll-admin-domain.patch
**WHY:** máy domain-joined: Add-LocalGroupMember resolve tên local user nhầm sang domain → fail thầm →
openitms KHÔNG vào Administrators → WinRS AccessDenied (dù đã fix token filter). (Debug thật.)
**WHAT:** winrs-enroll.ps1 dùng `net localgroup Administrators <user> /add` + verify (fallback
Add-LocalGroupMember COMPUTERNAME\user). Spec 0024.

## 2026-07-09 — 0023-enroll-localaccount-token.patch
**WHY:** sau cert-auth OK (0022), WinRS vẫn "AccessDenied (Code 5)" — local admin đăng nhập qua mạng
bị UAC lọc token (standard) → không đủ quyền tạo WinRM shell. (Debug thật.)
**WHAT:** winrs-enroll.ps1 set LocalAccountTokenFilterPolicy=1 (HKLM Policies\System) → full token cho
local-account remote logon. Mảnh cuối để WinRS chạy end-to-end. Spec 0023.

## 2026-07-08 — 0022-enroll-clientcert-negotiation.patch
**WHY:** WinRS connect FAIL 401 dù cert đúng — WinRM listener KHÔNG bật "Negotiate Client Certificate"
trên binding HTTP.sys → server không hỏi client cert → không map → 401. (Debug thật trên máy Windows.)
**WHAT:** winrs-enroll.ps1: enable clientcertnegotiation trên 0.0.0.0:5986 (netsh update, fallback
delete+add) + dọn cert/mapping CN=user cũ trước khi tạo (idempotent). Fix quyết định cho cert-auth. Spec 0022.

## 2026-07-08 — 0021-discovery-autoadd-schedule.patch
**WHY:** Network Discovery nâng cấp — quét định kỳ, Add all/Add-per-client vào inventory, tự thêm máy WinRS-connect được.
**WHAT:** quickwin_discovery.go +config(auto_scan/interval/auto_add_winrs) + addHostsToWinRSInventory +
winrsReachable/autoAddReachable + SetDiscoveryConfig/AddDiscoveryToInventory + RunDueScans; router +2 route;
root.go goroutine scheduler 60s; UI switch định kỳ+interval+auto-add + nút "Add all"/"Add". Spec 0021.

## 2026-07-08 — 0020-enroll-pem-pkcs1.patch
**WHY:** winrs-enroll.ps1 (chạy PS 5.1 do relaunch 0019) FAIL export key — ExportPkcs8PrivateKey là
API .NET Core, .NET Framework 4.x không có ("does not contain GetRSAPrivateKey").
**WHAT:** tự encode PKCS#1 DER từ RSAParameters (Enc-Len/Int/Seq + ConvertTo-Pkcs1B64) → PEM
"RSA PRIVATE KEY". Verify: PS 5.1 export OK + Go tls.X509KeyPair parse OK. Spec 0020.

## 2026-07-08 — 0019-enroll-ps51-compat.patch
**WHY:** script enroll FAIL trong PowerShell 7 (New-LocalUser/cert/WSMan không tương thích — lỗi
TelemetryAPI); + tiếng Việt trong chuỗi làm PS5.1 parse lỗi encoding.
**WHAT:** winrs-enroll.ps1 + ssh-enroll.ps1: guard tự relaunch bằng Windows PowerShell 5.1 khi chạy
PS>=6; chuyển toàn bộ chuỗi/comment sang ASCII/English. Giữ logic + placeholder 1-click. Spec 0019.

## 2026-07-08 — 0018-network-discovery.patch
**WHY:** autodiscovery subnet — quét dãy mạng tìm IP online + phân loại; default mỗi project có 192.168.0.0/16.
**WHAT:** quickwin_discovery.go (scanner TCP-probe bounded + deadline; phân loại managed[inventory
WinRS]/exception/gateway[.1/.254]/unmanaged; config JSON discovery dir); router +5 route /discovery/*;
UI NetworkDiscovery.vue (subnet CRUD, scan, bảng device, legend, ignore) + nav + tooltip + i18n. Spec 0018.

## 2026-07-07 — 0017-winrs-history.patch
**WHY:** WinRS Console nhớ lần gõ trước — last-run tự prefill + lịch sử execute lưu tạm ra file.
**WHAT:** quickwin_winrs_history.go (ghi JSONL ts/host/command/exit vào tmp dir project, giữ 100 dòng;
GetWinRSHistory mới-nhất-trước); WinRSExec ghi history mỗi lần chạy; router +GET /winrs/history;
UI localStorage nhớ last-run + bảng "Recent commands" + nút Reuse. Spec 0017.

## 2026-07-07 — 0016-enroll-1click.patch
**WHY:** enrollment 1-click — bỏ copy cert thủ công; máy đích tự nạp cert + tự thêm vào inventory.
**WHAT:** quickwin_enroll_token.go (token HMAC stateless key=CookieHash TTL30' + GetEnrollToken authed +
EnrollWinRS public nhận cert→lưu certs dir+thêm host inventory, lọc traversal/size); enroll.go chèn
server+token vào script; winrs-enroll.ps1 block tự POST cert; router +enroll-token(authed)+/enroll/winrs(public);
UI nút "Enroll 1-click". Spec 0016.

## 2026-07-07 — 0015-i18n-quickwin-ui.patch
**WHY:** UI QuickWin hardcode tiếng Việt trên giao diện English → lệch ngôn ngữ (tooltip, WinRS Console…).
**WHAT:** en.js +key (tooltip* menu, winrs* console/enroll, localRepo*, loading); App.vue navTooltips→$t;
WinRSConsole/InventoryForm/Repositories/OpenITMS chuyển VN→$t hoặc English. Chỉ frontend, fallback 'en'. Spec 0015.

## 2026-07-07 — 0014-endpoint-enroll-script.patch
**WHY:** cần cách chuẩn bị máy Windows (WinRM+cert / OpenSSH) để OpenITMS quản lý — tải script ngay trên web.
**WHAT:** api/projects/scripts/{winrs-enroll,ssh-enroll}.ps1 (bật WinRM HTTPS+cert / OpenSSH, xuất PEM +
hướng dẫn); quickwin_enroll.go (go:embed, GetEnrollScript tải theo kind=winrs|ssh); router.go +route
GET /endpoint/enroll-script/{kind}; UI WinRSConsole panel tải + InventoryForm link. Spec 0014.

## 2026-07-07 — 0013-menu-tooltips.patch
**WHY:** menu sidebar khó hiểu với người không chuyên IT — cần giải thích "cần tạo gì / ý nghĩa".
**WHAT:** App.vue computed navTooltips (mô tả tiếng Việt mỗi menu) + bọc 2 vòng render nav bằng
v-tooltip right. Chỉ frontend. Spec 0013.

## 2026-07-07 — 0012-view-local-repo.patch
**WHY:** repo local Gitea (0008) có token nhúng trong GitURL → cần link xem/mở repo per project (G-05).
**WHAT:** api/projects/quickwin_gitea_view.go (GetProjectGiteaRepo: tìm repo local, strip token,
trả web_url sạch); router.go +GET /gitea/repo; UI Repositories.vue banner "Repo local (Gitea)" +
nút "Mở trên Gitea". Spec 0012.

## 2026-07-07 — 0011-seed-host-project.patch
**WHY:** yêu cầu gốc "mặc định có 1 project quản lý chỉnh máy chính host" — có sẵn sau cài đặt.
**WHAT:** api/projects/quickwin_seed.go (SeedHostProjectIfEmpty: chưa có project + đã có user →
tạo project "Host" + None key + view All + auto-repo Gitea + inventory WinRS trỏ 127.0.0.1,
idempotent); cli/cmd/root.go gọi seed trong runService sau accessKeyService. Spec 0011.

## 2026-07-07 — 0010-winrs-console.patch
**WHY:** gõ lệnh nhanh xuống 1 host Windows mà không cần tạo template + inventory + task.
**WHAT:** api/projects/quickwin_winrs.go (GetWinRSCerts list *.pem + WinRSExec chạy 1 lệnh pwsh
qua winrsexec, cert-auth, chặn path-traversal, trả stdout/stderr/exit_code); router.go +2 route
(/winrs/certs, /winrs/exec) gate CanManageProjectResources; UI WinRSConsole.vue + route +
nav "WinRS Console". Đồng bộ, không lưu lịch sử. Spec 0010.

## 2026-07-07 — 0009-winrs-app.patch
**WHY:** OpenITMS chạy pwsh trực tiếp xuống Windows host qua WinRM (pwsh → winrs) — Semaphore gốc
chỉ có inventory/app hướng-Ansible, thiếu đường chạy này.
**WHAT:** db: InventoryWinRS + AppWinRS (+InventoryTypes). db_lib: WinRSApp (LocalApp) chạy script
pwsh xuống mỗi host qua quickwin.dev/winrsexec (WinRM cert-auth, -EncodedCommand); AppFactory
dispatch AppWinRS; api/projects/inventory.go Add/UpdateInventory chấp nhận type "winrs"
(thiếu → API 400, không tạo được inventory). UI: "WinRS Endpoints" trong NEW INVENTORY +
editor host/cert, app "WinRS" (constants). go.mod +winrsexec (replace ../winrs-exec). Spec 0009.

## 2026-07-07 — 0008-gitea-autorepo.patch
**WHY:** mỗi project mới tự có repo git local (ADR-0005) — Gitea bundle.
**WHAT:** api/projects/quickwin_gitea.go (hook AddProject → EnsureOrg+CreateRepo+CreateRepository,
token nhúng clone URL); projects.go +1 dòng; go.mod +gitea-manager. Non-fatal khi Gitea tắt. Spec 0008.

## 2026-07-07 — 0007-service-manager.patch
**WHY:** trang Admin xem status + restart service hạ tầng (MariaDB/Gitea/app).
**WHAT:** api/quickwin_services.go (/api/services list+restart); router +1; go.mod +servicemanager;
OpenITMS.vue +tab Services. Status via TCP (mọi OS); restart via systemctl (Linux). Spec 0007.

## 2026-07-06 — 0006-management-ui.patch
**WHY:** UI quản lý OpenITMS (Plugins/Registry/Hardening) — tính năng core theo plan.
**WHAT:** view mới OpenITMS.vue (3 tab, axios gọi /api/plugins + /api/registry + hardening);
router +route /openitms; App.vue +nav item. Thuần frontend. Verify: ui-build CI (cần node).

## 2026-07-06 — 0005-registry-client.patch
**WHY:** registry client = ngoại lệ core được duyệt (ADR-0003), hạ tầng plugin/template phụ thuộc.
**WHAT:** file mới api/quickwin_registry.go (search + install, verify sig+checksum, unpack plugin);
router.go +1 dòng; go.mod +require/replace quickwin.dev/registry. E2E: registry-through-core.sh.

## 2026-07-06 — 0003-default-password-banner.patch
**WHY:** admin mặc định admin/quickwin123 (yêu cầu gốc) → ép nhắc đổi mật khẩu lần đầu (plan 7.3).
**WHAT:** App.vue thêm 1 v-alert (banner cam) + computed showDefaultPasswordWarning +
data defaultPasswordDismissed (localStorage) + 2 method. Chỉ frontend, không đụng backend.
Spec: 0003-default-password-banner.md.

## 2026-07-06 — 0002-branding.patch
**WHY:** MIT không cấp quyền trademark — thay tên/logo "Semaphore" bằng OpenITMS-SMB (ADR-0004 #1).
**WHAT:** title index.html; logo.svg + favicon.svg (SVG text); key i18n `ansibleSemaphore`
→ 'OpenITMS-SMB' (5 file lang). Asset-replace thuần, không đổi code. Phần cần node
(favicon.png, chuỗi i18n dài, trang About attribution) → P1-07b. Spec: 0002-branding.md.

## 2026-07-06 — 0001-plugin-manager-hook.patch
**WHY:** cần 1 điểm hook duy nhất để start Plugin Manager + mount API động /api/plugins/
sau middleware authentication của Semaphore (plugin thừa hưởng authn core).
**WHAT:** file mới `api/quickwin_plugins.go` (keo ~60 dòng); `api/router.go` +1 dòng;
`go.mod` +require/replace `quickwin.dev/*`; `go.sum` generated (conflict → `go mod tidy`).
Sửa tay vào file upstream: **6 dòng** (1 router.go + 5 go.mod). E2E: tests/e2e/plugin-through-core.sh.
