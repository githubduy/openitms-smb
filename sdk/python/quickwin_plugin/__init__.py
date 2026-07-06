"""quickwin-plugin-sdk-py — viết plugin OpenITMS-SMB bằng Python.

Cùng file proto (proto/quickwin/plugin/v1/plugin.proto) làm nguồn chân lý — stub
sinh bằng scripts/gen-proto-py.sh (grpcio-tools). Plugin chỉ cần implement class
Plugin rồi gọi serve(); SDK lo handshake HashiCorp go-plugin + gRPC server.

Handshake go-plugin (in ra stdout, đúng 1 dòng):
    CORE-PROTOCOL-VERSION | APP-PROTOCOL-VERSION | network | address | protocol
    1|1|tcp|127.0.0.1:<port>|grpc
Magic cookie: env QUICKWIN_PLUGIN phải = giá trị SDK Go (handshake.go) — core set khi launch.
"""
from .plugin import Plugin, serve, PROTOCOL_VERSION

__all__ = ["Plugin", "serve", "PROTOCOL_VERSION"]
