# Sales Quotation — Feature Guide

**Audience:** anyone who needs to understand what this feature does and why it is
built the way it is — new developers, and the sales/finance people who own the
process.
**Status:** Phases 0–3 and the printed document are built and running on staging.
Not in production. See §12 for what is still open.
**Last updated:** 2026-09-08

Companion documents:

| Document | What it is for |
|---|---|
| `QUOTATION_FEATURE_ANALYSIS.md` | The original gap analysis against the reference prototype |
| `QUOTATION_DOCUMENT_ANALYSIS.md` | The real 2026 PDF template, field by field |
| `PHASE_0_ROLES_PROGRESS.md` | How roles and the permission layer were built |
| `PHASE_1_IMPORT_PLAN.md` | Customer/brand/assignment spreadsheet import |
| `STAGING_ENVIRONMENT_PLAN.md` | The staging server |

---

## 1. What this feature is for

A salesperson needs to give an advertiser a price for putting a video ad on TMN's
building screens. That price is not a simple list price:

- It is assembled from **buildings or packages**, each with its own weekly rate.
- It usually carries a **large discount** — 65% is normal in this business.
- Anything discounted must be **approved by someone senior**, and who that is
  depends on how deep the discount goes.
- The result is a **legal document** the client signs.

Before this feature, all of that happened in spreadsheets. The risks were the
obvious ones: prices typed by hand, discounts approved over chat, no record of who
agreed to what, and no way to answer "what did we actually quote this client in
March?".

This feature moves the whole thing into the application: **priced from a controlled
rate card, routed for approval by rule, and recorded permanently.**

### The one idea worth understanding first

**The system, not the user, calculates money.** The browser never does arithmetic on
prices. Every figure on the screen — including the live preview while typing — comes
from the server, from one function (`services/quotation/pricing.go`). This is why
you can trust that the number on the printed document is the number that was
approved.

---

## 2. Who uses it — actors and use cases

Five roles exist. They are **job titles from the org chart**, deliberately not
quotation-specific — so the next feature can reuse them without a rewrite.

| Role | In the business | In this feature |
|---|---|---|
| `sales` | Account executive | Writes quotations, owns customers |
| `head_of_sales` | Sales manager | Approves discounts up to 65% |
| `head_of_business_control` | Commercial control | Approves 65–75% |
| `ceo` | CEO | Approves above 75% |
| `admin` | System administrator | Maintains master data, rate cards, users. **Cannot approve** |

`admin` being unable to approve is intentional: approval authority follows the sales
hierarchy, not system administration. Someone who can edit the database should not
also be able to sign off a discount.

### Use case diagram

```
                          SALES QUOTATION SYSTEM
    ┌───────────────────────────────────────────────────────────────────┐
    │                                                                   │
    │   ┌──────────────────────┐        ┌──────────────────────────┐   │
    │   │ Browse rate card     │        │ Download / upload        │   │
    │   │ (read current prices)│        │ master data (spreadsheet)│   │
    │   └──────────────────────┘        └──────────────────────────┘   │
    │              ▲                                  ▲                 │
    │              │                                  │                 │
    │   ┌──────────────────────┐        ┌──────────────────────────┐   │
    │   │ Create quotation     │        │ Maintain customers,      │   │
    │   │  • pick customer     │        │ brands, sales assignments│   │
    │   │  • pick placement    │        └──────────────────────────┘   │
    │   │  • pick bonus        │                     ▲                 │
    │   │  • set discount      │        ┌──────────────────────────┐   │
    │   └──────────┬───────────┘        │ Publish rate card version│   │
    │              │ «includes»         └──────────────────────────┘   │
    │              ▼                                  ▲                 │
    │   ┌──────────────────────┐                      │                 │
    │   │ Price preview        │◄────────── server-side only            │
    │   └──────────────────────┘                                        │
    │              │                                                    │
    │   ┌──────────▼───────────┐  «includes»  ┌────────────────────┐   │
    │   │ Submit for approval  ├─────────────►│ Resolve approver   │   │
    │   └──────────┬───────────┘              │ (band + no self-   │   │
    │              │                          │  approval rule)    │   │
    │              │                          └────────────────────┘   │
    │              ▼                                                    │
    │   ┌──────────────────────┐        ┌──────────────────────────┐   │
    │   │ Approve quotation    │        │ Return with reason       │   │
    │   └──────────────────────┘        └──────────────────────────┘   │
    │              ▲                                  ▲                 │
    │              │                                  │                 │
    │   ┌──────────────────────┐        ┌──────────────────────────┐   │
    │   │ Revise and resubmit  │        │ Print / save as PDF      │   │
    │   │ (creates version n+1)│        └──────────────────────────┘   │
    │   └──────────────────────┘                                        │
    │                                                                   │
    └───────────────────────────────────────────────────────────────────┘

         ▲                    ▲                         ▲
         │                    │                         │
    ┌────┴─────┐    ┌─────────┴──────────┐    ┌─────────┴─────────┐
    │  Sales   │    │ Head of Sales      │    │      Admin        │
    │          │    │ Head of Bus.Control│    │                   │
    │          │    │ CEO                │    │                   │
    └──────────┘    └────────────────────┘    └───────────────────┘
     creates,        approve / return,          master data,
     revises,        read the queue,            rate cards, users.
     prints          also create quotations     NEVER approves
```

