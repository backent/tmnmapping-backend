-- Phase 3 prerequisite: make a sales package a first-class priced resource.
-- See docs/QUOTATION_DOCUMENT_ANALYSIS.md §4.1.
--
-- A package is not a view over its member buildings. The reference implementation
-- gives it its own price, traffic, impressions and screen count, set independently,
-- which is why a real quotation's package rate does not equal the sum of the rate
-- card entries for the buildings inside it.
--
-- The price itself lives in rate_card_package_prices, versioned like every other
-- price. Only the descriptive attributes belong here.

ALTER TABLE sales_packages
    ADD COLUMN IF NOT EXISTS package_code VARCHAR(50),
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS screen_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS traffic INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS impressions INTEGER NOT NULL DEFAULT 0;

-- Backfill a code for existing rows so the column can be made unique. SP-0001 style,
-- ordered by id, which keeps existing packages stable and readable.
UPDATE sales_packages
SET package_code = 'SP-' || LPAD(id::text, 4, '0')
WHERE package_code IS NULL OR package_code = '';

ALTER TABLE sales_packages
    ALTER COLUMN package_code SET NOT NULL;

ALTER TABLE sales_packages
    ADD CONSTRAINT unique_sales_package_code UNIQUE (package_code);

ALTER TABLE sales_packages
    ADD CONSTRAINT sales_packages_status_check CHECK (status IN ('active', 'inactive'));

-- Screen count, traffic and impressions are counts, never negative. They are also
-- allowed to be zero: a package can exist before its figures are known.
ALTER TABLE sales_packages
    ADD CONSTRAINT sales_packages_counts_non_negative
    CHECK (screen_count >= 0 AND traffic >= 0 AND impressions >= 0);

CREATE INDEX IF NOT EXISTS idx_sales_packages_status ON sales_packages(status);
