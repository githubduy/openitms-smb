---
level: L1
status: draft
owners: [maintainer]
updated: 2026-07-07
related-code: [gitea-manager/, core-patches/, installer/]
---

# Thiết kế: Tích hợp Gitea (git server local)

ADR-0005. Mục tiêu: mỗi project OpenITMS có sẵn 1 repo local; 1 project mặc định quản lý host.

## Kiến trúc
```
┌──────────────────────────────────────────────────────────┐
│  OpenITMS-SMB core (fork Semaphore)                       │
│   + project-create hook (patch 0007) ──┐                  │
│                                         ▼                  │
│  gitea-manager (package Go, ngoài core)                    │
│   - start/health/provision Gitea                          │
│   - GiteaClient: CreateRepo, EnsureOrg (API token)        │
└──────────────────────────────────────────────────────────┘
        │ localhost:3080 (HTTP API + git)
        ▼
┌──────────────────────┐   ┌──────────────────────┐
│  Gitea (binary MIT)  │──►│  MariaDB (db: gitea)  │
│  data: <prefix>/gitea│   └──────────────────────┘
└──────────────────────┘
```

## Thành phần
### 1. Bundle (installer)
- `installer/vendor/gitea/gitea(.exe)` — binary; deps.lock pin version + sha256.
- Gitea `app.ini` sinh lúc cài: HTTP port 3080 (nội bộ), DB mysql `gitea`, data dir,
  DISABLE_REGISTRATION=true, INSTALL_LOCK=true (bỏ wizard).
- systemd `openitms-gitea.service` (After=openitms-db).

### 2. gitea-manager (package Go `quickwin.dev/gitea-manager`)
- `Start()`: chạy gitea binary, chờ health `/api/healthz`.
- `Provision()`: tạo admin (`gitea admin user create`) + org `openitms` + API token (lần đầu).
- `GiteaClient`: `EnsureOrg`, `CreateRepo(org, name)`, `RepoCloneURL(...)` qua Gitea API + token.
- Token/admin pass: sinh random lúc cài, lưu `<prefix>/.gitea-secrets` (0600), KHÔNG commit.

### 3. Hook tạo project (core-patch 0007)
- Semaphore tạo project → hook gọi `giteaManager.CreateRepo("openitms", slug(project.Name))`
  → tạo repository record trong Semaphore trỏ tới clone URL local.
- Idempotent: project đã có repo local → bỏ qua.
- Hook mỏng: chèn 1 lời gọi sau `AddProject`, logic ở gitea-manager.

### 4. Project mặc định "host" (seed)
- Lúc cài (install.sh): tạo project id 1 "Host Management" + repo local `openitms/host`
  (seed sẵn playbook quản trị host: bật/tắt WinRS, cấu hình SSH, hardening...).
- Người dùng có ngay 1 project quản lý máy chính, không cần tạo tay.

## Quyết định phụ
- **DB Gitea**: MariaDB `gitea` (nhất quán) — Live USB minimal fallback SQLite.
- **Port**: 3080 nội bộ (không expose ra ngoài mặc định; reverse proxy nếu cần công khai).
- **Auth git**: token/deploy-key local giữa Semaphore ↔ Gitea (không cần user nhập).

## Backlog (Phase — chèn vào roadmap)
- G-01: bundle Gitea (vendor + deps.lock + app.ini + systemd) + fetch-deps hỗ trợ gitea.
- G-02: gitea-manager (start/health/provision + GiteaClient) + test.
- G-03: patch 0007 hook project-create → auto CreateRepo.
- G-04: seed project "host" + repo mặc định (playbook quản trị host).
- G-05: UI — tab/khu vực repository local trong project (hoặc link sang Gitea UI).
- G-06: E2E: tạo project → repo tự sinh trong Gitea → Semaphore chạy playbook từ repo local.
