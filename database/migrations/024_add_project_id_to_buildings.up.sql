-- Link a building to its project.
--
-- Added ALONGSIDE buildings.project_name, which is not dropped and must keep ERP's
-- exact spelling: letters_of_intent keeps syncing from ERP after the building cutover
-- and the dashboard joins it on text --
--   LEFT JOIN buildings b ON b.project_name = l.building_project
-- (repositories/dashboard/repository_dashboard_impl.go). A rename that diverged from
-- ERP's spelling would silently empty the LOI dashboard for those buildings.
--
-- So project_id is the internal relationship and project_name stays the ERP
-- correlation key. They coexist on purpose.
--
-- Nullable, and left nullable: 24 of the 3,771 buildings have no project name at all,
-- and a building with no project must stay loadable. Backfill is migration 025.
ALTER TABLE buildings
    ADD COLUMN IF NOT EXISTS project_id BIGINT;

ALTER TABLE buildings
    DROP CONSTRAINT IF EXISTS fk_buildings_project;

-- SET NULL, not CASCADE: deleting a project must never delete buildings. Everything
-- commercial hangs off buildings.id -- prices, package membership and restrictions all
-- cascade from it -- so a cascade here would silently destroy far more than a project.
ALTER TABLE buildings
    ADD CONSTRAINT fk_buildings_project
    FOREIGN KEY (project_id) REFERENCES building_projects(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_buildings_project_id ON buildings(project_id);
