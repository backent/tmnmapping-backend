DROP TABLE IF EXISTS user_proxy_sales;

DROP INDEX IF EXISTS idx_users_role;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_sales_group_check;

ALTER TABLE users
    DROP COLUMN IF EXISTS sales_group,
    DROP COLUMN IF EXISTS can_create_quotations;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_check;

-- The legacy vocabulary cannot be recovered (the mapping is lossy: both 'user' and
-- 'author' collapsed into 'admin'). Roll back to the pre-migration default instead.
UPDATE users
SET role = CASE
    WHEN role = 'head_of_sales' THEN 'approver'
    WHEN role = 'admin'         THEN 'admin'
    ELSE 'user'
END;

ALTER TABLE users
    ALTER COLUMN role DROP NOT NULL,
    ALTER COLUMN role SET DEFAULT 'user';

ALTER TABLE users
    ALTER COLUMN role TYPE VARCHAR(20);
