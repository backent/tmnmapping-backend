-- Building projects: the landlord side of a building.
--
-- A "project" is the site a contract is signed for -- Gading Resort Residence, not
-- its 31 individual towers. It already exists in the data as buildings.project_name,
-- a free-text column, and as building_project on the three ERP feeds, which join to
-- it by string. This promotes it to an entity.
--
-- One flat table, not project + contracts. The source workbook carries 36 projects
-- in 36 rows with no repeated project id, project name or contract number, and the
-- business does not track renewal history: a project has one current contract. Where
-- a previous value is needed, building_project_changes answers it. If contract
-- history is ever wanted as rows, splitting this table is a contained migration.
--
-- Money here is COST -- what TMN pays the landlord -- unlike every other figure in
-- this schema, which is revenue. annual_rental, company_name and contract_no are
-- gated by the building-projects.finance permission, not building-projects.view.

CREATE TABLE IF NOT EXISTS building_projects (
    id BIGSERIAL PRIMARY KEY,

    -- Assigned upstream in IRIS and required on every path, including the form.
    -- One key for one project across IRIS, the workbook and this app.
    project_id_iris VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,

    building_type VARCHAR(255),
    grade VARCHAR(100),
    pic VARCHAR(255),
    tmn_project_status VARCHAR(50),

    -- Declared counts. Deliberately NOT derived from the buildings or acquisitions
    -- rows and never cross-validated against them: the business maintains them
    -- independently. Reconciliation is a report, not a constraint.
    no_of_tower INTEGER,
    no_of_screen INTEGER,

    created_date DATE,
    remark TEXT,

    -- The current contract.
    contract_type VARCHAR(50),
    contract_no VARCHAR(100),
    contract_date DATE,
    contract_start DATE,
    contract_end DATE,
    period_month INTEGER,
    annual_rental BIGINT,
    payment_term VARCHAR(50),
    company_name VARCHAR(255),
    exclusivity VARCHAR(50),
    doc_type VARCHAR(50),
    contract_status VARCHAR(50),

    -- Set when a contract ends early. The workbook captured this but no formula
    -- read it, so a cancellation zeroed months that had already elapsed. Stored
    -- here so any future projection can bound itself by
    -- COALESCE(cancelled_at, contract_end) and leave history alone.
    cancelled_at DATE,
    cancel_last_status VARCHAR(50),
    cancel_reason TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_building_project_iris UNIQUE (project_id_iris),
    CONSTRAINT building_project_iris_not_blank CHECK (btrim(project_id_iris) <> ''),
    CONSTRAINT building_project_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT building_project_rental_non_negative CHECK (annual_rental IS NULL OR annual_rental >= 0),
    CONSTRAINT building_project_period_positive CHECK (period_month IS NULL OR period_month > 0),
    CONSTRAINT building_project_dates_ordered CHECK (
        contract_start IS NULL OR contract_end IS NULL OR contract_end >= contract_start)
);

CREATE INDEX IF NOT EXISTS idx_building_projects_name ON building_projects(name);
CREATE INDEX IF NOT EXISTS idx_building_projects_status ON building_projects(tmn_project_status);
CREATE INDEX IF NOT EXISTS idx_building_projects_pic ON building_projects(pic);
CREATE INDEX IF NOT EXISTS idx_building_projects_contract_end ON building_projects(contract_end);

-- Every change to project data, attributed and timestamped.
--
-- One row per changed FIELD, not per edit: knowing a project was touched is not
-- useful, knowing annual_rental went from 54,000,000 to 61,500,000 is. Unchanged
-- values are not written, so a partial edit produces a short readable log.
--
-- App-level, not a trigger. This schema has no triggers, and a trigger cannot see
-- the application user without SET LOCAL plumbing on every transaction. RequireAuth
-- already puts the id and role on the request context.
CREATE TABLE IF NOT EXISTS building_project_changes (
    id BIGSERIAL PRIMARY KEY,

    -- SET NULL, never CASCADE: deleting a project must not erase the record of who
    -- deleted it. project_id_iris is copied in so the history stays readable after
    -- the row it describes is gone.
    project_id BIGINT,
    project_id_iris VARCHAR(100) NOT NULL,

    actor_user_id BIGINT,
    -- The role AT THE TIME. Roles change; the audit trail should not change with
    -- them. Same reasoning as quotation_approvals.actor_role.
    actor_role VARCHAR(40),

    action VARCHAR(20) NOT NULL,
    source VARCHAR(20) NOT NULL,

    -- Groups one upload. A 2,000-row import touching five fields each is 10,000
    -- rows; without this the history is unreadable. Null for a form edit.
    batch_id UUID,

    -- Null for created and deleted, which describe the whole row.
    field VARCHAR(64),
    old_value TEXT,
    new_value TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (project_id) REFERENCES building_projects(id) ON DELETE SET NULL,
    FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT building_project_changes_action_check CHECK (
        action IN ('created', 'updated', 'deleted')),
    CONSTRAINT building_project_changes_source_check CHECK (
        source IN ('form', 'import')),
    -- An update that names no field records nothing useful.
    CONSTRAINT building_project_changes_update_has_field CHECK (
        action <> 'updated' OR field IS NOT NULL),
    -- An import writes under a batch; a form edit never does.
    CONSTRAINT building_project_changes_batch_matches_source CHECK (
        (source = 'import' AND batch_id IS NOT NULL) OR
        (source = 'form' AND batch_id IS NULL))
);

CREATE INDEX IF NOT EXISTS idx_bpc_project ON building_project_changes(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bpc_iris ON building_project_changes(project_id_iris, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bpc_batch ON building_project_changes(batch_id);
CREATE INDEX IF NOT EXISTS idx_bpc_actor ON building_project_changes(actor_user_id, created_at DESC);
