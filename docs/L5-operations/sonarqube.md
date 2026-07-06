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

## Quality gate
- Mục tiêu (plan 9.4): **0 new bug/vulnerability**, coverage code mới ≥ 70%.
- Baseline scan đầu (2026-07-06): 0 bug, 0 vuln, 0 hotspot, coverage 57.1%, dup 0%,
  6 code smell (5 cognitive-complexity + 1 TODO có chủ đích). Gate: **OK**.

## Việc còn lại để đóng P4-03
- [ ] Gắn scan vào CI (self-hosted runner reach được instance, hoặc instance expose có auth).
- [ ] Bật quality-gate "fail merge khi đỏ" trên PR.
- [ ] Backup + upgrade định kỳ instance (thuộc `sonarqube-instance.md` khi có).
