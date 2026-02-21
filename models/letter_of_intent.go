package models

type LetterOfIntent struct {
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

var LetterOfIntentTable string = "letters_of_intent"
