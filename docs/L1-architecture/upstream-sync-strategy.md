---
level: L1
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [upstream/, core-patches/, scripts/sync-upstream.sh, scripts/apply-patches.sh]
---

# Chiến lược đồng bộ Upstream

**Mục tiêu:** nhận bản vá bảo mật Semaphore trong ≤ 48h, không conflict-hell.

## Cơ chế
- `upstream/` = submodule pin tag baseline (hiện **v2.18.16**), CẤM sửa tay.
- Mọi thay đổi = patch trong `core-patches/` (bộ-4), apply lúc build bằng `scripts/apply-patches.sh`.
- Header patch bắt buộc có `# WHY:` — là ngữ cảnh để AI rebase giữ Ý ĐỊNH khi upstream đổi code.

## Quy trình khi upstream release (scripts/sync-upstream.sh <tag>)
1. Fetch + checkout tag mới → 2. Apply series → 3. Build → 4. Test.
Patch fail → skill `sync-upstream` (AI) đọc WHY + spec L2 → rebase → PR liệt kê:
patch giữ nguyên / patch rebase / diff hành vi (nếu có). Rebase làm ĐỔI HÀNH VI → dừng, hỏi maintainer.

## Quy tắc viết patch ít conflict
1 patch = 1 mục đích, càng nhỏ càng tốt; kỹ thuật "hook mỏng" — chèn vài dòng gọi ra package
ngoài cây upstream, logic thật nằm ngoài; branding ưu tiên asset-replace/build-flag.
