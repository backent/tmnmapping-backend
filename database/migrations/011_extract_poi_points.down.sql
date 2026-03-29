-- Re-add poi_id column to poi_points
ALTER TABLE poi_points ADD COLUMN poi_id BIGINT;

-- Restore relationships from junction table (pick first POI if multiple)
UPDATE poi_points pp SET poi_id = (
    SELECT poi_id FROM poi_point_pois ppp WHERE ppp.poi_point_id = pp.id LIMIT 1
);

-- Drop junction table
DROP TABLE IF EXISTS poi_point_pois;

-- Remove updated_at column
ALTER TABLE poi_points DROP COLUMN IF EXISTS updated_at;

-- Re-add foreign key constraint
ALTER TABLE poi_points ADD FOREIGN KEY (poi_id) REFERENCES pois(id) ON DELETE CASCADE;

-- Drop indexes
DROP INDEX IF EXISTS idx_poi_points_poi_name;
