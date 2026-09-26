-- Phase 1: advertiser customers, their brands, and the sales PIC assignment.
-- See docs/PHASE_1_IMPORT_PLAN.md and docs/QUOTATION_FEATURE_ANALYSIS.md §2.2, §4.1.
--
-- These are ADVERTISER customers. They are unrelated to `mother_brands`, which is a
-- retail/competitor concept used by the POI map layer.

CREATE TABLE IF NOT EXISTS customers (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    industry VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_customer_code UNIQUE (code),
    CONSTRAINT customers_status_check CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_customers_name ON customers(name);
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);

CREATE TABLE IF NOT EXISTS brands (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL,
    customer_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- RESTRICT, not CASCADE: a customer with brands must be dealt with explicitly.
    -- Quotations will reference brands, so silent cascade deletion would rewrite history.
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE RESTRICT,
    CONSTRAINT unique_brand_code UNIQUE (code),
    CONSTRAINT brands_status_check CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_brands_customer_id ON brands(customer_id);
CREATE INDEX IF NOT EXISTS idx_brands_name ON brands(name);
CREATE INDEX IF NOT EXISTS idx_brands_status ON brands(status);

-- One sales PIC per customer+brand pair. This is what scopes the quotation wizard's
-- customer list -- not users.sales_group, which is a reporting attribute only.
CREATE TABLE IF NOT EXISTS sales_assignments (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL,
    brand_id BIGINT NOT NULL,
    sales_user_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    registration_date DATE,
    expiry_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
    FOREIGN KEY (brand_id) REFERENCES brands(id) ON DELETE CASCADE,
    -- RESTRICT: deleting a user who still holds accounts must be a deliberate act.
    FOREIGN KEY (sales_user_id) REFERENCES users(id) ON DELETE RESTRICT,

    CONSTRAINT unique_sales_assignment UNIQUE (customer_id, brand_id),
    CONSTRAINT sales_assignments_status_check CHECK (status IN ('active', 'inactive')),
    CONSTRAINT sales_assignments_date_order CHECK (
        registration_date IS NULL OR expiry_date IS NULL OR expiry_date >= registration_date
    )
);

CREATE INDEX IF NOT EXISTS idx_sales_assignments_sales_user ON sales_assignments(sales_user_id);
CREATE INDEX IF NOT EXISTS idx_sales_assignments_customer ON sales_assignments(customer_id);
CREATE INDEX IF NOT EXISTS idx_sales_assignments_brand ON sales_assignments(brand_id);
