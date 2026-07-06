---
name: sync-upstream
description: >
  Đồng bộ baseline lên tag Semaphore mới, rebase core-patches khi conflict. Trigger khi có
  release upstream mới hoặc task sync. Enforce: rebase giữ Ý ĐỊNH (header WHY) không giữ từng dòng;
  đổi hành vi → dừng hỏi maintainer; SLA vá bảo mật ≤ 48h.
---

# Skill: sync-upstream

Bản thực thi playbook 5F + runbook. Script `scripts/sync-upstream.sh`.

## Bước
1. `scripts/sync-upstream.sh <new-tag>` — fetch + checkout + apply series + build + test.
2. Patch conflict → đọc `# WHY:` header của patch + spec L2 của nó → rebase GIỮ Ý ĐỊNH
   (tìm lại điểm hook tương đương, không dán nguyên dòng cũ). go.sum conflict → `go mod tidy`.
3. Sau rebase: full build + test + SonarQube. Nếu rebase làm **ĐỔI HÀNH VI** → DỪNG, mở issue hỏi maintainer.
4. PR sync liệt kê: patch giữ nguyên / patch rebase / diff hành vi (nếu có) + bump baseline (submodule + ADR-0004).
5. SLA: bản vá bảo mật upstream → PR trong ≤ 48h.

## Cấm
- Tự merge PR sync. Bỏ qua test sau rebase.
