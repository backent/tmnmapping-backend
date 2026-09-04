# Phase 0 — Roles & Authorization

Tracking doc for the first phase of the sales quotation feature.
Source analysis: `docs/QUOTATION_FEATURE_ANALYSIS.md` in the `tmn-mapping` project root (§5, §8).

**Status:** ✅ Complete — backend and frontend, tests green, not yet deployed.
**Branch:** `feat/phase-0-roles-authorization` (both `backend/` and `frontend/`)
**Started / finished:** 2026-09-04

---

## 1. Why this came first

Phase 0 was pulled ahead of the quotation feature because the analysis turned up an
existing hole, not just a missing capability:

> `users.role` existed, and the frontend had `isAdmin` / `isAuthor` / `isApprover` /
> `isGuest` getters — but `role` appeared nowhere in the backend except the auth
> response struct, the user `SELECT`, and a test fixture. `middlewares/auth.go` had
> exactly one guard, `RequireAuth`, which checked *authentication* only.
> **Every authenticated user could call every endpoint**, including `DELETE /buildings`
> and every import/export route. The four frontend getters were dead code.

Everything in phases 1–5 depends on role identity, so this had to land first regardless.

---

## 2. Decisions taken

| # | Decision | Choice | Rationale |
|---|---|---|---|
| 1 | Role vocabulary | `admin`, `sales`, `head_of_sales`, `head_of_business_control`, `ceo` | Matches the authoritative spec (2026-07-13 + 2026-07-23). Replaces `admin`/`author`/`approver`/`guest`/`user`. |
| 2 | Backfill of existing users | Unrecognised → `admin`; `approver` → `head_of_sales`; `guest` → `sales` | **Continuity over least privilege.** No existing account is locked out by the migration. Roles get narrowed deliberately, per user, after deploy. |
| 3 | Enforcement scope | Every existing write endpoint, now | Adding roles only for quotations while leaving `/buildings` and `/pois` open to any logged-in user would be a half-measure. |
| 4 | Where the role is read from | Database, on every request | Not carried in the JWT: a role change takes effect immediately without forcing re-login, and no token migration is needed. Costs one indexed lookup per request. |
| 5 | Unknown role at runtime | Denies (normalises to empty) | The migration handles existing data. A role written later by something outside the app must fail closed, not inherit a default. |
| 6 | Capabilities | Columns + table added, no middleware yet | `can_create_quotations`, `sales_group`, `user_proxy_sales` are modelled and surfaced by the API so Phase 3 can consume them. Enforcing them before quotation routes exist would be speculative. |

⚠️ **Decision 2 is the one to revisit after deploy.** Everyone who was using the app is
now an `admin`. See §7.

---

## 3. Backend — what changed

**Branch:** `feat/phase-0-roles-authorization`

| File | Change |
|---|---|
| `models/role.go` | **New.** Role + sales-group constants, `Roles`, `ApproverRoles`, `NormalizeRole`, `IsValidRole`, `IsApproverRole`, `HasRole`. |
| `models/user.go` | `+ CanCreateQuotations`, `+ SalesGroup`; `UserProxySalesTable`. |
| `database/migrations/015_add_user_roles_and_capabilities.{up,down}.sql` | **New.** See §4. |
| `exceptions/forbidden.go` | **New.** `ForbiddenError`, so authorization failures are 403, not 401. |
| `exceptions/http.go` | 403 branch in `RouterPanicHandler`. |
| `middlewares/auth.go` | `RequireAuth` now resolves and attaches the role; **new** `RequireRole(roles...)`. Exported `ContextKeyUserId` / `ContextKeyUserRole`. |
| `repositories/user/repository_user_impl.go` | `SELECT` includes the two new columns. |
| `web/auth/web_auth_response.go` | `+ can_create_quotations`, `+ sales_group`. |
| `services/auth/service_auth_impl.go`, `controllers/auth/controller_auth_impl.go` | Role normalised at the API boundary, so the frontend only ever sees canonical values. |
| `libs/router.go` | 38 write routes wrapped in `RequireRole(models.RoleAdmin)`. |
| `injector/wire_gen.go` | `NewAuthMiddleware` now takes `db` and the user repository. Hand-edited — `wire.go`'s provider set is unchanged, so regenerating produces the same result. |

### Route policy

| Group | Reads (GET) | Writes (POST/PUT/DELETE, import, export, sync) |
|---|---|---|
| Buildings | all roles | `admin` |
| POIs | all roles | `admin` |
| Sales packages | all roles | `admin` |
| Building restrictions | all roles | `admin` |
| Categories, sub-categories, mother brands, branches | all roles | `admin` |
| Saved polygons | all roles | **all roles** — see below |
| Dashboard, mapping search, dropdowns, filter options, ERP images | all roles | — |

