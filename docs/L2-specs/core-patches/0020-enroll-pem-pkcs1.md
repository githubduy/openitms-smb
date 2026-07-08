---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-08
related-code: [core-patches/0020-enroll-pem-pkcs1.patch, core-patches/0019-enroll-ps51-compat.patch, core-patches/0014-endpoint-enroll-script.patch]
---

# Spec patch 0020 — Enroll PEM export tương thích .NET Framework (PKCS#1)

## Mục đích (WHY)
Sau 0019, `winrs-enroll.ps1` relaunch sang Windows PowerShell 5.1 (để `New-LocalUser`/WinRM cmdlet
chạy được). Nhưng bước xuất private key dùng `GetRSAPrivateKey().ExportPkcs8PrivateKey()` — API chỉ
có ở **.NET Core / PowerShell 7**, KHÔNG có ở **.NET Framework 4.x (WinPS 5.1)** → lỗi runtime:
`does not contain a method named 'GetRSAPrivateKey'`.

**Nghịch lý**: `LocalAccounts` cần PS 5.1; export key hiện đại cần PS 7. Giải: giữ ở 5.1 + tự
encode private key theo cách .NET Framework hỗ trợ.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/scripts/winrs-enroll.ps1` | sửa (bước 5) | Bỏ `ExportPkcs8PrivateKey`. Thêm helper `Enc-Len`/`Enc-Int`/`Enc-Seq` (ASN.1 DER) + `ConvertTo-Pkcs1B64`: lấy `RSAParameters` qua static `[RSACertificateExtensions]::GetRSAPrivateKey($cert).ExportParameters($true)`, tự encode **PKCS#1 `RSAPrivateKey`** DER (version,n,e,d,p,q,dp,dq,coeff) → base64. PEM đổi header `PRIVATE KEY` → `RSA PRIVATE KEY` |

## Vì sao dùng được ở cả 5.1 lẫn 7
- `RSACertificateExtensions::GetRSAPrivateKey` (static) + `ExportParameters($true)` có từ .NET 4.6 (5.1) VÀ .NET Core (7).
- Tự encode DER = số học thuần, không phụ thuộc API export mới.
- Go `crypto/tls.X509KeyPair` chấp nhận PKCS#1 `RSA PRIVATE KEY` (winrsexec dùng).

## Verify (thật, PS 5.1 trên máy Windows)
- Chạy đúng block bước 5 (New-SelfSignedCertificate Custom + PFX round-trip Exportable + ConvertTo-Pkcs1B64)
  → xuất PEM thành công (không lỗi GetRSAPrivateKey).
- `go run` `tls.X509KeyPair(pem, pem)` → **OK** (PEM hợp lệ, cert+key khớp).
- Parser PS 5.1: winrs-enroll.ps1 PARSE OK; `go build ./api/...` OK (embed).
- Chain 0001–0020 apply + build.

## Liên quan
- Relaunch 5.1 + ASCII: [[0019-enroll-ps51-compat]]. Script gốc: [[0014-endpoint-enroll-script]].
