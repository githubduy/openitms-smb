---
name: dev-core-patch
description: >
  Thực hiện thay đổi CORE của OpenITMS-SMB đúng quy trình bộ-4. Trigger khi task có label
  area:core-patch, hoặc khi cần sửa hành vi core Semaphore. Enforce: cấm sửa upstream/ trực tiếp;
  ưu tiên plugin trước; mỗi patch phải có patch+series+CHANGELOG+spec L2.
---

# Skill: dev-core-patch

Bản thực thi của playbook 5A (guideline). ĐỌC `docs/L3-development/AI-ENGINEER-GUIDELINE.md` trước.

## Bước
1. **Chặn trước tiên:** hỏi "làm bằng plugin được không?". Nếu được → dừng, đề xuất chuyển area:plugin.
   Chỉ 2 ngoại lệ core được duyệt (Endpoint Script Manager, Registry client); thêm ngoại lệ → cần ADR.
2. `scripts/reset-upstream.sh` → sửa code trong `upstream/` (working tree) theo kỹ thuật **hook mỏng**
   (chèn tối thiểu, logic thật ở package ngoài cây upstream).
3. `git -C upstream add -A && scripts/export-patch.sh <NNNN-ten>` → điền header `# WHY:` (bắt buộc —
   để rebase khi upstream đổi) + `# WHAT:`.
4. **Bộ-4 bắt buộc:** thêm dòng vào `core-patches/series` + entry `core-patches/CHANGELOG.md`
   + spec `docs/L2-specs/core-patches/<NNNN-ten>.md` (mục đích, thay đổi, hướng dẫn rebase, verify).
5. `scripts/reset-upstream.sh && scripts/apply-patches.sh && scripts/build-all.sh` → PHẢI xanh.
   Thêm E2E nếu chạm hành vi (tham khảo tests/e2e/plugin-through-core.sh).
6. Commit Conventional + DCO (`-s`); branch `ai/<issue>-<slug>`. KHÔNG merge, KHÔNG push main.

## Cấm
- Sửa file trong `upstream/` rồi commit thẳng (phải là patch).
- Patch thiếu 1 trong bộ-4 → CI patch-hygiene fail.
