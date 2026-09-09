# Quotation document — analysis of the real TMN template

**Status:** Analysis, and now also the spec for the rework. The document was built
from this; its layout does **not** match the template — see §9.
**Date:** 2026-09-07
**Source:** `260101. Template Quotation 2026.pdf` — a filled example, one page.
**Related:** [`QUOTATION_FEATURE_ANALYSIS.md`](QUOTATION_FEATURE_ANALYSIS.md) §1.2, §6.5, §7 ·
[`PHASE_2_RATE_CARD_PLAN.md`](PHASE_2_RATE_CARD_PLAN.md)

This is the first artefact from the actual business rather than the reference
prototype. Every figure on it reconciles, so the pricing model can be treated as
confirmed rather than inferred.

---

## 1. The worked example, verified

| Line | Rate / week | Weeks | Gross | Discount | Nett |
|---|---|---|---|---|---|
| **Package A** (Apartment, 950) | 380,000,000 | 4 | 1,520,000,000 | 65% | 532,000,000 |
| **Bonus** (Retail & Hotel, 70) | 70,000,000 | 4 | 280,000,000 | FREE | 0 |
| **TOTAL** | | | **1,800,000,000** | | **532,000,000** |

```
VAT @ 11%                58,520,000
Total (VAT included)    590,520,000
Saving Value          1,268,000,000
Total Discount             70.44%
```

All eight figures check out arithmetically:

```
gross        = rate_per_week * weeks
nett         = gross * (1 - customer_discount)      # placement only
total_gross  = placement_gross + bonus_gross         # bonus counts toward gross
total_nett   = placement_nett                        # bonus is free
saving       = total_gross - total_nett
effective    = saving / total_gross
vat          = total_nett * 0.11
```

Both TVC lines are `15 Secs` at `180 Spot` per day per screen, over `4 Weeks`.

---

## 2. What this settles

### 2.1 ✅ VAT is 11%

`58,520,000 / 532,000,000 = 11.00%` exactly. This closes open decision 1 in
`QUOTATION_FEATURE_ANALYSIS.md` §7, where the prototype's self-labelled *simulated*
6% was flagged as the widest-blast-radius unknown.

VAT is charged on **nett**, not gross.

Still worth confirming whether the rate is fixed or configurable per quotation —
Indonesian PPN moved to 12% for some categories, and a hardcoded 11% would need a
code change rather than a settings change.

### 2.2 ✅ The rate is per week, and multiplies straight through

The column is literally `Gross Rate / Week`, and `380,000,000 × 4 = 1,520,000,000`.
No division by four anywhere on the document.

This confirms the rate card rename from `price_idr_per_4_weeks` to
`price_idr_per_week`, and the formula `gross = price × weeks`. The prototype's
`gross = price × weeks / 4` reflected its own 4-week rate card, not this business.

### 2.3 ✅ Placement / Bonus behaves exactly as specified

Bonus is free, the discount applies to placement only, and bonus still counts toward
`total_gross` — which is what drags the effective discount up. The spec's model
holds.

---

## 3. ⚠️ Where the document contradicts the spec

The document shows **Total Discount 70.44%** with **"NEED APPROVAL!"** beside it.

70.44% is the **effective** discount, inflated by the free bonus. The **customer
discount is 65%**, shown on the Package A line.

Spec `2026-07-23-final-approval-proxy-entry-design.md` is explicit:

> Effective discount, including the value of a free Bonus, remains a commercial-risk
> and analytics metric only. **It never determines the approver.**

Under that rule this quotation routes on 65% → **Head of Sales** (≤65% inclusive).
The document instead flags approval against 70.44%, which under the same bands is
**Head of Business Control**.

**Same quotation, two different approvers.** This must be settled before Phase 3
builds `ResolveApprovalRoute`: it is one line of code and a materially different
approval path in practice.

It is also possible "NEED APPROVAL!" is a blanket flag meaning *any* discount needs
sign-off, and carries no band information at all. The example alone cannot
distinguish the two readings.

