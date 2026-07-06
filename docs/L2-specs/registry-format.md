---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [registry/]
---

# Spec: Registry format (index tĩnh ký số)

ADR-0003: registry = **static index + chữ ký**, không server động. Module `quickwin.dev/registry`.
Mặc định cấu hình 2 source: `local` (thư mục cài, air-gapped) + `public` (GitHub Pages).

## Layout
```
<registry-root>/
├── index.json          # danh mục (được KÝ)
├── index.json.sig      # chữ ký ed25519 của index.json canonical
├── plugin/<name>/<name>-<ver>.tar.gz
├── template/<name>/<name>-<ver>.tar.gz
└── endpoint-script/<name>/<name>-<ver>.tar.gz
```

## index.json
```json
{
  "schema": 1,
  "name": "public",
  "generated": "2026-07-06T00:00:00Z",
  "artifacts": [
    {"name":"winrs-cert","version":"0.1.0","type":"plugin","license":"MIT",
     "path":"plugin/winrs-cert/winrs-cert-0.1.0.tar.gz",
     "sha256":"sha256:<64hex>","size":12345,"min_core_version":""}
  ]
}
```
- `type` ∈ `plugin | template | endpoint-script`.
- `license` **bắt buộc** — thiếu → từ chối build/index.
- `schema` versioned; client từ chối index schema mới hơn nó hỗ trợ.

## Mô hình tin cậy (ed25519 — stdlib, không cần cosign/minisign binary)
1. `index.json` được KÝ bằng **private key maintainer** → `index.json.sig` (`ed25519:<base64>`).
2. Client verify chữ ký index bằng **trusted public key** (nhúng trong core).
3. Tin sha256 trong index đã verify → verify sha256 mỗi tarball tải về.
4. Một chữ ký phủ toàn danh mục; sửa 1 byte index hoặc 1 tarball → phát hiện.
- **Ký (chuỗi canonical)**: `Index.Canonical()` sort artifact tất định rồi marshal — byte phải khớp tuyệt đối khi verify.
- **Private key**: env `REGISTRY_PRIVATE_KEY`, CI secret — **KHÔNG commit** (publishing-policy).
  AI KHÔNG BAO GIỜ đụng private key (guideline escalation #6).

## API package
- Build (maintainer/CI): `PackArtifact` / `BuildIndex` / `WriteRegistry`; CLI `registryctl keygen|build`.
- Client (core): `NewClient(trusted...)` → `LoadIndex` (verify sig) / `Search` / `Install`
  (verify sig + checksum) / `Unpack` (chống path-traversal). Fetcher: `file://` + `http(s)://`.

## Test (PASS 2026-07-06)
Round-trip build→sign→install→unpack; từ chối: index bị sửa, ký sai key, tarball bị sửa,
thiếu license; pick version cao nhất theo semver (không string-sort).

## Còn lại (P3-02/03 phần tích hợp)
- Patch 0005: hook registry client vào core + API `/api/registry/{search,install}` (như patch 0001).
- UI trang Plugins: cài từ registry / upload thủ công.
- Installer sinh `local` registry lúc cài; CI publish `public` lên GitHub Pages khi release.