Reading the actor lines:

- **Sales** — everything on the left: create, revise, resubmit, print. Sees only
  their own quotations.
- **Approvers** — the three senior roles. They can also *create* quotations (they
  sell too), which is exactly why the no-self-approval rule in §6 exists.
- **Admin** — master data and rate cards. Sees every quotation, approves none.

---

## 3. The lifecycle

A quotation is always in exactly one of six states.

```
                  ┌─────────┐
     create ─────►│  draft  │◄──────────────────┐
                  └────┬────┘                   │
                       │ submit                 │ revise
                       │ (prices recalculated,  │ (version + 1)
                       │  approver resolved)    │
                       ▼                        │
        ┌──────────────────────────────┐        │
        │  pending_manager             │        │
        │  pending_business_control    │        │
        │  pending_ceo                 │        │
        │  (which one = discount band) │        │
        └───────┬──────────────┬───────┘        │
                │              │                │
        approve │              │ return         │
                │              │ (reason        │
                ▼              ▼  REQUIRED)     │
          ┌──────────┐   ┌──────────┐           │
          │ approved │   │ returned ├───────────┘
          └──────────┘   └──────────┘
             final
```

Rules that hold at every step:

1. **Prices are recalculated on submit.** Not trusted from the browser. A quotation
   sitting in draft for a month is re-priced against the rate card when it is
   finally submitted.
2. **Returning requires a written reason.** Enforced in the service *and* by a
   database `CHECK` constraint — an empty reason is rejected with a 400.
3. **Resubmitting creates a new version.** Version 1 is never edited; a new
   immutable snapshot is appended. You can always see what version 1 said.
4. **`approved` is final.** There is no un-approve. A changed deal is a new
   quotation.

---

## 4. What goes into a quotation

Every quotation has exactly **two selections**, mirroring the paper template:

| Selection | Meaning | Charged? |
|---|---|---|
| **Placement** | What the client is buying | Yes, discount applies |
| **Bonus** | Free inventory thrown in to close the deal | **Never** — always free |

Each selection is independently either:

- **`building` mode** — pick specific buildings, and their weekly rates are summed.
- **`package` mode** — pick one pre-defined sales package at its package rate.

So a deal can be "Package A as placement, plus three named buildings free as bonus",
or any other combination.

Each selection also carries its own **TVC duration** (ad length in seconds),
**spots per day per screen**, and **campaign duration in weeks**.

### Two identities per quotation

This trips people up, so it is worth being explicit:

| Field | Meaning |
|---|---|
| `sales_user_id` | The **commercial owner**. Drives visibility, approval routing and reporting |
| `created_by_user_id` | Who **physically typed it in**. Audit only, never changes |

They differ when someone enters a quotation on a colleague's behalf. The routing
rules always follow the *owner*, not the typist.

---

## 5. How the money is calculated

One function, `CalculatePricing`, with no database access — which is why it can be
tested directly against the real signed document.

```
placement_gross = weekly_rate × weeks
placement_net   = placement_gross − round(placement_gross × discount%)

bonus_gross     = weekly_rate × weeks      (shown, to prove the giveaway's value)
bonus_net       = 0                        (always free)

total_gross     = placement_gross + bonus_gross
total_net       = placement_net            (bonus adds nothing to nett)

tax             = round(total_net × 11%)   ← on NETT, not gross
total_incl_tax  = total_net + tax
```

Two things to note:

- **Rates are per week and multiply straight through.** A 4-week campaign at
  380,000,000/week is 1,520,000,000. There is no division anywhere. (The reference
  prototype divided by four because its rate card was quoted per-4-weeks; ours is
  not, and copying that would have quartered every price.)
- **VAT is charged on nett, not gross.** Verified against the real document:
  58,520,000 / 532,000,000 = exactly 11.00%.

Money is stored as whole rupiah (`NUMERIC(18,0)`) — no cents, because the currency
has no practical minor unit here.

### Customer discount vs effective discount — do not confuse these

| | Meaning | Routes approval? |
|---|---|---|
| **Customer discount** | The % the salesperson typed. Applies to placement | **Yes** |
| **Effective discount** | Total giveaway including the free bonus, as % of total gross | **Never** |

