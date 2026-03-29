-- Add updated_at to poi_points for consistency with other entities
ALTER TABLE poi_points ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Create many-to-many junction table between poi_points and pois
CREATE TABLE IF NOT EXISTS poi_point_pois (
    id BIGSERIAL PRIMARY KEY,
    poi_point_id BIGINT NOT NULL,
    poi_id BIGINT NOT NULL,
    FOREIGN KEY (poi_point_id) REFERENCES poi_points(id) ON DELETE CASCADE,
    FOREIGN KEY (poi_id) REFERENCES pois(id) ON DELETE CASCADE,
    UNIQUE (poi_point_id, poi_id)
);

-- Migrate existing one-to-many relationships into junction table
INSERT INTO poi_point_pois (poi_point_id, poi_id)
SELECT id, poi_id FROM poi_points WHERE poi_id IS NOT NULL;

-- Remove the old foreign key column
ALTER TABLE poi_points DROP COLUMN poi_id;

-- Indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_poi_point_pois_poi_point_id ON poi_point_pois(poi_point_id);
CREATE INDEX IF NOT EXISTS idx_poi_point_pois_poi_id ON poi_point_pois(poi_id);
CREATE INDEX IF NOT EXISTS idx_poi_points_poi_name ON poi_points(poi_name);
