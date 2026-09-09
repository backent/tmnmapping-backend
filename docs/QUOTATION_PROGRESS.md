# Sales Quotation — progress tracker

**This is the one document to follow.** Everything else in `docs/` is either a
frozen reference or a plan written before the work; none of them track status.
If you are picking the project up, read this file, then the guide.

**Last verified:** 2026-09-09, against the running staging environment.

---

## 1. Where things stand

| | Staging | Production |
|---|---|---|
| Backend | `2.31.0-staging.7` | `2.30.0` |
| Frontend | `2.46.0-staging.9` | `2.45.0` |
| Migration | 019 | **014** |

**Nothing is in production, and nothing is pushed to any git remote.** Production
still runs migration 014, so none of the quotation tables exist there.

---

## 2. Phases

| Phase | What it covers | Status |
|---|---|---|
| **0** | Roles, permission layer, user management UI | ✅ Done, on staging |
| **1** | Customers, brands, sales assignments + spreadsheet import | ✅ Done, on staging |
| **2** | Rate card: versions, per-week building and package prices, publish | ✅ Done, on staging |
| **3** | Quotation core: wizard, server-side pricing, submit | ✅ Done, on staging |
| **4** | Approvals: routing, approve/return, audit trail | ⚠️ **Almost** — see §3.1 |
| **5** | Printed document | ⚠️ **Works, wrong layout** — see §3.2 |
| **6** | Notifications, server-side PDF, localization, dashboards | ❌ Not started |

Phases are not a formal plan; they are how the work was sequenced. The numbered
sections inside `QUOTATION_FEATURE_ANALYSIS.md` (§4.1 Migrations, §4.2 Models …)
are a **different** numbering — that whole section is complete and absorbed into
Phases 0–3. Do not read "§4.1" there as "the next task".

---

## 3. What is actually left

### 3.1 Version history is written but never read — **the one real Phase 4 gap**

Every submit marshals a full JSON snapshot into `quotation_versions`. Nothing reads
it back: no repository method, no service method, no endpoint, no UI. The audit
trail shows that a v1 existed and was returned, but not **what v1 said**.

Needs: a repository `FindVersions`, a service method, one endpoint, one panel on the
detail page. Self-contained.

### 3.2 The printed document does not match the template

Content and arithmetic are right; the layout is not. The source is A4 **landscape**;
ours is a portrait reflow. **Deferred by the user on 2026-09-09** — do not start it
unprompted. Full twelve-item gap list: `QUOTATION_DOCUMENT_ANALYSIS.md` §9.

### 3.3 Data gaps that block features

| Gap | Blocks |
|---|---|
| **No package prices in any rate card** (all three versions have 0) | Package mode is unusable in the wizard — it says "No package has a price in the published rate card yet" |
| **The sales package form exposes only `Name`** | `screen_count`, `traffic`, `impressions`, `package_code`, `status` and `description` exist in the schema (migration 018) and are read by the quotation service, but cannot be entered through the UI. All 10 staging packages are 0, so a package quotation would print "0 screens" and no audience. See `QUOTATION_DOCUMENT_ANALYSIS.md` §4.1 |
| **PIC Finance, sales team/phone not stored** | Two fields the printed template has (§9 items 10–11). A schema decision, not CSS |

### 3.4 Smaller, known, not urgent

- No manual **"Add price"** on a rate card — import only, though the API already
  upserts, so it is a missing button rather than a missing capability.
- A **published** rate card hides its edit controls with no explanation of why, which
  reads as a bug. Needs a line of copy.
- **`sales_group`** is stored, editable and returned by the API, and read by nothing.
  Decide: wire it up or drop it.
- **Building contract import** — planned in `BUILDING_CONTRACT_IMPORT_PLAN.md`,
  deliberately off the critical path.

### 3.5 Phase 6, untouched

No notifications of any kind (there is no mail code in the backend at all), no
server-side PDF, no localization — the document is English only.

---

## 4. Before production

1. Migration 014 → 019 on the production database. Rehearsed against a copy of real
   production data (3,762 buildings, 19 users) and it applied cleanly.
2. Load real customers, brands and sales assignments.
3. Publish a real rate card **including package prices** (§3.3).
4. Assign the approver roles — exactly **one** user per role, or submit fails loudly
   by design.
5. Decide the `can_create_quotations` default for real users. Admin is exempt.

---

## 5. The other documents

| Doc | Kind | Read it when |
|---|---|---|
| [`QUOTATION_FEATURE_GUIDE.md`](QUOTATION_FEATURE_GUIDE.md) | Explanation | You want to understand or use the feature. §3 is the how-to |
| [`QUOTATION_FEATURE_ANALYSIS.md`](QUOTATION_FEATURE_ANALYSIS.md) | Reference, frozen 2026-08-25 | You want the original reasoning vs the prototype |
| [`QUOTATION_DOCUMENT_ANALYSIS.md`](QUOTATION_DOCUMENT_ANALYSIS.md) | Reference + §9 spec | You are working on the printed document |
| [`PHASE_0_ROLES_PROGRESS.md`](PHASE_0_ROLES_PROGRESS.md) | Historical, Phase 0 only | You want the roles/permission reasoning |
| [`PHASE_1_IMPORT_PLAN.md`](PHASE_1_IMPORT_PLAN.md) | Plan, now built | You are changing the import format |
| [`BUILDING_CONTRACT_IMPORT_PLAN.md`](BUILDING_CONTRACT_IMPORT_PLAN.md) | Plan, not built | You pick up building contract import |
| [`STAGING_ENVIRONMENT_PLAN.md`](STAGING_ENVIRONMENT_PLAN.md) | Plan, superseded by reality | Historical only — staging is live; see §1 |
