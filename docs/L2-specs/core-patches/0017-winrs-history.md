---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0017-winrs-history.patch, core-patches/0010-winrs-console.patch]
---

# Spec patch 0017 — Lịch sử WinRS Console (last-run + history file)

## Mục đích (WHY)
WinRS Console (0010) chạy 1 lần rồi mất input/kết quả. Người dùng muốn: (1) mở lại console thấy
**last-run tự điền** (host/port/cert/lệnh), (2) xem **lịch sử các lệnh đã chạy** để lặp lại nhanh.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_winrs_history.go` | MỚI | `appendWinRSHistory` ghi JSONL `{ts,host,command,exit_code,ok,error}` vào `<project tmp>/winrs-history.jsonl`, giữ ≤100 dòng (cắt cũ nhất); `GetWinRSHistory` đọc + trả mới-nhất-trước; **KHÔNG lưu stdout** (nhẹ + tránh lộ dữ liệu) |
| `api/projects/quickwin_winrs.go` | sửa (0010) | `WinRSExec` đọc project từ context + gọi `appendWinRSHistory` sau mỗi lần chạy (cả thành công lẫn lỗi) |
| `api/router.go` | +1 route | `GET /project/{id}/winrs/history` (projectUserAPI) |
| `web/src/views/project/WinRSConsole.vue` | sửa | `localStorage['winrs-lastrun-<projectId>']` lưu/khôi phục last-run (prefill khi mở); tải + hiển thị history (bảng "Recent commands": time/host/exit/command + nút **Reuse** điền lại host+command); reload history sau mỗi lần chạy |
| `web/src/lang/en.js` | +key | `winrsHistory`, `winrsHistoryEmpty`, `winrsReuse` |

## Thiết kế / lưu ý
- **Last-run** = client-side (localStorage), gồm cả cert/port (file cục bộ, không cần lên server).
- **History** = server-side file tạm trong tmp dir project → chia sẻ giữa các phiên/trình duyệt, tự
  dọn khi ClearProjectTmpDir. Chỉ metadata (không stdout).
- History gate `CanManageProjectResources`. Reuse chỉ điền host+command (cert giữ lựa chọn hiện tại).
- Cắt 100 dòng: đọc-thêm-ghi lại (file nhỏ, chấp nhận được cho console tương tác).

## Verify (E2E app thật)
- Chạy lệnh (host giả) → history ghi 1 entry (host/command/ok=false/error); chạy lệnh 2 →
  `GET /winrs/history` trả 2 entry **mới nhất trước** (whoami/10.88 ở đầu). stdout không lưu.
- eslint sạch; chain 0001–0017 build; bundle chứa UI history.

## Liên quan
- Console gốc: [[0010-winrs-console]].
