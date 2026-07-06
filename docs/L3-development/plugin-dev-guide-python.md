---
level: L3
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [sdk/python/, plugins/hello-py/]
---

# Viết plugin Python — từ zero đến chạy được

Mẫu: `plugins/hello-py/` (~40 dòng). Cùng file `.proto` với plugin Go — hợp đồng không
phụ thuộc ngôn ngữ (ADR-0001). Plugin Manager launch plugin Python **y hệt** plugin Go.

## Cơ chế
HashiCorp go-plugin cho phép plugin non-Go: core spawn `python3 main.py`, plugin in **1 dòng
handshake** ra stdout (`1|1|tcp|127.0.0.1:<port>|grpc`) rồi serve gRPC theo `plugin.proto`.
SDK (`quickwin_plugin`) lo toàn bộ phần này. (Hiện core chưa bật AutoMTLS nên plugin Python
chỉ cần serve insecure trên localhost.)

## 1. Sinh stub từ proto (1 lần / khi proto đổi)
```bash
pip install grpcio grpcio-tools
scripts/gen-proto-py.sh          # → proto/gen/python/quickwin/plugin/v1/plugin_pb2*.py
```

## 2. Viết plugin
```python
from quickwin_plugin import Plugin, serve
from quickwin.plugin.v1 import plugin_pb2 as pb

class MyPlugin(Plugin):
    def metadata(self):        # PHẢI khớp plugin.yaml (routes, name, version)
        return pb.Metadata(name="my", version="0.1.0", routes=[...])
    def handle_request(self, req):   # API động, xử lý theo req.path
        return pb.HttpResponse(status=200, body=b"...")
    def run_task(self, spec):        # (tuỳ chọn) generator yield TaskEvent, kết bằng result
        ...
    def health(self):                # (tuỳ chọn) mặc định HEALTHY
        ...

if __name__ == "__main__":
    serve(MyPlugin())
```

## 3. plugin.yaml
Như plugin Go nhưng `entrypoint: { python: main.py }`. Host cần `python3 ≥ 3.10` +
`quickwin_plugin` (sdk/python) + stub (proto/gen/python) trên PYTHONPATH.

## 4. Chạy / test
- Env: `PYTHONPATH=sdk/python:proto/gen/python`.
- Plugin Manager phát hiện `entrypoint.python` → chạy `python3 main.py` (manager.go `entrypointCmd`).
- Integration test: `plugin-manager/python_integration_test.go` (build tag `python_integration`);
  CI job `python-plugin` cài grpcio + gen stub + chạy — chứng minh echo trả `"lang":"python"`.

## Luật
- KHÔNG viết stub tay — luôn sinh từ `.proto` (nguồn chân lý duy nhất).
- Metadata routes khớp plugin.yaml, nếu không Plugin Manager từ chối load.
