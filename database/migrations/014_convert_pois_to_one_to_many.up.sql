-- Convert pois <-> poi_points from many-to-many to one-to-many,
-- and relocate category/sub_category/mother_brand from poi_points to pois.

-- 1. Add metadata FKs to pois (nullable; backfilled below)
ALTER TABLE pois ADD COLUMN category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL;
ALTER TABLE pois ADD COLUMN sub_category_id BIGINT REFERENCES sub_categories(id) ON DELETE SET NULL;
ALTER TABLE pois ADD COLUMN mother_brand_id BIGINT REFERENCES mother_brands(id) ON DELETE SET NULL;

-- 2. Re-add poi_id FK on poi_points (nullable temporarily)
ALTER TABLE poi_points ADD COLUMN poi_id BIGINT REFERENCES pois(id) ON DELETE CASCADE;

-- 3. Safety check: confirm each point belongs to at most one POI in the junction.
--    Aborts the migration if violated.
DO $$
DECLARE
    multi_count BIGINT;
BEGIN
    SELECT COUNT(*) INTO multi_count FROM (
        SELECT poi_point_id
        FROM poi_point_pois
        GROUP BY poi_point_id
        HAVING COUNT(DISTINCT poi_id) > 1
    ) t;
    IF multi_count > 0 THEN
        RAISE EXCEPTION 'Migration aborted: % poi_points belong to multiple POIs in poi_point_pois. Resolve manually before running 014.', multi_count;
    END IF;
END $$;

-- 4. Backfill poi_points.poi_id from junction
UPDATE poi_points pp
SET poi_id = jp.poi_id
FROM poi_point_pois jp
WHERE jp.poi_point_id = pp.id;

-- 5. Backfill pois.{category_id,sub_category_id,mother_brand_id} from each POI's first associated point
UPDATE pois p
SET category_id = sub.category_id,
    sub_category_id = sub.sub_category_id,
    mother_brand_id = sub.mother_brand_id
FROM (
    SELECT DISTINCT ON (jp.poi_id)
        jp.poi_id,
        pp.category_id,
        pp.sub_category_id,
        pp.mother_brand_id
    FROM poi_point_pois jp
    JOIN poi_points pp ON pp.id = jp.poi_point_id
    ORDER BY jp.poi_id, pp.id
) sub
WHERE p.id = sub.poi_id;

-- 6. Delete orphan poi_points (not linked to any POI via the junction).
--    These would otherwise violate the upcoming NOT NULL constraint.
DELETE FROM poi_points WHERE poi_id IS NULL;

-- 7. Enforce poi_id NOT NULL on poi_points
ALTER TABLE poi_points ALTER COLUMN poi_id SET NOT NULL;

-- 8. Drop point-level metadata FK columns (now lives on pois)
ALTER TABLE poi_points DROP COLUMN category_id;
ALTER TABLE poi_points DROP COLUMN sub_category_id;
ALTER TABLE poi_points DROP COLUMN mother_brand_id;

-- 9. Drop junction table
DROP TABLE IF EXISTS poi_point_pois;

-- 10. Indexes
CREATE INDEX IF NOT EXISTS idx_poi_points_poi_id ON poi_points(poi_id);
CREATE INDEX IF NOT EXISTS idx_pois_category_id ON pois(category_id);
CREATE INDEX IF NOT EXISTS idx_pois_sub_category_id ON pois(sub_category_id);
CREATE INDEX IF NOT EXISTS idx_pois_mother_brand_id ON pois(mother_brand_id);
