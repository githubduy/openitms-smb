---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [plugins/hardening/]
---

# Spec: Plugin hardening

Quét cấu hình bảo mật của chính host OpenITMS-SMB, báo cáo finding + fix được mục fix được.
Plugin mẫu thứ 2 (cùng winrs-cert) — ship kèm bản cài.

## API động
| Route | Mô tả |
|---|---|
| `GET /api/plugins/hardening/scan` | trả `{findings[], issues, total}` — mỗi finding có id/title/severity/detail/fixable/passed |
| `POST /api/plugins/hardening/fix` (admin) | `{check_id}` → chmod về mode an toàn (chỉ mục fixable) |

Cũng hỗ trợ `RunTask` — task "đỏ" (FAILED) nếu có vấn đề để gây chú ý trên UI.

## Các check (v1)
| ID | Severity | Kiểm | Fixable |
|---|---|---|---|
| `default-admin-password` | high | config còn chứa `quickwin123`? (bổ sung banner patch 0003) | — |
| `config-perms` | high | config.json ≤ 0600 (chứa DB pass + keys) | ✅ chmod 0600 |
| `certs-dir-perms` | high | thư mục certs ≤ 0750 (chứa private key) | ✅ chmod 0750 |
| `ui-tls` | medium | config có dấu hiệu TLS/HTTPS | — (dùng reverse proxy) |
| `db-pass-perms` | medium | .db-pass ≤ 0600 (nếu có) | ✅ chmod 0600 |

- Windows: check quyền file bỏ qua (ACL khác Unix mode) → `passed` + ghi chú.
- Không kết luận được (file thiếu) → `passed=true` để tránh báo động giả.
- Đường dẫn từ env: `OPENITMS_PREFIX`, `QUICKWIN_CONFIG`, `QUICKWIN_CERTS_DIR`.

## Test (PASS 2026-07-06)
Phát hiện mật khẩu default; không false-positive trên host sạch (config 0600 + TLS);
phát hiện config 0644 quá mở → fix chmod 0600 → scan lại pass; fix check lạ → lỗi.
