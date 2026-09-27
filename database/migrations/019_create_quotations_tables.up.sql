-- Phase 3: quotations.
-- See docs/QUOTATION_FEATURE_ANALYSIS.md §1, §4.1 and
--     docs/QUOTATION_DOCUMENT_ANALYSIS.md for the verified worked example.
--
-- Money is NUMERIC(18,0): whole rupiah, no minor unit. Rates are per week, so
-- gross = rate_per_week * weeks.

CREATE TABLE IF NOT EXISTS quotations (
    id BIGSERIAL PRIMARY KEY,
    quote_number VARCHAR(50) NOT NULL,

    -- Two identities per quotation, per spec 2026-07-23.
    --   sales_user_id      : the commercial OWNER. Drives customer visibility,
    --                        approval routing and reporting.
    --   created_by_user_id : who physically entered it (proxy entry). Audit only,
    --                        and immutable for the life of the quotation.
    sales_user_id BIGINT NOT NULL,
    created_by_user_id BIGINT NOT NULL,

    customer_id BIGINT NOT NULL,
    brand_id BIGINT NOT NULL,

    -- The rate card the quotation was priced against, so an approved quotation
    -- never re-prices when a new rate card is published.
    rate_card_version_id BIGINT,

    -- Contact details are per quotation, not per customer: the same customer may be
    -- quoted through different people. Taken from the real document template.
    attention_to VARCHAR(255),
    job_title VARCHAR(255),
    contact_phone VARCHAR(50),
    contact_email VARCHAR(255),

    campaign_year INTEGER,
    valid_until DATE,

    -- The discount the salesperson entered. This is what routes the approval;
    -- effective_discount_rate below is analytics only and never routes.
    discount NUMERIC(5,2) NOT NULL DEFAULT 0,

    -- Per quotation, so an approved one keeps the rate it was priced at even if
    -- PPN changes later.
    tax_rate NUMERIC(5,4) NOT NULL DEFAULT 0.11,

    status VARCHAR(30) NOT NULL DEFAULT 'draft',
    required_approver_user_id BIGINT,
    version INTEGER NOT NULL DEFAULT 1,

    -- Pricing snapshot, recomputed server-side on every submit.
    placement_gross NUMERIC(18,0) NOT NULL DEFAULT 0,
    placement_discount_amount NUMERIC(18,0) NOT NULL DEFAULT 0,
    placement_net NUMERIC(18,0) NOT NULL DEFAULT 0,
    bonus_gross NUMERIC(18,0) NOT NULL DEFAULT 0,
    bonus_net NUMERIC(18,0) NOT NULL DEFAULT 0,
    total_gross NUMERIC(18,0) NOT NULL DEFAULT 0,
    total_net NUMERIC(18,0) NOT NULL DEFAULT 0,
    effective_discount_amount NUMERIC(18,0) NOT NULL DEFAULT 0,
    effective_discount_rate NUMERIC(7,4) NOT NULL DEFAULT 0,
    tax NUMERIC(18,0) NOT NULL DEFAULT 0,
    total_including_tax NUMERIC(18,0) NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP,

    FOREIGN KEY (sales_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE RESTRICT,
    FOREIGN KEY (brand_id) REFERENCES brands(id) ON DELETE RESTRICT,
    FOREIGN KEY (rate_card_version_id) REFERENCES rate_card_versions(id) ON DELETE RESTRICT,
    FOREIGN KEY (required_approver_user_id) REFERENCES users(id) ON DELETE SET NULL,

    CONSTRAINT unique_quote_number UNIQUE (quote_number),
    CONSTRAINT quotations_status_check CHECK (status IN (
        'draft', 'pending_manager', 'pending_business_control', 'pending_ceo',
        'returned', 'approved')),
    CONSTRAINT quotations_discount_range CHECK (discount >= 0 AND discount <= 100),
    CONSTRAINT quotations_tax_rate_range CHECK (tax_rate >= 0 AND tax_rate <= 1),
    CONSTRAINT quotations_amounts_non_negative CHECK (
        placement_gross >= 0 AND placement_net >= 0 AND bonus_gross >= 0
        AND total_gross >= 0 AND total_net >= 0 AND tax >= 0),
    CONSTRAINT quotations_approved_at_agrees CHECK (
        (status = 'approved' AND approved_at IS NOT NULL)
        OR (status <> 'approved' AND approved_at IS NULL))
);

CREATE INDEX IF NOT EXISTS idx_quotations_sales_user ON quotations(sales_user_id);
CREATE INDEX IF NOT EXISTS idx_quotations_status ON quotations(status);
CREATE INDEX IF NOT EXISTS idx_quotations_approver ON quotations(required_approver_user_id);
CREATE INDEX IF NOT EXISTS idx_quotations_customer ON quotations(customer_id);

-- Placement and Bonus are two independent selections. Each picks its own mode,
-- its own resources and its own campaign parameters.
CREATE TABLE IF NOT EXISTS quotation_selections (
    id BIGSERIAL PRIMARY KEY,
    quotation_id BIGINT NOT NULL,
    kind VARCHAR(20) NOT NULL,
    mode VARCHAR(20) NOT NULL,

    -- Set only when mode = 'package'. A package selection is exactly one package.
    sales_package_id BIGINT,
    sales_package_name VARCHAR(255),

    tvc_duration_seconds INTEGER NOT NULL DEFAULT 0,
    weeks INTEGER NOT NULL DEFAULT 0,
    spots INTEGER NOT NULL DEFAULT 0,

    gross_price NUMERIC(18,0) NOT NULL DEFAULT 0,
    traffic BIGINT NOT NULL DEFAULT 0,
    impressions BIGINT NOT NULL DEFAULT 0,
    screen_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE,
    FOREIGN KEY (sales_package_id) REFERENCES sales_packages(id) ON DELETE RESTRICT,
    CONSTRAINT unique_quotation_selection_kind UNIQUE (quotation_id, kind),
    CONSTRAINT quotation_selections_kind_check CHECK (kind IN ('placement', 'bonus')),
    CONSTRAINT quotation_selections_mode_check CHECK (mode IN ('building', 'package')),
    CONSTRAINT quotation_selections_package_present CHECK (
        (mode = 'package' AND sales_package_id IS NOT NULL)
        OR (mode = 'building' AND sales_package_id IS NULL))
);

CREATE INDEX IF NOT EXISTS idx_quotation_selections_quotation ON quotation_selections(quotation_id);

-- The buildings behind a selection, SNAPSHOT at submit time.
--
-- Names, types and prices are copied rather than joined, so editing or re-pricing a
-- building later cannot rewrite what a historical quotation was sold. building_id is
-- kept for traceability but is not what the document renders from.
CREATE TABLE IF NOT EXISTS quotation_selection_items (
    id BIGSERIAL PRIMARY KEY,
    quotation_selection_id BIGINT NOT NULL,
    building_id BIGINT,

    building_name VARCHAR(255) NOT NULL,
    building_iris_code VARCHAR(100),
    building_type VARCHAR(50),
    citytown VARCHAR(100),

    unit_price_idr NUMERIC(18,0) NOT NULL DEFAULT 0,
    traffic BIGINT NOT NULL DEFAULT 0,
    impressions BIGINT NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (quotation_selection_id) REFERENCES quotation_selections(id) ON DELETE CASCADE,
    FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_quotation_selection_items_selection
    ON quotation_selection_items(quotation_selection_id);

-- An immutable snapshot per submitted version. Resubmitting after a return
-- increments quotations.version and appends a row here.
CREATE TABLE IF NOT EXISTS quotation_versions (
    id BIGSERIAL PRIMARY KEY,
    quotation_id BIGINT NOT NULL,
    version INTEGER NOT NULL,
    snapshot JSONB NOT NULL,
    required_approver_user_id BIGINT,
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE,
    FOREIGN KEY (required_approver_user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT unique_quotation_version UNIQUE (quotation_id, version)
);

-- The approval audit trail. Every proxy action stays visible here.
CREATE TABLE IF NOT EXISTS quotation_approvals (
    id BIGSERIAL PRIMARY KEY,
    quotation_id BIGINT NOT NULL,
    version INTEGER NOT NULL,
    actor_user_id BIGINT,
    actor_role VARCHAR(40),
    action VARCHAR(20) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE,
    FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT quotation_approvals_action_check CHECK (
        action IN ('submitted', 'resubmitted', 'approved', 'returned')),
    -- Returning without saying why is the one thing the spec forbids outright,
    -- so the database enforces it too.
    CONSTRAINT quotation_approvals_return_needs_comment CHECK (
        action <> 'returned' OR (comment IS NOT NULL AND btrim(comment) <> ''))
);

CREATE INDEX IF NOT EXISTS idx_quotation_approvals_quotation ON quotation_approvals(quotation_id);