Two deliberate exceptions:

- **Saved polygons stay open to every role.** They are a working tool on the map, not
  master data, and the table is not user-scoped. Locking them to `admin` would break the
  mapping page for everyone else.
- **Master-data reads stay open to every role.** `POIPickerDialog.vue` and
  `BuildingRestrictionFilter.vue` on the mapping page load categories, sub-categories,
  mother brands, branches and restrictions. Gating those reads would break mapping for
  non-admins.

---

## 4. Migration 015

```
migrate -path database/migrations \
  -database "postgres://USER:PASS@HOST:5432/tmn_backend?sslmode=disable" up
```

What it does, in order:

1. Widens `users.role` to `VARCHAR(40)` — `head_of_business_control` is 24 characters
   and does not fit the original `VARCHAR(20)`.
2. Backfills the legacy vocabulary (decision 2).
3. Sets `role` `NOT NULL DEFAULT 'sales'` and adds a `CHECK` constraint on the five values.
4. Adds `can_create_quotations BOOLEAN NOT NULL DEFAULT FALSE` and
   `sales_group VARCHAR(30)` (`CHECK`-constrained to the three spec values).
5. Sets `can_create_quotations = TRUE`, `sales_group = 'sales_team'` for `sales` and
   `head_of_sales`. Admins do not sell.
6. Creates `idx_users_role` — the role is read on every authenticated request.
7. Creates `user_proxy_sales(actor_user_id, owner_user_id)`, the proxy-entry allow-list,
   with a unique pair constraint and a no-self-proxy check.

**The down migration is lossy and says so.** `user` and `author` both collapsed into
`admin`, so rolling back cannot distinguish them; it restores the pre-migration default
of `user` for everything except `admin` and `head_of_sales`.

---

## 5. Frontend — what changed

**Branch:** `feat/phase-0-roles-authorization`

| File | Change |
|---|---|
| `src/config/roles.ts` | **New.** `ROLES`, `ALL_ROLES`, `APPROVER_ROLES`, legacy aliases, `normalizeRole`, `isValidRole`, the `PERMISSIONS` map, `roleCan`, `roleHasAny`. Mirrors `models/role.go`. |
| `src/stores/auth.ts` | The four dead getters are gone. Now: `role`, `isAdmin`, `isSales`, `isApprover`, `can(permission)`, `hasAnyRole(roles)`, `canCreateQuotations`. |
| `src/http/auth.ts` | `User` carries `can_create_quotations` and `sales_group`; `LoginResponse.user` reuses `User` instead of redeclaring it. |
| `src/plugins/router/guards.ts` | **New.** `resolveNavigation` extracted from the router so it is unit-testable without mounting an app. |
| `src/plugins/router/index.ts` | Reduced to router construction; delegates to `resolveNavigation`. |
| `src/plugins/router/routes.ts` | `meta: { permission }` on 20 routes; `+ /not-authorized`. |
| `src/pages/not-authorized.vue` | **New.** 403 page, so a denied user lands somewhere instead of bouncing. |
| `src/layouts/components/NavItems.vue` | Restrictions and Master Data sections hidden from roles that cannot open them. |

`isApprover` changed meaning: it used to be `role === 'approver'`, it now means "holds one
of the three approval bands". `admin` is deliberately **not** an approver — approval
authority follows the sales hierarchy, not system administration.

### Honest note on the frontend gate

`*.manage` permissions correspond one-to-one with a `RequireRole` on the server.
`*.view` permissions (`master-data.view`, `building-restrictions.view`) are **UI-only** —
the API still serves those reads to every authenticated role, because the mapping page
needs the same records. Hiding the management screens is a navigation decision, not a
security boundary. Every mutation is server-enforced.

---

## 6. Tests

All green as of the last run.

**Backend** — `go test ./...`

| File | Covers |
|---|---|
| `models/role_test.go` (**new**, 5 tests) | Canonical pass-through, all four legacy aliases, unknown/empty/wrong-case → no role, `admin` is not an approver, `HasRole` never matches an empty role. |
| `middlewares/auth_test.go` (**new**, 11 tests) | Runs through a real `httprouter` with the real panic handler, so it asserts on the HTTP status a client receives: no token → 401, invalid token → 401, deleted account → 401, context carries id + role, legacy role normalised, `Authorization` header fallback, matching role → 200, wrong role → **403** with the handler not running, any-of-several-roles, unknown stored role → 403, `RequireRole` wired without `RequireAuth` → 401 (fails closed). |

