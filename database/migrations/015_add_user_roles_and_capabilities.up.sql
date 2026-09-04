-- Phase 0: role vocabulary + quotation capabilities.
-- See docs/QUOTATION_FEATURE_ANALYSIS.md §5.3 and docs/PHASE_0_ROLES_PROGRESS.md.

-- 'head_of_business_control' is 24 characters and does not fit VARCHAR(20).
ALTER TABLE users
    ALTER COLUMN role TYPE VARCHAR(40);

-- Backfill the legacy vocabulary (admin/author/approver/guest/user) onto the new one.
-- Continuity over least privilege: anything unrecognised becomes 'admin' so that no
-- existing account is locked out of the app by this migration. Roles are then
-- narrowed deliberately, per user, after deploy.
UPDATE users
SET role = CASE
    WHEN role = 'approver' THEN 'head_of_sales'
    WHEN role = 'guest'    THEN 'sales'
    WHEN role IN ('sales', 'head_of_sales', 'head_of_business_control', 'ceo') THEN role
    ELSE 'admin'
END;

ALTER TABLE users
    ALTER COLUMN role SET DEFAULT 'sales',
    ALTER COLUMN role SET NOT NULL;

ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'sales', 'head_of_sales', 'head_of_business_control', 'ceo'));

-- Capabilities, orthogonal to role: a Head of Sales both approves and owns quotations.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS can_create_quotations BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS sales_group VARCHAR(30);

ALTER TABLE users
    ADD CONSTRAINT users_sales_group_check
    CHECK (sales_group IS NULL OR sales_group IN ('sales_team', 'everyone_sales', 'freelancer'));

-- Roles that own quotations get the capability by default; admins do not sell.
UPDATE users
SET can_create_quotations = TRUE,
    sales_group = 'sales_team'
WHERE role IN ('sales', 'head_of_sales');

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Proxy entry allow-list: actor_user_id may create quotations owned by owner_user_id.
CREATE TABLE IF NOT EXISTS user_proxy_sales (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id BIGINT NOT NULL,
    owner_user_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT unique_user_proxy_sales UNIQUE (actor_user_id, owner_user_id),
    CONSTRAINT user_proxy_sales_no_self CHECK (actor_user_id <> owner_user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_proxy_sales_actor ON user_proxy_sales(actor_user_id);
CREATE INDEX IF NOT EXISTS idx_user_proxy_sales_owner ON user_proxy_sales(owner_user_id);
