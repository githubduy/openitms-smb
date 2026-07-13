---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0042-enroll-custom-account.patch, core-patches/0014-endpoint-enroll-script.patch, core-patches/0016-enroll-1click.patch]
---

# Spec patch 0042 — Enroll Windows: chọn account customer có sẵn (regen script)

## Mục đích (WHY)
Script enroll WinRS mặc định **tạo account mới `openitms`** trên máy đích. Khách hàng nhiều khi đã có
**account riêng** (customer account) và muốn OpenITMS dùng account đó (không tạo account lạ). Cần: nhập
account trên UI + **regen script enroll** theo account đó (tạo mới HOẶC dùng account có sẵn).

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/scripts/winrs-enroll.ps1` | sửa | +biến inject `@@ENROLL_USER@@`/`@@ENROLL_CREATE@@` → `$User`/`$CreateUser`. Bước tạo user bọc trong `if ($CreateUser)`; nhánh else = **dùng account có sẵn** (không New-LocalUser/đổi mật khẩu, Fail nếu account chưa tồn tại). Add-admin + LocalAccountTokenFilterPolicy + map cert chạy cho cả 2 |
| `api/projects/quickwin_enroll.go` | sửa | `GetEnrollScript` đọc `?user=<name>&create=<0\|1>` (mặc định openitms/1) → inject |
| `api/projects/quickwin_enroll_token.go` | sửa | +placeholder `enrollUserPlaceholder`/`enrollCreatePlaceholder`; `enrollScriptInject` +user/create |
| `web/src/views/project/WinRSConsole.vue` | sửa | Panel enroll +ô "Account Windows" + checkbox "Dùng account có sẵn"; computed `enrollAccountQuery`/`winrsScriptUrl` → wire vào nút download + `enroll1click` URL |
| `web/src/lang/en.js`, `vi.js` | +key | `enrollAccountLabel`, `enrollUseExisting`, `enrollAccountHint` |

## Hành vi
- **Mặc định** (openitms + tạo mới): như cũ — tạo account `openitms`, mật khẩu random, admin, map cert.
- **Custom account có sẵn** (`create=0`): script KHÔNG tạo/đổi account; yêu cầu account tồn tại; vẫn
  add vào Administrators + set LocalAccountTokenFilterPolicy + map cert client → account đó.
- Cả download thủ công lẫn nút 1-click đều regen theo account đã nhập.

## Giới hạn
- Account **local** (map cert local). Domain account map-cert khác (chưa hỗ trợ). "Dùng account có
  sẵn" vẫn cấp quyền admin + bật token policy cho account đó (cần cho WinRM cert-auth).

## Verify (E2E)
- `GET .../enroll-script/winrs` (plain) → `$__InjUser="openitms"` `$__InjCreate="1"`.
- `?user=acme_admin&create=0` → `$__InjUser="acme_admin"` `$__InjCreate="0"`. PS parse OK; eslint sạch;
  chain 0001–0042 build.

## Liên quan
- Enroll script: [[0014-endpoint-enroll-script]]. 1-click: [[0016-enroll-1click]].
