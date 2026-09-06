# Sales Quotation Feature — Gap Analysis for TMN Mapping

**Status:** Analysis only — no code changes made.
**Date:** 2026-08-25
**Reference implementation:** `sales_quotation_example/` (Next.js + Drizzle/PostgreSQL prototype)
**Target:** `backend/` (Go, httprouter + manual DI via wire) and `frontend/` (Vue 3 + Vuetify + Pinia)

---

## 1. What the reference app actually does

The example is a **quote-to-approval** workflow. Read in dependency order, its authoritative specs are:

| Spec | What it establishes |
|---|---|
| `docs/superpowers/specs/2026-07-10-sales-quotation-prototype-design.md` | Original flow: 6-step wizard, roles, statuses, quotation document |
| `docs/superpowers/specs/2026-07-11-stage-2-data-import-design.md` | Master data import (customer/brand, building, package, rate card) |
| `docs/superpowers/specs/2026-07-13-placement-bonus-direct-approval-design.md` | **Supersedes** the original approval model — Placement/Bonus split, single-step approval |
| `docs/superpowers/specs/2026-07-23-final-approval-proxy-entry-design.md` | Final routing thresholds, no-self-approval, proxy entry |

> ⚠️ The first spec describes a *cascading* manager → CEO approval at a 70% threshold. That model was **replaced**. The current, authoritative model is single-step direct routing at 65% / 75% (spec 2026-07-13 + 2026-07-23). Build against the later specs.

### 1.1 End-to-end flow

```
Login (role-based workspace)
   │
   ├─ Sales dashboard  ── counts: returned / pending / approved / all
   │
   └─ New quotation wizard (6 steps)
        1. [proxy only] choose commercial owner  →  choose customer  →  choose brand
                (only customers/brands assigned to that sales owner are listed)
        2. Placement: mode = individual buildings (1..n) OR one sales package
        3. Bonus: "No Bonus" OR an independent selection (own mode, own resources)
        4. Campaign parameters, separately for Placement and Bonus:
                TVC duration (seconds), weeks, spots
        5. Discount (%) → live pricing summary + "who will approve this" hint
        6. Review → Submit
   │
   ├─ Routing resolved at submit time (single approver, no cascade)
   │
   ├─ Approver workspace → Approve  |  Return (reason mandatory)
   │      Return → status `returned` → sales edits → resubmit = new version snapshot
   │
   └─ Approved → version locked → formal Quotation document (print/preview)
```

### 1.2 Pricing model (`lib/quotation.ts`)

```
resourceGross      = Σ(resource.priceIdr for selected buildings | package price)
selection.grossPrice = round(resourceGross × weeks / 4)      # rate card is per-4-weeks

placementDiscountAmount = round(placementGross × discount/100)
placementNet            = placementGross − placementDiscountAmount
bonusNet                = 0                                   # bonus is always free
totalGross              = placementGross + bonusGross
totalNet                = placementNet
effectiveDiscountRate   = (totalGross − totalNet) / totalGross × 100
tax                     = round(totalNet × taxRate)           # prototype uses 0.06
totalIncludingTax       = totalNet + tax
```

Key rules:
- All money is **integer IDR** (`Number.isSafeInteger` asserted everywhere). No floats, no cents.
- The discount applies to **Placement only**. Bonus inflates `totalGross`, so it drags `effectiveDiscountRate` up.
- **`effectiveDiscountRate` is an analytics metric only.** Routing uses the *sales-entered customer discount*.
- The server recomputes `grossPrice`, `traffic`, and `impressions` from the referenced resources and rejects the submission on mismatch (`validateQuoteReferences`). The client is never trusted for money.

### 1.3 Approval routing (`resolveApprovalRoute`)

| Customer discount | Approver role | Status |
|---|---|---|
| `≤ 65%` | Head of Sales | `pending_manager` |
| `> 65%` and `≤ 75%` | Head of Business Control | `pending_business_control` |
| `> 75%` | CEO | `pending_ceo` |

