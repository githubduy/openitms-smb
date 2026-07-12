# core-patches CHANGELOG

Mỗi patch thêm/sửa/xóa phải có 1 entry ở đây (mới nhất lên đầu).
Format: `## <ngày> — <patch-file>` + WHY (vì sao cần) + WHAT (đổi gì, mức cao).

## 2026-07-11 — 0040-host-system-facts-ui.patch
**WHY:** Host thu thêm facts (network/route/dns/domain/ntp/user/group/profile/env) — cần hiển thị UI.
**WHAT:** `DeviceInventory.vue`: +tab "Hệ thống" ở detail host, hiện facts nhóm theo category
(`factCategories` + method `factsIn`). +i18n `devTabSystem`/`fact*`. Backend (plugin, commit riêng):
`di_device_fact` + osquery (network/route/user/group) + PowerShell (dns/env/ntp/domain/profile).
Verify E2E: local collect → tab Hệ thống hiện network=3/route=18/user=14/.../env=20. Spec 0040.

## 2026-07-11 — 0039-auto-collect-ui.patch
**WHY:** Thu định kỳ tự động không nhập creds lại. Kết nối đã lưu trong di_device (0038) → UI chỉ cần
bật/tắt scheduler + chu kỳ.
**WHAT:** `DeviceInventory.vue` +switch "Tự động thu" + ô interval → GET/POST `/plugins/device-inventory/
config`. +i18n `devAutoCollect*`/`devInterval`. Scheduler backend (plugin, commit riêng): di_config +
GET/POST /config + goroutine `runScheduler` (mỗi phút, tới hạn → thu mọi device có conn_type qua
`doCollect`). Verify E2E: GET/POST config lưu (enabled+interval); scheduler goroutine chạy. Spec 0039.

## 2026-07-11 — 0038-unify-device-model.patch
**WHY:** User muốn gộp Inventory + CMDB thành 1 model quản lý: device = identity + kết nối + tài sản.
Menu Inventory native và Thiết bị (CMDB) dễ nhầm.
**WHAT:** `DeviceInventory.vue` thành hub thống nhất: nút "Thêm thiết bị" (host + conn_type
local/winrs/snmp → lưu kết nối vào di_device + thu ngay), per-row Collect/Delete. `App.vue`: đổi menu
Inventory native → "Máy đích (task)" + tooltip rõ (de-emphasize, giữ cho task/Ansible). +i18n.
Backend (plugin đã commit riêng): di_device +cột conn_*, POST/DELETE /device, POST /collect {id}
dispatch local/winrs/snmp. Verify E2E: add device (local/winrs/snmp) lưu+thu; per-row collect/delete;
menu đổi tên. Spec 0038.

## 2026-07-11 — 0037-local-self-collect.patch
**WHY:** Phần mềm chạy local nên inventory seed "Máy host (WinRS)" 127.0.0.1 (đi vòng WinRS tới
loopback, thực tế fail AccessDenied) là thừa. Server nên tự kiểm kê trực tiếp.
**WHAT:** `quickwin_seed.go` bỏ seed inventory 127.0.0.1. `DeviceInventory.vue` +nút "Thu server này
(local)" → POST `/collect {local:true}` (plugin device-inventory chạy osquery LOCAL trên server, không
WinRS/cert). +i18n `devCollectLocal*`. Verify E2E: nút hiện; local collect đi đường local exec (không
WinRS, 1.3s, osquery-not-found báo rõ). Spec 0037.

## 2026-07-11 — 0036-device-collect-host-ui.patch
**WHY:** UI Devices (0035) chỉ thu được switch; cần nút thu HOST (osquery) từ giao diện để dùng Phase 5
(plugin tự cài osquery khi máy đích chưa có).
**WHAT:** `DeviceInventory.vue` +nút "Collect host" + dialog (host, chọn cert từ `/winrs/certs`, checkbox
auto-deploy) → POST `/plugins/device-inventory/collect`. +i18n `devCollectHost`/`devAutoDeploy*`.
Verify E2E: nút hiện, dialog nạp cert list, POST collect. Spec 0036.

## 2026-07-11 — 0035-device-inventory-ui.patch
**WHY:** Plugin `device-inventory` (CMDB) chưa có UI — frontend Semaphore không có cơ chế render
plugin UI (`menu_title` chưa ai đọc). User cần thấy device/asset (host + switch).
**WHAT:** View "Devices" (`DeviceInventory.vue`) gọi API động `/api/plugins/device-inventory`
(devices/device/changes/collect-switch): bảng device (host/switch), detail theo kind (host→software/
services/patches; switch→interfaces/neighbors/fdb), tab Changes, form "Collect switch (SNMP)" v2c/v3.
+nav item "Devices" + tooltip (App.vue), +route (router). Logic ở plugin, core chỉ là vỏ UI.
+i18n dev*/tooltipDevices (en+vi). Verify E2E: menu Devices hiện, list rỗng, dialog collect. Spec 0035.

## 2026-07-11 — 0034-repo-file-browser.patch
**WHY:** Ở ô script filename user phải gõ tay tên file, không biết repo có gì; cần duyệt cây thư mục
git để chọn file + tạo mới thư mục/file ngay đó.
**WHAT:** Backend hook mỏng `ListTemplateRepoDir` + route `GET /gitea/dir` (logic `ListDir` ở
giteamanager). UI TemplateForm.vue: nút Browse + dialog duyệt (folder/file, lên cấp, chọn file →
điền filename), nút "New folder" (tạo `<dir>/.gitkeep`) + "New file" (mở editor 0029 tại thư mục
hiện tại). +i18n `scriptBrowse*`. Verify E2E: liệt kê repo, tạo folder, chọn file. Spec 0034.