On the real 2026 template these read **65%** and **70.44%** — different numbers, and
they fall in *different approval bands*. Routing on the wrong one would send deals to
the wrong approver. The effective rate exists for commercial-risk reporting only.

---

## 6. Who approves what

Two rules, applied in order at submit time.

**Rule 1 — the discount picks the approver.**

| Customer discount | Band | Approver | Resulting status |
|---|---|---|---|
| ≤ 65% | standard | Head of Sales | `pending_manager` |
| > 65% and ≤ 75% | elevated | Head of Business Control | `pending_business_control` |
| > 75% | executive | CEO | `pending_ceo` |

Both thresholds are **configurable**, because they are commercial policy rather than
a system rule:

```
QUOTATION_DISCOUNT_THRESHOLD_STANDARD=65
QUOTATION_DISCOUNT_THRESHOLD_ELEVATED=75
```

**Rule 2 — nobody approves their own quotation.**

If the resolved approver *is* the quotation's owner, it escalates one level: Head of
Sales → Head of Business Control → CEO. Escalation chains, so a Head of Business
Control who writes her own 70% deal goes to the CEO.

If the owner is the CEO, there is nobody above — and the system **refuses the
submission** rather than quietly letting it self-approve.

**Resolution is strict.** Exactly one user must hold the target role. Zero holders
means nobody can approve; more than one means the system would be guessing. Both are
rejected loudly at submit time. The approver is resolved *at submit*, from the live
users table, so a role change takes effect immediately.

---

## 7. Who can see what

Visibility is scoped in the service, not by hiding buttons:

- **Admin** sees every quotation.
- **Everyone else** sees only quotations they own…
- **…unless** they are asking for their approval queue, which returns quotations
  awaiting their decision. An approver must be able to read a quotation they did not
  write in order to decide on it.

This is why `quotations.view` and `quotations.manage` are open to *all* roles. The
endpoint is not the security boundary — the scope is. Restricting the endpoint by
role would break approvers.

Acting on an approval is separate and *is* role-restricted:

| Permission | Held by |
|---|---|
| `quotations.view` / `quotations.manage` | all roles (scoped per user) |
| `quotations.approve` | `head_of_sales`, `head_of_business_control`, `ceo` |
| `customers.*`, `brands.*`, `rate-cards.manage`, `rate-cards.publish` | `admin` only (view is open) |

All of this lives in **one file**, `models/permission.go`. Adding a feature means
adding keys there and listing the roles — routes and the role vocabulary stay
untouched. The frontend reads the current user's permissions from `/current-user`
and keeps **no copy** of the policy, so the two can never drift.

A safety net worth knowing about: `RequirePermission` **panics at startup** if a
route names a permission that does not exist. A typo takes the server down
immediately rather than silently rejecting every request to that route forever.

---

## 8. Rate cards — why quotations do not re-price themselves

Prices live in **versioned rate cards**. A rate card is edited as a draft, then
**published**; publishing is its own permission because it changes what every future
quotation is priced against.

Each quotation stores `rate_card_version_id` — the card it was priced against. When
a new card is published, **existing quotations keep their prices.** An approved
quotation is a commitment; it must not silently change because someone updated a
price list.

The same principle appears twice more:

- Each quotation stores its own `tax_rate`. If PPN changes to 12%, already-approved
  quotations keep 11%.
- Building names, types and prices are **copied** into
  `quotation_selection_items` at submit time rather than joined. Renaming or
  re-pricing a building later cannot rewrite what a client was sold.

This is the general rule of the feature: **a submitted quotation is a historical
record, not a live query.**

---

## 9. Data model

```
users ────────────┬─── sales_user_id ──────┐
                  └─── created_by ─────────┤
                                           │
customers ──────────── customer_id ────────┤
    │                                      │
    └── brands ──────── brand_id ──────────┤
                                           ▼
rate_card_versions ─── priced_against ─► quotations
    ├── rate_card_building_prices            │
    ├── rate_card_package_prices             ├──► quotation_selections ──► quotation_selection_items
    └── rate_card_package_buildings          │      (placement | bonus)       (building snapshots)
                                             │
buildings ───────────────────────────────────┼──► quotation_versions   (immutable snapshot per submit)
sales_packages ──────────────────────────────┤
                                             └──► quotation_approvals  (who did what, when, why)
```

Key tables:

| Table | Holds |
|---|---|
| `quotations` | The header: parties, discount, status, version, and the full pricing snapshot |
| `quotation_selections` | Exactly two rows max per quotation — placement and bonus |
| `quotation_selection_items` | The buildings behind a selection, snapshotted |
| `quotation_versions` | One immutable JSONB snapshot per submitted version |
| `quotation_approvals` | Append-only audit: `submitted`, `resubmitted`, `approved`, `returned` |

