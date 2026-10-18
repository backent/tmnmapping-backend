-- A sales package carries its own price.
--
-- Until now a package price lived in rate_card_package_prices, versioned alongside
-- building prices. Pricing a package therefore meant: create the package, open a
-- rate card draft, download a template, fill it, import it, publish. Seven steps
-- through a spreadsheet to set one number, and the result was that no package on
-- any rate card ever got a price -- the path people were expected to use most was
-- the one that did not work.
--
-- The price moves onto the package itself. Nothing is lost: an approved quotation
-- has never re-read a rate card. It stores its own pricing snapshot at submit and
-- reading it never recalculates, so historical quotations keep their figures
-- whatever happens to the price here.

ALTER TABLE sales_packages
    ADD COLUMN IF NOT EXISTS price_idr_per_week NUMERIC(18,0) NOT NULL DEFAULT 0;

-- Carry across anything already priced on the CURRENT rate card, so no existing
-- package silently loses a price it had.
UPDATE sales_packages p
SET price_idr_per_week = rp.price_idr_per_week
FROM rate_card_package_prices rp
JOIN rate_card_versions v ON v.id = rp.rate_card_version_id
WHERE rp.sales_package_id = p.id
  AND v.status = 'current';

-- Whole rupiah, never negative. Zero is allowed and means "not priced yet", which
-- the quotation service refuses to quote rather than selling for nothing.
ALTER TABLE sales_packages
    ADD CONSTRAINT sales_packages_price_non_negative
    CHECK (price_idr_per_week >= 0);
