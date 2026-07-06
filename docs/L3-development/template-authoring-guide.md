---
level: L3
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [templates/]
---

# Viết template playbook

Template = kịch bản chạy 1-click từ UI (endpoint-provisioning hoặc app-deployment).
Phân phối qua registry (type `template`). 5 template preload: `templates/`.

## Cấu trúc
```
templates/<name>/
├── template.yaml    # metadata + inputs (UI form) + license (bắt buộc)
└── <entrypoint>     # setup.ps1 (Windows/pwsh) | deploy.sh (Linux/bash) | playbook.yml (ansible)
```

## template.yaml
- `name` (kebab), `version` (semver), `type: template`, **`license`** (bắt buộc).
- `runner`: `powershell` | `bash` | `ansible` — quyết định kênh thực thi.
- `category`: `endpoint-provisioning` | `app-deployment`.
- `inputs[]`: mỗi input `{name, label, type, required?, default?}` → sinh form trên UI;
  giá trị inject vào entrypoint qua env (UPPER_SNAKE) hoặc param (pwsh).
- `entrypoint`: file chạy.

## Quy tắc
- Bash: `set -euo pipefail`, idempotent (chạy lại không hỏng), kiểm tra sẵn có trước khi cài.
- pwsh xuống Windows: đi qua winrs-cert (cert auth) — thứ tự ưu tiên pwsh → SSH → WinRS.
- **Docker trong template app-deployment là cài LÊN MÁY ĐÍCH** theo yêu cầu người dùng —
  OpenITMS-SMB không dùng Docker (luật #5).
- Lint bắt buộc (CI job `templates-lint`): `shellcheck` cho .sh, `ansible-lint` cho playbook.
- Secret (password…) qua input required, KHÔNG hardcode trong file.

## 5 template preload
| Template | Runner | Việc |
|---|---|---|
| `jea-winrs-setup` | powershell | Enroll cert + WinRM HTTPS + cert-auth + map JEA trên Win11 |
| `docker-cluster-setup` | bash | Cài Docker Engine + compose lên host Linux |
| `odoo-deploy` | bash | Odoo + Postgres qua compose |
| `mariadb-standalone` | bash | MariaDB standalone qua compose |
| `clickhouse-standalone` | bash | ClickHouse standalone qua compose |
