# BACKLOG CHI TIẾT — QuickWin (fork Semaphore UI)

> ## 📌 TRẠNG THÁI (cập nhật 2026-07-05)
> | Task | Trạng thái |
> |---|---|
> | P0-01 | 🟡 4/5 chốt (ADR-0004) — còn tên sản phẩm |
> | P0-02, 03, 04, 05 | ✅ xong — repo + submodule v2.18.16 + Go 1.24.6 trong `./Go/` + patch system test PASS; binary `quickwin-dev-sem2.18.16` build OK trên Windows (Linux verify khi CI chạy) |
> | P0-06 | 🟡 script sync-upstream có, chưa diễn tập với tag thật |
> | P0-07, 09, 10, 12 | ✅ xong — license/NOTICE + docs L0/L1 + ADR 0001–0004 + guideline + OWNERS.yaml + governance/issue forms |
> | P0-08, 13 | 🟡 file CI + policy đã viết — cần push GitHub để chạy thật + bật push protection/mirror |
> | P0-11 | ✅ testing-strategy.md + `tests/e2e/smoke.sh` chạy PASS thật (server bolt + /api/ping + UI 200); tầng 3-4 dựng ở P1/P2 |
> | P1-01 | 🟡 schema `plugin.schema.json` + spec + example xong — validator Go (bắt 6 case lỗi) làm cùng P1-04 |
> | P1-02 | 🟡 `plugin.proto` v1 + buf.yaml/buf.gen.yaml + CI job buf xong — `buf lint/breaking` verify khi CI chạy trên GitHub |
> | Còn lại | ⬜ chưa bắt đầu |
>
> **Chờ user:** tạo repo GitHub (org/tên) + chốt tên sản phẩm chính thức.

> Sinh từ `PLAN.md` (đã review 2026-07-05).
> Guideline làm việc: `docs/L3-development/AI-ENGINEER-GUIDELINE.md`.
>
> **Quy ước:**
> - ID: `P<phase>-<số>` — ví dụ `P0-03`.
> - **Owner**: `👤` = người (maintainer) quyết/làm · `🤖` = AI agent làm được trọn vẹn · `🤖+👤` = AI làm, người duyệt.
> - **Deps**: task phải xong trước.
> - **AC** = Acceptance Criteria (tiêu chí nghiệm thu — đo được, không cảm tính).
> - Ước lượng: S (< ½ ngày) · M (½–2 ngày) · L (3–5 ngày) · XL (> 1 tuần, nên tách nhỏ tiếp).

---

## PHASE 0 — Nền móng (tuần 1–2)

### P0-01 · Chốt quyết định mở 👤 · S · Deps: —  ⚡ LÀM ĐẦU TIÊN
~~5 quyết định~~ → **4/5 ĐÃ CHỐT 2026-07-05** (xem plan mục 13): baseline = tag stable mới nhất;
registry public = GitHub Pages; SonarQube = **self-host**; Live USB = **Alpine** (control từng package,
lưu ý musl). **CÒN LẠI 1:** tên sản phẩm chính thức + domain (kiểm tra trademark/domain trước khi chốt;
"QuickWin" là tên tạm — chặn P1-07 branding).
**AC:** toàn bộ quyết định (4 đã chốt + tên) ghi thành `docs/L1-architecture/adr/ADR-0004-initial-decisions.md`, status approved.

### P0-02 · Khởi tạo repo + cấu trúc thư mục 🤖+👤 · S · Deps: P0-01
Tạo repo GitHub, cây thư mục theo plan mục 4.1 (upstream/, core-patches/, plugin-manager/, plugins/, proto/, registry/, installer/, templates/, docs/, website/, scripts/, .github/). README tối thiểu + branch protection cho `main`.
**AC:** repo public/private theo quyết định P0-01; cây thư mục đúng plan; push trực tiếp vào `main` bị chặn.