Plus:
- **No self-approval:** if the resolved approver *is* the commercial owner of the quote, it escalates one level (Head of Sales → Head of Business Control).
- Statuses: `draft`, `pending_manager`, `pending_business_control`, `pending_ceo`, `returned`, `approved`.
- Approval actions logged as events: `submitted`, `resubmitted`, `approved`, `returned`. `returned` **requires** a non-empty comment.
- Resubmit increments `version` and appends an immutable `QuoteVersionSnapshot` (customer, brand, both selections, full pricing, discount, approver, timestamp).

### 1.4 Two identities per quote

- `salesId` — the **commercial owner**. Drives customer/brand visibility, routing, reporting.
- `createdById` — who physically entered it (proxy entry). Immutable. Audit only.

Enabled by three user attributes: `canCreateQuotations`, `canCreateOnBehalfOfSalesIds[]`, `salesGroup` (`sales_team` | `everyone_sales` | `freelancer`).

---

## 2. Where TMN stands today

### 2.1 What we already have and can reuse

| Need | TMN today | Verdict |
|---|---|---|
| Buildings master | `buildings` table + full CRUD + ERP sync (`backend/models/building.go`) | ✅ Reuse |
| Building traffic | `buildings.audience` | ✅ Maps to `traffic` |
| Building impressions | `buildings.impression` | ✅ Maps to `impressions` |
| Building location/type | `subdistrict`, `citytown`, `province`, `cbd_area`, `building_type`, `grade_resource` | ✅ Richer than the example |
| Sales packages | `sales_packages` + `sales_package_buildings` junction (`004_create_sales_packages_tables.up.sql`) | ✅ Structure reusable, needs pricing |
| Sellability filter | `buildings.sellable`, `buildings.lcd_presence_status` | ✅ Use to gate quotable inventory |
| Auth | JWT cookie + `RequireAuth` middleware (`backend/middlewares/auth.go`) | ✅ Reuse |
| CRUD layering | controller → service → repository → web req/resp, DI in `injector/wire_gen.go` | ✅ Pattern to follow |
| Import/export | XLSX import/export already implemented per entity | ✅ Reuse for rate card |
| Frontend shell | Vuetify admin layout, Pinia stores, `src/http/*` API clients, list+form page pairs | ✅ Reuse |

### 2.2 What does not exist at all

| Missing concept | Consequence |
|---|---|
| **Customer** (advertiser) | Nothing to quote *for* |
| **Brand** (advertiser brand) | Step 1 of the wizard has no data source |
| **Sales assignment** (customer/brand → sales PIC) | "only my customers" rule cannot be enforced |
| **Rate card / any price** | Zero prices anywhere in the schema — no gross can be computed |
| **Quotation** + versions + approval events | The entire feature |
| **Role authorization** | `users.role` exists but is never checked (see §5) |

### 2.3 ⚠️ `mother_brands` is *not* the brand you need

`backend/models/masterdata.go:61` defines `MotherBrand`, and `pois.mother_brand_id` references it. Confirmed from `010_revamp_pois_table.up.sql` and `models/poi.go`: **`pois.brand` / `mother_brands` are retail-location brands used for competitor & proximity mapping** (the POI layer on the map), not advertiser customers.

Do not overload them. The quotation feature needs a separate `customers` → `brands` hierarchy. Reusing `mother_brands` would couple the POI map layer to sales data and break both.

---

## 3. Building data — what needs to change

You asked specifically about building data. Summary: **the building table itself is nearly complete; the gap is pricing and package metadata.**

### 3.1 `buildings` — small additions

| Field | Status | Action |
|---|---|---|
| `audience` | exists | Reuse as `traffic`. **Verify semantics** — is it daily traffic? The quotation document prints "Daily Traffic". |
| `impression` | exists | Reuse. **Verify** — the document prints "Monthly Impressions". |
| `sellable` | exists (`VARCHAR(20)`) | Decide the exact values that make a building quotable and filter the selector on it |
| `lcd_presence_status` | exists | Decide whether non-installed buildings are quotable |
| `address` | ❌ missing | Optional — the quotation appendix prints region/type, which we already have |
| price | ❌ missing | **Do not add a price column to `buildings`.** Price is rate-card-versioned (§3.3) |

