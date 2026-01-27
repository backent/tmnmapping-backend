-- Create POI table (main entity)
CREATE TABLE IF NOT EXISTS pois (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(7),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create POI Points table (one-to-many relationship)
CREATE TABLE IF NOT EXISTS poi_points (
    id BIGSERIAL PRIMARY KEY,
    poi_id BIGINT NOT NULL,
    place_name VARCHAR(255),
    address VARCHAR(500),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    location GEOGRAPHY(POINT, 4326),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (poi_id) REFERENCES pois(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_pois_name ON pois(name);
CREATE INDEX IF NOT EXISTS idx_pois_created_at ON pois(created_at);
CREATE INDEX IF NOT EXISTS idx_poi_points_poi_id ON poi_points(poi_id);
CREATE INDEX IF NOT EXISTS idx_poi_points_location_gist ON poi_points USING GIST(location);
