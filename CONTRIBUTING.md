# Đóng góp

1. **Đọc trước:** [docs/L3-development/AI-ENGINEER-GUIDELINE.md](docs/L3-development/AI-ENGINEER-GUIDELINE.md)
   — luật cứng, bản đồ repo, playbook theo loại task, Definition of Done. Áp dụng cho cả người lẫn AI.
2. **Nạp yêu cầu:** GitHub Issues (form Feature / Bug / Plugin proposal). Issue đủ thông tin
   có thể được AI agent tự phát triển.
3. **Quy tắc nhanh:** không sửa `upstream/`; thay đổi core = bộ-4 trong `core-patches/`;
   tính năng mới = plugin; Conventional Commits; PR cần CI xanh + 1 review người.
4. **DCO:** mọi commit cần `Signed-off-by` (`git commit -s`) — xác nhận bạn có quyền đóng góp
   code theo MIT.
5. Branch: người dùng `feat/* fix/* docs/* chore/*`; AI agent dùng `ai/*`.
