# Phase 1 — Import-driven master data

**Status:** Plan. No code written.
**Date:** 2026-09-06
**Decision that triggered this:** customers, brands and building data will be
**uploaded by users**, not synced from ERP and not hand-entered.
**Related:** [`QUOTATION_FEATURE_ANALYSIS.md`](QUOTATION_FEATURE_ANALYSIS.md) §2.2, §3, §7 ·
[`PHASE_0_ROLES_PROGRESS.md`](PHASE_0_ROLES_PROGRESS.md)

---

## 1. Analysis of `MOCK_20260807_TMN_Building.xlsx`

Three sheets, 36 data rows each (synthetic, all values prefixed `MOCK-`).

| Sheet | Rows × Cols | What it holds |
|---|---|---|
| `TMN_Project_Detail` | 36 × 250 | Landlord contract per project + a 66-month screen/rental forecast |
| `Cancelled_Terminate List` | 12 × 17 | Contracts that ended early, with a reason |
| `TMN_Building` | 36 × 18 | Screen installation progress per building |

### 1.1 ⚠️ The headline: this file contains no advertising prices

Three money columns, and all three are **money TMN pays out**, not money it charges:

| Column | What it is | Verified |
|---|---|---|
| `Annual Rental` / `Annual Rental/Building` | What TMN pays the landlord per year for the site | — |
| `Price Per Screen` | Not an entered price. It is `Annual Rental ÷ No of Screen` | 8/8 rows match exactly |
| `Contract Value` | `Annual Rental × contract years` | 5/5 rows match |

The quotation needs the **rate an advertiser pays TMN** for a slot. That is a different
number flowing in the opposite direction, and it cannot be derived from rent — margin,
utilisation and campaign length all sit in between.

**So this workbook does not supply the rate card.** Phase 2 still needs its own price
source. Naming a column `Price Per Screen` makes this easy to get wrong, which is
precisely why it is called out first here.

### 1.2 It is supply-side data, not inventory data

Compared against the existing `buildings` table:

**Overlaps (4):** building name, building type, grade, IRIS ids.

**Present in `buildings`, absent from the workbook — everything the quotation and the
map need:**
`audience` (daily traffic), `impression` (monthly impressions), `latitude`,
`longitude`, `subdistrict`, `citytown`, `province`, `cbd_area`, `sellable`,
`lcd_presence_status`, `connectivity`, `resource_type`, `competitor_*`, `images`.

The quotation appendix prints name, region/type, daily traffic and monthly
impressions per building. Three of those four are not in this file. Uploaded rows
also could not appear on the mapping page, because there are no coordinates.

**Present in the workbook, absent from `buildings` — the contract and rollout story:**
`Record Type` (New Acquisition / Additional Order / Reactivation), `Deal Type`
(Greenfield / Win-Over / Expansion), `Order Screen Number`, `Installed Screen Number`,
`Remaining Screen Installed`, `Installation Percentage`, `Installation Progress`,
`Take Over From` / `Take Over Date`, and the full contract block (`Contract No`,
`Contract Type`, `Doc. Type`, `Payment Term`, `Exclusivity`, `Contract Start`/`End`,
`Contract Status`, `Company Name`, `PIC`).

**Conclusion: this is the acquisition pipeline, not the sellable inventory master.**
It sits much closer to the existing `acquisitions` / `building_proposals` /
`letters_of_intent` tables — which already carry `external_id`, `workflow_state`,
`acquisition_person`, `building_project` and sync from ERP — than to `buildings`.

### 1.3 Structural traps for the importer

Anything parsing this file has to survive:

1. **`TMN_Project_Detail` has a two-row header**, then a **summary row at row 3**
   (`Create Date` = "36", `Remark` = "MOCK DATA ONLY"). Real data starts at row 4.
   A naive "row 1 is the header, row 2 starts data" parser silently ingests two
   garbage rows.
2. **The 66-month block is not uniform.** From 2026-07 to 2030-03 each month has four
   columns (`Confirmed Screen`, `Confirmed Rental`, `Signed Screen`, `Signed Rental`);
   from 2030-04 onward it drops to two (`Screen`, `Rental`). Column position cannot
   be assumed — the month must be read from the row-1 group header.
3. **`Cancelled_Terminate List` duplicates rows already in `TMN_Project_Detail`** with
   status `Cancelled`/`Terminated`. Two sources for one fact; decide which wins, or
   the same contract lands twice.
4. **Trailing whitespace in names** — `'Graha Puspa Puspa Jaya Demo '`. Trim on import
   or matching against existing buildings will silently miss.
5. **`Installation Percentage` is a 0–1 fraction**, not a percentage. `0.2727…` means
   27%. Storing it as-is and rendering it as "0%" is the obvious bug.