## 2026-07-11 — 0033-repo-scaffold.patch
**WHY:** Repo local mới tạo (0008) trống trơn → user không biết bắt đầu từ đâu.
**WHAT:** `quickwin_gitea.go` sau `CreateRepo` gọi `scaffoldRepo()` PutFile README.md + scripts/ +
playbooks/ + deploys/ (mỗi folder 1 README.md vì git không track folder rỗng), dùng `PutFile` (0029).
Non-fatal (chỉ log nếu fail). Verify E2E: project mới → repo có README + 3 folder. Spec 0033.

## 2026-07-11 — 0032-template-form-ux.patch
**WHY:** Tạo Task Template: SMB thường chỉ 1 repo local → nên mặc định chọn sẵn (khỏi để trống);
3 tab Task/Build/Deploy khó hiểu với người không chuyên → cần tooltip.
**WHAT:** `TemplateForm.vue`: `afterLoadData` default `repository_id` = repo local (Gitea) khi tạo
mới (hoặc repo duy nhất); thêm `:title` tooltip 3 tab type (`templateTypeTooltip`). +i18n
`tooltipTemplateTask/Build/Deploy`. Verify E2E. Spec 0032.

## 2026-07-11 — 0031-vietnamese-language.patch
**WHY:** OpenITMS-SMB hướng tới IT doanh nghiệp VN nhưng bộ chọn ngôn ngữ chưa có tiếng Việt.
**WHAT:** Thêm `web/src/lang/vi.js` (dịch full 538/538 key từ en.js, giữ nguyên placeholder/comment);
đăng ký `vi: 'Tiếng Việt'` vào map `LANGUAGES` trong `App.vue` (hiện trong switcher); thêm cờ Việt Nam
`web/public/flags/vi.svg`. `vi.js` tự load qua `require.context` (index.js không cần sửa).
Verify E2E: set lang=vi → UI dịch (Lịch sử/Thống kê/Hoạt động/Cài đặt); flag served 200. Spec 0031.

## 2026-07-11 — 0030-discovery-separate-inventory.patch
**WHY:** Network Discovery đang trộn máy remote vào inventory local "Máy host (WinRS)" (dùng
`invs[0]` = 127.0.0.1). Máy host mặc định phải hiểu là máy local; máy discovery phải vào inventory riêng.
**WHAT:** `addHostsToWinRSInventory` chọn/tạo inventory tên "Discovered machines (WinRS)" theo tên,
không còn lấy `invs[0]` → không đụng inventory local host. `managedHosts` giữ nguyên (quét mọi WinRS
inventory nên máy đã thêm vẫn được đánh dấu managed). Verify E2E: add discovery → inventory mới tách
riêng, "Máy host (WinRS)" vẫn chỉ 127.0.0.1. Spec 0030.

## 2026-07-11 — 0029-template-script-editor.patch
**WHY:** SMB IT chưa có script trong repo — cần tạo/sửa file .yml/.sh/.ps1 ngay trên UI và
commit vào git local, khỏi phải clone repo + dùng git ngoài.
**WHAT:** Template form thêm nút "New/Edit file" cạnh ô script filename → mở editor codemirror,
điền sample theo app type (ansible/bash/powershell/python/terraform); Lưu → POST /gitea/file →
commit vào repo local (Gitea). Backend `quickwin_gitea_commit.go` (hook mỏng) + route GET/POST
`/gitea/file`; logic Gitea Contents API (`GetFile`/`PutFile`) ở giteamanager ngoài cây upstream.
Test: gitea-manager GetFile/PutFile (create→POST, update→PUT kèm sha). Spec 0029.

## 2026-07-11 — DEPRECATE core inventory (Phase 4): gỡ 0025(inventory)/0026/0027/0028
**WHY:** device/asset inventory đã chuyển sang **plugin `device-inventory`** (osquery host + SNMP switch,
CMDB trong MariaDB) — đúng nguyên tắc plugin-first. Inventory tự-viết trong core thành thừa/trùng.
**WHAT:** XÓA hẳn patch `0026-inventory-extra-schedule`, `0027-inventory-filter-diff-export`,
`0028-inventory-export-detail`; và phần inventory của `0025` (collect-inventory.ps1,
quickwin_winrs_inventory.go, routes /winrs/inventory*, scheduler root.go, section Inventory +i18n inv*
trong WinRSConsole.vue). Chain 0029-0035 apply sạch không cần các patch này (đã kiểm chứng).
Config inventory (host/cred) vẫn ở DB core (Semaphore native). Device data → plugin.

## 2026-07-11 — 0025-cert-security-panel.patch (tách từ 0025 cũ)
**WHY:** file cert `.pem` chứa private key = credential admin toàn fleet; cần cảnh báo rõ rủi ro khi lộ.
**WHAT:** WinRSConsole.vue +panel cảnh báo bảo mật cert (.pem) + en.js `certSecurity*`. (Phần bảo mật
cert giữ lại từ 0025 cũ; phần inventory đã bỏ ở Phase 4.) Verify: panel hiện trên WinRS Console. Spec 0025.

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
