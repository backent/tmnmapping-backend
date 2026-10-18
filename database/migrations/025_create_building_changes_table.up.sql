-- Every change to building data, attributed and timestamped.
--
-- This ships WITH the spreadsheet importer, not after it. A blank cell clears a value
-- on that import, so a careless upload can empty columns across thousands of rows;
-- without this table the only recovery is a database restore. The change log is the
-- undo trail that makes blank-means-clear survivable.
--
-- Same shape as building_project_changes (migration 023), deliberately: one row per
-- changed FIELD, app-level rather than a trigger, and the history outlives the row it
-- describes.
CREATE TABLE IF NOT EXISTS building_changes (
    id BIGSERIAL PRIMARY KEY,

    -- SET NULL, never CASCADE. Buildings cascade to prices, package membership and
    -- restrictions, so a building row is already dangerous to delete; the record of
    -- what happened to it must not go with it.
    building_id BIGINT,
    -- Denormalised so the history stays readable afterwards. external_building_id is
    -- the import key; 24 buildings have none, so the name is kept as well.
    external_building_id VARCHAR(100),
    building_name VARCHAR(255) NOT NULL,

    actor_user_id BIGINT,
    -- The role AT THE TIME. Roles change; the audit trail should not change with them.
    actor_role VARCHAR(40),

    action VARCHAR(20) NOT NULL,
    -- 'sync' exists because ERP still writes here: photos keep syncing after the
    -- cutover, on a narrowed FetchBuildings. A change nobody made by hand should say
    -- so rather than appear as an anonymous edit.
    source VARCHAR(20) NOT NULL,

    -- Groups one upload. Null for a form edit or a sync.
    batch_id UUID,

    field VARCHAR(64),
    old_value TEXT,
    new_value TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE SET NULL,
    FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT building_changes_action_check CHECK (
        action IN ('created', 'updated', 'deleted')),
    CONSTRAINT building_changes_source_check CHECK (
        source IN ('form', 'import', 'sync')),
    CONSTRAINT building_changes_update_has_field CHECK (
        action <> 'updated' OR field IS NOT NULL),
    -- An import writes under a batch; nothing else does.
    CONSTRAINT building_changes_batch_matches_source CHECK (
        (source = 'import' AND batch_id IS NOT NULL) OR
        (source <> 'import' AND batch_id IS NULL))
);

CREATE INDEX IF NOT EXISTS idx_building_changes_building ON building_changes(building_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_building_changes_external ON building_changes(external_building_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_building_changes_batch ON building_changes(batch_id);
CREATE INDEX IF NOT EXISTS idx_building_changes_actor ON building_changes(actor_user_id, created_at DESC);
