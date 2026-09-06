# Building contract & installation import — plan

**Status:** Plan. No code written.
**Date:** 2026-09-06
**Approach:** Option A from [`PHASE_1_IMPORT_PLAN.md`](PHASE_1_IMPORT_PLAN.md) §3 —
the workbook lands in **new tables**, keyed to `buildings`, and ERP keeps owning
`buildings`.
**Source file analysed:** `MOCK_20260807_TMN_Building.xlsx`

---

## 1. Why new tables rather than writing to `buildings`

`buildings` is ERP-synced every 30 minutes, and the scheduler also runs a full sync
on startup — the last observed run touched **3,731 rows**. An upload writing the same
columns is silently overwritten within the half hour, and the operator gets no signal
that it happened.

The workbook also isn't the same entity. It shares four fields with `buildings`
(name, type, grade, IRIS ids) and carries none of `audience`, `impression`,
coordinates, region, `sellable` or `lcd_presence_status`. What it *does* carry —
contract terms, screen counts, installation progress, take-over history — has no home
in `buildings` at all.

So the two never contend for a row. `buildings` stays the sellable inventory master
that ERP owns; the new tables describe the landlord contract and the screen rollout
behind each one.

**This work is not on the quotation critical path.** It replaces whatever manual
tracking the spreadsheet represents. It does not unblock Phase 2, because the
workbook contains no advertising prices (`PHASE_1_IMPORT_PLAN.md` §1.1).

---

## 2. Schema — migration `017`

```
building_contracts
    id                     BIGSERIAL PK
    project_iris_id        VARCHAR(50)  UNIQUE NOT NULL   -- "Project ID IRIS"
    project_name           VARCHAR(255) NOT NULL
    building_id            BIGINT NULL REFERENCES buildings(id) ON DELETE SET NULL
    pic                    VARCHAR(100)
    project_status         VARCHAR(30)   -- Active/Confirmed/Installation/Expired/Cancelled/Terminated
    contract_type          VARCHAR(30)   -- Initial/Renewal/Addendum
    contract_no            VARCHAR(100)
    doc_type               VARCHAR(20)   -- MOU/PKS/PO
    contract_status        VARCHAR(30)   -- Draft/Under Review/Signed/Closed
    company_name           VARCHAR(255)
    building_type          VARCHAR(50)
    grade                  VARCHAR(30)
    exclusivity            VARCHAR(20)   -- Exclusive/Non-Exclusive
    payment_term           VARCHAR(20)   -- Monthly/Quarterly/Semesterly/Annually
    tower_count            INT
    screen_count           INT
    annual_rental_idr      NUMERIC(18,0) -- what TMN PAYS the landlord
    contract_value_idr     NUMERIC(18,0)
    period_month           INT
    contract_start         DATE
    contract_end           DATE
    contract_date          DATE
    cancelled_date         DATE
    cancellation_reason    TEXT
    last_status_before_end VARCHAR(30)
    remark                 TEXT
    created_date           DATE          -- the sheet's "Create Date"
    import_batch_id        BIGINT REFERENCES import_batches(id)
    created_at, updated_at

building_installations
    id                        BIGSERIAL PK
    iris_building_id          VARCHAR(50) UNIQUE NOT NULL   -- "IRIS Building ID"
    building_contract_id      BIGINT REFERENCES building_contracts(id) ON DELETE CASCADE
    building_id               BIGINT NULL REFERENCES buildings(id) ON DELETE SET NULL
    building_name             VARCHAR(255) NOT NULL
    building_type             VARCHAR(50)
    grade                     VARCHAR(30)
    record_type               VARCHAR(30)   -- New Acquisition/Additional Order/Reactivation
    deal_type                 VARCHAR(30)   -- Greenfield/Win-Over/Expansion
    order_screen_count        INT
    installed_screen_count    INT
    remaining_screen_count    INT
    installation_percentage   NUMERIC(5,4)  -- 0..1, NOT a percent (see §4)
    installation_progress     VARCHAR(40)
    annual_rental_idr         NUMERIC(18,0)
    take_over_from            VARCHAR(255)
    take_over_date            DATE
    estimated_installation_at DATE
    confirmation_date         DATE
    created_date              DATE
    import_batch_id           BIGINT REFERENCES import_batches(id)
    created_at, updated_at
```

**Deliberately not stored: the 66-month forecast.** ~66 rows per contract of
Confirmed/Signed screen and rental figures, regenerated on every export, that nothing
in the app would read. Open question 5 in `PHASE_1_IMPORT_PLAN.md` — decide before
building. If it is wanted, it belongs in a
`building_contract_forecasts(contract_id, month, confirmed_screens, confirmed_rental,
signed_screens, signed_rental)` child table, not in wide columns.

