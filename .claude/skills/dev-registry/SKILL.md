---
name: dev-registry
description: >
  Thay đổi registry/index/schema hoặc thêm artifact vào registry OpenITMS-SMB. Trigger khi task
  có label area:registry. Enforce: schema versioned, ký số bắt buộc, KHÔNG đụng private key.
---

# Skill: dev-registry

Bản thực thi playbook 5C-registry. Module `registry/`, spec `docs/L2-specs/registry-format.md`.

## Bước
1. Đổi schema Index/Artifact → bump `SchemaVersion`, giữ backward-compat (client cũ đọc index mới).
2. Artifact mới: `license` bắt buộc; đóng gói `PackArtifact`; index ký ed25519.
3. **KHÔNG BAO GIỜ đụng private key** (env REGISTRY_PRIVATE_KEY / CI secret — maintainer giữ, escalation #6).
   Ký thật do maintainer/CI làm; AI chỉ chuẩn bị artifact + code.
4. Test: round-trip build→sign→verify→install; case từ chối (tamper/wrong-key/no-license).
5. Cập nhật spec registry-format.md. Commit Conventional + DCO.

## Cấm
- Đổi/xóa field schema đã phát hành (chỉ thêm). Ký bằng key tự sinh trong code sản phẩm.