**Data quality risk:** `audience` and `impression` are `INTEGER` and nullable, and are populated by ERP sync. If a large share of rows are `NULL`/`0`, the quotation's traffic and impression totals will be wrong and the appendix will look broken. **Action: run a coverage check before committing to the design.**

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE audience IS NULL OR audience = 0)   AS no_audience,
       count(*) FILTER (WHERE impression IS NULL OR impression = 0) AS no_impression,
       count(*) FILTER (WHERE sellable IS NULL OR sellable = '')  AS no_sellable
FROM buildings;
```

### 3.2 `sales_packages` — needs real master-data fields

Today it is only `id`, `name`, timestamps. The quotation flow needs:

| Field | Why |
|---|---|
| `package_code` (unique) | Stable reference for rate card import and quotation snapshots |
| `status` (`active`/`inactive`) | Retired packages must stop appearing in the wizard without deleting history |
| `description` | Shown in the wizard when choosing a package |
| price | Comes from the rate card, not this table |

Note the current `sales_package_buildings` uses `ON DELETE CASCADE`. Once packages are quotable, deleting a package or building would silently mutate the membership behind historical quotes. Quotation line items must therefore **snapshot** the resolved buildings, not just store IDs (see §4.1).

### 3.3 New: rate card (the real missing piece)

The example versions its price list rather than storing prices inline. Recommended for TMN too:

- `rate_card_versions` — `version_code`, `currency` (IDR), `status` (`current` | `historical`), published_by/at
- `rate_card_building_prices` — (version, building) → `price_idr`
- `rate_card_package_prices` — (version, package) → `price_idr`
- `rate_card_package_buildings` — (version, package, building), the package composition **as priced**

Publishing a new version flips the previous `current` to `historical` and keeps all child rows read-only. Quotations store `rate_card_version_id` so an approved quote never changes price when a new rate card lands.

Use `NUMERIC(18,0)` for IDR (matching the example) — integer rupiah, no decimals.

---

## 4. Backend work required (Go)

Follows the existing controller → service → repository → web pattern, one package per entity, registered in `libs/router.go` and `injector/wire.go`.

### 4.1 Migrations (new, continuing from `014_`)

```
015_create_customers_brands_tables
        customers(id, code UNIQUE, name, industry, status, timestamps)
        brands(id, code UNIQUE, customer_id FK, name, category, status, timestamps)

016_create_sales_assignments_table
        sales_assignments(id, customer_id FK, brand_id FK, sales_user_id FK -> users,
                          status, registration_date, expiry_date, timestamps)
        UNIQUE(customer_id, brand_id)   -- one PIC per customer+brand

017_alter_sales_packages_add_master_fields
        + package_code UNIQUE, + description, + status

018_create_rate_card_tables
        rate_card_versions / rate_card_building_prices
        rate_card_package_prices / rate_card_package_buildings

019_create_quotations_tables
        quotations(id, quote_number UNIQUE, sales_user_id, created_by_user_id,
                   customer_id, brand_id, rate_card_version_id,
                   discount NUMERIC(5,2), tax_rate NUMERIC(5,4),
                   status, required_approver_user_id, version,
                   -- pricing snapshot, all NUMERIC(18,0):
                   placement_gross, placement_discount_amount, placement_net,
                   bonus_gross, bonus_net, total_gross, total_net,
                   effective_discount_amount, effective_discount_rate NUMERIC(7,4),
                   tax, total_including_tax,
                   created_at, updated_at, approved_at)

        quotation_selections(id, quotation_id FK, kind 'placement'|'bonus',
                             mode 'building'|'package', package_id,
                             tvc_duration_seconds, weeks, spots,
                             gross_price, traffic, impressions)
        quotation_selection_items(id, selection_id FK, building_id,
                                  -- snapshot columns so history survives master-data edits:
                                  building_name, building_type, citytown, subdistrict,
                                  unit_price_idr, traffic, impressions)

        quotation_versions(id, quotation_id FK, version, snapshot JSONB,
                           required_approver_user_id, submitted_at)
        quotation_approvals(id, quotation_id FK, version, actor_user_id, actor_role,
                            action 'submitted'|'resubmitted'|'approved'|'returned',
                            comment, created_at)

