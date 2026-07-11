---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0030-discovery-separate-inventory.patch, core-patches/0018-network-discovery.patch, core-patches/0021-discovery-periodic-autoadd.patch]
---

# Spec patch 0030 — Discovery: máy phát hiện vào inventory riêng (tách khỏi máy local)

## Mục đích (WHY)
Inventory seed "Máy host (WinRS)" trỏ `127.0.0.1` = **máy local chạy OpenITMS**. Nhưng
`addHostsToWinRSInventory` (0021) chọn `invs[0]` — chính là inventory local đó — rồi **append máy
remote phát hiện qua Network Discovery vào chung**. Kết quả: máy local bị lẫn IP remote, sai ngữ nghĩa.

Yêu cầu: "Máy host (WinRS)" mặc định = máy local; **máy discovery phải tạo inventory RIÊNG**.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_discovery.go` | sửa | Thêm const `winrsDiscoveredInvName = "Discovered machines (WinRS)"`. `addHostsToWinRSInventory` chọn inventory theo **tên** này (tạo nếu chưa có), **không còn `invs[0]`** → không đụng inventory local host |

`managedHosts` **giữ nguyên**: vẫn quét mọi WinRS inventory (local + discovered) để đánh dấu IP đã
managed → máy vừa thêm vào inventory discovery vẫn hiển thị "Managed" ở lần quét sau.

## Hành vi sau patch
- Discovery "Add all" / "Add" từng máy / auto-add WinRS-reachable → đổ vào **"Discovered machines
  (WinRS)"** (tạo lần đầu, các lần sau append vào đúng inventory này).
- "Máy host (WinRS)" (seed 127.0.0.1) **không bị chạm tới**.
- Cả 2 inventory đều là type WinRS → dùng chung cho WinRS Console / task template.

## Rebase
- Neo: hàm `addHostsToWinRSInventory` trong `quickwin_discovery.go` (do 0021 tạo). Nếu 0021 đổi chữ ký
  hàm → giữ Ý ĐỊNH: chọn inventory discovery theo tên, đừng lấy phần tử đầu danh sách.

## Verify (E2E)
- Project có sẵn "Máy host (WinRS)" (127.0.0.1). Chạy discovery, Add máy → xuất hiện inventory mới
  "Discovered machines (WinRS)" chứa IP remote; "Máy host (WinRS)" vẫn chỉ 127.0.0.1.
- Add lần 2 máy khác → append vào cùng "Discovered machines (WinRS)", không tạo trùng.
- `go build ./api/projects/` OK; chain 0001–0030 build.

## Liên quan
- Network Discovery: [[0018-network-discovery]], [[0021-discovery-periodic-autoadd]].