### P0-03 · Pin upstream làm submodule 🤖 · S · Deps: P0-02
`git submodule add https://github.com/semaphoreui/semaphore upstream/` + checkout đúng tag baseline. Thêm CI check "submodule không có diff".
**AC:** `git submodule status` trả đúng tag; CI fail nếu working tree trong `upstream/` bẩn.

### P0-04 · Go toolchain vào /project/Go/ + build vanilla 🤖 · M · Deps: P0-03
Script `scripts/setup-toolchain.sh` tải Go (version pin trong `go-version.txt`) vào `/project/Go/`; `scripts/build-all.sh` build Semaphore NGUYÊN BẢN (0 patch) bằng toolchain đó, output vào `/project/dist/`.
**AC:** máy sạch chạy 2 script → ra binary `dist/bin/semaphore` chạy được `--version`; không dùng Go hệ thống (CI kiểm bằng PATH rỗng).

### P0-05 · Khung patch system (apply/export) 🤖 · M · Deps: P0-04
`core-patches/series` (rỗng), `scripts/apply-patches.sh` (apply tuần tự, dừng + in hunk khi fail), `scripts/export-patch.sh` (từ commit → file patch đúng format có header VÌ SAO). Định nghĩa format header patch.
**AC:** apply chuỗi rỗng PASS; thêm 1 patch giả (đổi 1 comment) → apply PASS, build PASS; patch hỏng → exit code ≠ 0 + báo đúng hunk.

### P0-06 · Khung sync-upstream 🤖 · M · Deps: P0-05
`scripts/sync-upstream.sh <tag>`: fetch tag → branch `build/<tag>` → apply series → build + test. Chưa cần AI-rebase (Phase 4), chỉ cần fail rõ ràng.
**AC:** chạy với chính tag baseline → PASS end-to-end; chạy với tag không tồn tại → lỗi rõ ràng.

### P0-07 · Hồ sơ pháp lý + license gate 🤖+👤 · M · Deps: P0-02
`LICENSE` (MIT của ta), `LICENSE-SEMAPHORE` (nguyên văn), `NOTICE.md`; CI job `go-licenses check ./...` + sinh `THIRD_PARTY_LICENSES.md`; fail build nếu có GPL/AGPL link trực tiếp.
**AC:** 3 file đúng nội dung plan mục 2.1; CI đỏ khi thêm thử 1 dep GPL; báo cáo third-party sinh tự động trong artifact build.

