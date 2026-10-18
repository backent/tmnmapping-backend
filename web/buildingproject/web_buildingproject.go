package buildingproject

import "strings"

// SaveBuildingProjectRequest is the form payload for both create and update.
//
// One struct for both, because the form REPLACES the record: every column is sent
// explicitly and a blank means null. The importer does not use this type -- it merges
// a parsed row instead -- but both end up in the same service call, because blank now
// means "clear" on both paths. See BUILDING_PROJECT_ENTITY_SPEC.md §11.
//
// Only the key and the name are required. A project may be raised from a partly
// known site and completed later; refusing that would push people back to the
// spreadsheet for exactly the records that need attention.
type SaveBuildingProjectRequest struct {
	ProjectIdIris string `json:"project_id_iris" validate:"required,max=100"`
	Name          string `json:"name" validate:"required,max=255"`

	BuildingType     string `json:"building_type" validate:"omitempty,max=255"`
	Grade            string `json:"grade" validate:"omitempty,max=100"`
	Pic              string `json:"pic" validate:"omitempty,max=255"`
	TmnProjectStatus string `json:"tmn_project_status" validate:"omitempty,max=50"`

	NoOfTower  int `json:"no_of_tower" validate:"omitempty,gte=0"`
	NoOfScreen int `json:"no_of_screen" validate:"omitempty,gte=0"`

	CreatedDate string `json:"created_date" validate:"omitempty,datetime=2006-01-02"`
	Remark      string `json:"remark" validate:"omitempty"`

	ContractType   string `json:"contract_type" validate:"omitempty,max=50"`
	ContractNo     string `json:"contract_no" validate:"omitempty,max=100"`
	ContractDate   string `json:"contract_date" validate:"omitempty,datetime=2006-01-02"`
	ContractStart  string `json:"contract_start" validate:"omitempty,datetime=2006-01-02"`
	ContractEnd    string `json:"contract_end" validate:"omitempty,datetime=2006-01-02"`
	PeriodMonth    int    `json:"period_month" validate:"omitempty,gt=0"`
	AnnualRental   int64  `json:"annual_rental" validate:"omitempty,gte=0"`
	PaymentTerm    string `json:"payment_term" validate:"omitempty,max=50"`
	CompanyName    string `json:"company_name" validate:"omitempty,max=255"`
	Exclusivity    string `json:"exclusivity" validate:"omitempty,max=50"`
	DocType        string `json:"doc_type" validate:"omitempty,max=50"`
	ContractStatus string `json:"contract_status" validate:"omitempty,max=50"`

	CancelledAt      string `json:"cancelled_at" validate:"omitempty,datetime=2006-01-02"`
	CancelLastStatus string `json:"cancel_last_status" validate:"omitempty,max=50"`
	CancelReason     string `json:"cancel_reason" validate:"omitempty"`
}

// BuildingProjectResponse mirrors the model. The three finance fields are omitted
// from the JSON when empty rather than sent as zero, so a caller without
// building-projects.finance cannot tell a redacted rental from a rental of nothing --
// and more importantly, never receives one.
type BuildingProjectResponse struct {
	Id int `json:"id"`

	ProjectIdIris string `json:"project_id_iris"`
	Name          string `json:"name"`

	BuildingType     string `json:"building_type"`
	Grade            string `json:"grade"`
	Pic              string `json:"pic"`
	TmnProjectStatus string `json:"tmn_project_status"`

	NoOfTower  int `json:"no_of_tower"`
	NoOfScreen int `json:"no_of_screen"`

	CreatedDate string `json:"created_date"`
	Remark      string `json:"remark"`

	ContractType   string `json:"contract_type"`
	ContractDate   string `json:"contract_date"`
	ContractStart  string `json:"contract_start"`
	ContractEnd    string `json:"contract_end"`
	PeriodMonth    int    `json:"period_month"`
	PaymentTerm    string `json:"payment_term"`
	Exclusivity    string `json:"exclusivity"`
	DocType        string `json:"doc_type"`
	ContractStatus string `json:"contract_status"`

	CancelledAt      string `json:"cancelled_at"`
	CancelLastStatus string `json:"cancel_last_status"`
	CancelReason     string `json:"cancel_reason"`

	// Gated by building-projects.finance. Omitted entirely for callers without it:
	// the repository does not select the columns, so there is nothing to leak.
	AnnualRental *int64  `json:"annual_rental,omitempty"`
	CompanyName  *string `json:"company_name,omitempty"`
	ContractNo   *string `json:"contract_no,omitempty"`

	// Derived on read, never stored, so a stale spreadsheet value cannot enter.
	PricePerScreen *int64 `json:"price_per_screen,omitempty"`
	ContractValue  *int64 `json:"contract_value,omitempty"`

	BuildingCount int `json:"building_count"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// BuildingProjectChangeResponse is one row of the history.
type BuildingProjectChangeResponse struct {
	Id            int    `json:"id"`
	ProjectId     int    `json:"project_id"`
	ProjectIdIris string `json:"project_id_iris"`
	ActorUserId   int    `json:"actor_user_id"`
	ActorName     string `json:"actor_name"`
	ActorRole     string `json:"actor_role"`
	Action        string `json:"action"`
	Source        string `json:"source"`
	BatchId       string `json:"batch_id"`
	Field         string `json:"field"`
	OldValue      string `json:"old_value"`
	NewValue      string `json:"new_value"`
	CreatedAt     string `json:"created_at"`
}

type BuildingProjectRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string

	Status       string
	ContractType string
	Pic          string
}

func (r *BuildingProjectRequestFindAll) SetSkip(skip int)          { r.skip = skip }
func (r *BuildingProjectRequestFindAll) SetTake(take int)          { r.take = take }
func (r *BuildingProjectRequestFindAll) GetSkip() int              { return r.skip }
func (r *BuildingProjectRequestFindAll) GetTake() int              { return r.take }
func (r *BuildingProjectRequestFindAll) SetOrderBy(orderBy string) { r.orderBy = orderBy }
func (r *BuildingProjectRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *BuildingProjectRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}

	return r.orderBy
}

func (r *BuildingProjectRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}

	return r.orderDirection
}

func (r *BuildingProjectRequestFindAll) SetSearch(search string) { r.search = search }
func (r *BuildingProjectRequestFindAll) GetSearch() string       { return r.search }
