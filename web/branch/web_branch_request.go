package branch

import "strings"

type CreateBranchRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateBranchRequest struct {
	Name string `json:"name" validate:"required"`
}

type BranchRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
}

func (r *BranchRequestFindAll) SetSkip(skip int)                        { r.skip = skip }
func (r *BranchRequestFindAll) SetTake(take int)                        { r.take = take }
func (r *BranchRequestFindAll) GetSkip() int                            { return r.skip }
func (r *BranchRequestFindAll) GetTake() int                            { return r.take }
func (r *BranchRequestFindAll) SetOrderBy(orderBy string)               { r.orderBy = orderBy }
func (r *BranchRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}
func (r *BranchRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}
func (r *BranchRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}
func (r *BranchRequestFindAll) SetSearch(search string) { r.search = search }
func (r *BranchRequestFindAll) GetSearch() string       { return r.search }
