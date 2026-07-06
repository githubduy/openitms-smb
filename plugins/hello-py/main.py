#!/usr/bin/env python3
"""Plugin Python mẫu — chứng minh hợp đồng .proto không phụ thuộc ngôn ngữ.
Được OpenITMS-SMB core (Plugin Manager) launch qua entrypoint: python."""
import json

from quickwin_plugin import Plugin, serve
from quickwin.plugin.v1 import plugin_pb2 as pb

VERSION = "0.1.0"


class HelloPy(Plugin):
    def metadata(self):
        return pb.Metadata(
            name="hello-py",
            version=VERSION,
            routes=[
                pb.Route(method="POST", path="echo", description="Echo request body back"),
                pb.Route(method="GET", path="info", description="Plugin info"),
            ],
        )

    def handle_request(self, req):
        if req.path == "echo":
            body = json.dumps({"echo": req.body.decode("utf-8", "replace"),
                               "caller": req.caller.username, "lang": "python"})
            return _json(200, body)
        if req.path == "info":
            return _json(200, json.dumps({"name": "hello-py", "version": VERSION, "lang": "python"}))
        return _json(404, json.dumps({"error": "unknown route"}))

    def run_task(self, spec):
        yield pb.TaskEvent(task_id=spec.task_id, status=pb.TASK_STATUS_RUNNING)
        yield pb.TaskEvent(task_id=spec.task_id, log_line="hello-py task running")
        yield pb.TaskEvent(task_id=spec.task_id,
                           result=pb.TaskResult(status=pb.TASK_STATUS_SUCCESS, message="done"))


def _json(status, body):
    return pb.HttpResponse(status=status,
                           headers={"Content-Type": "application/json"},
                           body=body.encode("utf-8"))


if __name__ == "__main__":
    serve(HelloPy())
