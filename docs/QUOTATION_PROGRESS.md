# Sales Quotation — progress tracker

**This is the one document to follow.** Everything else in `docs/` is either a
frozen reference or a plan written before the work; none of them track status.
If you are picking the project up, read this file, then the guide.

**Last verified:** 2026-09-10, against the running staging environment.

---

## 1. Where things stand

| | Staging | Production |
|---|---|---|
| Backend | `2.31.0-staging.10` | `2.30.0` |
| Frontend | `2.46.0-staging.17` | `2.45.0` |
| Migration | 021 | **014** |

**Nothing is in production, and nothing is pushed to any git remote.** Production
still runs migration 014, so none of the quotation tables exist there.

---

## 2. Phases

| Phase | What it covers | Status |
|---|---|---|
| **0** | Roles, permission layer, user management UI | ✅ Done, on staging |
| **1** | Customers, brands, sales assignments + spreadsheet import | ✅ Done, on staging |
| **2** | Prices: per-week building prices on a Prices page with upload preview, package prices on the package. Replaced versioned rate cards 2026-09-10 | ✅ Done, on staging |
| **3** | Quotation core: wizard, server-side pricing, submit | ✅ Done, on staging |
| **4** | Approvals: routing, approve/return, audit trail | ⚠️ **Almost** — see §3.1 |
| **5** | Printed document | ✅ Rebuilt to the template in landscape; logo files still needed — see §3.2 |
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

### 3.2 ✅ The printed document — rebuilt, three items left

Rebuilt in **A4 landscape** on 2026-09-10 to match the template: CHECKER panel,
bordered party box, two tinted line-item tables sharing a colgroup so the TOTAL row
lines up, IDR-prefixed money, and the blue totals box. Verified by rendering real
PDFs of both a placement-only and a placement-plus-bonus quotation and comparing
against the source.

What remains, none of it a layout problem:

| Item | Why it is open |
|---|---|
| **Logos** | The TMN, Focus Media Group and Sinarmas marks have **no asset files in this project** — `public/logo.png` is the admin template's mark and `tmn-logo-small.png` is a TapOn logo. The letterhead leaves their space. **Needs the files from the business.** |
| First column | Prints a building count rather than a screen count, because screen count is still unverified — §4.2 of the analysis |
| Sales team / phone, finance PIC | Print as labelled blanks. Not stored, and the source template prints them blank too |

Also fixed while verifying: a **draft** printed a complete-looking document with
IDR 0 everywhere, because pricing is computed on submit, not on save. Drafts now
carry a red "NOT YET PRICED — do not send to a client" banner.

### 3.3 Data gaps that block features

| Gap | Blocks |
|---|---|
| **Real package prices not loaded** | Packages carry their own price since migration 020, set on the package form, and package mode works end to end. Only the test package SP-0018 has one |
| **Audience figures (traffic / impressions) are hidden** | Removed from display 2026-09-10 because ERP has them for only 86 of 1,581 sellable buildings, so every quotation printed zeros. Columns, API and summing all kept. **Option A is the agreed next step if users start uploading the figures** — see `QUOTATION_DOCUMENT_ANALYSIS.md` §4.3.1 |
| **`screen_count` may not be a real requirement** | Nothing confirms it. Not in the reference prototype; the template's "Spot/Day/Screen" column is satisfied by the spots the salesperson types. It rests on an unlabelled "950" I inferred was screens. Meanwhile building mode counts one screen per building, which the rate card shows undercounts by 79% — so the printed document carries a number we cannot justify. **One question to the business settles it.** See `QUOTATION_DOCUMENT_ANALYSIS.md` §4.2 |
| ~~The sales package form exposes only `Name`~~ | ✅ Fixed 2026-09-09. The form could not save at all — the API required `package_code` and `status`, which it never sent, so every create and update returned 400. All six fields are now there |
| **PIC Finance, sales team/phone not stored** | Two fields the printed template has (§9 items 10–11). A schema decision, not CSS |

### 3.4 Smaller, known, not urgent

- ~~No manual "Add price" on a rate card; a published card hid its edit controls
  unexplained~~ — ✅ gone with the Rate Cards page. The Prices page adds, edits and
  removes by hand.
- **Remove the dormant rate card backend.** The `rate_card_*` tables and `/rate-cards*`
  endpoints are unused since 2026-09-10 but still present, deliberately kept for one
  release as history. Drop them in a follow-up migration.
- **`sales_group`** is stored, editable and returned by the API, and read by nothing.
  Decide: wire it up or drop it.
- **Building contract import** — planned in `BUILDING_CONTRACT_IMPORT_PLAN.md`,
  deliberately off the critical path.

### 3.5 Phase 6, untouched

No notifications of any kind (there is no mail code in the backend at all), no
server-side PDF, no localization — the document is English only.

---

## 4. Before production

1. Migrations 014 → 021 on the production database. 014 → 019 was rehearsed against
   a copy of real production data (3,762 buildings, 19 users) and applied cleanly;
   **020 and 021 have only been applied to staging** and need the same rehearsal.
2. Load real customers, brands and sales assignments.
3. Upload real building prices on the Prices page, and set a price on every package
   that should be quotable (§3.3).
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
