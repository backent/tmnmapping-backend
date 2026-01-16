-- Enable PostGIS extension for spatial queries
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS buildings (
    id BIGSERIAL PRIMARY KEY,
    external_building_id VARCHAR(100) UNIQUE,
    iris_code VARCHAR(100),
    name VARCHAR(255) NOT NULL,
    project_name VARCHAR(255),
    audience INTEGER,
    impression INTEGER,
    cbd_area VARCHAR(255),
    building_status VARCHAR(255),
    competitor_location BOOLEAN DEFAULT FALSE,
    sellable VARCHAR(20),
    connectivity VARCHAR(50),
    resource_type VARCHAR(255),
    subdistrict VARCHAR(255),
    citytown VARCHAR(255),
    province VARCHAR(255),
    grade_resource VARCHAR(255),
    building_type VARCHAR(255),
    completion_year INTEGER,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    location GEOGRAPHY(POINT, 4326),
    images JSONB,
    synced_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_buildings_external_id ON buildings(external_building_id);
CREATE INDEX IF NOT EXISTS idx_buildings_iris_code ON buildings(iris_code);
CREATE INDEX IF NOT EXISTS idx_buildings_name ON buildings(name);
CREATE INDEX IF NOT EXISTS idx_buildings_created_at ON buildings(created_at);
CREATE INDEX IF NOT EXISTS idx_buildings_competitor_location ON buildings(competitor_location);
CREATE INDEX IF NOT EXISTS idx_buildings_sellable ON buildings(sellable);
CREATE INDEX IF NOT EXISTS idx_buildings_connectivity ON buildings(connectivity);
CREATE INDEX IF NOT EXISTS idx_buildings_resource_type ON buildings(resource_type);
CREATE INDEX IF NOT EXISTS idx_buildings_subdistrict ON buildings(subdistrict);
CREATE INDEX IF NOT EXISTS idx_buildings_citytown ON buildings(citytown);
CREATE INDEX IF NOT EXISTS idx_buildings_province ON buildings(province);
CREATE INDEX IF NOT EXISTS idx_buildings_grade_resource ON buildings(grade_resource);
CREATE INDEX IF NOT EXISTS idx_buildings_building_type ON buildings(building_type);
CREATE INDEX IF NOT EXISTS idx_buildings_location_gist ON buildings USING GIST(location);

