-- Phase 2: the advertising rate card.
-- See docs/QUOTATION_FEATURE_ANALYSIS.md §3.3 and docs/PHASE_2_RATE_CARD_PLAN.md.
--
-- Prices here are what an ADVERTISER PAYS TMN. They are unrelated to the landlord
-- rent in the building contract workbook, which is money flowing the other way.
--
-- Every price is a PER-WEEK rate, matching the rate card source spreadsheet, whose
-- own arithmetic derives a weekly figure (Total Cost / Month / 4). The quotation
-- computes
--     gross = price_idr_per_week * weeks
-- Storing the same period the source uses means a number in the app can be checked
-- against the spreadsheet directly, with no factor of four in between.

CREATE TABLE IF NOT EXISTS rate_card_versions (
    id BIGSERIAL PRIMARY KEY,
    version_code VARCHAR(50) NOT NULL,
    description TEXT,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    published_by_user_id BIGINT,
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (published_by_user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT unique_rate_card_version_code UNIQUE (version_code),
    CONSTRAINT rate_card_versions_status_check CHECK (status IN ('draft', 'current', 'historical')),
    CONSTRAINT rate_card_versions_published_together CHECK (
        (status = 'draft' AND published_at IS NULL)
        OR (status <> 'draft' AND published_at IS NOT NULL)
    )
);

-- At most one current version. Enforced by the database rather than by the service
-- alone, because two "current" rate cards would make every quotation ambiguous.
CREATE UNIQUE INDEX IF NOT EXISTS idx_rate_card_versions_single_current
    ON rate_card_versions (status) WHERE status = 'current';

CREATE INDEX IF NOT EXISTS idx_rate_card_versions_status ON rate_card_versions(status);

CREATE TABLE IF NOT EXISTS rate_card_building_prices (
    id BIGSERIAL PRIMARY KEY,
    rate_card_version_id BIGINT NOT NULL,
    building_id BIGINT NOT NULL,
    price_idr_per_week NUMERIC(18,0) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (rate_card_version_id) REFERENCES rate_card_versions(id) ON DELETE CASCADE,
    -- RESTRICT: a published price list must not lose rows underneath it. The ERP
    -- sync only creates and updates buildings, so this never blocks it.
    FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE RESTRICT,
    CONSTRAINT unique_rate_card_building UNIQUE (rate_card_version_id, building_id),
    CONSTRAINT rate_card_building_price_non_negative CHECK (price_idr_per_week >= 0)
);

CREATE INDEX IF NOT EXISTS idx_rate_card_building_prices_version
    ON rate_card_building_prices(rate_card_version_id);

CREATE TABLE IF NOT EXISTS rate_card_package_prices (
    id BIGSERIAL PRIMARY KEY,
    rate_card_version_id BIGINT NOT NULL,
    sales_package_id BIGINT NOT NULL,
    price_idr_per_week NUMERIC(18,0) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (rate_card_version_id) REFERENCES rate_card_versions(id) ON DELETE CASCADE,
    FOREIGN KEY (sales_package_id) REFERENCES sales_packages(id) ON DELETE RESTRICT,
    CONSTRAINT unique_rate_card_package UNIQUE (rate_card_version_id, sales_package_id),
    CONSTRAINT rate_card_package_price_non_negative CHECK (price_idr_per_week >= 0)
);

CREATE INDEX IF NOT EXISTS idx_rate_card_package_prices_version
    ON rate_card_package_prices(rate_card_version_id);

-- The package composition AS PRICED.
--
-- sales_package_buildings is live master data and uses ON DELETE CASCADE, so editing
-- or deleting a package would silently rewrite what a historical quotation was sold.
-- This table freezes the membership at the moment the version was published.
CREATE TABLE IF NOT EXISTS rate_card_package_buildings (
    id BIGSERIAL PRIMARY KEY,
    rate_card_version_id BIGINT NOT NULL,
    sales_package_id BIGINT NOT NULL,
    building_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (rate_card_version_id) REFERENCES rate_card_versions(id) ON DELETE CASCADE,
    FOREIGN KEY (sales_package_id) REFERENCES sales_packages(id) ON DELETE RESTRICT,
    FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE RESTRICT,
    CONSTRAINT unique_rate_card_package_building
        UNIQUE (rate_card_version_id, sales_package_id, building_id)
);

CREATE INDEX IF NOT EXISTS idx_rate_card_package_buildings_version
    ON rate_card_package_buildings(rate_card_version_id, sales_package_id);