---

## 4. ✅ A placement is buildings OR one named package — settled

**Resolved 2026-09-08** by reading the reference implementation. This section
previously listed three possibilities; the prototype answers it outright.

`lib/types.ts`:

```ts
export type PlacementMode = "building" | "package";
```

`lib/quotation.ts` selects the resource pool by mode, then prices **both the same
way** — a package is not a special pricing rule, it is simply a resource that carries
its own price:

```ts
const resources = selection.mode === "building" ? references.buildings
                : selection.mode === "package"  ? references.packages
                : undefined;

const expected = Math.round(
  selected.reduce((sum, r) => sum + (r?.priceIdr ?? 0), 0) * (weeks / 4)
);
if (selection.grossPrice !== expected) errors[...] = basePriceMismatch;
```

So "category plus screen count" is **not** a selection mode. On the example document,
`950` and `Apartment` are attributes of Package A shown for information — not values
the seller entered.

Three details that follow:

1. **A package selection is exactly one package.** Enforced explicitly:
   `(selection.mode === "package" && ids.length !== 1)` is a validation error.
   Buildings are 1..n. A package cannot be combined with another package, or mixed
   with loose buildings, inside one selection. Placement and Bonus are separate
   selections, so each chooses its mode independently — which is exactly the shape of
   the example document.
2. **A package carries its own price, traffic and impressions**, set independently
   rather than derived from its member buildings. This is why the example's per-screen
   economics do not match the rate card sum, and why its two lines differ from each
   other (400,000 vs 1,000,000 per screen per week).
3. **A package stores `buildingIds`**, which is what feeds the separate Building List
   attachment of clause 7, and matches the `rate_card_package_buildings` snapshot our
   schema freezes at publish.

Our data model already supports this: `rate_card_building_prices` and
`rate_card_package_prices` are two pools of priced resources, and
`quotation_selections.mode` chooses between them.

### 4.1 What package pricing still needs

Supporting the mode is not the same as being able to use it. Three gaps:

| Need | Where | Status (re-verified 2026-09-09) |
|---|---|---|
| Package prices | `rate_card_package_prices` | ⛔ Table exists, **still empty on every rate card version**. Import pipeline is built (template → import → export → edit → delete); only the data is missing, and only the business can supply it. |
| Package traffic + impressions | `sales_packages` | ✅ Columns added by migration `018`. The quotation service reads them into the selection. **But every row is 0** — see below. |
| Package screen count | `sales_packages` | ✅ Column added by migration `018`, read by the service. **Every row is 0** — see below. |
| A way to enter those figures | `frontend/src/pages/sales-package-form.vue` | ⛔ **Missing.** The form exposes only `Name`. `package_code`, `status`, `description`, `screen_count`, `traffic` and `impressions` have no input, so they cannot be set at all through the UI. |

⚠️ **Consequence:** the service copies `screen_count`, `traffic` and `impressions`
from the package onto the quotation selection. All 10 packages on staging have zeros,
so a package quotation would print **"0 screens"** and a zero audience even after the
prices are loaded. Loading prices alone is not enough to make package mode usable.

Plus the master-data fields already flagged in `QUOTATION_FEATURE_ANALYSIS.md` §3.2:
`package_code`, `status`, `description`.

So `sales_packages` needs a migration adding: `package_code` (unique), `status`,
`description`, `screen_count`, `traffic`, `impressions`. That is a prerequisite for
Phase 3's package mode, and it is small — but the **package prices themselves are a
data gap only the business can close**, in the same way the building prices arrived
as a spreadsheet.

## 5. Fields the document has that we do not model

| Group | Fields |
|---|---|
| **Sales side** | Sales Person, Team, No. HP, Email |
| **Client side** | Attention To, Job Title, Company Name, Brand Name, Handphone, Email |
| **Document** | No. Quotation, Date Prepared, **Validation Date**, Campaign Plan (year), Saving Value |
| **Finance** | PIC Finance Department — Name, Email, HP |
| **Line item** | screen count, building-type label, Spot/Day/Screen |

