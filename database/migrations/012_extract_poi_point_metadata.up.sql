-- Create master data tables
CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sub_categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE mother_brands (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE branches (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Populate from existing distinct values
INSERT INTO categories (name) SELECT DISTINCT category FROM poi_points WHERE category IS NOT NULL AND category != '' ON CONFLICT DO NOTHING;
INSERT INTO sub_categories (name) SELECT DISTINCT sub_category FROM poi_points WHERE sub_category IS NOT NULL AND sub_category != '' ON CONFLICT DO NOTHING;
INSERT INTO mother_brands (name) SELECT DISTINCT mother_brand FROM poi_points WHERE mother_brand IS NOT NULL AND mother_brand != '' ON CONFLICT DO NOTHING;
INSERT INTO branches (name) SELECT DISTINCT branch FROM poi_points WHERE branch IS NOT NULL AND branch != '' ON CONFLICT DO NOTHING;

-- Add FK columns (nullable)
ALTER TABLE poi_points ADD COLUMN category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL;
ALTER TABLE poi_points ADD COLUMN sub_category_id BIGINT REFERENCES sub_categories(id) ON DELETE SET NULL;
ALTER TABLE poi_points ADD COLUMN mother_brand_id BIGINT REFERENCES mother_brands(id) ON DELETE SET NULL;
ALTER TABLE poi_points ADD COLUMN branch_id BIGINT REFERENCES branches(id) ON DELETE SET NULL;

-- Backfill FK columns from existing string values
UPDATE poi_points SET category_id = c.id FROM categories c WHERE poi_points.category = c.name;
UPDATE poi_points SET sub_category_id = sc.id FROM sub_categories sc WHERE poi_points.sub_category = sc.name;
UPDATE poi_points SET mother_brand_id = mb.id FROM mother_brands mb WHERE poi_points.mother_brand = mb.name;
UPDATE poi_points SET branch_id = b.id FROM branches b WHERE poi_points.branch = b.name;

-- Drop old VARCHAR columns
ALTER TABLE poi_points DROP COLUMN category;
ALTER TABLE poi_points DROP COLUMN sub_category;
ALTER TABLE poi_points DROP COLUMN mother_brand;
ALTER TABLE poi_points DROP COLUMN branch;
