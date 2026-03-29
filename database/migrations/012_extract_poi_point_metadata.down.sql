-- Re-add VARCHAR columns
ALTER TABLE poi_points ADD COLUMN category VARCHAR(255);
ALTER TABLE poi_points ADD COLUMN sub_category VARCHAR(255);
ALTER TABLE poi_points ADD COLUMN mother_brand VARCHAR(255);
ALTER TABLE poi_points ADD COLUMN branch VARCHAR(255);

-- Backfill from FK tables
UPDATE poi_points SET category = c.name FROM categories c WHERE poi_points.category_id = c.id;
UPDATE poi_points SET sub_category = sc.name FROM sub_categories sc WHERE poi_points.sub_category_id = sc.id;
UPDATE poi_points SET mother_brand = mb.name FROM mother_brands mb WHERE poi_points.mother_brand_id = mb.id;
UPDATE poi_points SET branch = b.name FROM branches b WHERE poi_points.branch_id = b.id;

-- Drop FK columns
ALTER TABLE poi_points DROP COLUMN category_id;
ALTER TABLE poi_points DROP COLUMN sub_category_id;
ALTER TABLE poi_points DROP COLUMN mother_brand_id;
ALTER TABLE poi_points DROP COLUMN branch_id;

-- Drop master data tables
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS sub_categories;
DROP TABLE IF EXISTS mother_brands;
DROP TABLE IF EXISTS branches;
