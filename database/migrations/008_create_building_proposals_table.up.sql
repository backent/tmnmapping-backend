CREATE TABLE building_proposals (
    id BIGSERIAL PRIMARY KEY,
    external_id VARCHAR(255) NOT NULL,
    workflow_state VARCHAR(100),
    acquisition_person VARCHAR(255),
    building_project VARCHAR(255),
    status VARCHAR(100),
    number_of_screen INT DEFAULT 0,
    modified TIMESTAMP,
    created_at_erp TIMESTAMP,
    synced_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_building_proposals_person ON building_proposals(acquisition_person);
CREATE INDEX idx_building_proposals_workflow ON building_proposals(workflow_state);
CREATE INDEX idx_building_proposals_erp_date ON building_proposals(created_at_erp);
CREATE INDEX idx_building_proposals_building_project ON building_proposals(building_project);
CREATE INDEX idx_building_proposals_modified ON building_proposals(modified);
