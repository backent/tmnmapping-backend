package models

type Acquisition struct {
	Id                int
	ExternalId        string
	WorkflowState     string
	AcquisitionPerson string
	BuildingProject   string
	Status            string
	Modified          string
	CreatedAtErp      string
	SyncedAt          string
	CreatedAt         string
	UpdatedAt         string
}

var AcquisitionTable string = "acquisitions"
