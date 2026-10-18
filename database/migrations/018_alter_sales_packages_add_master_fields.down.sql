DROP INDEX IF EXISTS idx_sales_packages_status;

ALTER TABLE sales_packages
    DROP CONSTRAINT IF EXISTS sales_packages_counts_non_negative;

ALTER TABLE sales_packages
    DROP CONSTRAINT IF EXISTS sales_packages_status_check;

ALTER TABLE sales_packages
    DROP CONSTRAINT IF EXISTS unique_sales_package_code;

ALTER TABLE sales_packages
    DROP COLUMN IF EXISTS impressions,
    DROP COLUMN IF EXISTS traffic,
    DROP COLUMN IF EXISTS screen_count,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS package_code;
