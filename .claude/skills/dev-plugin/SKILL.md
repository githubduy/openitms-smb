---
name: dev-plugin
description: >
  Viết/sửa plugin OpenITMS-SMB (Go hoặc Python) qua SDK + proto hiện hành. Trigger khi task có
  label area:plugin. Enforce: dùng SDK không viết stub tay; manifest đủ field + license;
  plugin chịu được crash/kill/input rác; kèm test.
---

# Skill: dev-plugin

Bản thực thi playbook 5B. Mẫu: `plugins/hello` (Go), `plugins/hello-py` (Python), `plugins/winrs-cert`.

## Bước
1. Đọc `docs/L2-specs/plugin-manifest.md` + `proto-contract.md`. Chọn ngôn ngữ:
   - Go: `quickwin.dev/sdk` — implement interface Plugin, `sdk.Serve(...)`. go.mod replace nội bộ.
   - Python: `quickwin_plugin` + stub từ `scripts/gen-proto-py.sh`. entrypoint `python: main.py`.
2. Cấu trúc `plugins/<name>/{plugin.yaml, code}`. plugin.yaml: name(kebab), version(semver),
   **license**, protocol_version:1, entrypoint, routes (KHỚP Metadata — lệch → từ chối load), permissions.
3. Plugin phải: healthcheck, chịu core restart / bị kill / input rác không panic.
4. Test: unit + integration chạy plugin THẬT qua go-plugin (mẫu integration_test.go / python_integration_test.go).
5. `scripts/run-tests.sh` xanh. Commit Conventional + DCO, branch `ai/*`.

## Cấm
- Viết proto stub tay (luôn sinh từ .proto).
- Thêm tính năng vào core khi làm được bằng plugin.
