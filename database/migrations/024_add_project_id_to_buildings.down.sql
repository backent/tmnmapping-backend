DROP INDEX IF EXISTS idx_buildings_project_id;
ALTER TABLE buildings DROP CONSTRAINT IF EXISTS fk_buildings_project;
ALTER TABLE buildings DROP COLUMN IF EXISTS project_id;
