---
level: L1
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: []
---

# ADR-0001 — HashiCorp go-plugin + gRPC/Protobuf làm cơ chế plugin duy nhất

**Trạng thái:** approved (2026-07-05)

## Bối cảnh
Cần plugin đa ngôn ngữ (Go, Python), isolate crash khỏi core, hợp đồng versioned.

## Quyết định
Plugin chạy process riêng qua go-plugin (handshake + mTLS tự động), giao tiếp gRPC theo proto/quickwin/plugin/v1. Một file .proto là nguồn chân lý cho mọi SDK.

## Phương án đã loại
Loại - embed Go plugin package (không isolate crash, không đa ngôn ngữ); REST sidecar tự chế (thiếu handshake/lifecycle chuẩn).

## Hệ quả
Xem PLAN.md mục 3, 5 và GUIDELINE (luật cứng) — mọi thay đổi ngược quyết định này cần ADR mới supersede.
