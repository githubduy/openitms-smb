"""Handshake go-plugin + gRPC server cho plugin Python."""
import abc
import os
import sys
import socket
from concurrent import futures

import grpc

# Stub sinh bởi scripts/gen-proto-py.sh vào cùng package (gen/python trên PYTHONPATH).
from quickwin.plugin.v1 import plugin_pb2 as pb
from quickwin.plugin.v1 import plugin_pb2_grpc as pb_grpc

PROTOCOL_VERSION = 1  # APP-PROTOCOL-VERSION — khớp sdk/go handshake.go
_MAGIC_COOKIE_KEY = "QUICKWIN_PLUGIN"
_MAGIC_COOKIE_VALUE = "quickwin-plugin-v1-2f8a4e"


class Plugin(abc.ABC):
    """Plugin Python implement 4 method (đối xứng sdk.Plugin của Go)."""

    @abc.abstractmethod
    def metadata(self) -> pb.Metadata:
        """Trả Metadata — PHẢI khớp plugin.yaml (core đối chiếu, lệch → từ chối load)."""

    @abc.abstractmethod
    def handle_request(self, req: pb.HttpRequest) -> pb.HttpResponse:
        """Xử lý 1 request từ API động /api/plugins/<name>/<route>."""

    def run_task(self, spec: pb.TaskSpec):
        """Generator yield TaskEvent (log_line/status), return TaskResult cuối.
        Mặc định: không hỗ trợ task."""
        yield pb.TaskEvent(task_id=spec.task_id,
                           result=pb.TaskResult(status=pb.TASK_STATUS_FAILED,
                                                message="plugin không hỗ trợ RunTask"))

    def health(self) -> pb.Health:
        return pb.Health(status=pb.Health.STATUS_HEALTHY)


class _Servicer(pb_grpc.PluginServicer):
    def __init__(self, impl: Plugin):
        self._impl = impl

    def GetMetadata(self, request, context):
        return self._impl.metadata()

    def HandleRequest(self, request, context):
        return self._impl.handle_request(request)

    def HealthCheck(self, request, context):
        return self._impl.health()

    def RunTask(self, request, context):
        emitted_result = False
        for ev in self._impl.run_task(request):
            if ev.HasField("result"):
                emitted_result = True
            yield ev
        if not emitted_result:
            # hợp đồng: event cuối phải là result (proto-contract.md)
            yield pb.TaskEvent(task_id=request.task_id,
                               result=pb.TaskResult(status=pb.TASK_STATUS_SUCCESS))


def _check_magic_cookie():
    if os.environ.get(_MAGIC_COOKIE_KEY) != _MAGIC_COOKIE_VALUE:
        sys.stderr.write("plugin này phải được OpenITMS-SMB core khởi chạy (thiếu magic cookie)\n")
        sys.exit(1)


def serve(impl: Plugin):
    """Điểm vào plugin Python. Gọi trong __main__:  serve(MyPlugin())."""
    _check_magic_cookie()

    # bind cổng động trên localhost
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.bind(("127.0.0.1", 0))
    port = sock.getsockname()[1]
    sock.close()

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=8))
    pb_grpc.add_PluginServicer_to_server(_Servicer(impl), server)
    server.add_insecure_port(f"127.0.0.1:{port}")
    server.start()

    # dòng handshake go-plugin — PHẢI in ra stdout rồi flush
    sys.stdout.write(f"{1}|{PROTOCOL_VERSION}|tcp|127.0.0.1:{port}|grpc\n")
    sys.stdout.flush()

    server.wait_for_termination()