Quote numbers are `Q-<year>-<sequence>`, sequence counted within the year.

Constraints worth knowing (the database enforces business rules, not just shapes):

- `status` is restricted to the six valid states.
- `approved_at` must be set if and only if status is `approved`.
- A `returned` approval row **must** have a non-blank comment.
- Package mode requires a package; building mode forbids one.
- Amounts cannot be negative.

Migrations: `015` roles, `016` customers/brands/assignments, `017` rate card,
`018` sales package master fields, `019` quotations.

---

## 10. Screens and endpoints

### Screens

| Route | Screen |
|---|---|
| `/quotations` | Pipeline list, filterable; approvers get an "awaiting me" view |
| `/quotations/new` | 6-step wizard |
| `/quotations/:id` | Detail: figures, selections, approval history, actions |
| `/quotations/:id/edit` | Revise a draft or returned quotation |
| `/quotations/:id/document` | **The printable A4 document** |
| `/rate-cards`, `/customers`, `/brands`, `/sales-assignments` | Supporting master data |

The wizard's six steps: **Customer & brand → Placement → Bonus → Campaign →
Discount → Review**. The pricing panel is visible throughout and updates ~300ms
after you stop typing, always from the server.

### API

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/quotations` | List, scoped per §7 |
| `GET` | `/quotations-dashboard` | Pipeline summary |
| `GET` | `/quotations/:id` | Detail with selections, versions, approvals |
| `POST` | `/quotations` | Create draft |
| `PUT` | `/quotations/:id` | Update draft/returned |
| `DELETE` | `/quotations/:id` | Delete draft |
| `POST` | `/quotations/:id/submit` | Re-price, resolve approver, submit |
| `POST` | `/quotations/:id/approve` | Approve (must be the assigned approver) |
| `POST` | `/quotations/:id/return` | Return — reason required |
| `POST` | `/quotations-pricing-preview` | Price without saving. Powers the wizard |

Master data additionally supports `-template`, `-import` and `-export` for
spreadsheet round-trips: download a template, fill it in, upload it.

---

## 11. The printed document

`/quotations/:id/document` renders the A4 page the client signs, laid out to match
the 2026 template: letterhead and saving value, party blocks, the total discount with
its approval flag, the placement/bonus/TOTAL table, terms and conditions, the
nett/VAT/total block, payment terms, invoicing documents, and a signature area.

Notes:

- It renders on the **blank layout**, so the app's navigation never reaches the page.
- It reads **only** the stored pricing snapshot — it performs no arithmetic.
- Static content lives in `frontend/src/config/quotationDocument.ts`, transcribed
  from the source PDF **including its typos** ("maximun", "Cancelation", "foreit").
  These clauses carry legal force; silently editing them is a business decision, not
  a developer's.
- `TERMS_VERSION` (`2026.1`) versions that wording, and a test pins the clause text
  to it — changing the terms fails the test until the version is bumped deliberately.

---

## 12. What is not built yet

| Gap | Impact |
|---|---|
| **PIC Finance, sales phone/team not stored** | Those template fields are omitted from the printed document |
| **No dedicated approval queue screen** | Approvers filter the list instead |
| **No version-history browsing UI** | Versions are stored and visible in the audit trail, but not diffable |
| **No notifications** | An approver is not told a quotation is waiting |
| **No server-side PDF** | Printing relies on the browser's print dialog |
| **Building contract import** | Planned in `BUILDING_CONTRACT_IMPORT_PLAN.md`, off the critical path |
| **Not in production** | Production is still on migration 014 |

---

## 13. Where the code lives

**Backend**

```
models/permission.go              the whole authorization policy, one map
models/role.go                    role vocabulary + legacy aliases
middlewares/auth.go               RequireAuth, RequirePermission
services/quotation/pricing.go     money. pure, no DB
services/quotation/routing.go     approval bands + no-self-approval. pure, no DB
services/quotation/roles.go       which roles hold an approval queue
services/quotation/service_*.go   orchestration, visibility scoping
repositories/quotation/           persistence, quote numbering
database/migrations/015..019      schema
libs/router_conflict_test.go      catches route conflicts before they crash startup
```

**Frontend**

```
src/pages/quotation-form.vue          the 6-step wizard
src/pages/quotation-detail.vue        detail + approval actions
src/pages/quotation-document.vue      the printable A4 document
src/pages/quotations.vue              pipeline list
src/components/quotation/             PricingSummary and friends
src/config/quotationDocument.ts       company details + legal terms
src/stores/quotation.ts               Pinia store
```

The two pure files — `pricing.go` and `routing.go` — are the ones to read first.
They contain the entire commercial logic, take no database handle, and are tested
directly against the real signed quotation.
