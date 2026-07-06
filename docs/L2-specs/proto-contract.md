---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [proto/]
---

# Spec: Hợp đồng Protobuf core ↔ plugin (v1)

File nguồn chân lý: `proto/quickwin/plugin/v1/plugin.proto`. Mọi SDK (Go, Python)
generate stub từ CHÍNH file này — cấm viết stub tay.

## 4 RPC
| RPC | Ai gọi | Khi nào | Ghi chú |
|---|---|---|---|
| `GetMetadata` | Core | Ngay sau launch | Core đối chiếu với `plugin.yaml` — lệch name/version → từ chối load |
| `HandleRequest` | Core | Mỗi HTTP request tới API động | Core đã authenticate; `Caller` cho plugin biết ai gọi. Plugin KHÔNG tự nhận request từ ngoài |
| `RunTask` | Core | Tác vụ dài (UI task runner) | Server-stream: `log_line`* → (`status`)* → `result` (bắt buộc, cuối cùng) |
| `HealthCheck` | Core | Định kỳ 30s | 3 lần fail liên tiếp → restart plugin (backoff 1s/5s/25s) |

## Luật tiến hóa (enforce bằng `buf breaking` trong CI)
1. CHỈ thêm field với tag number MỚI; chỉ thêm rpc mới; chỉ thêm enum value mới (không đổi số 0).
2. KHÔNG: đổi/xóa field, đổi type, đổi tag, đổi tên rpc/message đã phát hành.
3. Breaking thực sự cần → package `quickwin.plugin.v2` MỚI (giữ v1 chạy song song ≥ 1 major release) + bump APP-PROTOCOL-VERSION + ADR được duyệt.

## Handshake (HashiCorp go-plugin)
- CORE-PROTOCOL-VERSION: 1 (của go-plugin) · **APP-PROTOCOL-VERSION: 1** (của ta)
- Magic cookie key/value: định nghĩa trong SDK (`quickwin-plugin-sdk-go`), plugin Python in
  dòng handshake chuẩn ra stdout: `1|1|tcp|127.0.0.1:<port>|grpc`
- Transport: gRPC, go-plugin tự cấp mTLS.

## Quy ước semantics
- `HttpResponse.status` = HTTP status thật trả cho client; body là bytes tùy ý (thường JSON).
- `TaskEvent` cuối cùng PHẢI là `result`; core coi stream đóng không có result = TASK_STATUS_FAILED.
- Timeout: core cắt stream sau `timeout_seconds` (0 → mặc định plugin tự quản, tối đa 24h).
- Health `STATUS_DEGRADED`: core log warning, KHÔNG restart.