020_alter_users_add_quotation_capabilities
        + can_create_quotations BOOLEAN DEFAULT FALSE
        + sales_group VARCHAR(30)
        user_proxy_sales(id, actor_user_id FK, owner_user_id FK)  -- allow-list
```

Add a CHECK constraint mirroring the example: `returned` approval rows must have a non-empty comment.

### 4.2 Models — `backend/models/`

New files following the existing `NullAbleX` + `NullAbleXToX()` convention:
`customer.go`, `brand.go`, `sales_assignment.go`, `rate_card.go`, `quotation.go`.

⚠️ `models.Building` currently exposes `Audience`/`Impression`. Keep those names in the building domain and map to `traffic`/`impressions` only at the quotation response boundary — do not rename the existing field and break the ERP sync.

### 4.3 New service packages

| Package | Responsibility |
|---|---|
| `services/customer`, `services/brand` | Master-data CRUD + XLSX import/export (mirror `services/salespackage`) |
| `services/salesassignment` | Assignment CRUD; `FindCustomersBySalesUser` for wizard step 1 |
| `services/ratecard` | Upload → validate → publish; `GetCurrentVersion`, `GetBuildingPrices`, `GetPackagePrice` |
| **`services/quotation`** | The core. See below. |
| `services/quotationapproval` | Approve / return, queue queries per approver |

**`services/quotation` must own the pricing and routing logic server-side.** Port `lib/quotation.ts` as pure, table-driven Go functions so they are unit-testable without a DB:

```go
// services/quotation/pricing.go
func CalculatePricing(in PricingInput) (PricingSummary, error)
func ValidateSelection(sel Selection, refs ReferenceData) []ValidationError

