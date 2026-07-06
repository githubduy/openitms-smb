# KẾ HOẠCH CHI TIẾT — Fork Semaphore UI thành nền tảng quản lý hạ tầng IT cho SMB

> Tên sản phẩm: **OpenITMS-SMB** (chốt 2026-07-06, đã kiểm tra trùng tên/trademark — ADR-0004).
> Upstream: [semaphoreui/semaphore](https://github.com/semaphoreui/semaphore) — MIT License.
> Ngày lập: 2026-07-04. Trạng thái: DRAFT v1.

---

## 0. PRIMARY PLAN — Tóm tắt 1 trang (tự chứa, cắt riêng gửi được)

### Sản phẩm là gì?
**OpenITMS-SMB** — nền tảng mã nguồn mở giúp doanh nghiệp **SMB quản lý hạ tầng IT**:
tự động hóa triển khai, chạy lệnh/script xuống máy trạm Windows 11 và server Linux từ một
Web UI duy nhất. Xây dựng bằng cách **fork [Semaphore UI](https://github.com/semaphoreui/semaphore)**
(Go + Vue, giấy phép MIT — được phép fork thương mại/cộng đồng hợp pháp).

**Giá trị cốt lõi:** cài trong **10 phút, 1 lệnh, zero-config** — SMB không cần biết
Docker hay Ansible vẫn dùng được ngay.

### 4 trụ cột thiết kế

| Trụ cột | Nội dung |
|---|---|
| **1. Core sạch** | Không sửa trực tiếp mã Semaphore; mọi thay đổi là **bộ patch script-hóa** (5 patch mỏng) → khi Semaphore ra bản mới, chạy script là đồng bộ xong, SLA vá bảo mật ≤ 48h |
| **2. Plugin-first** | Mở rộng tính năng qua **HashiCorp go-plugin** (gRPC/Protobuf); core tự quét `/plugins` và sinh API động; hỗ trợ viết plugin bằng **Go lẫn Python**; có **registry** (local cho air-gapped + public) phân phối plugin & template có ký số |
| **3. Zero-config install** | Bộ cài **toàn binary native — tuyệt đối không Docker**: core + MariaDB (tune sẵn) + PowerShell Core đóng gói chung; admin mặc định `admin/quickwin123` (ép đổi khi login đầu); ném file cert `.pfx/.pem` vào thư mục `./certs` là gọi được lệnh xuống Windows qua WinRS certificate auth |
| **4. AI-driven dev** | Cộng đồng nạp yêu cầu qua GitHub Issues → **AI tự đọc, tự dev, mở PR** → quality gate (SonarQube, license check, test) → người duyệt merge → release note/website/registry cập nhật tự động |

### Tính năng nổi bật với người dùng cuối
- **Template 1-click**: JEA/WinRS cho Windows 11, triển khai Docker/Odoo/MariaDB/ClickHouse lên máy đích.
- **Endpoint Script Manager**: bật/tắt WinRS, cấu hình SSH, enroll certificate cho máy trạm ngay từ UI.
- **Plugin hardening**: tự quét cấu hình bảo mật của chính server và cảnh báo/fix.
- Kênh thực thi xuống client: **pwsh → SSH → WinRS (cert)** — phủ cả Linux lẫn Windows.

### Pháp lý — an toàn để phát hành
MIT chỉ yêu cầu giữ nguyên copyright + license gốc của Semaphore trong mọi bản phân phối (đã
thiết kế sẵn `LICENSE-SEMAPHORE` + `NOTICE.md` + CI tự kiểm license mọi dependency).
Hai điểm đã xử lý: **đổi thương hiệu hoàn toàn** (MIT không cấp quyền dùng tên/logo "Semaphore")
và **MariaDB (GPLv2) đóng gói dạng binary độc lập giao tiếp qua socket** — hợp lệ, kèm hồ sơ GPL.

### Lộ trình (16 tuần đến v1.0)

| Giai đoạn | Tuần | Kết quả |
|---|---|---|
| 0 — Nền móng ✅ | 1–2 | Repo + patch system + CI + hồ sơ pháp lý — **xong 2026-07-05** |
| 1 — Plugin Manager 🟡 | 3–6 | Plugin chạy được ✅ (SDK Go + Plugin Manager + hello, test thật pass); còn: hook core, branding, winrs-cert tới Windows 11 |
| 2 — Đóng gói Linux | 7–9 | Bộ cài 1 lệnh hoàn chỉnh (core + MariaDB + pwsh) |
| 3 — Registry + Templates | 10–13 | Cài plugin/template từ registry, 5 template 1-click |
| 4 — AI pipeline + **Release v1.0 Linux** | 14–16 | Quy trình AI-dev vận hành, phát hành công khai |
| 5 — Sau v1.0 | — | Windows installer (v1.x) → Live USB Linux minimal (v2.x) |

### Hiện trạng & cần quyết
- ✅ Đã có: plan chi tiết này, guideline AI engineer, backlog chi tiết (TASKS.md — bảng trạng thái ở đầu file).
- ✅ Đã chốt (mục 13): baseline = **v2.18.16**; registry public = GitHub Pages;
  SonarQube self-host; Live USB = Alpine (control từng package).
- ✅ **Phase 0 xong** (2026-07-05): repo + submodule pin + patch system (test PASS) + build binary
  + smoke e2e PASS + pháp lý + docs L0/L1/ADR + governance + CI/policy (chạy thật khi lên GitHub).
- ✅ **Phase 1 lõi plugin xong**: proto v1 + SDK Go + Plugin Manager + plugin hello —
  integration test thật pass (API động, stream, tự restart). Còn: hook core (patch 0001),
  branding, config, winrs-cert.
- ✅ Tên chốt: **OpenITMS-SMB** (2026-07-06). ⏳ Còn chờ: **repo GitHub** (org/URL) để push + đăng ký domain.

> *Chi tiết từng mảng: mục 2 (pháp lý) · 3–5 (kiến trúc & plugin) · 7 (đóng gói) · 9 (quy trình
> dev & publish) · 10 (roadmap) · 12 (hệ thống tài liệu) — trong tài liệu này.*

---

## 1. Mục tiêu & phạm vi

### 1.1 Mục tiêu sản phẩm
- SMB triển khai được hệ thống quản lý hạ tầng IT trong **< 10 phút**, không cần biết Docker/Ansible.
- Cài đặt kiểu "next-next-finish": loại bỏ bước cấu hình, giảm tối đa tùy chọn.
- Mở rộng tính năng **chỉ qua plugin**, không đụng core.
- Cộng đồng đóng góp qua GitHub Issues; AI agent tự động đọc yêu cầu/bug và phát triển.

### 1.2 Phạm vi nền tảng
- **Giai đoạn đầu (v1.0): Linux** — theo đúng lộ trình phân phối Linux → Windows → Live USB (mục 7.5).

### 1.3 Ngoài phạm vi
- Multi-tenant SaaS.
- HA/cluster (single-node trước, ghi chú kiến trúc để mở sau).
- macOS: KHÔNG hỗ trợ, không nằm trong lộ trình.

---

## 2. Pháp lý — Tuân thủ giấy phép khi release

### 2.1 Nghĩa vụ MIT License (bắt buộc, không thương lượng)
MIT chỉ yêu cầu **một** điều khi phân phối (source hoặc binary):

> Giữ nguyên **copyright notice + toàn văn MIT license** của Semaphore trong mọi bản phân phối.

Checklist thực thi:
- [x] Giữ file `LICENSE` gốc của Semaphore, KHÔNG xóa/sửa dòng copyright của tác giả gốc
      (✔ `LICENSE-SEMAPHORE` nguyên văn, copyright Denis Gukov/Castaway Labs giữ nguyên).
- [x] Thêm license của mình thành file riêng (✔ `LICENSE` + `NOTICE.md` đã có trong repo), cấu trúc:
  ```
  LICENSE                  ← MIT của fork (copyright của ta, cho phần code mới)
  LICENSE-SEMAPHORE        ← MIT gốc của Semaphore, nguyên văn
  NOTICE.md                ← "This product is a fork of Semaphore UI (https://github.com/semaphoreui/semaphore), © Semaphore UI authors, MIT License. Modifications © 2026 <ta>."
  ```
- [ ] Trong bộ cài binary: nhúng thư mục `licenses/` chứa MIT gốc + license mọi dependency
      (dùng `go-licenses report ./... > THIRD_PARTY_LICENSES.md` trong CI). *(CI job đã viết — chạy ở Phase 2 khi đóng gói)*
- [ ] Trang "About" trên UI hiển thị attribution: "Based on Semaphore UI (MIT)". *(thuộc patch 0002 branding)*

### 2.2 Dependency có license KHÁC MIT — cần xử lý riêng khi ĐÓNG GÓI

| Thành phần | License | Rủi ro | Cách xử lý |
|---|---|---|---|
| **MariaDB** (đóng gói kèm) | GPLv2 | GPL lây nếu link | ✅ An toàn: phân phối dạng **binary độc lập, giao tiếp qua socket/TCP** (mere aggregation). KHÔNG static-link libmariadb vào Go binary — dùng driver `go-sql-driver/mysql` (MPL 2.0, OK). Kèm link source MariaDB + toàn văn GPLv2 trong `licenses/`. |
| **Ansible** (runtime) | GPLv3 | Tương tự | ✅ Semaphore gốc đã gọi Ansible qua **exec process riêng** — giữ nguyên mô hình này. Nếu bundle ansible vào Live USB: kèm GPLv3 text + offer source. |
| **pwsh (PowerShell Core)** | MIT | Không | Bundle thoải mái, kèm license. |
| Go stdlib + deps | BSD/MIT/Apache đa số | Thấp | `go-licenses check` trong CI, fail build nếu xuất hiện GPL/AGPL link trực tiếp. |

### 2.3 Trademark & branding (quan trọng, hay bị bỏ sót)
MIT **không** cấp quyền dùng tên/logo. Bắt buộc:
- [ ] Đổi tên sản phẩm (không dùng "Semaphore" trong tên sản phẩm/domain/logo).
- [ ] Không được ngụ ý Semaphore UI "chứng thực" fork này.
- [ ] Được phép ghi mô tả thực tế: *"a fork of / based on Semaphore UI"*.
- [ ] Thay logo, favicon, tên trong UI ở **layer branding patch** (xem mục 4).

### 2.4 License cho code mới của ta
- Core-patches, Plugin Manager, plugins, installer, registry: phát hành **MIT** (đồng nhất, dễ nhận đóng góp).
- Mỗi plugin trong registry bắt buộc khai báo `license` trong manifest; registry từ chối plugin thiếu license.

---

## 3. Kiến trúc tổng thể

```
┌────────────────────────────────────────────────────────────────┐
│                    QUICKWIN NODE (native, no Docker)           │
│                                                                │
│  ┌──────────────────────────┐    ┌──────────────────────────┐  │
│  │  Semaphore Core (fork)   │    │  MariaDB (bundled)       │  │
│  │  Go binary + Vue UI      │◄──►│  binary riêng, socket    │  │
│  │                          │    │  tuned my.cnf sẵn        │  │
│  │  + Plugin Manager module │    └──────────────────────────┘  │
│  │  + Endpoint Script Mgr   │                                  │
│  │  + Registry client       │    ┌──────────────────────────┐  │
│  └───────────┬──────────────┘    │  ./certs (thư mục sẵn)   │  │
│              │ HashiCorp go-plugin (gRPC + Protobuf)          │
│  ┌───────────▼──────────────────────────────────────────────┐  │
│  │  /plugins/                                               │  │
│  │   ├── winrs-cert/     (Go — gọi WinRS qua certificate)   │  │
│  │   ├── hardening/      (Go — quét bảo mật host)           │  │
│  │   └── <python-plugin>/ (Python, nói chuyện qua .proto)   │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
        │ pwsh / SSH / WinRS(cert)              ▲
        ▼                                       │
   Windows 11 / Linux endpoints        Registry (local + public)
                                       playbook templates + plugins
```

### 3.1 Nguyên tắc vàng
1. **Core = upstream + patch mỏng.** Mọi diff so với upstream phải nằm trong bộ patch có script hóa (mục 4).
2. **Tính năng mới = plugin**, trừ 2 ngoại lệ được duyệt tích hợp thẳng lõi:
   - Quản lý script triển khai endpoint (bật/tắt WinRS, cấu hình SSH, ...)
   - Plugin Registry client
3. **1 giao thức duy nhất giữa core ↔ plugin**: gRPC + Protobuf qua HashiCorp go-plugin.

---

## 4. Chiến lược Fork & đồng bộ Upstream (Harness)

### 4.1 Cấu trúc repo

```
quickwin/
├── upstream/                  # git submodule hoặc subtree → semaphoreui/semaphore, KHÔNG sửa tay
├── core-patches/              # TOÀN BỘ thay đổi vào core, đánh số thứ tự
│   ├── 0001-plugin-manager-hook.patch
│   ├── 0002-branding.patch
│   ├── 0003-default-config-hardcode.patch
│   ├── 0004-endpoint-script-manager.patch
│   ├── 0005-registry-client.patch
│   └── series                 # thứ tự apply (kiểu quilt/git-am)
├── plugin-manager/            # module Go mới (được patch 0001 wire vào core)
├── plugins/                   # plugin chính chủ (winrs-cert, hardening)
├── proto/                     # file .proto — hợp đồng core↔plugin, versioned
├── registry/                  # code registry server + schema manifest
├── installer/                 # script đóng gói Linux/Windows/LiveUSB
├── templates/                 # playbook templates mẫu (JEA/WinRS, Odoo, MariaDB, ClickHouse…)
├── docs/                      # TOÀN BỘ tài liệu L0→L5 (docs-as-code, xem mục 12)
├── website/                   # landing + render docs/ ra web, auto-publish
├── scripts/
│   ├── sync-upstream.ps1/.sh  # QUY TRÌNH CẬP NHẬT UPSTREAM (xem 4.2)
│   ├── apply-patches.sh
│   └── build-all.sh
└── .github/workflows/         # CI/CD + AI-dev pipeline
```

**Luật cứng:** không ai (kể cả AI agent) commit trực tiếp vào cây `upstream/`. Muốn đổi core →
tạo/sửa file trong `core-patches/` + cập nhật `series` + ghi chú vào `core-patches/CHANGELOG.md`.

### 4.2 Quy trình cập nhật khi Semaphore ra release mới

```
scripts/sync-upstream.sh <new-tag>
  1. git -C upstream fetch && checkout <new-tag>
  2. Tạo branch build mới: build/<new-tag>
  3. Apply lần lượt core-patches/series (git am / patch)
  4. Nếu patch FAIL → dừng, in ra hunk conflict
       → AI agent được trigger đọc conflict, đề xuất patch đã rebase, mở PR
  5. Chạy full test + build + SonarQube
  6. PASS → tag quickwin-vX.Y.Z-sem<upstream-ver>, sinh release note tự động
```

- Mục tiêu SLA: **cập nhật bản vá bảo mật upstream trong ≤ 48h** kể từ khi upstream release.
- Mỗi patch phải nhỏ, 1 mục đích, có header mô tả *vì sao* để AI rebase được khi conflict.

### 4.3 Quy tắc thiết kế patch để ít conflict
- Patch 0001 (plugin-manager-hook) chỉ chèn **vài dòng hook** vào `main.go`/router init — toàn bộ
  logic nằm ở package `plugin-manager/` ngoài cây upstream.
- Branding: ưu tiên override qua build flag / asset replace, hạn chế sửa file Vue gốc.
- Config mặc định: file config riêng + patch chỉ đổi đường dẫn tìm config.

---

## 5. Plugin System — HashiCorp go-plugin + Registry

### 5.1 Plugin Manager (module Go mới trong core)
Chức năng khi Semaphore khởi động:
1. Quét thư mục `/plugins/*/` — mỗi plugin là 1 thư mục chứa:
   ```
   plugins/winrs-cert/
   ├── plugin.yaml        # manifest: name, version, license, api routes, permissions, checksum
   ├── winrs-cert(.exe)   # binary thực thi (go-plugin host process)
   └── ui/                # (optional) static assets nhúng vào UI
   ```
2. Verify checksum + manifest schema → khởi chạy binary qua `hashicorp/go-plugin`
   (handshake token, gRPC, mTLS tự động của go-plugin).
3. **API động:** đọc `plugin.yaml → routes[]`, tự đăng ký endpoint
   `POST/GET /api/plugins/<name>/<route>` trên router của Semaphore → proxy sang plugin qua gRPC.
4. Health-check định kỳ, restart plugin crash, log tách riêng theo plugin.
5. UI trang "Plugins": bật/tắt, xem log, cài từ registry, upload thủ công.

### 5.2 Hợp đồng Protobuf (bí quyết kết nối, kể cả plugin Python)

`proto/quickwin/plugin/v1/plugin.proto` — **versioned, backward-compatible only**:

```proto
service Plugin {
  rpc GetMetadata(Empty) returns (Metadata);          // name, version, routes, ui hooks
  rpc HandleRequest(HttpRequest) returns (HttpResponse); // API động
  rpc RunTask(TaskSpec) returns (stream TaskEvent);   // tích hợp vào task runner Semaphore
  rpc HealthCheck(Empty) returns (Health);
}
```

- **Plugin Go**: dùng trực tiếp go-plugin SDK ta phát hành (`quickwin-plugin-sdk-go`).
- **Plugin Python**: HashiCorp go-plugin hỗ trợ non-Go plugin qua chuẩn handshake stdout
  (`CORE-PROTOCOL-VERSION|APP-PROTOCOL-VERSION|network|addr|grpc`). Ta phát hành
  `quickwin-plugin-sdk-py` (grpcio + generated stubs từ cùng file .proto) → dev Python
  chỉ implement class `PluginServicer` là chạy được. **Cùng 1 file .proto là nguồn chân lý duy nhất.**
- Quy tắc tiến hóa proto: chỉ thêm field (số tag mới), không đổi/không xóa; bump
  `APP-PROTOCOL-VERSION` khi buộc phải breaking → core chạy song song 2 version trong 1 major release.

### 5.3 Plugin Registry (tích hợp thẳng lõi — ngoại lệ được duyệt)
- **2 registry cấu hình mặc định song song:**
  1. `local` — thư mục/HTTP server nội bộ sinh ra lúc cài đặt (air-gapped OK).
  2. `public` — `https://registry.<domain>` do CI publish tự động.
- Registry là **static index**: `index.json` + tarball plugin + chữ ký (cosign/minisign) —
  host được trên GitHub Pages/object storage, không cần server động.
- Client trong core: search / install / update / verify signature / pin version.
- Playbook templates dùng **cùng cơ chế registry** (loại artifact `template` bên cạnh `plugin`).

### 5.4 Plugin mẫu (ship kèm bản cài)
| Plugin | Ngôn ngữ | Chức năng |
|---|---|---|
| `winrs-cert` | Go | Gọi lệnh WinRS xuống Windows 11 qua **certificate auth**; tự nhận cert từ `./certs` (`.pfx`/`.pem`); expose API động `POST /api/plugins/winrs-cert/exec` |
| `hardening` | Go | Quét cấu hình bảo mật host Semaphore (port, quyền file, TLS, mật khẩu default chưa đổi, firewall) → báo cáo + nút "fix" trên UI |

---

## 6. Tính năng tích hợp thẳng lõi (chỉ 2, qua core-patches)

### 6.1 Endpoint Script Manager (patch 0004)
- Kho script chuẩn bị endpoint: **bật/tắt WinRS**, cấu hình SSH, enroll certificate, cài pwsh…
- UI: chọn endpoint → chọn script → chạy → log realtime (tái dùng task runner của Semaphore).
- Script lưu dạng template có version trong registry (loại `endpoint-script`).

### 6.2 Registry Client (patch 0005) — như mục 5.3.

---

## 7. Đóng gói & Trải nghiệm cài đặt (Zero-config)

### 7.1 Thành phần bộ cài (tất cả nằm trong `/project/dist/` sau khi build)

**Nguyên tắc tuyệt đối: KHÔNG Docker, KHÔNG container runtime.** Mọi thành phần
(core, MariaDB, pwsh, plugins) đều là **binary native đóng gói sẵn trong dist** —
máy đích chỉ cần giải nén + chạy `install.sh`, không cài thêm bất kỳ dependency nào.

```
dist/
├── quickwin-linux-amd64.tar.gz     # hoặc .deb/.rpm ở phase 2
│   ├── bin/semaphore               # core binary native
│   ├── bin/pwsh/…                  # PowerShell Core — binary bundled trong dist
│   ├── mariadb/                    # MariaDB binary bundled trong dist + my.cnf tuned sẵn
│   ├── plugins/winrs-cert, hardening
│   ├── templates/                  # playbook mẫu preload
│   ├── certs/                      # volume certificate — NÉM .pfx/.pem VÀO LÀ CHẠY
│   ├── config/config.json          # env hardcode lần đầu (xem 7.3)
│   ├── licenses/                   # MIT gốc + third-party
│   └── install.sh                  # 1 lệnh duy nhất
└── quickwin-windows-amd64.zip      # phase 2
```

- Toolchain Go cài tại `/project/Go/` (không dùng Go hệ thống — build reproducible).
- Build cross-platform bằng chính Go (`GOOS/GOARCH`), CI giữ checksum + SBOM.

### 7.2 MariaDB đóng gói mặc định
- Chọn MariaDB LTS (11.4.x) binary tarball, chạy user riêng `quickwin-db`, **chỉ listen unix
  socket/localhost** (không mở TCP ra ngoài).
- `my.cnf` tune sẵn cho SMB (buffer pool theo % RAM, utf8mb4, slow-log bật).
- `install.sh` tự: init datadir → tạo DB `semaphore` + user + password random → ghi vào config
  → systemd unit `quickwin-db.service` + `quickwin.service` (After=quickwin-db).
- GPL compliance như mục 2.2.

### 7.3 Cấu hình lần đầu — hardcode, sau đó sửa qua UI
- `config/config.json` sinh sẵn khi cài: port 3000, DB socket, đường dẫn certs, admin mặc định.
- **Admin mặc định: `admin` / `quickwin123`** — UI hiển thị banner đỏ bắt đổi mật khẩu ở lần
  đăng nhập đầu (plugin `hardening` cũng cảnh báo nếu chưa đổi).
- Trang **Settings UI** (patch 0003 + plugin-manager) cho sửa các env/config sau cài đặt —
  người dùng không bao giờ phải mở file text.

### 7.4 Thư mục certificates (`./certs`)
- Thư mục `./certs` là thư mục local nằm sẵn trong bộ cài, được khai báo sẵn trong config
  (không phải Docker volume); watcher tự nhận file `.pfx`/`.pem`
  mới → nạp vào cert store nội bộ → plugin `winrs-cert` và inventory WinRM dùng ngay,
  không cần restart.

### 7.5 Lộ trình phân phối (thứ tự nền tảng — khác với "Phase" của roadmap mục 10)
| Thứ tự | Nền tảng | Ứng với roadmap |
|---|---|---|
| 1 | **Linux** (tar.gz + install.sh, systemd) | Phase 2 → release v1.0 ở Phase 4 |
| 2 | **Windows** (zip + install.ps1, Windows Service qua NSSM/sc) | Phase 5, v1.x — MariaDB Windows build |
| 3 | **Live USB Linux minimal** (Alpine — package list tường minh, control từng gói; boot là có server chạy sẵn — "cứu hộ/triển khai tại chỗ") | Phase 5, v2.x |

---

## 8. Template Playbooks (Registry: local + public, mặc định bật cả 2)

Preload trên UI, chạy "1 click":

1. **JEA/WinRS Automation** — kịch bản chuẩn hóa xử lý certificate để gọi PowerShell
   xuống Windows 11 (enroll cert → cấu hình WinRM HTTPS cert-auth → JEA endpoint → test).
2. **App Deployment** — template Ansible/Bash (lưu ý: Docker ở đây là thứ template
   cài **lên máy client đích** theo yêu cầu người dùng — bản thân OpenITMS-SMB không dùng Docker):
   - Docker cluster setup (host chuẩn bị + compose)
   - Odoo one-click
   - MariaDB standalone
   - ClickHouse standalone
3. **Endpoint provisioning** — bật/tắt WinRS, cấu hình SSH key, cài pwsh trên client.

Chuẩn template: `template.yaml` (metadata, inputs có mô tả UI, license) + playbook/script +
docs — publish lên registry qua CI giống plugin.

---

## 9. Quy trình phát triển cộng đồng + AI

### 9.1 Nơi nạp yêu cầu
- **GitHub Issues** với Issue Forms bắt buộc: `feature-request.yaml`, `bug-report.yaml`,
  `plugin-proposal.yaml` — form có field máy-đọc-được (component, severity, repro).
- Label routing: `area:core-patch` / `area:plugin` / `area:registry` / `area:installer` / `area:template`.

### 9.2 AI tự động đọc & dev
```
Issue mới → workflow ai-triage:
  1. AI đọc issue → phân loại, gắn label, hỏi lại nếu thiếu thông tin (comment)
  2. Nếu đủ điều kiện (label ai-approved bởi maintainer HOẶC bug có repro rõ):
     → spawn AI dev agent (Claude Code headless) trên branch ai/<issue-number>
  3. AI dev theo AI Skill tương ứng loại thay đổi (xem 9.3) → mở PR
  4. Gate tự động: build + test + SonarQube + license-check + patch-hygiene-check
  5. Maintainer (người) review & merge — AI KHÔNG tự merge
```

### 9.3 AI Skills kiểm soát quy trình (mỗi loại thay đổi 1 skill riêng)

Mọi AI agent (và engineer mới) bắt buộc nạp `docs/L3-development/AI-ENGINEER-GUIDELINE.md`
làm ngữ cảnh đầu tiên trước khi nhận task — guideline chứa luật cứng, bản đồ repo,
playbook theo loại task, Definition of Done và luật escalation. Các skill dưới đây
là bản "thực thi được" của từng playbook trong guideline đó.

| Skill | Phạm vi | Ràng buộc chính |
|---|---|---|
| `dev-core-patch` | Sửa core | CHỈ được sửa qua `core-patches/`; patch mới phải vào `series` + CHANGELOG; cấm sửa `upstream/` |
| `dev-plugin` | Plugin mới/sửa plugin | Chỉ trong `plugins/`; phải qua SDK + proto hiện hành; kèm test + manifest |
| `dev-registry` | Registry/index | Schema versioned, ký số bắt buộc |
| `dev-template` | Playbook templates | Lint ansible-lint/shellcheck; test container |
| `sync-upstream` | Rebase patch khi upstream release | Chạy quy trình 4.2, tự xử conflict, mở PR |
| `release` | Đóng gói phát hành | Sinh release note, bump version, publish registry, cập nhật website |

### 9.4 Chất lượng & tự động hóa
- **SonarQube** quét mọi PR (quality gate: 0 new bug/vulnerability, coverage phần code mới ≥ 70%).
- Conventional Commits → **release note tự sinh** (git-cliff/release-please).
- Semantic version: `quickwin-vMAJOR.MINOR.PATCH-sem<upstream-version>`.
- CI release pipeline: build dist → sign → GitHub Release → publish registry index →
  **auto-deploy website/docs** (docs-as-code, cùng repo, GitHub Pages).

---

### 9.5 Quy định publish & lựa chọn nền tảng Git

#### 9.5.1 Nền tảng — GitHub chính, mirror sang nền tảng đang lên

| Nền tảng | Vai trò | Lý do |
|---|---|---|
| **GitHub** | **PRIMARY** — source of truth, Issues, PR, CI (Actions), Releases, Pages (registry public + website) | Cộng đồng lớn nhất → nhận contributor dễ nhất; Issue Forms + Actions + Pages là hạ tầng miễn phí toàn bộ pipeline AI-dev của ta |
| **Codeberg (Forgejo)** | **MIRROR** — push-mirror tự động, read-only | Nền tảng đang lên đáng chú ý nhất: phi lợi nhuận EU, chạy Forgejo (fork cộng đồng của Gitea, đang phát triển federation ForgeFed); chống phụ thuộc 1 nhà cung cấp; cộng đồng FOSS thuần ưa chuộng |
| **Forgejo/Gitea self-host** | TÙY CHỌN — bản sao nội bộ cho khách air-gapped | Nhẹ (1 binary Go — cùng triết lý OpenITMS-SMB), khách SMB có thể tự host mirror + registry local |

Các lựa chọn khác đã cân nhắc và KHÔNG chọn làm primary:
- **GitLab**: đầy đủ nhưng nặng, cộng đồng OSS drive-by contributor kém hơn GitHub.
- **SourceHut**: tối giản, workflow email-patch — rào cản cao với audience SMB.
- **Radicle** (P2P, đang lên): thú vị về chống kiểm duyệt nhưng tooling/CI chưa đủ chín — theo dõi lại sau v1.0.

Quy tắc mirror: GitHub → Codeberg **một chiều, tự động sau mỗi push main + tag** (Forgejo có
pull-mirror sẵn, hoặc Action push-mirror). Issue/PR CHỈ nhận trên GitHub — README trên mirror
ghi rõ "mirror, đóng góp tại GitHub" để tránh phân mảnh.

#### 9.5.2 Quy định publish (bắt buộc, CI + branch protection enforce)

**Cái gì KHÔNG BAO GIỜ được publish:**
- Secrets, private keys (đặc biệt khóa ký registry — mục 5.3), certificates thật, `servers.json`,
  IP/hostname hạ tầng nội bộ, credentials khách hàng.
- Bật **GitHub secret scanning + push protection** ngay khi tạo repo; thêm `gitleaks` vào CI
  làm lớp 2; `.gitignore` chặn sẵn `*.pfx *.pem *.key servers.json .env`.
- Lỡ push secret → coi như ĐÃ LỘ: rotate ngay, không chỉ xóa commit.

**Quy định branch & tag:**
- `main` protected: không push trực tiếp, không force-push, PR + 1 review (người) + CI xanh mới merge.
- AI agent chỉ được push branch `ai/*`; người: `feat/*`, `fix/*`, `docs/*`, `chore/*`.
- Tag release (`quickwin-v*`) **ký GPG/sigstore, chỉ maintainer tạo** (hoặc CI tạo sau khi
  maintainer approve) — tag là thứ trigger pipeline release nên là điểm kiểm soát cuối.
- Maintainer bắt buộc bật 2FA; quyền admin repo tối thiểu 2 người, tối đa ít người nhất có thể.

**Checklist trước MỌI lần publish release (skill `release` enforce):**
1. `LICENSE`, `LICENSE-SEMAPHORE`, `NOTICE.md`, `THIRD_PARTY_LICENSES.md` có mặt trong artifact.
2. `gitleaks` + secret scanning sạch trên toàn bộ diff kể từ release trước.
3. Artifact kèm SHA256 checksum + SBOM + chữ ký.
4. Release note sinh tự động đã được maintainer đọc lại (không lộ thông tin nội bộ trong commit message).
5. Mirror Codeberg đã sync tag mới.

**Vệ sinh lịch sử commit:** commit message không chứa đường dẫn nội bộ/tên khách hàng;
squash-merge mặc định cho PR của AI để lịch sử `main` sạch.

## 10. Roadmap & Milestones

### Phase 0 — Nền móng (tuần 1–2) — ✅ DONE 2026-07-05 (phần local; CI/mirror chờ repo GitHub)
- [x] Lập repo cấu trúc mục 4.1; submodule upstream, pin tag Semaphore mới nhất (**v2.18.16**).
- [x] Cài Go toolchain (1.24.6) vào `Go/`, `build-all.sh` build core nguyên bản
      (binary `quickwin-dev-sem2.18.16`; backend-only dùng placeholder UI assets, `FULL_UI=1` cho bản đầy đủ).
- [x] Bộ khung `sync-upstream` + `apply-patches` chạy rỗng pass (+ test patch giả PASS / patch hỏng FAIL đúng).
- [x] Hồ sơ pháp lý: LICENSE/LICENSE-SEMAPHORE/NOTICE.md ✔; CI `go-licenses` gate đã viết
      (chạy thật khi push GitHub — cùng với P0-13 hardening/mirror).
- [x] *(bổ sung)* Docs L0/L1 + ADR-0001→0004, guideline AI, governance/DCO/issue forms,
      testing-strategy + `tests/e2e/smoke.sh` PASS (server thật, /api/ping + UI 200).

### Phase 1 — Core patches + Plugin Manager (tuần 3–6) — 🟡 đang làm
- [x] Package `plugin-manager/` quét `/plugins`, chạy go-plugin, API động từ manifest,
      health-restart — integration test thật PASS (commit 6913057).
      → [ ] còn: Patch 0001 hook vào core + enforce permissions từng quyền (P1-05/06).
- [x] `proto/` v1 + SDK Go + plugin `hello` chạy end-to-end qua go-plugin.
      → [ ] còn: plugin `winrs-cert` tới Windows 11 qua cert (P1-09 — cần Win11 lab).
- [ ] Patch 0002 branding (chờ chốt tên) + 0003 config hardcode + Settings UI tối thiểu.

### Phase 2 — Đóng gói Linux (tuần 7–9)
- [ ] Bundle MariaDB + systemd + `install.sh` 1 lệnh; output vào `/project/dist/`.
- [ ] Watcher thư mục certs; admin default + banner đổi mật khẩu.
- [ ] Plugin `hardening` v1.

### Phase 3 — Registry + Templates (tuần 10–13) — 🟡 phần lớn xong
- [x] Registry static index + ký số (ed25519) — package `registry/` + client + CLI, test pass.
      → [ ] còn: client trong core (patch 0005) + UI + local/public publish pipeline.
- [ ] Endpoint Script Manager (patch 0004) — cần core patch + UI.
- [x] 5 template preload (JEA/WinRS, Docker, Odoo, MariaDB, ClickHouse) + CI shellcheck.
- [x] SDK Python + plugin `hello-py` (chứng minh hợp đồng .proto đa ngôn ngữ) — verify CI.

### Phase 4 — Quy trình AI + Release công khai (tuần 14–16) — 🟡 khung xong
- [x] Issue forms + ai-triage (keyword) + 6 AI dev skills + SonarQube gate (OK).
      → [ ] AI-dev đầy đủ cần self-hosted runner + ANTHROPIC_API_KEY (maintainer bật).
- [x] Pipeline release (`release.yml`): release note + publish registry — [ ] cần secret + package.sh.
- [ ] Website public (render docs) — cần Pages/domain.
- [ ] **Public release v1.0 (Linux)** — chờ đóng gói Phase 2 hoàn chỉnh.

### Phase 5 — Windows & Live USB (sau v1.0) — scaffolding doc xong
- [ ] Windows installer (service) → v1.x. *(doc: L5 phase5-windows-installer.md)*
- [ ] Live USB **Alpine** minimal → v2.x. *(doc: phase5-liveusb-alpine.md — rủi ro musl/pwsh ghi rõ)*
- [ ] HA/cluster khảo sát. *(doc: phase5-ha-survey.md)*

---

## 11. Rủi ro chính & đối sách

| Rủi ro | Đối sách |
|---|---|
| Upstream refactor lớn làm vỡ patch | Patch mỏng + hook tối thiểu; skill `sync-upstream` AI-rebase; test hợp đồng hook |
| go-plugin overhead / plugin crash | Health-check + restart; plugin isolate process nên không sập core |
| GPL MariaDB khi đóng gói | Aggregation, socket-only, kèm license + link source (mục 2.2) |
| Mật khẩu default bị bỏ quên | Banner cưỡng chế + hardening plugin cảnh báo + doc |
| AI dev tạo PR kém chất lượng | Quality gate SonarQube + patch-hygiene check + human merge duy nhất |
| Trademark "Semaphore" | Rebrand toàn bộ từ Phase 1 (patch 0002) |

---

## 12. Hệ thống tài liệu — 6 level (docs-as-code)

Toàn bộ tài liệu sống trong `docs/` cùng repo, viết Markdown, render tự động ra website
khi release (mục 9.4). Nguyên tắc: **mỗi tài liệu thuộc đúng 1 level, có audience rõ,
có trigger cập nhật rõ** — AI skill và CI dựa vào đó để bắt buộc cập nhật docs khi code đổi.

### 12.1 Định nghĩa các level

| Level | Tên | Trả lời câu hỏi | Audience | Vòng đời |
|---|---|---|---|---|
| **L0** | Tầm nhìn & Tổng quan | *Sản phẩm này là gì, cho ai, vì sao?* | Tất cả (kể cả người không kỹ thuật) | Ít đổi, review theo quý |
| **L1** | Kiến trúc & Quyết định | *Hệ thống ghép với nhau thế nào, vì sao chọn vậy?* | Maintainer, AI dev agent | Đổi khi có ADR mới |
| **L2** | Đặc tả kỹ thuật (Spec) | *Từng thành phần hoạt động chính xác ra sao?* | Dev, AI dev agent | Đổi cùng PR code — CI gate |
| **L3** | Quy trình & Runbook dev | *Làm việc trên repo này như thế nào?* | Contributor, AI skills | Đổi khi quy trình đổi |
| **L4** | Tài liệu người dùng | *Cài đặt, dùng, mở rộng sản phẩm ra sao?* | Admin SMB, plugin author | Đổi mỗi release |
| **L5** | Vận hành & Xử lý sự cố | *Chạy production thế nào, hỏng thì làm gì?* | Ops/admin SMB | Đổi khi có sự cố mới/known issue |

### 12.2 Cây thư mục `docs/`

```
docs/
├── L0-overview/
│   ├── vision.md                    # tầm nhìn, đối tượng SMB, giá trị cốt lõi
│   ├── roadmap.md                   # sync từ mục 10 plan này
│   ├── glossary.md                  # thuật ngữ: core-patch, registry, plugin, template…
│   └── legal.md                     # tóm tắt nghĩa vụ MIT/GPL (mục 2) cho người phân phối
├── L1-architecture/
│   ├── system-overview.md           # sơ đồ tổng thể (mục 3) + data flow
│   ├── core-vs-plugin-boundary.md   # tiêu chí "cái gì được vào core" (2 ngoại lệ)
│   ├── upstream-sync-strategy.md    # chiến lược patch mỏng (mục 4)
│   └── adr/                         # Architecture Decision Records, đánh số
│       ├── ADR-0001-hashicorp-go-plugin.md
│       ├── ADR-0002-mariadb-bundled-socket-only.md
│       ├── ADR-0003-static-registry-signed-index.md
│       └── ADR-template.md
├── L2-specs/                        # 1 spec / 1 thành phần — nguồn chân lý cho AI dev
│   ├── plugin-manager.md            # lifecycle, quét /plugins, API động, health-check
│   ├── plugin-manifest.md           # schema plugin.yaml (JSON Schema kèm theo)
│   ├── proto-contract.md            # hợp đồng .proto v1 + quy tắc tiến hóa (mục 5.2)
│   ├── registry-format.md           # index.json schema, ký số, artifact types
│   ├── endpoint-script-manager.md   # patch 0004
│   ├── installer-linux.md           # layout dist/, install.sh, systemd units
│   ├── config-and-settings-ui.md    # config.json schema, hardcode lần đầu, Settings UI
│   ├── certs-directory.md           # thư mục ./certs: watcher, format hỗ trợ, nạp nóng
│   └── core-patches/                # MỖI PATCH 1 SPEC — bắt buộc
│       ├── 0001-plugin-manager-hook.md
│       ├── 0002-branding.md
│       └── …
├── L3-development/
│   ├── AI-ENGINEER-GUIDELINE.md     # onboarding tự chứa cho AI engineer/agent — ĐỌC ĐẦU TIÊN
│   │                                #   (draft hiện tại: GUIDELINE-AI-ENGINEER.md cạnh plan này)
│   ├── CONTRIBUTING.md              # symlink/include từ root
│   ├── repo-structure.md            # giải thích cây repo (mục 4.1)
│   ├── build-from-source.md         # /project/Go/, build-all.sh
│   ├── patch-authoring-guide.md     # cách viết core-patch đúng chuẩn (mục 4.3)
│   ├── plugin-dev-guide-go.md       # SDK Go, từ zero → plugin chạy được
│   ├── plugin-dev-guide-python.md   # SDK Python + handshake go-plugin
│   ├── template-authoring-guide.md  # viết playbook template + template.yaml
│   ├── ai-workflow.md               # pipeline issue → AI dev → PR (mục 9.2)
│   └── runbooks/
│       ├── sync-upstream.md         # quy trình 4.2 từng bước + xử lý conflict
│       ├── release.md               # quy trình phát hành + publish registry/website
│       └── security-patch-sla.md    # quy trình 48h vá bảo mật upstream
├── L4-user/
│   ├── quickstart-linux.md          # cài trong 10 phút
│   ├── first-login.md               # admin/quickwin123 → đổi mật khẩu → tour UI
│   ├── managing-endpoints.md        # thêm Windows 11/Linux client, script bật WinRS/SSH
│   ├── certificates.md              # ném .pfx/.pem vào ./certs → JEA/WinRS
│   ├── using-templates.md           # chạy template 1-click (Odoo, Docker…)
│   ├── installing-plugins.md        # cài từ registry local/public
│   └── settings-reference.md        # mọi option trên Settings UI
└── L5-operations/
    ├── backup-restore.md            # backup MariaDB + config + certs
    ├── upgrade.md                   # nâng cấp giữa các bản quickwin
    ├── security-hardening.md        # checklist + liên kết plugin hardening
    ├── troubleshooting.md           # lỗi thường gặp: DB không lên, plugin crash, cert sai
    └── known-issues.md              # sinh tự động từ label GitHub khi release
```

### 12.3 Luật liên kết docs ↔ code (CI enforce)

1. **PR đổi behavior mà không đổi docs L2 tương ứng → CI fail** (mapping qua
   `docs/L2-specs/OWNERS.yaml`: mỗi spec khai báo các đường dẫn code nó "sở hữu").
2. **Patch core mới bắt buộc có spec** trong `docs/L2-specs/core-patches/<số>-<tên>.md`
   trước khi được nhận vào `series` — đây là input để AI rebase khi upstream đổi.
3. **Quyết định kiến trúc chỉ có hiệu lực khi có ADR** (L1) — AI dev agent bị skill ràng buộc:
   gặp quyết định chưa có ADR thì dừng và mở issue, không tự quyết.
4. **L4/L5 sinh & review mỗi release** — skill `release` diff các spec L2 đã đổi kể từ bản
   trước và cập nhật user docs tương ứng; release note link tới trang docs liên quan.
5. Mỗi file docs có frontmatter chuẩn:
   ```yaml
   ---
   level: L2
   status: draft | approved | deprecated
   owners: [maintainer-handle]
   updated: 2026-07-04
   related-code: [plugin-manager/, core-patches/0001-*]
   ---
   ```
   → website render badge trạng thái; AI dùng `related-code` để biết doc nào cần sửa.

### 12.4 Thứ tự viết tài liệu theo phase (khớp mục 10)

| Phase | Docs phải có khi kết thúc phase |
|---|---|
| 0 | L0 đủ; L1 `system-overview` + `upstream-sync-strategy` + ADR-0001..0003; L3 `repo-structure`, `build-from-source`, `patch-authoring-guide` |
| 1 | L2: `plugin-manager`, `plugin-manifest`, `proto-contract`, spec patch 0001–0003; L3 `plugin-dev-guide-go` |
| 2 | L2: `installer-linux`, `config-and-settings-ui`, `certs-directory`; L4 `quickstart-linux`, `first-login`; L5 `backup-restore` |
| 3 | L2: `registry-format`, `endpoint-script-manager`; L3 `template-authoring-guide`, `plugin-dev-guide-python`; L4 phần còn lại |
| 4 | L3 `ai-workflow` + toàn bộ `runbooks/`; L5 hoàn chỉnh; website public cùng v1.0 |

---

## 13. Quyết định đã chốt & còn mở

### ĐÃ CHỐT (2026-07-05 — ghi vào ADR-0004 khi init repo)

| # | Quyết định | Chốt |
|---|---|---|
| 2 | Baseline upstream | **Tag release stable mới nhất** của semaphoreui/semaphore tại ngày init repo (không dùng HEAD) |
| 3 | Registry public | **GitHub Pages** (static index + artifact) |
| 4 | SonarQube | **Self-host** (cần thêm task dựng + vận hành instance — backlog P4-03) |
| 5 | Live USB base | **Alpine** — kiểm soát cẩn thận TỪNG package đưa vào image (danh sách package tường minh, có review); lưu ý kỹ thuật: musl libc → binary bundle (core/MariaDB/pwsh) phải build/chọn bản tương thích musl hoặc static, kiểm chứng ở P5-B |

### CÒN MỞ

1. **Tên sản phẩm chính thức + domain** — chặn patch branding (P1-07) và URL registry public.
   ĐÃ CHỐT 2026-07-06: **OpenITMS-SMB** (kiểm tra: không có project/trademark active trùng "openitms"; lưu ý "ITMS™" của WidePoint — xem ADR-0004).
