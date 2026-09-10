# Sales Quotation — Feature Guide

**Audience:** anyone who needs to understand what this feature does and why it is
built the way it is — new developers, and the sales/finance people who own the
process.
**Status:** Phases 0–3 and the printed document are built and running on staging.
Not in production. See §13 for what is still open.
**Last updated:** 2026-09-08

> **Just want to use it?** §3 is the step-by-step walkthrough. The rest
> explains why it behaves the way it does.

Companion documents:

| Document | What it is for |
|---|---|
| **`QUOTATION_PROGRESS.md`** | **Where the project stands and what is left. Start there.** |
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
price list, routed for approval by rule, and recorded permanently.**

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
| `admin` | System administrator | Maintains master data, prices, users. **Cannot approve** |

`admin` being unable to approve is intentional: approval authority follows the sales
hierarchy, not system administration. Someone who can edit the database should not
also be able to sign off a discount.

### Use case diagram

```
                          SALES QUOTATION SYSTEM
    ┌───────────────────────────────────────────────────────────────────┐
    │                                                                   │
    │   ┌──────────────────────┐        ┌──────────────────────────┐   │
    │   │ Browse prices        │        │ Download / upload        │   │
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
    │   └──────────┬───────────┘        │ Upload & preview prices  │   │
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
     revises,        read the queue,            prices, users.
     prints          also create quotations     NEVER approves
```

Reading the actor lines:

- **Sales** — everything on the left: create, revise, resubmit, print. Sees only
  their own quotations.
- **Approvers** — the three senior roles. They can also *create* quotations (they
  sell too), which is exactly why the no-self-approval rule in §7 exists.
- **Admin** — master data and prices. Sees every quotation, approves none.

---

## 3. How to use it — step by step

This section is the walkthrough. Everything here is what the screens actually say.

### 3.1 Before anyone can quote (admin, once)

A quotation cannot be priced out of thin air. Four things must exist first:

1. **Users with roles.** At minimum one `sales`, and **exactly one** user in each
   approver role you intend to use. "Exactly one" is not a style preference — the
   system refuses to submit if a role has zero or several holders (§7).
2. **Customers and brands.** A quotation is always *for* a brand *of* a customer.
3. **Sales assignments.** Which salesperson owns which customer.
4. **Prices.** A building or package with no price cannot be quoted. Load building
   prices on the **Prices** page (§3.8), and set each package's price on its own form.

Items 2 and 3 are fastest via spreadsheet — see §3.7.

### 3.2 Creating a quotation (sales)

**Quotations → New quotation.** Six steps, with the price panel visible the whole
way and updating about 300ms after you stop typing.

| Step | What you do |
|---|---|
| **1. Customer & brand** | Pick **Customer**, then **Brand**. Fill the client contact: **Attention to**, **Job title**, **Phone**, **Email**. These print on the document, and they live on the quotation rather than the customer — the same client may be quoted through different people |
| **2. Placement** | Choose **Individual buildings** or **Sales package**. Buildings: search and tick them; their weekly rates add up. Package: choose one — it is priced as a single resource |
| **3. Bonus** | Tick **Include a bonus selection** if you are giving inventory away. Same two modes. Bonus is **always free** — it shows its gross value on the document to prove what was given, and contributes nothing to the nett |
| **4. Campaign** | **TVC duration (seconds)**, **Campaign duration (weeks)**, **Spots / day / screen** — set per selection, so placement and bonus can differ |
| **5. Discount** | Drag the **Customer discount (%)** slider. The panel immediately shows the deduction and tells you **who will have to approve it**. **VAT rate** defaults to 11% |
| **6. Review** | Everything in one place before committing |

At the end you have two buttons: **Save draft** (come back later) and **Submit for
approval**.

> If a package you expect is missing in step 2, it has no price yet or it is
> inactive. Set its price on the sales package form. That is a data gap, not a bug.

### 3.3 Submitting

**Submit for approval** does three things in one transaction:

1. **Re-prices everything** at current prices. Whatever the browser
   showed is discarded and recalculated server-side. A draft left for a month is
   priced as of *today*.