// services/quotation/routing.go
func ResolveApprovalRoute(discount float64, ownerUserID int, dir ApprovalDirectory) (Route, error)
func GetDiscountBand(discount float64) DiscountBand
```

Two non-negotiables carried over from the example:
1. **Recompute `gross_price`, `traffic`, `impressions` on the server** from the rate card + selected buildings, and reject a submit whose client-supplied numbers disagree.
2. **Reject a customer/brand the acting sales owner is not assigned to.**

Go note: `float64` money is a bug factory. Use `int64` rupiah internally, or `shopspring/decimal`; `NUMERIC(18,0)` in Postgres. Keep `discount` and `effective_discount_rate` as the only decimal fields.

### 4.4 New routes — `backend/libs/router.go`

```
POST   /customers ... (+ GET, GET/:id, PUT/:id, DELETE/:id, -import, -export)
POST   /brands ...
GET    /brands-dropdown?customer_id=
POST   /sales-assignments ...
GET    /my-customers                       # wizard step 1, scoped to caller
POST   /rate-cards                         # upload
POST   /rate-cards/:id/publish
GET    /rate-cards/current
POST   /quotations                         # create draft
GET    /quotations                         # scoped by role
GET    /quotations/:id
PUT    /quotations/:id                     # edit draft/returned only
POST   /quotations/:id/submit
POST   /quotations/:id/approve
POST   /quotations/:id/return              # comment required
GET    /quotations/:id/versions
GET    /quotations/:id/document            # formal quotation payload
GET    /quotations/pricing-preview         # live wizard calculation
GET    /approvals/queue                    # approver's own queue
```

### 4.5 Authorization middleware — **new, and required**

See §5. A `RequireRole(...)` / `RequirePermission(...)` middleware must be added to `backend/middlewares/auth.go` and wrapped around every quotation route, the same way `RequireAuth` is today.

---

## 5. Roles — the honest answer: **not capable yet**

You asked whether the app already supports the roles this feature needs. It does not.

### 5.1 What exists

| Layer | State |
|---|---|
| `users.role VARCHAR(20) DEFAULT 'user'` (`001_create_users_table.up.sql`) | Column exists |
| `models.User.Role` | Field exists |
| `web/auth/web_auth_response.go` → `role` | Returned to the frontend on login and `/current-user` |
| `frontend/src/stores/auth.ts` → `isAdmin` / `isAuthor` / `isApprover` / `isGuest` | Getters exist |

### 5.2 What is missing — verified by grep

- **No authorization anywhere in the backend.** `grep -rn "role" backend --include=*.go` returns only the auth response struct, the user repository `SELECT`, and a test fixture. `middlewares/auth.go` has exactly one guard, `RequireAuth`, which checks *authentication* only. **Every authenticated user can call every endpoint**, including `DELETE /buildings`, imports, and exports.
- **The frontend role getters are dead code.** `isAdmin`, `isAuthor`, `isApprover`, `isGuest` are defined in `src/stores/auth.ts` and referenced nowhere outside their own unit test.
- **No route guards by role.** `src/plugins/router/index.ts` only checks authentication and redirects to `/login`.
- **Nav is not filtered.** `src/layouts/components/NavItems.vue` renders every link for everyone.
- **The role vocabulary does not match.** Today's implied set is `admin` / `author` / `approver` / `guest` / `user`. The quotation feature needs `sales`, `head_of_sales`, `head_of_business_control`, `ceo` — plus the ability to be *both* (Head of Sales can own and enter quotations, per spec 2026-07-23).
- **A single `role` string is too weak.** The final spec requires capabilities orthogonal to role: `canCreateQuotations`, `canCreateOnBehalfOfSalesIds[]`, `salesGroup`. A Head of Sales approves *and* sells. Model this as role + capability flags (or a `user_permissions` table like the example's) rather than stuffing it into one column.

### 5.3 Recommended role model

```
role (one per user, drives the workspace and the approval directory):
    admin                       — master data, rate card publish, user management
    sales                       — create/edit own quotations
    head_of_sales               — approve ≤65%  + may also own quotations
    head_of_business_control    — approve >65% ≤75%
    ceo                         — approve >75%

capabilities (independent of role):
    can_create_quotations        BOOLEAN
    sales_group                  sales_team | everyone_sales | freelancer
    user_proxy_sales             allow-list of owners this user may enter quotes for
