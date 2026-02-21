CREATE TABLE acquisitions (
    id BIGSERIAL PRIMARY KEY,
    external_id VARCHAR(255) NOT NULL,
    workflow_state VARCHAR(100),
    acquisition_person VARCHAR(255),
    building_project VARCHAR(255),
    status VARCHAR(100),
    modified TIMESTAMP,
    created_at_erp TIMESTAMP,
    synced_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_acquisitions_person ON acquisitions(acquisition_person);
CREATE INDEX idx_acquisitions_workflow ON acquisitions(workflow_state);
CREATE INDEX idx_acquisitions_erp_date ON acquisitions(created_at_erp);
CREATE INDEX idx_acquisitions_building_project ON acquisitions(building_project);
CREATE INDEX idx_acquisitions_modified ON acquisitions(modified);