Notes:

- **Validation Date** is 30 days after Date Prepared (22 Jun → 22 Jul 2026). Quotation
  validity is a real field, not a term-and-condition sentence.
- **Company Name / Brand Name** map onto the `customers` / `brands` tables built in
  Phase 1. The contact fields (Attention To, Job Title, Handphone, Email) do not exist
  yet and are per-quotation, not per-customer — the same customer may be quoted
  through different contacts.
- **Saving Value** and **Total Discount** are derived, and should be computed rather
  than stored.
- **Campaign Plan** is a year, and the payment terms reference *"unable to launch the
  ads at latest 30 Dec 2026"* — so a quotation is scoped to a campaign year.
- **PIC Finance Department** is blank (`0`) in the example. Possibly optional, or
  filled at invoicing.

---

## 6. Two structural differences from the prototype

### 6.1 The building list is a separate document

> 7. Building List & Summary Media Performance Attached in Separated Document

The prototype prints a per-building appendix (name, region/type, daily traffic,
monthly impressions) **inside** the quotation. TMN does not: it ships a separate
attachment.

That materially reduces §6.5 of the analysis doc. It also means daily traffic and
monthly impressions may not belong on the quotation at all — which is worth knowing,
because `buildings.audience` / `buildings.impression` coverage was flagged as a data
risk specifically for that appendix.

### 6.2 Terms and conditions are substantial and static

Ten numbered clauses plus payment terms. They include real commercial rules:

- **Payment:** 50% down payment before campaign start (payable within 30 days of
  invoice), 50% after campaign end.
- **Cancellation:** 50% charged if cancelled before PO, 100% after PO.
- **Screen error tolerance:** up to 5% of total screens may be in error.
- **Content:** delivered on the Monday one week before placement.
- **Invoicing documents:** Signed Quotation, Purchase Order, NPWP.

These are static text today, but they are versioned in practice — a quotation signed
under 2026 terms must keep those terms. Storing a `terms_version` alongside the
quotation is cheaper than reconstructing which wording applied.

---

## 7. What to do with this

**Confirmed, no action needed:** per-week rate card, `gross = rate × weeks`,
placement/bonus mechanics, VAT on nett.

**Feeds Phase 3 (`services/quotation`):**
- Tax rate constant 11%, applied to nett — decide fixed or configurable.
- Pricing functions can be tested against this exact example as a fixture. It is a
  real, fully reconciled worked case, which is worth more than invented numbers.
- Placement and Bonus each select **either** 1..n buildings **or** exactly one
  package, priced identically by summing the selected resources. §4.
- `sales_packages` needs the columns in §4.1 before package mode can work.

**Feeds Phase 5 (document):** the field list in §5 and the terms in §6.2 are the
template. Note the appendix is *not* part of it.

**Blocking questions raised by this document:**

1. Does approval route on the **customer** discount (65%) or the **effective**
   discount (70.44%)? §3. *Still the one that changes code shape.*
2. ~~Is a placement buildings, a package, or a category plus count?~~
   ✅ **Answered — buildings or one named package.** §4.
3. Are the per-quotation contact fields (Attention To, Job Title, Handphone, Email)
   required, and do they belong to the quotation or the customer?
4. Is VAT fixed at 11% or configurable per quotation?
5. **Where do package prices come from?** §4.1. Buildings arrived as a spreadsheet;
   packages have no equivalent yet, and package mode cannot be used without them.

---

## 8. Open decisions, refreshed

From `QUOTATION_FEATURE_ANALYSIS.md` §7:

