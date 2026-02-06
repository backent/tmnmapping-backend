-- Create building_restrictions table (main entity)
CREATE TABLE IF NOT EXISTS building_restrictions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create building_restriction_buildings junction table (many-to-many)
CREATE TABLE IF NOT EXISTS building_restriction_buildings (
    id BIGSERIAL PRIMARY KEY,
    building_restriction_id BIGINT NOT NULL,
    building_id BIGINT NOT NULL,
    FOREIGN KEY (building_restriction_id) REFERENCES building_restrictions(id) ON DELETE CASCADE,
    FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE CASCADE,
    UNIQUE (building_restriction_id, building_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_building_restrictions_name ON building_restrictions(name);
CREATE INDEX IF NOT EXISTS idx_building_restrictions_created_at ON building_restrictions(created_at);
CREATE INDEX IF NOT EXISTS idx_building_restriction_buildings_building_restriction_id ON building_restriction_buildings(building_restriction_id);
CREATE INDEX IF NOT EXISTS idx_building_restriction_buildings_building_id ON building_restriction_buildings(building_id);
