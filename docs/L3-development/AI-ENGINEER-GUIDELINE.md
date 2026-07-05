# GUIDELINE CHO AI ENGINEER — Tham gia dự án QuickWin (fork Semaphore UI)

> Đọc file này TRƯỚC KHI làm bất cứ việc gì. Đây là tài liệu onboarding tự chứa (self-contained):
> đọc xong là đủ ngữ cảnh để nhận task và làm việc đúng quy trình.
> Tài liệu gốc chi tiết: `PLAN-quickwin-semaphore-fork.md` (cùng thư mục).
> Áp dụng cho cả AI agent chạy tự động lẫn engineer làm việc cùng AI.

---

## 1. Dự án này là gì (30 giây)

**QuickWin** = fork của [Semaphore UI](https://github.com/semaphoreui/semaphore) (Go + Vue, MIT)
thành nền tảng quản lý hạ tầng IT cho doanh nghiệp SMB. Khác biệt cốt lõi so với upstream:

1. **Core sạch** — sửa upstream tối thiểu, mọi thay đổi là patch script-hóa để merge upstream nhanh.
2. **Plugin-first** — mở rộng qua HashiCorp go-plugin (gRPC/Protobuf), có registry.
3. **Zero-config install** — toàn binary native trong `dist/` (core + MariaDB + pwsh), **tuyệt đối không Docker**, cài 1 lệnh.
4. **AI-driven dev** — issue → AI đọc → AI dev → quality gate → người merge.

Lộ trình: Linux (v1.0) → Windows → Live USB. Kênh thực thi xuống client: pwsh → SSH → WinRS (cert auth).

---

## 2. LUẬT CỨNG — vi phạm là PR bị từ chối tự động

| # | Luật | Vì sao |
|---|---|---|
| 1 | **KHÔNG BAO GIỜ sửa file trong `upstream/`** — kể cả 1 ký tự | `upstream/` là submodule Semaphore gốc. Sửa vào đó phá vỡ toàn bộ chiến lược sync. Muốn đổi core → viết patch trong `core-patches/` |
| 2 | **Mỗi thay đổi core = 1 patch trong `core-patches/` + 1 dòng trong `series` + 1 entry `CHANGELOG.md` + 1 spec trong `docs/L2-specs/core-patches/`** | Thiếu 1 trong 4 → CI fail. Spec là input để AI khác rebase patch khi upstream release |
| 3 | **Tính năng mới mặc định là PLUGIN, không phải core** | Chỉ 2 thứ được ở core: Endpoint Script Manager + Registry Client. Muốn thêm ngoại lệ thứ 3 → phải có ADR được maintainer duyệt, không tự quyết |
| 4 | **Protobuf chỉ được THÊM field (tag mới), không đổi/xóa field cũ** | Backward compatibility với plugin đã phát hành. Breaking change → bump APP-PROTOCOL-VERSION + ADR |
| 5 | **KHÔNG đưa Docker/container vào sản phẩm hay bộ cài** | Nguyên tắc đóng gói: binary native trong dist. Docker chỉ xuất hiện trong templates cài lên máy client đích |
| 6 | **KHÔNG xóa/sửa copyright & LICENSE gốc của Semaphore**; dependency mới phải qua `go-licenses check` | Nghĩa vụ MIT + tránh GPL link trực tiếp (MariaDB/Ansible chỉ giao tiếp qua process/socket) |
| 7 | **AI không tự merge PR, không tự tạo ADR có hiệu lực** | Gặp quyết định kiến trúc chưa có ADR → DỪNG, mở issue hỏi maintainer |
| 8 | **PR đổi behavior phải cập nhật docs L2 tương ứng** (mapping trong `docs/L2-specs/OWNERS.yaml`) | Docs-as-code là hợp đồng, CI enforce |
| 9 | **AI chỉ push branch `ai/*` — KHÔNG BAO GIỜ push `main`, tạo tag, tạo release, hay commit secrets** (key, cert thật, `.pfx/.pem/.key`, `servers.json`, IP nội bộ) | Tag ký số là trigger release, chỉ maintainer tạo. Lỡ push secret = ĐÃ LỘ → báo ngay để rotate, không chỉ xóa commit. Chi tiết: plan mục 9.5 |

---

## 3. Bản đồ repo — sửa ở đâu cho việc gì

```
quickwin/
├── upstream/          ⛔ CẤM SỬA — submodule Semaphore gốc
├── core-patches/      ✏️ thay đổi core (patch đánh số + series + CHANGELOG.md)
├── plugin-manager/    ✏️ module Go: quét /plugins, chạy go-plugin, API động
├── plugins/           ✏️ plugin chính chủ (winrs-cert, hardening, …)
├── proto/             ⚠️ hợp đồng core↔plugin — chỉ thêm, không sửa (luật #4)
├── registry/          ✏️ registry server/index + schema manifest
├── installer/         ✏️ đóng gói Linux/Windows/LiveUSB → output /project/dist/
├── templates/         ✏️ playbook templates (JEA/WinRS, Odoo, MariaDB, ClickHouse…)
├── docs/              ✏️ tài liệu L0→L5 (xem mục 6)
├── website/           🤖 auto-publish, ít khi sửa tay
├── scripts/           ✏️ sync-upstream, apply-patches, build-all
└── .github/workflows/ ⚠️ sửa phải được maintainer duyệt
```

**Quy tắc định vị task nhanh:** đọc label issue → `area:core-patch | area:plugin | area:registry |
area:installer | area:template` → chỉ được sửa trong thư mục tương ứng + docs của nó.
Task đụng ≥ 2 area → tách PR.

---

## 4. Quy trình làm 1 task (từ issue đến PR)

```
1. ĐỌC issue → xác định area, đọc spec L2 liên quan trong docs/L2-specs/
   └─ Thiếu thông tin để làm? → comment hỏi trên issue, DỪNG. Không đoán.
2. TẠO branch: ai/<issue-number>-<slug>   (người: feat/ hoặc fix/ + slug)
3. LÀM theo playbook đúng loại task (mục 5 dưới đây)
4. TỰ KIỂM trước khi mở PR — checklist mục 7
5. MỞ PR: title theo Conventional Commits, body có "Closes #<issue>",
   mô tả cái gì đổi + vì sao + cách test
6. CI chạy: build + test + SonarQube + go-licenses + patch-hygiene + docs-mapping
   └─ Đỏ → tự đọc log, tự fix, push lại. Không ping người khi chưa tự thử.
7. Maintainer review & merge. AI KHÔNG merge.
```

**Conventional Commits** (bắt buộc — release note sinh tự động từ đây):
`feat(plugin-manager): ...` | `fix(installer): ...` | `docs(L2): ...` | `chore(sync): ...`

---

## 5. Playbook theo loại task

### 5A. Sửa CORE (area:core-patch)
1. Đọc `docs/L1-architecture/upstream-sync-strategy.md` + spec các patch hiện có.
2. Tự hỏi: *"Việc này làm bằng plugin được không?"* — nếu được → đề xuất chuyển hướng trên issue.
3. Viết code thay đổi trên bản upstream đã apply patch → export thành patch mới
   `core-patches/00XX-<ten>.patch`. Patch phải: **nhỏ, 1 mục đích, có header mô tả VÌ SAO**
   (header này là thứ AI khác dùng để rebase khi upstream đổi).
4. Ưu tiên kỹ thuật "hook mỏng": patch chỉ chèn vài dòng gọi ra package ngoài cây upstream
   (ví dụ `plugin-manager/`), logic thật nằm ngoài.
5. Cập nhật đủ bộ 4: patch + `series` + `CHANGELOG.md` + spec L2.
6. Chạy `scripts/apply-patches.sh` từ upstream sạch → phải PASS toàn bộ chuỗi patch.

### 5B. Viết/sửa PLUGIN (area:plugin)
1. Đọc `docs/L2-specs/plugin-manifest.md` + `proto-contract.md`.
2. Go: dùng `quickwin-plugin-sdk-go`. Python: dùng `quickwin-plugin-sdk-py`
   (cùng 1 file `.proto` — KHÔNG tự viết stub tay).
3. Cấu trúc bắt buộc: `plugins/<name>/{plugin.yaml, binary, ui/ (optional)}`.
   `plugin.yaml` phải khai: name, version (semver), license, routes[], permissions, checksum.
4. Plugin phải chịu được: core restart, bị kill giữa chừng, gọi API với input rác.
   Health-check endpoint là bắt buộc.
5. Kèm: unit test + 1 integration test chạy plugin thật qua go-plugin handshake.

### 5C. Sửa PROTO (hiếm — cẩn trọng tối đa)
1. Chỉ THÊM field với tag number mới / THÊM rpc mới. Chạy `buf breaking` trong CI.
2. Regenerate stub cho CẢ 2 SDK (Go + Python) trong cùng PR.
3. Breaking change thực sự cần thiết → viết ADR draft, mở issue, DỪNG chờ duyệt.

### 5D. INSTALLER / đóng gói (area:installer)
1. Nhớ luật #5: mọi thứ là binary native. Build dùng Go toolchain tại `/project/Go/`.
2. Output duy nhất vào `/project/dist/`. Layout chuẩn xem `docs/L2-specs/installer-linux.md`.
3. Thay đổi thành phần bundle (thêm/bớt binary) → cập nhật `licenses/` + SBOM + spec.
4. Test bắt buộc: cài trên VM Linux sạch (không internet) → login được UI với
   `admin/quickwin123` → banner đổi mật khẩu hiện → chạy được 1 task mẫu.

### 5E. TEMPLATE playbook (area:template)
1. Cấu trúc: `templates/<name>/{template.yaml, playbook/script, docs}`.
2. `template.yaml`: metadata + inputs có mô tả UI-friendly + license.
3. Lint: `ansible-lint` / `shellcheck` phải sạch. Test trong môi trường isolated.
4. Template thao tác Windows: đường đi chuẩn là pwsh → SSH → WinRS(cert), đúng thứ tự ưu tiên đó.

### 5F. SYNC UPSTREAM (khi Semaphore ra release mới)
1. Chạy `scripts/sync-upstream.sh <new-tag>` — apply lần lượt `core-patches/series`.
2. Patch conflict → đọc header "VÌ SAO" của patch + spec L2 của nó → rebase patch giữ nguyên
   Ý ĐỊNH, không giữ nguyên từng dòng code.
3. Sau rebase: full test + build + SonarQube. SLA bản vá bảo mật: **≤ 48h**.
4. PR sync phải liệt kê: patch nào giữ nguyên / patch nào rebase / diff hành vi (nếu có).

---

## 6. Tài liệu — đọc gì trước, cập nhật gì sau

**Thứ tự đọc khi mới vào dự án (~30 phút):**
1. File này.
2. `docs/L0-overview/vision.md` + `glossary.md` — hiểu thuật ngữ (core-patch, registry, manifest…).
3. `docs/L1-architecture/system-overview.md` + `core-vs-plugin-boundary.md` + toàn bộ ADR (ngắn).
4. Spec L2 của **đúng area mình sắp làm** — không cần đọc hết.

**Khi làm xong task — cập nhật docs theo bảng:**

| Bạn đổi gì | Docs phải cập nhật |
|---|---|
| Hành vi 1 thành phần | Spec L2 tương ứng (frontmatter `related-code` chỉ đúng file đó) |
| Thêm core-patch | Spec mới `docs/L2-specs/core-patches/00XX-*.md` |
| Quyết định kiến trúc | ADR draft (status: draft) + issue xin duyệt |
| Thứ người dùng nhìn thấy | Ghi chú vào PR để skill `release` cập nhật L4 khi phát hành |
| Quy trình dev | `docs/L3-development/` tương ứng |

---

## 7. Definition of Done — tự kiểm trước khi mở PR

- [ ] Build sạch từ đầu: `scripts/build-all.sh` PASS (upstream sạch + full patch chain).
- [ ] Test: unit mới cho code mới; integration nếu đụng plugin/installer; coverage phần mới ≥ 70%.
- [ ] `go-licenses check` PASS; không thêm dependency GPL/AGPL link trực tiếp.
- [ ] Không sửa `upstream/` (git status phải sạch trong submodule).
- [ ] Đủ bộ 4 nếu là core-patch (luật #2).
- [ ] Docs L2 cập nhật theo bảng mục 6; frontmatter `updated:` mới.
- [ ] Commit theo Conventional Commits; PR body có `Closes #<issue>`.
- [ ] Tự đọc lại diff một lượt với câu hỏi: *"Người review lạnh lùng nhất sẽ chê chỗ nào?"*

---

## 8. Khi nào DỪNG và hỏi người (escalation)

DỪNG NGAY và mở issue/comment (không tự quyết, không làm tiếp) khi gặp:

1. Quyết định kiến trúc chưa có ADR (thêm ngoại lệ core, đổi giao thức, đổi cấu trúc dist…).
2. Cần breaking change proto hoặc registry schema.
3. Cần thêm dependency có license lạ / GPL / AGPL.
4. Task yêu cầu sửa `upstream/` trực tiếp hoặc yêu cầu mâu thuẫn với luật cứng mục 2
   → nêu mâu thuẫn trên issue, đề xuất phương án thay thế.
5. Sync upstream có conflict mà rebase làm ĐỔI HÀNH VI (không chỉ đổi vị trí code).
6. Bất cứ thao tác nào đụng credentials, ký số registry, hạ tầng release.

Nguyên tắc: **đoán sai kiến trúc đắt hơn nhiều so với hỏi một câu.** Nhưng lỗi build/test/lint
thì tự fix — đừng hỏi những thứ log đã trả lời.

---

## 9. Tra cứu nhanh

| Cần | Ở đâu |
|---|---|
| Plan tổng thể | `PLAN-quickwin-semaphore-fork.md` |
| Go toolchain | `/project/Go/` (không dùng Go hệ thống) |
| Output build | `/project/dist/` |
| Hợp đồng plugin | `proto/quickwin/plugin/v1/plugin.proto` |
| Schema manifest plugin | `docs/L2-specs/plugin-manifest.md` |
| Admin mặc định (dev/test) | `admin` / `quickwin123` |
| Quality gate | SonarQube: 0 new bug/vuln, coverage mới ≥ 70% |
| Versioning | `quickwin-vMAJOR.MINOR.PATCH-sem<upstream-ver>` |
| SLA vá bảo mật upstream | ≤ 48h |
| Git hosting | GitHub (primary — Issues/PR/CI/Releases) · Codeberg (mirror read-only, tự động) |
| Branch của AI | chỉ `ai/<issue>-<slug>` — luật #9 |
