-- Reverse 014: restore many-to-many junction and move metadata back to poi_points.
-- Best-effort: only the first point of each POI gets the POI's metadata back; this
-- is lossy if data was edited after the up migration ran.

-- 1. Recreate junction table
CREATE TABLE IF NOT EXISTS poi_point_pois (
    id BIGSERIAL PRIMARY KEY,
    poi_point_id BIGINT NOT NULL,
    poi_id BIGINT NOT NULL,
    FOREIGN KEY (poi_point_id) REFERENCES poi_points(id) ON DELETE CASCADE,
    FOREIGN KEY (poi_id) REFERENCES pois(id) ON DELETE CASCADE,
    UNIQUE (poi_point_id, poi_id)
);

-- 2. Re-add metadata FK columns to poi_points
ALTER TABLE poi_points ADD COLUMN category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL;
ALTER TABLE poi_points ADD COLUMN sub_category_id BIGINT REFERENCES sub_categories(id) ON DELETE SET NULL;
ALTER TABLE poi_points ADD COLUMN mother_brand_id BIGINT REFERENCES mother_brands(id) ON DELETE SET NULL;

-- 3. Backfill junction from current 1:N relationship
INSERT INTO poi_point_pois (poi_point_id, poi_id)
SELECT id, poi_id FROM poi_points WHERE poi_id IS NOT NULL;

-- 4. Backfill point-level metadata from each point's POI
UPDATE poi_points pp
SET category_id = p.category_id,
    sub_category_id = p.sub_category_id,
    mother_brand_id = p.mother_brand_id
FROM pois p
WHERE pp.poi_id = p.id;

-- 5. Drop poi_id from poi_points
ALTER TABLE poi_points DROP COLUMN poi_id;

-- 6. Drop new POI metadata columns
ALTER TABLE pois DROP COLUMN category_id;
ALTER TABLE pois DROP COLUMN sub_category_id;
ALTER TABLE pois DROP COLUMN mother_brand_id;

-- 7. Drop indexes from up
DROP INDEX IF EXISTS idx_poi_points_poi_id;
DROP INDEX IF EXISTS idx_pois_category_id;
DROP INDEX IF EXISTS idx_pois_sub_category_id;
DROP INDEX IF EXISTS idx_pois_mother_brand_id;

-- 8. Restore junction indexes (matches 011)
CREATE INDEX IF NOT EXISTS idx_poi_point_pois_poi_point_id ON poi_point_pois(poi_point_id);
CREATE INDEX IF NOT EXISTS idx_poi_point_pois_poi_id ON poi_point_pois(poi_id);
