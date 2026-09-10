ALTER TABLE sales_packages DROP CONSTRAINT IF EXISTS sales_packages_price_non_negative;
ALTER TABLE sales_packages DROP COLUMN IF EXISTS price_idr_per_week;