2. **Resolves the approver** from your discount and the no-self-approval rule (§7).
3. **Snapshots the version** and writes the audit trail.

The status changes from `draft` to one of the three `pending_*` states depending on
who it went to.

### 3.4 Approving or returning (approvers)

Approvers see two tabs on the quotation list: **My quotations** and **Awaiting my
approval**. The second is your queue.

Open one and you get the full picture — figures, selections, and the history of what
happened so far. Two actions:

- **Approve** — final. There is no un-approve; a changed deal is a new quotation.
- **Return** — sends it back to the salesperson. **A reason is mandatory.** Leaving
  it blank is rejected, and it is rejected in two independent places: the service
  and a database constraint. This is deliberate: "returned with no explanation" is
  the failure mode the whole approval trail exists to prevent.

You can only act on quotations **assigned to you**. Approving someone else's queue
item returns 403 even if you hold an approver role.

### 3.5 Revising after a return

A returned quotation goes back to the owner, who edits and resubmits. On resubmit:

- The version number increments — v1 is never overwritten.
- Prices are recalculated again.
- The approver is resolved again. **If the discount changed enough to cross a band,
  it goes to a different person.**

The audit trail keeps every version, every actor and every reason.

### 3.6 Printing the document

From the quotation detail, **Document** → then **Print / Save as PDF**.

The page is A4 with 12mm margins, and the app's navigation is suppressed in print
output. Use your browser's "Save as PDF" destination to produce a file to email.

### 3.7 Master data by spreadsheet (admin)

Customers, brands, sales assignments and building prices all have the same three
buttons: **Template**, **Export**, **Import**.

The normal flow is: **Template** → fill it in → **Import**. Use **Export** to pull
current data down, edit, and push it back.

**Imports are all-or-nothing.** If any row fails, nothing is written — you get the
errors and the data is untouched. There is no partial import to clean up.

### 3.8 Setting prices (admin)

**Building prices** live on the **Prices** page. Upload a spreadsheet — the
business's own rate card workbook works as-is, its "Round Up" column is read as the
price — or add, edit and remove prices one building at a time.

An upload is **previewed before it applies**: rows read, new, changed, already
correct and skipped, then **Apply**. Nothing changes until you confirm. Buildings not
in the file keep their price, and a row priced 0 is skipped rather than offered for
nothing. Like every import, one bad row refuses the whole file.

**Package prices** are set on each sales package's own form, which shows what its
member buildings add up to as a starting point.

There are no versions and no publish step. A change takes effect for the next
quotation priced and **never** alters a quotation already submitted — see §9.

### 3.9 When something is refused

The system fails loudly rather than guessing. What the common refusals mean:

| What you see | What it means | Fix |
|---|---|---|
| *No user holds the required approver role* | The role that should approve this discount has no user | Assign someone that role |
| *More than one user holds the required approver role* | Two people hold it; the system will not pick | Leave exactly one |
| Submit refused, owner is the CEO | Nobody outranks the CEO, so there is no valid approver | Someone else must own the quotation |
| **403** on approve | You are not the assigned approver | Check whose queue it is in |
| **400** on return | The reason was blank | Write why |
| No packages offered | No active package has a price | Set one on the sales package form |
| *"X has no price yet, so it cannot be quoted"* | That building or package has no price | Add it on the Prices page, or on the package form |

---

## 4. The lifecycle

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
   sitting in draft for a month is re-priced at current prices when it is
   finally submitted.
2. **Returning requires a written reason.** Enforced in the service *and* by a
   database `CHECK` constraint — an empty reason is rejected with a 400.
3. **Resubmitting creates a new version.** Version 1 is never edited; a new
   immutable snapshot is appended. You can always see what version 1 said.
4. **`approved` is final.** There is no un-approve. A changed deal is a new
   quotation.

---

## 5. What goes into a quotation

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

## 6. How the money is calculated

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

## 7. Who approves what

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

## 8. Who can see what

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
| `customers.*`, `brands.*`, `building-prices.manage`, `sales-packages.manage` | `admin` only (view is open) |

