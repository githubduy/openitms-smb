---
level: L1
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [core-patches/, plugins/]
---

# Ranh giới Core vs Plugin

**Mặc định: mọi tính năng mới là PLUGIN.** Câu hỏi kiểm tra trước khi đề xuất sửa core:
1. Có làm được bằng plugin (API động + RunTask + UI hooks) không? → Nếu có: plugin.
2. Có phải hạ tầng mà plugin PHỤ THUỘC vào không (lifecycle, registry, security)? → Mới cân nhắc core.

**Ngoại lệ đã duyệt (chỉ 2):**
| Tính năng | Patch | Lý do được vào core |
|---|---|---|
| Endpoint Script Manager | 0004 | Cần tái dùng sâu task runner + inventory của Semaphore |
| Registry client | 0005 | Là hạ tầng phân phối mà mọi plugin phụ thuộc |

Muốn thêm ngoại lệ thứ 3: viết ADR draft → maintainer duyệt. AI agent KHÔNG tự quyết (luật #7).
Patch hạ tầng (0001 hook, 0002 branding, 0003 config) không tính là "tính năng".
