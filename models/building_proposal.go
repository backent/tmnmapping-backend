package models

type BuildingProposal struct {
	Id                int
	ExternalId        string
	WorkflowState     string
	AcquisitionPerson string
	BuildingProject   string
	Status            string
	NumberOfScreen    int
	Modified          string
	CreatedAtErp      string
	SyncedAt          string
	CreatedAt         string
	UpdatedAt         string
}

var BuildingProposalTable string = "building_proposals"
