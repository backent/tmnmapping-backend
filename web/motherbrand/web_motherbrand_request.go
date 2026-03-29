package motherbrand

import "strings"

type CreateMotherBrandRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateMotherBrandRequest struct {
	Name string `json:"name" validate:"required"`
}

type MotherBrandRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
}

func (r *MotherBrandRequestFindAll) SetSkip(skip int)                        { r.skip = skip }
func (r *MotherBrandRequestFindAll) SetTake(take int)                        { r.take = take }
func (r *MotherBrandRequestFindAll) GetSkip() int                            { return r.skip }
func (r *MotherBrandRequestFindAll) GetTake() int                            { return r.take }
func (r *MotherBrandRequestFindAll) SetOrderBy(orderBy string)               { r.orderBy = orderBy }
func (r *MotherBrandRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}
func (r *MotherBrandRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}
func (r *MotherBrandRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}
func (r *MotherBrandRequestFindAll) SetSearch(search string) { r.search = search }
func (r *MotherBrandRequestFindAll) GetSearch() string       { return r.search }
