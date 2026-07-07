# OpenITMS-SMB

**Quản lý toàn bộ máy tính Windows & Linux của công ty từ một trang web duy nhất.**

OpenITMS-SMB là phần mềm mã nguồn mở giúp doanh nghiệp nhỏ và vừa (SMB) tự động hoá công việc IT:
cài phần mềm, cập nhật, khởi động lại dịch vụ, chạy lệnh/script xuống nhiều máy cùng lúc — tất cả
qua giao diện web, không cần gõ lệnh trên từng máy.

> **Trạng thái: đang hoàn thiện (pre-release).** Bản cài đóng gói sẵn đang được chuẩn bị. Hiện tại
> bạn có thể tự dựng bản cài từ mã nguồn — xem [DEVELOPMENT.md](DEVELOPMENT.md).

---

## Làm được gì?

- 🖥️ **Quản lý máy Windows từ xa** — chạy PowerShell xuống máy Windows qua WinRM (chứng chỉ,
  không cần mật khẩu), hoặc SSH. Hỗ trợ cả máy Linux.
- ⚡ **WinRS Console** — gõ nhanh 1 lệnh xuống 1 máy và xem kết quả ngay, không cần tạo gì.
- 📋 **Mẫu tác vụ (Task Templates)** — định nghĩa "việc cần chạy" một lần, rồi bấm nút chạy lại
  bất cứ lúc nào, xuống nhiều máy.
- ⏰ **Hẹn giờ tự động** — chạy tác vụ theo lịch (sao lưu hằng đêm, cập nhật hằng tuần…).
- 🚀 **Thêm máy chỉ 1 lần bấm** — tải script cài đặt, chạy trên máy đích; chứng chỉ tự động được
  nạp lên và máy tự xuất hiện trong danh sách.
- 📦 **Kèm sẵn mọi thứ** — cơ sở dữ liệu (MariaDB), git server nội bộ (Gitea), PowerShell.
  Không cần Docker, không cần cài thêm gì.
- 🔒 **An toàn mặc định** — mật khẩu/khoá được mã hoá, cảnh báo khi còn dùng cấu hình mặc định.

---

## Cài đặt

Cần một máy chủ (server) để chạy OpenITMS — có thể là 1 máy Windows hoặc Linux trong mạng nội bộ.
Người dùng khác truy cập qua trình duyệt.

### Linux

```bash
# 1. Tải và giải nén bản cài
tar -xzf openitms-smb-*-linux-amd64.tar.gz
cd openitms-smb-*-linux-amd64

# 2. Cài (1 lệnh, cần quyền root)
sudo ./install.sh
```

Xong. Trình cài đặt tự lo mọi thứ (database, service, admin, git server). Cuối màn hình sẽ in ra
địa chỉ truy cập. Yêu cầu: Linux 64-bit có `systemd`. Không cần internet.

- Gỡ cài: `sudo ./uninstall.sh`
- Service: `systemctl status openitms openitms-db`

### Windows

Mở **PowerShell với quyền Administrator** (chuột phải → *Run as administrator*):

```powershell
# 1. Giải nén bản cài, vào thư mục vừa giải nén
cd openitms-smb-<phiên-bản>-windows

# 2. Cho phép chạy script + cài (1 lệnh)
Set-ExecutionPolicy -Scope Process Bypass -Force
.\installer\windows\install.ps1
```

Trình cài đặt tự khởi tạo database, tạo tài khoản quản trị, và đăng ký chạy nền lúc khởi động máy.
Cuối màn hình in ra địa chỉ truy cập. Yêu cầu: Windows 10/11 hoặc Windows Server 64-bit.

- Gỡ cài: `.\installer\windows\uninstall.ps1` (thêm `-PurgeData` để xoá cả dữ liệu)
- Trạng thái: `Get-Service OpenITMS-DB` và `Get-ScheduledTask OpenITMS`

---

## Bắt đầu sử dụng

1. Mở trình duyệt tới địa chỉ in ra khi cài xong (ví dụ `http://<ip-máy-chủ>:3000`).
2. Đăng nhập bằng tài khoản mặc định **`admin` / `quickwin123`** — **đổi mật khẩu ngay** ở lần đầu.
3. Đã có sẵn project **"Host"** quản lý chính máy chủ. Vào **WinRS Console** để thử.
4. **Thêm máy Windows cần quản lý:** WinRS Console → *Enroll this machine (1-click)* → tải script →
   chạy trên máy đó bằng PowerShell (Administrator). Máy sẽ tự xuất hiện trong Inventory.

> 💡 Rê chuột vào từng mục menu để xem giải thích ngắn gọn mục đó dùng làm gì.

---

## Bảo mật cần biết

- Tài khoản mặc định `admin / quickwin123` và mật khẩu database mặc định là **known default** để cài
  nhanh — hãy đổi mật khẩu admin ngay, và cân nhắc đổi mật khẩu DB (biến môi trường khi cài).
- Cơ sở dữ liệu chỉ nghe nội bộ (localhost), không mở ra internet.
- Chứng chỉ để tại thư mục `certs/` của máy chủ; khoá/mật khẩu trong app được mã hoá.

---

## Tài liệu & đóng góp

- 🛠️ **Tự dựng từ mã nguồn / phát triển:** [DEVELOPMENT.md](DEVELOPMENT.md)
- 📚 Tài liệu chi tiết: [docs/](docs/)
- 🤝 Đóng góp: [CONTRIBUTING.md](CONTRIBUTING.md) · Quy tắc: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

---

## Giấy phép

MIT — [LICENSE](LICENSE). Sản phẩm là bản fork của
[Semaphore UI](https://github.com/semaphoreui/semaphore) © Denis Gukov / Castaway Labs LLC (MIT) —
[LICENSE-SEMAPHORE](LICENSE-SEMAPHORE), [NOTICE.md](NOTICE.md). Không liên kết và không được chứng
thực bởi Semaphore UI.