### P0-08 · CI skeleton 🤖 · M · Deps: P0-04, P0-05
GitHub Actions: job build (toolchain pin + vanilla + full-patch-chain), job test, job license (P0-07), job patch-hygiene (đủ bộ-4 khi có patch mới — xem luật #2 guideline). SonarCloud gắn ở P4.
**AC:** PR mẫu chạy đủ 4 job; PR thêm patch thiếu spec L2 → job patch-hygiene ĐỎ.

### P0-09 · Docs L0 + L1 khởi điểm 🤖+👤 · M · Deps: P0-01
`docs/L0-overview/{vision,roadmap,glossary,legal}.md`; `docs/L1-architecture/{system-overview,core-vs-plugin-boundary,upstream-sync-strategy}.md`; ADR-0001 (go-plugin), ADR-0002 (MariaDB bundled socket-only), ADR-0003 (static registry + ký số), ADR-template. Nội dung chưng cất từ plan — KHÔNG copy nguyên văn, viết theo audience từng level.
**AC:** đủ file, frontmatter chuẩn mục 12.3; maintainer duyệt ADR 0001–0003 → status approved.

### P0-10 · Chuyển guideline AI vào repo + OWNERS.yaml 🤖 · S · Deps: P0-09
Chuyển `GUIDELINE-AI-ENGINEER.md` → `docs/L3-development/AI-ENGINEER-GUIDELINE.md` (cập nhật path nội bộ); tạo `docs/L2-specs/OWNERS.yaml` (schema mapping spec ↔ code path) + CI job docs-mapping.
**AC:** PR sửa `plugin-manager/` mà không sửa spec tương ứng → CI đỏ (test bằng PR giả).

### P0-11 · Chiến lược test tổng thể (bổ sung từ review) 🤖+👤 · M · Deps: P0-08
`docs/L3-development/testing-strategy.md`: tầng unit (Go test), integration (plugin handshake thật), E2E installer (VM Linux sạch, không internet — dùng ảnh VM chuẩn), smoke Windows endpoint (1 máy Win11 lab cho winrs-cert). Kèm khung `tests/e2e/` chạy được bằng script.
**AC:** doc approved; `tests/e2e/run.sh` bootstrap VM + assert "UI login được" chạy PASS trên vanilla build.

### P0-12 · Governance cộng đồng (bổ sung từ review) 👤 · S · Deps: P0-02
DCO (sign-off) hoặc CLA — khuyến nghị DCO cho nhẹ; `CODE_OF_CONDUCT.md`; `GOVERNANCE.md` (ai là maintainer, quy trình thêm maintainer); Issue Forms (`feature-request.yaml`, `bug-report.yaml`, `plugin-proposal.yaml`) + labels `area:*`.
**AC:** DCO check bật trên PR; 3 issue form dùng được với field máy-đọc.

### P0-13 · Chính sách publish + hardening repo + mirror Codeberg 🤖+👤 · M · Deps: P0-02
Thực thi plan mục 9.5: bật secret scanning + push protection, `gitleaks` vào CI, `.gitignore` chặn `*.pfx *.pem *.key servers.json .env`; branch protection `main` (PR + 1 human review + CI xanh, cấm force-push); quy ước branch `ai/*` cho agent; 2FA bắt buộc maintainer; tạo mirror Codeberg (push-mirror tự động main + tags, README mirror ghi "đóng góp tại GitHub"); tag release ký số chỉ maintainer/CI-sau-approve tạo.
**AC:** push thử 1 fake secret → bị push protection chặn; push trực tiếp `main` → bị từ chối; push main trên GitHub → mirror Codeberg tự sync ≤ 15 phút; doc `docs/L3-development/publishing-policy.md` approved.

**🏁 Phase 0 DONE khi:** build vanilla + chuỗi patch rỗng xanh trên CI, license gate hoạt động, docs L0/L1 approved, e2e skeleton chạy, repo hardening + mirror hoạt động (P0-13).

---

## PHASE 1 — Core patches + Plugin Manager (tuần 3–6)

### P1-01 · Spec + JSON Schema plugin.yaml 🤖+👤 · M · Deps: P0-09
`docs/L2-specs/plugin-manifest.md` + `plugin-manager/schema/plugin.schema.json`: name, version (semver), license (bắt buộc), routes[], permissions, checksum, ui hooks.
**AC:** schema validate được manifest mẫu hợp lệ + bắt được 5 case lỗi (thiếu license, version sai semver, route trùng…).

### P1-02 · Proto v1 + codegen 🤖+👤 · M · Deps: P0-09
`proto/quickwin/plugin/v1/plugin.proto` (GetMetadata, HandleRequest, RunTask stream, HealthCheck); `buf` config + `buf breaking` trong CI; script regenerate stub.
**AC:** `buf lint` + `buf breaking` chạy trong CI; sửa thử 1 field cũ → CI đỏ.

### P1-03 · SDK Go cho plugin 🤖 · M · Deps: P1-02
Package `quickwin-plugin-sdk-go`: wrap go-plugin handshake, serve gRPC từ interface đơn giản; kèm plugin "hello" ví dụ + test handshake.
**AC:** dev viết plugin hello < 50 dòng dùng SDK; integration test spawn hello qua go-plugin PASS.

### P1-04 · Plugin Manager: scan + launch + lifecycle 🤖 · L · Deps: P1-01, P1-03
Package `plugin-manager/`: quét `/plugins/*/`, validate manifest + checksum, launch qua go-plugin (mTLS tự động), health-check định kỳ, restart khi crash (backoff), log tách theo plugin.
**AC:** integration test: plugin hello được phát hiện + chạy; kill process plugin → tự restart ≤ 5s; manifest checksum sai → từ chối load + log rõ.

### P1-05 · Plugin Manager: API động 🤖 · L · Deps: P1-04
Đọc `routes[]` từ manifest → đăng ký `/api/plugins/<name>/<route>` trên router core → proxy sang plugin qua `HandleRequest`. Enforce permissions từ manifest; auth dùng session/token sẵn có của Semaphore.
**AC:** curl endpoint động của hello plugin trả đúng; gọi route không khai trong manifest → 404; user không đủ quyền → 403.

### P1-06 · Patch 0001 — hook Plugin Manager vào core 🤖+👤 · M · Deps: P1-04
Patch chèn TỐI THIỂU số dòng vào init của Semaphore để gọi `plugin-manager`. Đủ bộ-4 (patch + series + CHANGELOG + spec `docs/L2-specs/core-patches/0001-*.md`).
**AC:** `apply-patches.sh` từ upstream sạch PASS; diff patch < 30 dòng vào file upstream; core chạy KHÔNG có thư mục /plugins vẫn bình thường (degrade gracefully).

### P1-07 · Patch 0002 — Branding 🤖+👤 · M · Deps: P0-01, P1-06
Đổi tên/logo/favicon theo tên chốt ở P0-01, ưu tiên asset-replace + build flag, hạn chế sửa file Vue. Giữ attribution "Based on Semaphore UI (MIT)" ở trang About.
**AC:** UI không còn chữ/logo "Semaphore" ngoài trang About; patch apply sạch; screenshot đính kèm PR.

### P1-08 · Patch 0003 — Config hardcode lần đầu 🤖+👤 · M · Deps: P1-06
`config/config.json` sinh sẵn (port 3000, DB socket, đường dẫn certs, admin mặc định `admin/quickwin123`); banner đỏ bắt đổi mật khẩu ở lần login đầu.
**AC:** khởi động từ thư mục giải nén không cần wizard cấu hình nào; login lần đầu thấy banner; đổi mật khẩu xong banner biến mất.

### P1-09 · Plugin winrs-cert v1 🤖+👤 · L · Deps: P1-05, P0-11 (Win11 lab)
Plugin Go: load cert `.pfx`/`.pem` từ thư mục certs, gọi WinRS qua certificate auth xuống Windows 11; API động `POST /api/plugins/winrs-cert/exec` (host, command, timeout); log/audit mỗi lệnh.
**AC:** E2E trên Win11 lab: chạy `hostname` qua API trả đúng; cert sai → lỗi phân loại rõ (cert vs network vs WinRM config); unit + integration test PASS.

### P1-10 · Docs L2 Phase 1 + guide plugin Go 🤖 · M · Deps: P1-05, P1-06..08
`docs/L2-specs/{plugin-manager,proto-contract}.md`, spec 3 patch; `docs/L3-development/plugin-dev-guide-go.md` (zero → plugin chạy, dựa trên hello).
**AC:** một AI agent mới CHỈ đọc guide + spec viết được plugin echo pass integration test (test thực tế bằng 1 phiên agent).

**🏁 Phase 1 DONE khi:** upstream sạch + 3 patch build xanh; hello + winrs-cert chạy qua API động; lệnh thật xuống Win11 bằng cert thành công.

---

## PHASE 2 — Đóng gói Linux (tuần 7–9)

### P2-01 · Bundle MariaDB 🤖+👤 · L · Deps: P1-08
Chọn MariaDB LTS binary tarball; `my.cnf` tune SMB (buffer pool %RAM, utf8mb4, slow-log, **chỉ listen socket/localhost**); script init datadir + user `quickwin-db` + password random ghi vào config; kèm GPLv2 text + link source trong `licenses/`.
**AC:** trên VM sạch: DB init + start + core kết nối qua socket; `ss -tlnp` không thấy MariaDB mở TCP ra ngoài; licenses/ đủ hồ sơ GPL.

### P2-02 · install.sh + systemd 🤖 · L · Deps: P2-01
1 lệnh: giải nén → init DB → sinh config → `quickwin-db.service` + `quickwin.service` (After=quickwin-db) → in URL + tài khoản mặc định. Idempotent (chạy lại không phá dữ liệu). Kèm `uninstall.sh` (hỏi trước khi xóa data).
**AC:** VM Debian + VM RHEL-family sạch, KHÔNG internet: cài < 5 phút, login UI được; chạy `install.sh` lần 2 không mất dữ liệu; reboot VM → 2 service tự lên đúng thứ tự.

### P2-03 · Bundle pwsh + layout dist chuẩn 🤖 · M · Deps: P2-02
PowerShell Core binary vào `bin/pwsh/`; chốt layout `dist/` đúng plan 7.1; build-all.sh đóng gói tar.gz + checksum + SBOM.
**AC:** tar.gz giải nén đúng layout; `bin/pwsh/pwsh -v` chạy trên VM sạch; SBOM + checksum trong artifact CI.

### P2-04 · Watcher thư mục ./certs 🤖 · M · Deps: P1-09
Watch `./certs`, nạp nóng `.pfx`/`.pem` vào cert store nội bộ (không cần restart); winrs-cert + inventory WinRM dùng chung store; validate file rác không làm crash.
**AC:** copy cert mới vào thư mục → gọi được WinRS bằng cert đó trong ≤ 10s, không restart; file không phải cert → log warning, bỏ qua.

### P2-05 · Plugin hardening v1 🤖+👤 · L · Deps: P1-05
Quét host: mật khẩu admin còn default, port thừa, quyền file config/certs, TLS UI, firewall. Báo cáo mức độ + nút "fix" cho mục fix được (đổi quyền file…).
**AC:** trên bản cài mới tinh phát hiện ≥ 5 finding chuẩn (trong đó có "mật khẩu default"); fix quyền file qua UI hoạt động; không finding giả trên host đã harden theo doc.

### P2-06 · Settings UI 🤖+👤 · M · Deps: P1-08
Trang Settings sửa config sau cài (port, đường dẫn certs, registry URLs…) — người dùng không mở file text. Ghi config + báo mục nào cần restart.
**AC:** đổi port qua UI → restart service → UI lên port mới; giá trị không hợp lệ bị chặn kèm thông báo.

### P2-07 · Docs Phase 2 🤖 · M · Deps: P2-02..06
L2: `installer-linux`, `config-and-settings-ui`, `certs-directory`; L4: `quickstart-linux`, `first-login`; L5: `backup-restore`.
**AC:** người ngoài dự án theo `quickstart-linux.md` cài thành công không cần hỏi (test bằng 1 người/agent chưa từng đụng repo).

**🏁 Phase 2 DONE khi:** E2E VM offline: tar.gz → install.sh → login → đổi mật khẩu → ném cert → chạy lệnh WinRS xuống Win11 — toàn chuỗi xanh trong CI (chạy nightly).

---

## PHASE 3 — Registry + Templates (tuần 10–13)

### P3-01 · Schema registry + ký số 🤖+👤 · M · Deps: P0-09
`index.json` schema (artifact types: plugin, template, endpoint-script; version, checksum, signature, min-core-version); chọn cosign/minisign; **quy trình quản lý khóa ký: 👤 maintainer giữ, CI ký qua secret — AI không bao giờ đụng private key**.
**AC:** schema versioned + validator; doc key-management approved; artifact sửa 1 byte → verify FAIL.

### P3-02 · Registry client trong core (patch 0005) 🤖+👤 · L · Deps: P3-01, P1-06
Search/install/update/verify-signature/pin-version; cấu hình mặc định 2 registry: `local` + `public`. UI trang Plugins: cài từ registry, upload thủ công.
**AC:** cài plugin hello từ registry local qua UI; artifact chữ ký sai → từ chối + thông báo rõ; đủ bộ-4 cho patch 0005.

### P3-03 · Registry local khi cài + pipeline publish public 🤖 · M · Deps: P3-01, P2-02
`install.sh` sinh registry local (thư mục + index đã ký) từ artifact bundle sẵn — air-gapped dùng được; CI publish registry public (GitHub Pages) khi release.
**AC:** VM offline cài plugin từ registry local OK; push tag release → index public cập nhật tự động.

### P3-04 · Endpoint Script Manager (patch 0004) 🤖+👤 · L · Deps: P1-06, P3-02
Kho script chuẩn bị endpoint (bật/tắt WinRS, cấu hình SSH, enroll cert, cài pwsh) — artifact type `endpoint-script` từ registry; UI chọn endpoint → script → chạy → log realtime (tái dùng task runner Semaphore).
**AC:** từ UI bật WinRS trên Win11 lab thành công rồi winrs-cert gọi lệnh được ngay; đủ bộ-4 cho patch 0004.

### P3-05 · 5 templates preload 🤖+👤 · L · Deps: P3-03
JEA/WinRS Automation (enroll cert → WinRM HTTPS cert-auth → JEA → test), Docker cluster setup (cài lên client đích), Odoo, MariaDB standalone, ClickHouse standalone. Mỗi cái: `template.yaml` + playbook + docs; ansible-lint/shellcheck sạch.
**AC:** mỗi template chạy 1-click từ UI thành công trên môi trường test tương ứng; JEA/WinRS template pass trên Win11 lab.

### P3-06 · SDK Python + plugin demo 🤖 · M · Deps: P1-02
`quickwin-plugin-sdk-py` (grpcio + stub generate từ CÙNG file proto, handshake go-plugin chuẩn stdout); 1 plugin Python demo chứng minh hợp đồng.
**AC:** plugin Python được Plugin Manager load y như plugin Go, API động hoạt động; guide `plugin-dev-guide-python.md` được agent mới làm theo thành công.

### P3-07 · Docs Phase 3 🤖 · M · Deps: P3-02..06
L2: `registry-format`, `endpoint-script-manager`; L3: `template-authoring-guide`, `plugin-dev-guide-python`; L4: `managing-endpoints`, `certificates`, `using-templates`, `installing-plugins`, `settings-reference`.
**AC:** đủ file, frontmatter chuẩn, CI docs-mapping xanh.

**🏁 Phase 3 DONE khi:** cài plugin từ 2 registry, bật WinRS từ UI, 5 template 1-click chạy, plugin Python sống chung với plugin Go.

---

## PHASE 4 — Quy trình AI + Release v1.0 (tuần 14–16)

### P4-01 · Workflow ai-triage 🤖+👤 · M · Deps: P0-12
Issue mới → AI đọc → gắn label area/severity → hỏi lại nếu thiếu info → đủ điều kiện (label `ai-approved` hoặc bug có repro) → trigger AI dev.
**AC:** 10 issue thử nghiệm được phân loại đúng ≥ 8; issue thiếu info nhận được câu hỏi cụ thể, không chung chung.

### P4-02 · 6 AI Skills 🤖+👤 · L · Deps: P4-01, guideline
`dev-core-patch`, `dev-plugin`, `dev-registry`, `dev-template`, `sync-upstream` (AI-rebase conflict theo header VÌ SAO + spec), `release`. Mỗi skill enforce đúng playbook mục 5 của guideline.
**AC:** chạy thử mỗi skill với 1 task thật → PR đúng chuẩn (bộ-4, docs, DoD); skill `dev-core-patch` từ chối task sửa được bằng plugin (test bằng task bẫy).

### P4-03 · SonarQube self-host + gate 🤖+👤 · L · Deps: P0-08, P0-01
Đã chốt **self-host** (không SonarCloud): dựng instance SonarQube (👤 chọn server đặt ở đâu — cần
reachable từ CI runner; nếu CI là GitHub Actions cloud thì instance phải expose có auth, hoặc dùng
self-hosted runner cùng mạng); cấu hình backup + upgrade định kỳ cho instance; gắn vào CI qua
`sonar-scanner` + token trong GitHub secrets; quality gate: 0 new bug/vulnerability, coverage phần
mới ≥ 70%, chặn merge khi đỏ.
**AC:** PR chứa bug cấy thử bị chặn merge; instance có backup tự động + doc vận hành
`docs/L5-operations/sonarqube-instance.md`; token chỉ nằm trong secrets (gitleaks xanh).

### P4-04 · Pipeline release tự động 🤖+👤 · L · Deps: P3-03, P2-03
Conventional Commits → release note (git-cliff/release-please); version `quickwin-vX.Y.Z-sem<upstream>`; build dist → sign → GitHub Release → publish registry index → deploy website/docs.
**AC:** push tag → trong 1 pipeline ra đủ: release + note + registry cập nhật + website cập nhật; release note nhóm đúng feat/fix/docs.

### P4-05 · Website/docs public 🤖+👤 · M · Deps: P4-04, docs các phase
Render `docs/` ra website (badge status từ frontmatter), landing page, hướng dẫn tải.
**AC:** website live trên domain đã chốt; mọi trang L4 truy cập được; link từ release note tới docs đúng.

### P4-06 · Diễn tập sync-upstream + Release v1.0 👤 · M · Deps: TẤT CẢ
Diễn tập: giả lập upstream release mới → skill sync-upstream chạy → đo thời gian (mục tiêu SLA 48h). Checklist release: license files trong artifact, hardening pass trên bản build, E2E full-chain xanh, docs đủ theo bảng 12.4 → **phát hành v1.0 Linux công khai**.
**AC:** v1.0 tải được công khai; 1 người ngoài cài thành công theo quickstart; diễn tập sync hoàn tất ≤ 48h.

---

## PHASE 5 — Sau v1.0 (chưa bẻ chi tiết — sẽ lên task khi tới)
- P5-A: Windows installer (service NSSM/sc, MariaDB Windows build) → v1.x
- P5-B: Live USB **Alpine minimal** → v2.x — package list tường minh trong git (mỗi package thêm vào phải có lý do + review), verify binary bundle (core/MariaDB/pwsh) tương thích musl hoặc build static; pwsh trên Alpine cần kiểm chứng sớm (hỗ trợ musl của PowerShell hạn chế — nếu kẹt: fallback container-less .NET self-contained hoặc đưa pwsh ra optional)
- P5-C: Khảo sát HA/cluster (ghi chú kiến trúc từ plan 1.3)

---

## Sơ đồ phụ thuộc mức phase

```
P0-01 ─┬─► P0-02 ─► P0-03 ─► P0-04 ─► P0-05 ─► P0-06
       │      │                  │
       │      ├─► P0-07          └─► P0-08 ─► P0-11
       │      └─► P0-12                │
       └─► P0-09 ─► P0-10 ◄────────────┘
                │
                ▼
P1: 01,02 ─► 03 ─► 04 ─► 05 ─► 06(patch1) ─► 07,08(patch2,3) ─► 09(winrs) ─► 10(docs)
                                    │
P2: 01(mariadb) ─► 02(install) ─► 03(pwsh/dist) · 04(certs) · 05(hardening) · 06(settings) ─► 07
                                    │
P3: 01(schema+ký) ─► 02(client) ─► 03(local+publish) ─► 04(endpoint mgr) · 05(templates) · 06(py sdk) ─► 07
                                    │
P4: 01(triage) ─► 02(skills) · 03(sonar) · 04(release pipe) ─► 05(website) ─► 06(v1.0 🚀)
```

## ⚡ Có thể bắt đầu NGAY (không chờ gì)
1. **P0-01** — chốt 5 quyết định (👤, 30 phút với bảng khuyến nghị sẵn).
2. **P0-02 → P0-06** — chuỗi khởi tạo repo/toolchain/patch-system (🤖 chạy tuần tự được ngay sau P0-01).
3. **P0-07, P0-09, P0-12, P0-13** — song song với chuỗi trên.
