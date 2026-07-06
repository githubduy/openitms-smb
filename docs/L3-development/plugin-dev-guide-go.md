---
level: L3
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [sdk/go/, plugins/hello/]
---

# Viết plugin Go — từ zero đến chạy được

Mẫu hoàn chỉnh: `plugins/hello/` (~100 dòng). Làm theo 4 bước:

## 1. Tạo module
```
plugins/<ten-plugin>/
├── go.mod        # module quickwin.dev/plugins/<ten-plugin>
├── main.go
└── plugin.yaml
```
go.mod cần 2 replace (xem `plugins/hello/go.mod`): `quickwin.dev/sdk` → `../../sdk/go`,
`quickwin.dev/proto` → `../../proto/gen/go`.

## 2. Implement 4 method của `sdk.Plugin`
```go
type myPlugin struct{}
func (p *myPlugin) Metadata(ctx) (*pluginv1.Metadata, error)      // PHẢI khớp plugin.yaml
func (p *myPlugin) HandleRequest(ctx, req) (*pluginv1.HttpResponse, error) // API động
func (p *myPlugin) RunTask(ctx, spec, emit) (*pluginv1.TaskResult, error)  // task dài, emit.Log()
func (p *myPlugin) Health(ctx) (*pluginv1.Health, error)
func main() { sdk.Serve(&myPlugin{}) }
```
Quy tắc: HandleRequest xử lý theo `req.Path` (đã strip prefix); RunTask tôn trọng
`ctx.Done()` (core cắt khi timeout); Health trả DEGRADED nếu chạy được nhưng có vấn đề.

## 3. Viết plugin.yaml
Schema: `plugin-manager/schema/plugin.schema.json` · spec: `docs/L2-specs/plugin-manifest.md`.
Bắt buộc: name (kebab-case), version (semver), **license**, protocol_version: 1, entrypoint.
`routes` phải khớp 100% Metadata — lệch là Plugin Manager từ chối load.

## 4. Test
- Unit test logic của bạn như Go thường.
- Integration: bắt chước `plugin-manager/integration_test.go` — hoặc đơn giản:
  build binary vào `<plugins-dir>/<ten>/` cùng plugin.yaml, chạy core (hoặc Manager
  standalone), gọi `POST /api/plugins/<ten>/<route>`.

## Checklist trước khi mở PR (kèm DoD guideline)
- [ ] `go test ./...` xanh; plugin sống sót khi core restart / kill giữa chừng.
- [ ] Không panic với input rác (fuzz nhẹ HandleRequest).
- [ ] plugin.yaml validate qua schema; license khai đúng SPDX.
- [ ] Cập nhật spec L2 nếu plugin có hành vi hệ thống mới.