All of this lives in **one file**, `models/permission.go`. Adding a feature means
adding keys there and listing the roles — routes and the role vocabulary stay
untouched. The frontend reads the current user's permissions from `/current-user`
and keeps **no copy** of the policy, so the two can never drift.

A safety net worth knowing about: `RequirePermission` **panics at startup** if a
route names a permission that does not exist. A typo takes the server down
immediately rather than silently rejecting every request to that route forever.

---

## 9. Prices — why quotations do not re-price themselves

Prices are **current**, not versioned. A building's weekly price sits on the Prices
page; a package's sits on the package. Change one and the next quotation priced uses
the new figure.

That is safe because **a submitted quotation carries its own copy of every price**.
Submit stores the full pricing snapshot on the quotation and a unit price per
building on `quotation_selection_items`, and reading a quotation never recalculates.
An approved quotation is a commitment, and it keeps its figures whatever happens to
the price list afterwards.

This replaced versioned rate cards (draft → publish → current → historical) on
2026-09-10. The versions looked like the protection, but the snapshot always was; the
versions were ceremony nobody outside the system saw. The rate card tables remain in
the database, unused, as history.

The same principle appears twice more:

- Each quotation stores its own `tax_rate`. If PPN changes to 12%, already-approved
  quotations keep 11%.
- Building names, types and prices are **copied** into
  `quotation_selection_items` at submit time rather than joined. Renaming or
  re-pricing a building later cannot rewrite what a client was sold.

This is the general rule of the feature: **a submitted quotation is a historical
record, not a live query.**

---

## 10. Data model

```
users ────────────┬─── sales_user_id ──────┐
                  └─── created_by ─────────┤
                                           │
customers ──────────── customer_id ────────┤
    │                                      │
    └── brands ──────── brand_id ──────────┤
                                           ▼
building_prices ─────── priced from ──► quotations
    (one per building)                       │
                                             ├──► quotation_selections ──► quotation_selection_items
                                             │      (placement | bonus)       (building + price snapshots)
                                             │
buildings ───────────────────────────────────┼──► quotation_versions   (immutable snapshot per submit)
sales_packages (carry their own price) ──────┤
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

Migrations: `015` roles, `016` customers/brands/assignments, `017` rate card (now
unused), `018` sales package master fields, `019` quotations, `020` package price,
`021` building prices.

---

## 11. Screens and endpoints

### Screens

| Route | Screen |
|---|---|
| `/quotations` | Pipeline list, filterable; approvers get an "awaiting me" view |
| `/quotations/new` | 6-step wizard |
| `/quotations/:id` | Detail: figures, selections, approval history, actions |
| `/quotations/:id/edit` | Revise a draft or returned quotation |
| `/quotations/:id/document` | **The printable A4 document** |
| `/building-prices` | Prices: search, add, edit, remove, upload with preview |
| `/customers`, `/advertiser-brands`, `/sales-assignments`, `/sales-packages` | Supporting master data |

The wizard's six steps: **Customer & brand → Placement → Bonus → Campaign →
Discount → Review**. The pricing panel is visible throughout and updates ~300ms
after you stop typing, always from the server.

### API

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/quotations` | List, scoped per §8 |
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

## 12. The printed document

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

## 13. What is not built yet

| Gap | Impact |
|---|---|
| **Printed document does not match the template layout** | The source is A4 **landscape**, with a different header, two tinted line-item tables and a CHECKER box. Ours is a portrait reflow: right content and arithmetic, wrong layout. Deferred — gap list in `QUOTATION_DOCUMENT_ANALYSIS.md` §9 |
| **PIC Finance, sales phone/team not stored** | Those template fields are omitted from the printed document |
| **No dedicated approval queue screen** | Approvers filter the list instead |
| **No version-history browsing UI** | Versions are stored and visible in the audit trail, but not diffable |
| **No notifications** | An approver is not told a quotation is waiting |
| **No server-side PDF** | Printing relies on the browser's print dialog |
| **Building contract import** | Planned in `BUILDING_CONTRACT_IMPORT_PLAN.md`, off the critical path |
| **Not in production** | Production is still on migration 014 |

---

## 14. Where the code lives

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
