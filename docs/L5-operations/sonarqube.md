---
level: L5
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [sonar-project.properties, scripts/sonar-scan.sh]
---

# SonarQube (self-host) — quét chất lượng code

Chốt: **self-host** (ADR-0004 #4), SonarQube Community Build. Project key: **`openitms-smb`**.

## Nguyên tắc bảo mật
- **KHÔNG commit** host URL / IP nội bộ / token vào repo (public). `sonar-project.properties`
  chỉ chứa project key + sources + exclusions.
- Host + token truyền qua env khi scan:
  `SONAR_HOST_URL`, `SONAR_TOKEN` (project analysis token). CI: lưu trong GitHub secrets.

## Phạm vi quét
- CHỈ code của dự án (`plugin-manager`, `sdk`, `plugins`, `registry`, `proto/quickwin`,
  `scripts`, `installer`). **KHÔNG quét `upstream/`** (Semaphore gốc — luật #1), generated
  (`proto/gen`, `*.pb.go`), toolchain (`Go/`), `dist/`, vendor, patch.
- Coverage: Go — `scripts/sonar-scan.sh` sinh `coverage.out` mỗi module rồi nạp qua
  `sonar.go.coverage.reportPaths`.

## Chạy
```bash
export SONAR_HOST_URL=http://<host>:9000
export SONAR_TOKEN=<project-analysis-token>
export SONAR_SCANNER=/path/to/sonar-scanner    # nếu không có trong PATH
scripts/sonar-scan.sh
```
Windows: chạy `sonar-scanner.bat` với `-Dsonar.host.url` `-Dsonar.token`
`-Dsonar.scanner.skipJreProvisioning=true` (dùng Java local — tránh lỗi extract JRE cache).

## Quality gate — "OpenITMS-SMB Gate" (khớp plan 9.4)
Gate riêng gán cho project (không dùng "Sonar way" 80% mặc định), điều kiện trên **new code**:
- `new_violations = 0` (0 new bug/vuln/smell)
- `new_coverage ≥ 70%` (đúng plan 9.4; "Sonar way" mặc định 80% quá khắt cho giai đoạn đầu)
- `new_duplicated_lines_density < 3%`

Baseline (2026-07-06): **GATE OK** — 0 bug, 0 vuln, 0 hotspot, 0 new_violations,
coverage tổng 57.1% / new code 75.9%, dup 0%, 5 code smell còn lại (4 cognitive-complexity
tự nhiên: certstore.scan, manager.Scan, matchRoute, integration_test; 1 TODO có chủ đích).
Handler đã refactor 32→<15.

## Việc còn lại để đóng P4-03
- [ ] Gắn scan vào CI (self-hosted runner reach được instance, hoặc instance expose có auth).
- [ ] Bật quality-gate "fail merge khi đỏ" trên PR (webhook/GitHub check).
- [ ] Backup + upgrade định kỳ instance.
- [ ] (tùy chọn) refactor 4 complexity smell còn lại nếu chạm code đó.
