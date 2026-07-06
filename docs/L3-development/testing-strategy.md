---
level: L3
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [tests/, scripts/run-tests.sh]
---

# Chiến lược test

4 tầng, từ rẻ đến đắt — PR phải xanh tầng 1–2; tầng 3 nightly; tầng 4 trước release.

| Tầng | Gì | Chạy ở đâu / khi nào |
|---|---|---|
| **1. Unit** | `go test` các module của ta (plugin-manager, plugins/*, registry) — `scripts/run-tests.sh` | Mọi PR |
| **2. Smoke** | Build binary → chạy server thật (BoltDB tạm) → assert `/api/ping` + login page — `tests/e2e/smoke.sh` | Mọi PR |
| **3. Integration** | Plugin handshake THẬT qua go-plugin (spawn binary hello → GetMetadata/HandleRequest); patch-chain trên upstream sạch | Mọi PR (khi có plugin) |
| **4. E2E installer** | VM Linux sạch KHÔNG internet: tar.gz → install.sh → login UI → đổi mật khẩu → chạy task mẫu. + Smoke Windows endpoint: 1 máy Win11 lab cho winrs-cert (cert auth thật) | Nightly + trước release (`tests/e2e/run.sh`) |

## Quy tắc
- Test của TA test code của TA — không re-test upstream (upstream test chỉ chạy ở sync-upstream, `SYNC=1`).
- Coverage phần code mới ≥ 70% (SonarQube gate, Phase 4).
- Mỗi bug được fix phải kèm test tái hiện bug đó (regression).
- E2E installer chạy trên cả Debian-family + RHEL-family; image VM chuẩn pin version trong `tests/e2e/`.
- Máy Win11 lab: cần 1 máy/VM Windows 11 có WinRM HTTPS + cert enroll được (bổ sung khi tới P1-09).

## Tình trạng hiện tại
- Tầng 1: `scripts/run-tests.sh` sẵn sàng (chưa có module → no-op).
- Tầng 2: `tests/e2e/smoke.sh` hoạt động (binary backend-only, BoltDB).
- Tầng 3–4: dựng ở Phase 1 (P1-03/04) và Phase 2 (P2-02).
