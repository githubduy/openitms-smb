---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0015-i18n-quickwin-ui.patch, core-patches/0013-menu-tooltips.patch]
---

# Spec patch 0015 — i18n cho UI QuickWin (English, fallback locale)

## Mục đích (WHY)
Các phần UI QuickWin thêm ở 0006/0009–0014 (tooltip menu, WinRS Console, panel tải script,
banner repo local, lỗi OpenITMS) hardcode **tiếng Việt**, trong khi giao diện Semaphore mặc định
**English** → lệch ngôn ngữ. Chuyển sang i18n (`$t`) với chuỗi English trong `en.js`. Vì
`fallbackLocale: 'en'`, mọi locale (kể cả khi người dùng chưa có bản dịch riêng) đều hiển thị
English đồng nhất với phần còn lại của app.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `web/src/lang/en.js` | +key | `tooltip*` (10 menu), `winrsConsoleTitle/Hint`, `winrsEnroll*`, `winrsScriptBtn`, `sshScriptBtn`, `winrsCertLabel`, `winrsNoCerts`, `winrsRun`, `winrsUnknownError`, `winrsHostHint` (interpolation certsDir/envVar), `winrsDownloadScript(Tail)`, `localRepoLabel`, `openOnGitea`, `loading` |
| `web/src/App.vue` | sửa | `navTooltips` trả `this.$t('tooltip*')` thay vì chuỗi VN cứng |
| `web/src/views/project/WinRSConsole.vue` | sửa | tiêu đề/hint/panel tải/label/nút/lỗi → `$t`; placeholder host → English |
| `web/src/components/InventoryForm.vue` | sửa | caption editor WinRS → `$t('winrsHostHint', {...})` + link tải; placeholder → English |
| `web/src/views/project/Repositories.vue` | sửa | banner repo local → `$t('localRepoLabel')` + `$t('openOnGitea')` |
| `web/src/views/OpenITMS.vue` | sửa | 2 chuỗi lỗi VN → English |

## Ghi chú thiết kế
- **Chỉ frontend**, không đụng backend.
- Chuỗi chuẩn ở `en.js`; chưa tạo `vi.js` (app chưa có locale vi). Muốn thêm tiếng Việt đầy đủ =
  tạo `web/src/lang/vi.js` dịch toàn bộ (việc lớn, tách riêng) — khi đó tooltip QuickWin tự có sẵn key.
- `winrsHostHint` dùng named interpolation (`{certsDir}`, `{envVar}`) để chèn `certs/` +
  `QUICKWIN_WINRS_DEFAULT_CERT` không cần nối chuỗi.

## Verify
- eslint 6 file sạch.
- Chain 0001–0015 apply + build; bundle chứa chuỗi English ("The target machines to run tasks…"),
  KHÔNG còn chuỗi VN cũ ("Nơi định nghĩa", "Danh sách máy đích").

## Liên quan
- Tooltip gốc (VN): [[0013-menu-tooltips]].
