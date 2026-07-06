---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [plugin-manager/schema/]
---

# Spec: Plugin manifest (`plugin.yaml`)

Schema máy-đọc: `plugin-manager/schema/plugin.schema.json` (JSON Schema 2020-12) — nguồn
chân lý; doc này giải thích semantics. Manifest nằm ở gốc thư mục plugin:
`plugins/<name>/plugin.yaml`.

## Ví dụ đầy đủ
```yaml
name: winrs-cert
version: 0.1.0
license: MIT
description: Run commands on Windows 11 via WinRS with certificate authentication
protocol_version: 1
entrypoint:
  linux-amd64: winrs-cert
  windows-amd64: winrs-cert.exe
checksum:
  winrs-cert: "sha256:<64 hex>"
  winrs-cert.exe: "sha256:<64 hex>"
routes:
  - method: POST
    path: exec
    description: Execute a command on a Windows host
    require_admin: true
permissions: [certs:read, inventory:read, network:outbound]
min_core_version: "0.1.0"
ui:
  menu_title: WinRS
```

## Semantics quan trọng
- **`license` bắt buộc** — registry và Plugin Manager từ chối manifest thiếu license.
- **`checksum`**: Plugin Manager verify sha256 từng file TRƯỚC khi launch; lệch → từ chối load
  + log error. Registry ký trên toàn artifact (tarball) — 2 lớp độc lập.
- **`permissions`**: enforce ở core (deny-by-default). Plugin gọi thứ ngoài permissions khai
  → request bị chặn + audit log. UI hiển thị danh sách quyền khi người dùng cài plugin.
- **`entrypoint`**: chọn theo platform của core đang chạy; `python` entrypoint yêu cầu host có
  python3 ≥ 3.10 (Plugin Manager kiểm tra + báo lỗi rõ nếu thiếu).
- **`routes`** phải khớp `Metadata.routes` plugin trả qua gRPC — lệch → từ chối load (chống
  manifest nói dối để qua mặt review).
- Version core so sánh semver với `min_core_version`; không đạt → không load, hiện lý do trên UI.

## Case lỗi chuẩn (validator phải bắt được — test P1-01)
1. Thiếu `license` → reject.
2. `version` không phải semver (vd `1.0`) → reject.
3. `name` viết hoa/underscore (vd `WinRS_Cert`) → reject.
4. `routes[].method` ngoài GET/POST/PUT/DELETE → reject.
5. `checksum` không đúng format `sha256:<64 hex>` → reject.
6. Field lạ ngoài schema (`additionalProperties: false`) → reject.
