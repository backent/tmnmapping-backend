package models

import "database/sql"

// BuildingProject is the landlord side of a building: the site a contract is signed
// for, rather than one tower. Buildings point at it, and the money on it is what TMN
// PAYS -- unlike every other figure in this schema, which is revenue.
//
// One flat row per project, including its current contract. The business does not
// track renewal history; where a previous value is needed, BuildingProjectChange
// answers it. See docs/BUILDING_PROJECT_ENTITY_SPEC.md.
type BuildingProject struct {
	Id int `json:"id"`

	ProjectIdIris string `json:"project_id_iris"`
	Name          string `json:"name"`

	BuildingType     string `json:"building_type"`
	Grade            string `json:"grade"`
	Pic              string `json:"pic"`
	TmnProjectStatus string `json:"tmn_project_status"`

	// Declared counts, maintained by the business. Deliberately not derived from the
	// buildings rows and never cross-validated against them.
	NoOfTower  int `json:"no_of_tower"`
	NoOfScreen int `json:"no_of_screen"`

	CreatedDate string `json:"created_date"`
	Remark      string `json:"remark"`

	ContractType   string `json:"contract_type"`
	ContractNo     string `json:"contract_no"`
	ContractDate   string `json:"contract_date"`
	ContractStart  string `json:"contract_start"`
	ContractEnd    string `json:"contract_end"`
	PeriodMonth    int    `json:"period_month"`
	AnnualRental   int64  `json:"annual_rental"`
	PaymentTerm    string `json:"payment_term"`
	CompanyName    string `json:"company_name"`
	Exclusivity    string `json:"exclusivity"`
	DocType        string `json:"doc_type"`
	ContractStatus string `json:"contract_status"`

	CancelledAt      string `json:"cancelled_at"`
	CancelLastStatus string `json:"cancel_last_status"`
	CancelReason     string `json:"cancel_reason"`

	// BuildingCount is joined, not stored: how many buildings carry this project.
	// Compared against NoOfTower in a report, never enforced.
	BuildingCount int `json:"building_count"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type NullAbleBuildingProject struct {
	Id               sql.NullInt64
	ProjectIdIris    sql.NullString
	Name             sql.NullString
	BuildingType     sql.NullString
	Grade            sql.NullString
	Pic              sql.NullString
	TmnProjectStatus sql.NullString
	NoOfTower        sql.NullInt64
	NoOfScreen       sql.NullInt64
	CreatedDate      sql.NullString
	Remark           sql.NullString
	ContractType     sql.NullString
	ContractNo       sql.NullString
	ContractDate     sql.NullString
	ContractStart    sql.NullString
	ContractEnd      sql.NullString
	PeriodMonth      sql.NullInt64
	AnnualRental     sql.NullInt64
	PaymentTerm      sql.NullString
	CompanyName      sql.NullString
	Exclusivity      sql.NullString
	DocType          sql.NullString
	ContractStatus   sql.NullString
	CancelledAt      sql.NullString
	CancelLastStatus sql.NullString
	CancelReason     sql.NullString
	BuildingCount    sql.NullInt64
	CreatedAt        sql.NullString
	UpdatedAt        sql.NullString
}

func NullAbleBuildingProjectToBuildingProject(n NullAbleBuildingProject) BuildingProject {
	return BuildingProject{
		Id:               int(n.Id.Int64),
		ProjectIdIris:    n.ProjectIdIris.String,
		Name:             n.Name.String,
		BuildingType:     n.BuildingType.String,
		Grade:            n.Grade.String,
		Pic:              n.Pic.String,
		TmnProjectStatus: n.TmnProjectStatus.String,
		NoOfTower:        int(n.NoOfTower.Int64),
		NoOfScreen:       int(n.NoOfScreen.Int64),
		CreatedDate:      n.CreatedDate.String,
		Remark:           n.Remark.String,
		ContractType:     n.ContractType.String,
		ContractNo:       n.ContractNo.String,
		ContractDate:     n.ContractDate.String,
		ContractStart:    n.ContractStart.String,
		ContractEnd:      n.ContractEnd.String,
		PeriodMonth:      int(n.PeriodMonth.Int64),
		AnnualRental:     n.AnnualRental.Int64,
		PaymentTerm:      n.PaymentTerm.String,
		CompanyName:      n.CompanyName.String,
		Exclusivity:      n.Exclusivity.String,
		DocType:          n.DocType.String,
		ContractStatus:   n.ContractStatus.String,
		CancelledAt:      n.CancelledAt.String,
		CancelLastStatus: n.CancelLastStatus.String,
		CancelReason:     n.CancelReason.String,
		BuildingCount:    int(n.BuildingCount.Int64),
		CreatedAt:        n.CreatedAt.String,
		UpdatedAt:        n.UpdatedAt.String,
	}
}

// BuildingProjectChange is one field of one project changing, attributed and
// timestamped. One row per FIELD: knowing a project was touched is not useful,
// knowing annual_rental went from 54,000,000 to 61,500,000 is.
type BuildingProjectChange struct {
	Id int `json:"id"`

	// ProjectId is null once the project is deleted; ProjectIdIris always survives,
	// so the history stays readable after the row it describes is gone.
	ProjectId     int    `json:"project_id"`
	ProjectIdIris string `json:"project_id_iris"`

	ActorUserId int    `json:"actor_user_id"`
	ActorName   string `json:"actor_name"`
	// The role AT THE TIME. Roles change; the audit trail should not change with them.
	ActorRole string `json:"actor_role"`

	Action string `json:"action"`
	Source string `json:"source"`
	// BatchId groups one upload. Empty for a form edit.
	BatchId string `json:"batch_id"`

	Field    string `json:"field"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`

	CreatedAt string `json:"created_at"`
}

type NullAbleBuildingProjectChange struct {
	Id            sql.NullInt64
	ProjectId     sql.NullInt64
	ProjectIdIris sql.NullString
	ActorUserId   sql.NullInt64
	ActorName     sql.NullString
	ActorRole     sql.NullString
	Action        sql.NullString
	Source        sql.NullString
	BatchId       sql.NullString
	Field         sql.NullString
	OldValue      sql.NullString
	NewValue      sql.NullString
	CreatedAt     sql.NullString
}

func NullAbleBuildingProjectChangeToChange(n NullAbleBuildingProjectChange) BuildingProjectChange {
	return BuildingProjectChange{
		Id:            int(n.Id.Int64),
		ProjectId:     int(n.ProjectId.Int64),
		ProjectIdIris: n.ProjectIdIris.String,
		ActorUserId:   int(n.ActorUserId.Int64),
		ActorName:     n.ActorName.String,
		ActorRole:     n.ActorRole.String,
		Action:        n.Action.String,
		Source:        n.Source.String,
		BatchId:       n.BatchId.String,
		Field:         n.Field.String,
		OldValue:      n.OldValue.String,
		NewValue:      n.NewValue.String,
		CreatedAt:     n.CreatedAt.String,
	}
}

// Change actions and sources, matching the CHECK constraints in migration 023.
const (
	BuildingProjectActionCreated = "created"
	BuildingProjectActionUpdated = "updated"
	BuildingProjectActionDeleted = "deleted"

	BuildingProjectSourceForm   = "form"
	BuildingProjectSourceImport = "import"
)

var BuildingProjectTable string = "building_projects"
var BuildingProjectChangeTable string = "building_project_changes"