6. **`Price Per Screen` carries full float noise** (`2863636.363636…`). It is derived,
   so do not store it — recompute, or store rent as `NUMERIC(18,0)` and divide on read.
7. **No coordinates anywhere**, so nothing uploaded here can be placed on the map.

---

## 2. What the workbook means for the plan

It answers the "where does data come from" question for the **acquisition** side, and
it confirms — by omission — that the **quotation** side still has no price source.

| Data the quotation needs | Source | Status |
|---|---|---|
| Customers, brands, sales assignment | User upload | Phase 1, format not yet defined |
| Building inventory (name, region, type) | ERP sync today | Exists |
| Building traffic + impressions | ERP sync today (`audience`, `impression`) | Exists — coverage unverified |
| **Advertising rate card** | **Nothing yet** | **Still the blocker for Phase 2** |
| Contract / installation tracking | This workbook | New, and not on the quotation path |

---

## 3. ⚠️ Open decision: what does "upload building data" replace?

`buildings` is synced from ERP every 30 minutes and the scheduler also runs a full
sync on startup — the last run touched **3,731 rows**. Any upload writing to the same
columns is overwritten within half an hour, silently.

Three coherent options:

| Option | Shape | Cost |
|---|---|---|
| **A. Upload is a new table** | This workbook lands in new `building_contracts` / `building_installations` tables keyed to `buildings` by `iris_code`. ERP keeps owning `buildings`. | Lowest risk. Nothing conflicts. Does not give the quotation anything new. |
| **B. Field-level ownership** | Upload writes only columns ERP does not own; the sync is changed to never touch those. | Needs an explicit owner per column and a sync rewrite. Easy to get subtly wrong. |
| **C. Upload replaces ERP for `buildings`** | Turn the scheduler off, uploads become the source of truth. | Biggest change. Loses automatic freshness. Needs a full-coverage template including traffic, impressions and coordinates — which this workbook does not have. |

**Recommendation: A.** The workbook's content genuinely is a different entity, and
option A is the only one that does not put upload and ERP in conflict over the same
row. B and C both become real options later if the rate card upload proves the import
machinery.

---

## 4. Proposed build order

### 4.0 Shared import infrastructure (do this first)

Every importer below wants the same pipeline, and the reference prototype already
proves the shape at `/admin/imports`:

```
upload file  ->  parse  ->  validate (row by row)  ->  preview + error report
             ->  commit as a batch  ->  audit trail, re-runnable
```

Backend: `services/importer` with a per-entity parser interface, plus
`import_batches` and `import_batch_errors` tables recording who uploaded what, when,
how many rows succeeded and every rejection with its row number.

Non-negotiables, learned from the workbook:
- **Validate everything before writing anything.** A half-applied master-data import
  is worse than a rejected one.
- **Preview before commit.** The operator sees counts and rejections first.
- **Never trust column position** — resolve by header name (see §1.3 trap 2).
- Reuse the existing XLSX handling already in `services/*/…Import` rather than adding
  a second Excel library.

Permission keys: `data-import.upload`, `data-import.publish`, `data-import.audit`,
wired the same way as everything else in `models/permission.go`.

### 4.1 Customers, brands, sales assignments

Migration `016`, `017`. Three importers, published in dependency order:
customers → brands → assignments. Template columns need defining with whoever
produces the file; unlike the building workbook, no sample exists yet.

**This is the Phase 1 critical path** — the quotation wizard's step 1 has no data source
without it.

### 4.2 Rate card

Migration `018`. Still blocked: **no price source exists.** Needs a decision on where
advertising rates come from before anything can be built. Versioned tables per
`QUOTATION_FEATURE_ANALYSIS.md` §3.3 so approved quotations never re-price.

### 4.3 Contract / installation import (this workbook)

Migration `019` under option A. Off the quotation critical path — schedule it after
4.1, or in parallel if someone else picks it up. Value here is replacing whatever
manual tracking this spreadsheet represents, not enabling quotations.

---

## 5. Still open

1. **What does "upload building data" replace?** §3. Blocks 4.3, and shapes 4.0.
2. **Where do advertising rates come from?** Blocks Phase 2 entirely.
3. **Who produces the customer/brand file, and in what columns?** Blocks 4.1.
4. Is `Cancelled_Terminate List` authoritative over `TMN_Project_Detail` for ended
   contracts, or a derived view of it? §1.3 trap 3.
5. Is the 66-month forecast worth storing, or is it a reporting artifact regenerated
   each export? Storing it means ~66 rows per contract in a child table.
6. The eight ten-decisions in `QUOTATION_FEATURE_ANALYSIS.md` §7 that Phase 0 did not
   touch — tax rate and rate-card period remain the widest-blast-radius ones.
