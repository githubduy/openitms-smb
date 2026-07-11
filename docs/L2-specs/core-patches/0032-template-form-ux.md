---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0032-template-form-ux.patch, core-patches/0015-menu-tooltips.patch]
---

# Spec patch 0032 — Template form: mặc định repo local + tooltip Task/Build/Deploy

## Mục đích (WHY)
- Khi tạo Task Template, ô Repository để trống → user (SMB, thường chỉ có 1 repo local Gitea) phải tự
  chọn. Nên **mặc định chọn repo local** cho nhanh.
- 3 tab loại template **Task / Build / Deploy** không rõ nghĩa với người không chuyên → cần **tooltip**.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `web/src/components/TemplateForm.vue` | sửa | `afterLoadData`: nếu tạo mới + chưa chọn repo → set `item.repository_id` = repo tên chứa "Gitea", hoặc repo duy nhất. Thêm `:title` + method `templateTypeTooltip(key)` cho 3 v-tab type |
| `web/src/lang/en.js`, `vi.js` | +key | `tooltipTemplateTask/Build/Deploy` |

## Chi tiết
- **Default repo**: chỉ áp dụng khi `isNew && !item.repository_id`. Ưu tiên repo local (tên chứa
  "Gitea" — khớp quy ước 0008/0012); nếu không có nhưng chỉ có 1 repo → chọn repo đó. Copy template
  (sourceItemId) đã có repo_id nên không đụng.
- **Tooltip tab**: dùng `:title` (native) thay vì `v-tooltip` vì `v-tabs` yêu cầu `v-tab` là con trực
  tiếp — bọc tooltip dễ vỡ chỉ số tab. `:title` an toàn tuyệt đối.

## Rebase
- Neo: ô `v-autocomplete v-model="item.repository_id"` và vòng `v-for` render `v-tab` type trong
  `TemplateForm.vue`. Upstream đổi cấu trúc → gắn lại default + `:title`.

## Verify
- Tạo Task Template mới trong project có repo "Local (Gitea)" → ô Repository tự chọn sẵn nó.
- Hover 3 tab Task/Build/Deploy → hiện mô tả. eslint sạch; chain 0001–0032 build.

## Liên quan
- Tooltip menu: [[0015-menu-tooltips]].
