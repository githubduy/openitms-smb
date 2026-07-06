---
name: dev-template
description: >
  Viết/sửa playbook template (endpoint-provisioning hoặc app-deployment). Trigger khi task có
  label area:template. Enforce: template.yaml đủ inputs+license; shellcheck/ansible-lint sạch;
  secret qua input không hardcode.
---

# Skill: dev-template

Bản thực thi playbook 5E. Guide `docs/L3-development/template-authoring-guide.md`. Mẫu: `templates/`.

## Bước
1. `templates/<name>/{template.yaml, entrypoint}`. template.yaml: name, version, type:template,
   **license**, runner (bash|powershell|ansible), category, inputs[] (→ UI form), entrypoint.
2. Bash: `set -euo pipefail`, idempotent. pwsh xuống Windows: qua winrs-cert (cert auth).
   Docker trong app-deployment = cài LÊN máy đích (OpenITMS không dùng Docker — luật #5).
3. Secret qua input required, KHÔNG hardcode.
4. Lint: `shellcheck -S error` (.sh) / `ansible-lint` (playbook) — CI job templates-lint.
5. Commit Conventional + DCO.
