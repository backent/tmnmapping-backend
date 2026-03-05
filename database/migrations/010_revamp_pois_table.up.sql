-- Revamp POI: brand-centric model with new point metadata fields

-- Rename pois.name to brand
ALTER TABLE pois RENAME COLUMN name TO brand;

-- Rename poi_points.place_name to poi_name
ALTER TABLE poi_points RENAME COLUMN place_name TO poi_name;

-- Add new metadata columns to poi_points (nullable for existing data)
ALTER TABLE poi_points ADD COLUMN category VARCHAR(255);
ALTER TABLE poi_points ADD COLUMN sub_category VARCHAR(255);
ALTER TABLE poi_points ADD COLUMN mother_brand VARCHAR(255);
ALTER TABLE poi_points ADD COLUMN branch VARCHAR(255);

-- Update index
DROP INDEX IF EXISTS idx_pois_name;
CREATE INDEX IF NOT EXISTS idx_pois_brand ON pois(brand);
