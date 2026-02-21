CREATE TABLE letters_of_intent (
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

CREATE INDEX idx_loi_person ON letters_of_intent(acquisition_person);
CREATE INDEX idx_loi_workflow ON letters_of_intent(workflow_state);
CREATE INDEX idx_loi_erp_date ON letters_of_intent(created_at_erp);
CREATE INDEX idx_loi_building_project ON letters_of_intent(building_project);
CREATE INDEX idx_loi_modified ON letters_of_intent(modified);