**Frontend** — `npm test` (`vue-tsc --noEmit && vitest run`) → **289 passed, 15 files**

| File | Covers |
|---|---|
| `src/__tests__/config/roles.test.ts` (**new**, 34 tests) | Normalisation, validity, `APPROVER_ROLES` excludes admin, and a sweep asserting *every* `.manage` permission is admin-only and every `.view` shared permission is open to all roles. |
| `src/__tests__/plugins/routerGuards.test.ts` (**new**, 31 tests) | Login redirects, session restore, restore failure → `/login`, permitted/denied routes, unknown role denied, `/not-authorized` reachable by any role but still requires a session. Plus a route-table sweep: every `new`/`edit` form route is gated, and `/dashboard`, `/mapping`, `/buildings`, `/pois`, `/sales-packages` stay open. |
| `src/__tests__/stores/auth.test.ts` (**rewritten getters block**) | Fixtures moved to the new vocabulary; the dead-getter assertions are replaced with `role`, `isAdmin`, `isSales`, `isApprover`, `can`, `hasAnyRole`, `canCreateQuotations`. |

### Not covered

- No test asserts the `libs/router.go` policy table itself — the middleware tests cover
  the mechanism, not which routes it is attached to. A wiring mistake on a single route
  would not be caught. Worth an integration test when the quotation routes land.
- Migration 015 is not exercised by a test; there is no migration test harness in the repo.

---

## 7. Before deploying

1. **Run migration 015** against staging first and check the backfill:
   ```sql
   SELECT role, count(*), count(*) FILTER (WHERE can_create_quotations) AS can_quote
   FROM users GROUP BY role ORDER BY 2 DESC;
   ```
2. **Narrow the admins.** Everyone landed on `admin` by design. Decide who is actually
   `sales` / `head_of_sales` / `head_of_business_control` / `ceo` and update them.
   There is no user-management UI yet, so this is SQL for now (see §8).
3. **Smoke test as a non-admin:** dashboard, mapping (POI picker + restriction filter),
   buildings list, sales packages list should all work; the Master Data and Restrictions
   nav sections should be gone; `/categories/new` typed directly should land on
   `/not-authorized`; a `DELETE /pois/:id` should come back 403.
4. Deploy backend and frontend together — the frontend reads `can_create_quotations`
   from `/current-user`, which only exists after the backend ships.

---

## 8. Follow-ups this phase deliberately left open

| Item | Note |
|---|---|
| **User management UI** | There is no screen to set a user's role or capabilities. Roles must be changed by SQL. This is the biggest usability gap left by Phase 0. |
| **Capability enforcement** | `can_create_quotations`, `sales_group` and `user_proxy_sales` are stored and returned but nothing checks them yet. Phase 3. |
| **Approval directory** | Which concrete user holds each approver role must come from configuration or user data, never hardcoded names. Not yet built — Phase 4. |
| **Route policy test** | See §6. |
| **`docs/QUOTATION_FEATURE_ANALYSIS.md` is not version controlled.** | It lives in `tmn-mapping/docs/`, which is outside both git repos. It is the source of truth for phases 1–6 and currently exists only on one machine. |

---

## 9. Phase status

| Phase | Scope | Status |
|---|---|---|
| **0. Roles & authorization** | `RequireRole`, role model + capabilities, frontend guards, retrofit existing routes | ✅ **Done** |
| 1. Customer / Brand / Assignment | Migrations 016–017, CRUD + import/export, admin pages | Not started |
| 2. Rate card | Migrations 018–019, upload → validate → publish, versioning | Not started |
| 3. Quotation core | Migration 020, `services/quotation` pricing + routing, wizard | Not started |
| 4. Approval | Queue, approve/return, versions, timeline | Not started |
| 5. Document | Formal quotation page + print | Not started |
| 6. Polish | Notifications, PDF, localization, dashboards | Not started |

> Migration numbers for phases 1–3 have shifted by one from the analysis doc, which
> assumed Phase 0 would use `020`. Phase 0 took `015`.

**Blocking before Phase 1:** the ten business decisions in §7 of the analysis doc.
The three with the widest blast radius are the tax rate (the prototype hardcodes a
self-labelled "simulated" 6%; Indonesian PPN is 11%/12%), the rate card period
(`gross = price × weeks / 4` implies 4-week rates), and whether customers/brands come
from ERP or are entered manually.