### Linking to `buildings`

`building_id` is **nullable and best-effort**. The importer matches on
`buildings.iris_code`, falling back to a trimmed case-insensitive name match, and
leaves it NULL when neither hits. A row that cannot be matched is still worth
importing — it is a real contract — and the link can be filled in later.

`ON DELETE SET NULL`, not CASCADE: an ERP sync removing a building must not delete
contract history.

---

## 3. Money: name the columns for what they are

`annual_rental_idr` is what TMN **pays out**. Nothing in the quotation path may read
it as a price. The column name carries `rental`, never `price`, precisely because the
source spreadsheet calls its derived cost-per-screen "Price Per Screen" and that
naming is what makes this easy to get wrong.

`price_per_screen` is **not stored** — it is `annual_rental ÷ screen_count`, verified
8/8 on the sample. Storing a derived value invites the two drifting apart.
`contract_value` **is** stored despite also being derivable (`annual_rental × years`,
5/5 verified), because it is the number that appears on the contract document.

Use `NUMERIC(18,0)` and `int64` in Go, as with every other rupiah figure.

---

## 4. Parser requirements

From `PHASE_1_IMPORT_PLAN.md` §1.3, all seven traps, restated as requirements:

1. **`TMN_Project_Detail` header is two rows deep and row 3 is a summary row.** Data
   starts at row 4. The parser must skip rows 2 and 3 explicitly rather than assuming
   "header is row 1".
2. **Resolve every column by its row-1 header text**, never by position. The monthly
   block changes width partway through (four columns per month until 2030-03, two
   after), so positional offsets are wrong for the second half of the sheet.
3. **`Cancelled_Terminate List` duplicates rows already in `TMN_Project_Detail`.**
   Decide which is authoritative before writing the importer — open question 4. The
   default assumption should be that `TMN_Project_Detail` wins and the cancelled
   sheet only contributes `cancellation_reason` and `last_status_before_end`.
4. **Trim all names.** The sample contains `'Graha Puspa Puspa Jaya Demo '`; without
   trimming, the match to `buildings` silently misses.
5. **`Installation Percentage` is a 0–1 fraction.** Store as `NUMERIC(5,4)` and
   multiply for display. Storing it raw and rendering it as a percent shows "0%".
6. **Do not store `Price Per Screen`.** §3.
7. **There are no coordinates**, so imported rows can never place a pin on the map.
   The UI must not imply otherwise.

Reuse the `spreadsheets` package built for Phase 1: `ParseSpreadsheet`,
`MapHeaderColumns`, `MissingRequiredColumns`, `ColValue`, `BuildTemplate`,
`BuildExport`. This workbook needs one addition — multi-sheet reading, since all
three sheets arrive in one file, where the Phase 1 importers each read sheet 0.

---

## 5. Import semantics

Same contract as Phase 1, which is now proven:

- **Upsert by natural key** — `project_iris_id` for contracts, `iris_building_id` for
  installations.
- **All-or-nothing.** Validate every row across all three sheets first; one bad row
  rejects the file.
- **Report every problem in one pass**, with sheet name plus the row number the
  operator sees in Excel.
- **Blank rows skipped.**

One difference: this is a **three-sheet file**, so `ImportError` needs a `sheet`
field. Add it to `web.ImportError` — the Phase 1 importers leave it empty.

Ordering within a batch: contracts first, then installations, so
`building_contract_id` resolves. The cancelled sheet is applied last, as an update
over contracts already inserted.

---

## 6. Surface

| Piece | Detail |
|---|---|
| Permissions | `building-contracts.view` (all roles), `building-contracts.manage` (admin) |
| Routes | `GET/POST /building-contracts`, `GET/PUT/DELETE /building-contracts/:id`, plus `-import`, `-export`, `-template` |
| Pages | `building-contracts.vue` list + detail showing installation rows per contract |
| Nav | Under Buildings, admin-visible |

The list is worth designing around installation progress rather than contract terms —
"what is installed, what is pending" is the question the spreadsheet exists to answer.

---

## 7. Ordering

Do this **after** Phase 1's customers/brands/assignments have been used in anger, and
after the shared import machinery has survived real files. It is independent of the
quotation path, so it can also wait behind Phase 2 without blocking anything.

**Prerequisites:** open questions 4 and 5 in `PHASE_1_IMPORT_PLAN.md` (which sheet
wins for ended contracts; whether the 66-month forecast is stored).
