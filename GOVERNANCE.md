# Quản trị dự án

- **Maintainer:** người có quyền merge vào `main` và tạo tag release (ký số).
  Danh sách: *(bổ sung khi lập org GitHub — tối thiểu 2 admin, bật 2FA bắt buộc)*.
- **AI agents:** được mở PR trên branch `ai/*`, KHÔNG có quyền merge/tag/release/secret.
- **Thêm maintainer:** đề cử bởi 1 maintainer + đồng thuận của các maintainer còn lại,
  sau ≥ 3 tháng đóng góp đều.
- **Quyết định kiến trúc:** qua ADR (docs/L1-architecture/adr/) — chỉ maintainer approve.
- **Khóa ký registry/release:** chỉ maintainer giữ; CI dùng qua secrets; AI không bao giờ tiếp cận.
- Ứng xử: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
