---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [core-patches/0003-default-password-banner.patch]
---

# Spec patch 0003 — Banner nhắc đổi mật khẩu mặc định

## Mục đích (WHY — dùng khi rebase)
Bộ cài đặt admin mặc định `admin/quickwin123` (yêu cầu gốc, zero-config — plan 7.3).
Bắt buộc **nhắc người dùng đổi ngay lần đầu** bằng banner cảnh báo.

## Thay đổi (chỉ frontend — `web/src/App.vue`)
| Phần | Nội dung |
|---|---|
| Template | +1 `<v-alert type="warning">` (chèn sau banner BoltDB có sẵn, trong `<v-main>` đầu) — hiện khi `showDefaultPasswordWarning`; 2 nút: "Đổi mật khẩu" (mở `userDialog`) + "Đã đổi/Dismiss" |
| computed | `showDefaultPasswordWarning` = user là admin **và** chưa dismiss |
| data | `defaultPasswordDismissed` khởi tạo từ `localStorage['openitms_default_pwd_ack']` |
| methods | `openChangePassword` (mở `userDialog` — đã bind `:item-id="user.id"`), `dismissDefaultPasswordWarning` (set localStorage + flag) |

**KHÔNG đụng backend**: route đổi password `/users/{user_id}/password` đã có sẵn trong core;
banner chỉ điều hướng người dùng tới dialog đổi password hiện có.

## Giới hạn có chủ đích (v1)
- Banner là **nhắc nhở dismiss được** (không cưỡng chế cứng) — backend không track được
  "password hiện tại có phải mặc định" (đã hash). Đủ cho mục tiêu nhắc lần đầu.
- Nâng cấp tương lai (nếu cần cưỡng chế): thêm cờ `must_change_password` vào user khi
  installer tạo admin → cần backend (patch riêng + ADR). Ghi nhận, chưa làm.
- Text song ngữ inline (như banner BoltDB gốc) — không thêm i18n key mới.

## Verify
- `apply-patches.sh` chuỗi 0001+0002+0003 từ upstream sạch PASS; build backend PASS.
- UI thật: CI job `ui-build` (FULL_UI, cần node) build frontend — bắt lỗi cú pháp Vue.