| # | Decision | Status |
|---|---|---|
| 1 | Tax rate | ✅ **11%, on nett** |
| 2 | Rate card period | ✅ **Per week** (confirmed twice: rate card file and this document) |
| 3 | Approval thresholds 65% / 75% | Unconfirmed — and §3 above complicates it |
| 4 | Approver identities | Still open |
| 5 | Customer / brand source | ✅ **User upload** |
| 6 | Quotable inventory (`sellable`, `lcd_presence_status`) | Still open |
| 7 | `audience` / `impression` semantics | Possibly moot — §6.1 |
| 8 | PDF vs browser print | Print CSS fixed 2026-09-09; **layout does not match the template — see §9** |
| 9 | Localization | Document is English only |
| 10 | Notifications | Still open |

---

## 9. ⏸ Deviations from the source template — DEFERRED

**Status:** known, and deliberately left for now. Raised 2026-09-09; the user will
come back to it. **Do not treat the current document as finished.**

The built document (`frontend/src/pages/quotation-document.vue`) is a faithful
*portrait reflow* of the template's content. Every field and clause is present and
the arithmetic matches. But the **layout is not the template's**, and for a document
a client signs, resembling the real thing matters.

Source of truth: `260101. Template Quotation 2026.pdf`, verified 2026-09-09 from its
`/MediaBox`: **842 × 595 pt = 297 × 210 mm — A4 LANDSCAPE**, one page.

### 9.1 The gap list

| # | Source template | What we built | Severity |
|---|---|---|---|
| 1 | **A4 landscape** | A4 portrait | **Structural** — most of the rest follows from this |
| 2 | TMN logo, plus "Member company of" with the Focus Media Group and Sinarmas marks | No logos at all | High — it is letterhead |
| 3 | Separate bordered **CHECKER** box top-right: black bar, Saving Value, Total Discount on **yellow**, NEED APPROVAL! on **red** | Values scattered into the letterhead and party rows; plain red text | High |
| 4 | Title and both party columns inside **one bordered box** | Two borderless tables side by side | Medium |
| 5 | **Two separate line-item tables**, each with its own header row — placement tinted **blue** and labelled with the package name ("Package A"), bonus tinted **peach** | One table, grey header, placement and bonus as two rows | High |
| 6 | First column is screen count over building type ("950 / Apartment", "70 / Retail & Hotel") | Package name or "N buildings", screen count beneath | Medium |
| 7 | Money cells carry a left-aligned `IDR` prefix with the amount right-aligned in the same cell | `Rp` prefix, one right-aligned run | Low |
| 8 | TOTAL row highlights the summed weeks in **yellow** | Plain bold row | Low |
| 9 | Terms, Terms of Payment, Documents for Invoicing and PIC Finance all inside one bordered box | Separate ungrouped sections | Medium |
| 10 | **PIC Finance Department: Name / Email / HP** | Omitted — not stored anywhere | Medium — data decision, see §5 |
| 11 | Sales **Team**, **No. HP**, **Email** in the left party column | Omitted — not stored anywhere | Medium — data decision, see §5 |
| 12 | Totals block is a bordered, blue-tinted box | Borderless rows | Low |

### 9.2 What is NOT wrong

Stated so the rework does not "fix" what is already right:

- The **arithmetic** matches the template exactly, including VAT on nett.
- **Total Discount shows the effective rate** (70.44% on the template), not the
  customer discount (65%). Ours does the same. See §3 — these must not be swapped.
- All **ten clauses** are present and verbatim, typos included.
- The **bonus reads FREE** in both the Discount and Total Nett columns.

### 9.3 When picking this up

1. Landscape first (`@page { size: A4 landscape }`). The column widths only make
   sense at 297mm, and items 4, 5 and 9 depend on it.
2. Items 10 and 11 are **data**, not layout — they need somewhere to live before the
   document can render them. Schema decision, not CSS.
3. Logos need asset files from the business, including the two partner marks.
4. Verify by rendering a real PDF and comparing side by side with the source, not by
   reading the markup. `qlmanage -t -s 2400 -o . file.pdf` renders a page to PNG on
   macOS without installing anything.
