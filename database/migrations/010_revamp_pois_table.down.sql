-- Rollback POI revamp

ALTER TABLE poi_points DROP COLUMN IF EXISTS branch;
ALTER TABLE poi_points DROP COLUMN IF EXISTS mother_brand;
ALTER TABLE poi_points DROP COLUMN IF EXISTS sub_category;
ALTER TABLE poi_points DROP COLUMN IF EXISTS category;
ALTER TABLE poi_points RENAME COLUMN poi_name TO place_name;
ALTER TABLE pois RENAME COLUMN brand TO name;
DROP INDEX IF EXISTS idx_pois_brand;
CREATE INDEX IF NOT EXISTS idx_pois_name ON pois(name);
