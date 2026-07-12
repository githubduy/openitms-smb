---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0037-local-self-collect.patch, plugins/device-inventory/, core-patches/0036-device-collect-host-ui.patch]
---

# Spec patch 0037 — Server tự kiểm kê local + bỏ seed "Máy host (WinRS)" 127.0.0.1

## Mục đích (WHY)
OpenITMS chạy NGAY trên máy chủ, nên inventory seed "Máy host (WinRS)" trỏ `127.0.0.1` (thu qua WinRS
cert-auth tới loopback) là **thừa** — và thực tế fail `AccessDenied` (cert-mapping loopback). Thay vì
đi vòng WinRS, server **kết nối trực tiếp local** để tự kiểm kê.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_seed.go` | sửa | Bỏ block `CreateInventory("Máy host (WinRS)", 127.0.0.1)` — không seed nữa |
| `web/src/views/project/DeviceInventory.vue` | sửa | +nút "Thu server này (local)" (màu success) → POST `/collect {local:true, auto_deploy:true}`; method `collectLocal` + `localBusy` |
| `web/src/lang/en.js`, `vi.js` | +key | `devCollectLocal`, `devCollectLocalHint` |

**Backend (plugin, ngoài patch — commit thẳng):** `collectHostLocal` chạy `powershell -Command
<osqueryPS>` bằng `os/exec` NGAY trên server (không WinRS/cert), parse như host thường; `handleCollect`
nhận `local:true`. Server tự cài osquery (auto-deploy) nếu chưa có.

## Hành vi
- Bấm "Thu server này (local)" → plugin chạy osquery local → server xuất hiện trong CMDB (kind=host,
  host = hostname máy chủ).
- Project mới không còn inventory "Máy host (WinRS)" mặc định. Máy remote thêm qua enroll/discovery;
  server qua nút local.

## Rebase
- Neo: block seed trong `quickwin_seed.go`; toolbar `DeviceInventory.vue` (0035/0036).

## Verify (E2E)
- Trang Devices có nút "Thu server này (local)".
- Local collect: đi đường local exec (không WinRS) — verify osquery-not-found trả nhanh (~1.3s) khi
  server chưa có osquery. eslint sạch; chain 0001–0037 build.

## Liên quan
- Thu host qua WinRS: [[0036-device-collect-host-ui]]. Plugin: [[plugins-device-inventory]].
