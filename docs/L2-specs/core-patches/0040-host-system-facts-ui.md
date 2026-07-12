---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0040-host-system-facts-ui.patch, plugins/device-inventory/, core-patches/0035-device-inventory-ui.patch]
---

# Spec patch 0040 — Tab "Hệ thống": hiển thị network/route/dns/user/group/env/ntp/domain/profile

## Mục đích (WHY)
Khi chuyển host inventory sang osquery, mất các trường mà core cũ (0025-0028) thu: DNS, network/IP,
route table, user, group, domain, env, NTP, profile. Backend đã bổ sung (bảng `di_device_fact`); patch
này thêm UI hiển thị.

## Thay đổi (core patch — UI)
| File | Loại | Nội dung |
|---|---|---|
| `web/src/views/project/DeviceInventory.vue` | sửa | Detail host +tab "Hệ thống" hiện `facts` nhóm theo category (mỗi nhóm 1 bảng name/detail); computed `factCategories` + method `factsIn(cat)` |
| `web/src/lang/en.js`, `vi.js` | +key | `devTabSystem`, `devNoFacts`, `fact*` (Network/Route/Dns/Domain/Ntp/User/Group/Profile/Env) |

## Backend (plugin — commit thẳng)
- `di_device_fact(device_id, category, name, detail)`.
- Thu: **osquery** `interface_addresses`(network) / `routes` / `users` / `groups`; **PowerShell**
  `Get-DnsClientServerAddress`(dns) / `[Environment]::GetEnvironmentVariables('Machine')`(env) /
  registry `W32Time`(ntp) / CIM `Win32_ComputerSystem`(domain) / CIM `Win32_UserProfile`(profile).
  Chạy chung 1 script over WinRS/local (osquery + PS bù chỗ osquery Windows thiếu).
- `storeHost` lưu facts; `getDevice` trả `facts[]`.

## Verify (E2E máy thật, local self-collect)
- Facts thu được: network=3 / route=18 / user=14 / group=19 / dns=2 / env=20 / ntp=1 / domain=1 /
  profile=9. Tab "Hệ thống" hiện đủ nhóm. eslint sạch; chain 0001–0040 build.

## Liên quan
- UI Devices: [[0035-device-inventory-ui]]. Plugin: [[plugins-device-inventory]].
