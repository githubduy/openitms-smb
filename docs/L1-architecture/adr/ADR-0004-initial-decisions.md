---
level: L1
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [upstream/, .github/workflows/]
---

# ADR-0004 — Các quyết định khởi động dự án

**Trạng thái:** approved (chốt bởi maintainer, 2026-07-05)

| # | Quyết định | Chốt | Hệ quả |
|---|---|---|---|
| 1 | Tên sản phẩm | **OpenITMS-SMB** (Open IT Management System for SMB) — chốt 2026-07-06 | Kiểm tra trước khi chốt: (a) không có project/sản phẩm active nào tên "openitms" (chỉ 1 repo GitHub bỏ hoang 0 sao, 2 commit, không license — không phải trademark); (b) naming "open"+acronym là thông lệ FOSS (tiền lệ OpenNMS); (c) **lưu ý**: "ITMS™" là mark của WidePoint (IT asset mgmt) và Broadcom có "IT Management Suite (ITMS)" — mark tổng hợp "OpenITMS-SMB" khác biệt đủ cho dự án OSS, nhưng nếu thương mại hóa tại US thì tham vấn luật sư + đăng ký domain sớm. Internal identifiers (`quickwin.dev/*` module path, env `QUICKWIN_*`, thư mục repo local, password mặc định `quickwin123` — password là yêu cầu gốc của user, giữ nguyên) đổi trong 1 commit rename khi tạo org GitHub |
| 2 | Baseline upstream | **v2.18.16** (tag stable mới nhất tại ngày init; dòng 2.19.x còn beta) | Submodule `upstream/` pin tag này; mọi patch viết trên baseline này |
| 3 | Registry public | **GitHub Pages** (static index + artifact, ký số) | CI publish lên nhánh gh-pages / Pages workflow |
| 4 | Chất lượng code | **SonarQube self-host** | Cần dựng + vận hành instance (task P4-03); CI runner phải reach được instance |
| 5 | Live USB base | **Alpine**, control từng package (danh sách tường minh, có review) | musl libc → binary bundle phải build musl-compatible/static; pwsh trên musl phải kiểm chứng sớm (P5-B) |

Bối cảnh & các phương án đã loại: xem `PLAN.md` mục 13.