```

The approval directory (which concrete user holds each approver role) must be resolved from **configuration or user data, never hardcoded names.** The example is explicit about this; its own demo users are Ayu / April / Thomas but the routing code never mentions them.

### 5.4 Role work required

**Backend**
1. `RequireRole(roles ...string)` middleware in `middlewares/auth.go`; load the user's role in `RequireAuth` and put it in the request context alongside `userId` (today only `userId` is stored).
2. Migration `020` for capability columns + `user_proxy_sales`.
3. Decide what happens to existing rows with `role = 'user'` — a backfill is needed, otherwise every current account lands in an undefined role.
4. **Retrofit the existing endpoints.** Adding roles for quotations while leaving `/buildings`, `/pois`, `/sales-packages` open to any logged-in user is a half-measure. Treat this as its own slice.

**Frontend**
1. Replace the four dead getters with a real `hasRole()` / `can()` helper driven by the API response.
2. `meta: { roles: [...] }` on quotation routes + a role check in the `router.beforeEach` guard in `src/plugins/router/index.ts`.
3. Role-filter `NavItems.vue`.
4. Role-aware landing: sales → quotation dashboard, approvers → approval queue.

---

## 6. Frontend work required (Vue 3)

### 6.1 New API clients — `src/http/`
`customer.ts`, `brand.ts`, `salesassignment.ts`, `ratecard.ts`, `quotation.ts`, `approval.ts` — same axios wrapper style as `src/http/salespackage.ts`.

### 6.2 New Pinia stores — `src/stores/`
`customer.ts`, `brand.ts`, `ratecard.ts`, `quotation.ts` (wizard draft state + live pricing), `approval.ts` (queue + counts).

### 6.3 New pages — `src/pages/`

| Page | Route | Notes |
|---|---|---|
| `quotations.vue` | `/quotations` | Sales dashboard: counts (returned / pending / approved / all) + list |
| `quotation-form.vue` | `/quotations/new`, `/quotations/:id/edit` | The 6-step wizard, `VStepper` |
| `quotation-detail.vue` | `/quotations/:id` | Status, timeline, version history |
| `quotation-document.vue` | `/quotations/:id/document` | Formal print layout (§6.5) |
| `approvals.vue` | `/approvals` | Approver queue |
| `approval-detail.vue` | `/approvals/:id` | Approve / Return-with-reason |
| `customers.vue` + `customer-form.vue` | `/customers` | Master data CRUD |
| `brands.vue` + `brand-form.vue` | `/brands` | Master data CRUD |
| `sales-assignments.vue` + form | `/sales-assignments` | PIC mapping |
| `rate-cards.vue` + upload/publish | `/rate-cards` | Admin only |

### 6.4 New components — `src/components/quotation/`
`ResourceSelector.vue` (building multi-select w/ filters, or package single-select — reused for both Placement and Bonus), `CommercialSelectionForm.vue`, `PricingSummaryCard.vue` (sticky sidebar), `ApprovalRouteHint.vue` (discount band + who approves), `ApprovalTimeline.vue`, `VersionHistoryDialog.vue`, `DiscountBandChip.vue`.

Accessibility note carried from the spec: **discount risk must not be conveyed by colour alone.** Pair each band chip with a text label.

### 6.5 The quotation document

The example prints these sections (`components/quotation-screen.tsx`) — use as the field checklist:

1. **Reference** — quote number, issue date, version, currency
2. **Client & brand** — customer, brand, sales owner (+ "created by" when proxied), campaign period in weeks
3. **Resources & items** — Placement row + Bonus row (item, type/region, period, amount); metrics: Placement spots, Bonus spots, daily traffic, monthly impressions
4. **Price details** — placement gross, discount deduction, placement nett, bonus gross, bonus nett (FREE), total gross, total nett, tax, **total including tax**
5. **Terms** — validity, rate card reference, currency & tax note
6. **Appendix** — per-building table: name, region/type, daily traffic, monthly impressions
7. **Approval record** — action, approver, version, time & comment + approved stamp

Print is browser-native in the prototype (no server-side PDF). Decide whether TMN needs a real PDF — that changes the backend scope materially.

### 6.6 Existing frontend pages to touch
- `src/plugins/router/routes.ts` — register the new routes
- `src/plugins/router/index.ts` — role guard
- `src/layouts/components/NavItems.vue` — new nav section, role-filtered
- `src/stores/auth.ts` — replace the dead role getters
- `src/pages/sales-package-form.vue` — surface the new `package_code` / `status` fields

---

## 7. Open decisions (blocking)

These change the design materially and are business calls, not engineering ones:

1. **Tax rate.** The prototype hardcodes a *simulated* 6% and labels it as such. Indonesian PPN is 11%/12%. Confirm the real rate and whether it is configurable per quotation.
2. **Rate card period.** The example computes `gross = unitPrice × weeks / 4`, i.e. the rate card price is a **4-week** rate. Confirm this matches TMN's commercial reality.
3. **Approval thresholds.** Confirm 65% / 75% and confirm boundaries are inclusive (`≤65` → Head of Sales, `>65` → Business Control).
4. **Approver identities.** Who holds Head of Sales, Head of Business Control, CEO — and is it one person per role or a pool?
5. **Customer/brand source.** Are advertiser customers coming from ERP (like buildings/LOI/acquisitions already do via `services/erp`), from a CRM export, or entered manually? This decides whether §4.1 migration `015` needs `external_id` + a sync scheduler.
6. **Quotable inventory.** Which `sellable` and `lcd_presence_status` values make a building quotable?
7. **`audience` / `impression` semantics.** Daily traffic and monthly impressions? And what is the current data coverage (query in §3.1)?
8. **PDF.** Browser print (cheap) or server-side PDF generation (new dependency, new backend work)?
9. **Localization.** The example is EN + zh-CN. TMN's UI is English-only today. Does the quotation need Bahasa Indonesia?
10. **Notifications.** Out of scope in the prototype. Does an approver need email/WhatsApp on submit?

---

## 8. Suggested phasing

| Phase | Scope | Why first |
|---|---|---|
| **0. Roles & authorization** | `RequireRole` middleware, role model + capabilities, frontend guards, retrofit existing routes | Everything else depends on it, and it closes a real hole today |
| **1. Customer / Brand / Assignment** | Migrations 015–016, CRUD + import/export, admin pages | Wizard step 1 has no data without it |
| **2. Rate card** | Migrations 017–018, upload → validate → publish, versioning | No prices = no quotation |
| **3. Quotation core** | Migration 019, `services/quotation` pricing + routing (pure, unit-tested), draft/submit, wizard | The feature |
| **4. Approval** | Queue, approve/return, versions, timeline | Closes the loop |
| **5. Document** | Formal quotation page + print | Deliverable output |
| **6. Polish** | Notifications, PDF, localization, dashboards | Deferrable |

Phases 0–2 are pure enablement and contain no quotation logic. Do not compress them into phase 3.

---

## 9. Test impact

Per `CLAUDE.md`, both services need tests updated when this lands.

**Backend** (`go test ./...`) — existing pattern is `services/*/service_*_impl_test.go` with `testutil/fixtures.go`:
- `services/quotation/pricing_test.go` — table-driven; must cover the exact boundaries: 0%, 65%, 65.01%, 75%, 75.01%, 100%; No-Bonus vs Bonus effective-discount; integer-overflow guards
- `services/quotation/routing_test.go` — band boundaries + the **no-self-approval escalation**
- `services/quotation/service_quotation_impl_test.go` — client/server price mismatch rejection, unassigned-customer rejection, draft/returned-only editing, resubmit version increment
- `services/quotationapproval/..._test.go` — an approver can only act on their own queue; return without a comment fails
- `services/ratecard/..._test.go` — publish flips current→historical; historical rows immutable
- `middlewares/auth_test.go` — **new**: `RequireRole` allows/denies correctly
- `testutil/fixtures.go` — add quotation/customer/brand/rate-card fixtures

**Frontend** (`npm test` = `vue-tsc --noEmit && vitest run`) — existing tests live in `src/__tests__/`:
- `src/__tests__/stores/quotation.test.ts` — wizard state, live pricing mirror, validation
- `src/__tests__/stores/approval.test.ts` — queue filtering by role
- `src/__tests__/http/quotation.test.ts` — API client shapes
- `src/__tests__/stores/auth.test.ts` — **update**: the current test asserts the four dead getters; it must follow the new role model
- `src/__tests__/utils/` — IDR currency and percentage formatters

**Cross-cutting:** the frontend pricing preview and the backend calculation must agree exactly. Use one shared fixture set of inputs → expected outputs, asserted in both suites, so rounding can never drift between Go and TypeScript.

---

## 10. Summary

- The reference app's **authoritative** model is single-step approval at 65% / 75%, with independent Placement and Bonus selections. Ignore the cascading 70% model in the earliest spec.
- **Building data is in good shape** — `audience`, `impression`, type, and region already exist. The real gaps are **prices (rate card)** and **sales package master fields**. Verify `audience`/`impression` coverage before designing around them.
- **`mother_brands` is a POI/competitor concept, not an advertiser brand.** Customers and brands must be new tables.
- **Roles are not ready.** The column and the frontend getters exist, but nothing is enforced anywhere in the backend and the getters are dead code. This is both a prerequisite for the quotation feature and an existing security gap worth closing on its own.
- Ten business decisions in §7 should be answered before implementation starts — tax rate, rate-card period, and the customer/brand data source have the widest blast radius.
