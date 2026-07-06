# installer/vendor/ — stage binary dependency LOCAL (gitignored)

fetch-deps.sh COPY binary từ đây thay vì tải mạng (air-gapped / mạng chặn external).
Binary lớn → KHÔNG commit (gitignored). Maintainer stage thủ công trước khi package.

## MariaDB (bundle kèm bản cài — GPLv2, socket-only aggregation)
Đặt 1 trong 2:
- Thư mục đã giải nén: `installer/vendor/mariadb/` (chứa bin/, scripts/, share/...)
- Hoặc archive: `installer/vendor/mariadb.tar.gz` (sha256 phải khớp deps.lock)
Nguồn: copy từ máy có MariaDB, hoặc tải sẵn `mariadb-11.4.4-linux-systemd-x86_64.tar.gz`.

## pwsh (PowerShell Core — MIT)
- `installer/vendor/pwsh/` hoặc `installer/vendor/pwsh.tar.gz`.

Hoặc trỏ env: `DEPS_MARIADB_DIR=/path/to/mariadb DEPS_PWSH_DIR=/path/to/pwsh installer/fetch-deps.sh`
