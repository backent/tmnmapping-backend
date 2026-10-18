-- Recreates the tables empty. The rows are not recoverable here -- they were an ERP
-- mirror, and re-populating means re-fetching from ERP, which is code this migration's
-- companion removed.
CREATE TABLE IF NOT EXISTS acquisitions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    building_project VARCHAR(255),
    acquisition_person VARCHAR(255),
    workflow_state VARCHAR(255),
    modified TIMESTAMP,
    synced_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS building_proposals (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    building_project VARCHAR(255),
    acquisition_person VARCHAR(255),
    workflow_state VARCHAR(255),
    number_of_screen INTEGER,
    modified TIMESTAMP,
    synced_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
