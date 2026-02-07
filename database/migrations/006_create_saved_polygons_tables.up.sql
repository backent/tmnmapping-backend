-- Create saved_polygons table (main entity)
CREATE TABLE IF NOT EXISTS saved_polygons (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create saved_polygon_points table (one-to-many, ordered points)
CREATE TABLE IF NOT EXISTS saved_polygon_points (
    id BIGSERIAL PRIMARY KEY,
    saved_polygon_id BIGINT NOT NULL,
    ord INT NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (saved_polygon_id) REFERENCES saved_polygons(id) ON DELETE CASCADE,
    UNIQUE (saved_polygon_id, ord)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_saved_polygons_name ON saved_polygons(name);
CREATE INDEX IF NOT EXISTS idx_saved_polygons_created_at ON saved_polygons(created_at);
CREATE INDEX IF NOT EXISTS idx_saved_polygon_points_saved_polygon_id ON saved_polygon_points(saved_polygon_id);
CREATE INDEX IF NOT EXISTS idx_saved_polygon_points_ord ON saved_polygon_points(ord);
