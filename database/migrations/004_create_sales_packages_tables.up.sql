-- Create sales_packages table (main entity)
CREATE TABLE IF NOT EXISTS sales_packages (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create sales_package_buildings junction table (many-to-many)
CREATE TABLE IF NOT EXISTS sales_package_buildings (
    id BIGSERIAL PRIMARY KEY,
    sales_package_id BIGINT NOT NULL,
    building_id BIGINT NOT NULL,
    FOREIGN KEY (sales_package_id) REFERENCES sales_packages(id) ON DELETE CASCADE,
    FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE CASCADE,
    UNIQUE (sales_package_id, building_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_sales_packages_name ON sales_packages(name);
CREATE INDEX IF NOT EXISTS idx_sales_packages_created_at ON sales_packages(created_at);
CREATE INDEX IF NOT EXISTS idx_sales_package_buildings_sales_package_id ON sales_package_buildings(sales_package_id);
CREATE INDEX IF NOT EXISTS idx_sales_package_buildings_building_id ON sales_package_buildings(building_id);
