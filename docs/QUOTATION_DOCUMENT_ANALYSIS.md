# Quotation document — analysis of the real TMN template

**Status:** Analysis only. No code changed.
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

## 4. ⚠️ A placement may not be a list of buildings

The document describes the placement as **"Package A — 950 — Apartment"** and the
bonus as **"Bonus — 70 — Retail & Hotel"**. That reads as *a screen count within a
building category*, not a set of named buildings.

Our model (and the prototype's) offers placement as **either** individual buildings
**or** one named sales package. Neither matches "950 screens in Apartments" directly.

Three readings, and they lead to different wizards:

1. **"Package A" is a named sales package**, 950 is its screen count, and Apartment is
   a label. Then our model already fits and the number is derived.
2. **The seller picks a category and a screen count**, and the system chooses
   buildings. That is a different selection mode we have not built.
3. **The seller picks buildings**, and the document summarises them as a count and a
   dominant category. Then the summary is a rendering concern only.

Reading 3 is most consistent with clause 7 (below), but this needs answering by the
business, not inferred.

---

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

**Feeds Phase 5 (document):** the field list in §5 and the terms in §6.2 are the
template. Note the appendix is *not* part of it.

**Blocking questions raised by this document:**

1. Does approval route on the **customer** discount (65%) or the **effective**
   discount (70.44%)? §3.
2. Is a placement a set of buildings, a named package, or a category plus a screen
   count? §4.
3. Are the per-quotation contact fields (Attention To, Job Title, Handphone, Email)
   required, and do they belong to the quotation or the customer?
4. Is VAT fixed at 11% or configurable per quotation?

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
| 8 | PDF vs browser print | Still open; this is clearly a print-ready A4 layout |
| 9 | Localization | Document is English only |
| 10 | Notifications | Still open |
