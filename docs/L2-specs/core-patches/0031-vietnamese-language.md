---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0031-vietnamese-language.patch, core-patches/0015-menu-tooltips.patch]
---

# Spec patch 0031 — Thêm tiếng Việt vào bộ chọn ngôn ngữ

## Mục đích (WHY)
OpenITMS-SMB nhắm tới IT doanh nghiệp nhỏ Việt Nam, nhưng UI (kế thừa Semaphore) chỉ có các ngôn ngữ
en/de/es/fr/... — **chưa có tiếng Việt** trong menu chọn ngôn ngữ. Người dùng cuối không rành tiếng Anh
khó dùng. Patch thêm gói tiếng Việt đầy đủ + đăng ký vào switcher.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `web/src/lang/vi.js` | file mới | Dịch **full 538/538 key** từ `en.js` sang tiếng Việt (giữ nguyên key, placeholder `{...}`, chuỗi nối `+`, comment `//`, thuật ngữ IT giữ tiếng Anh khi phù hợp) |
| `web/src/App.vue` | sửa | Thêm `vi: { title: 'Tiếng Việt' }` vào map `LANGUAGES` → xuất hiện trong dropdown chọn ngôn ngữ |
| `web/public/flags/vi.svg` | file mới | Cờ Việt Nam (đỏ + sao vàng, 512×512) — switcher render `flags/vi.svg` |

## Cơ chế đăng ký
- `src/lang/index.js` dùng `require.context('.', false, /\.js$/)` → **tự load mọi `.js`** trong thư mục
  lang. Thêm `vi.js` là đủ để `messages.vi` có mặt; **không cần sửa index.js**.
- `plugins/i18.js` đặt `fallbackLocale: 'en'` + `silentFallbackWarn: true` → key nào thiếu tự lùi về
  tiếng Anh (hiện đủ 538/538 nên không lùi).
- Switcher trong `App.vue` build danh sách từ `Object.keys(LANGUAGES)` → phải có entry `vi` mới hiện.

## Verify (E2E máy thật)
- Parity: `en.js` 538 key == `vi.js` 538 key, 0 key thiếu; eslint sạch.
- Runtime: `localStorage.lang='vi'` → reload → UI tiếng Việt (tab "Lịch sử"/"Thống kê"/"Hoạt động"/
  "Cài đặt", nút "Sửa menu").
- `GET /flags/vi.svg` → 200; bundle chứa "Tiếng Việt".
- Chain 0001–0031 build.

## Rebase
- `vi.js` độc lập (file mới) — hiếm khi conflict.
- Neo `App.vue`: map `LANGUAGES`. Nếu upstream đổi cấu trúc switcher → thêm lại entry `vi`.
- Khi upstream thêm key mới vào `en.js` → bổ sung key tương ứng vào `vi.js` (hoặc để fallback en).

## Liên quan
- Tooltip menu (đã i18n): [[0015-menu-tooltips]].
