package category

import "strings"

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type CategoryRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
}

func (r *CategoryRequestFindAll) SetSkip(skip int)                        { r.skip = skip }
func (r *CategoryRequestFindAll) SetTake(take int)                        { r.take = take }
func (r *CategoryRequestFindAll) GetSkip() int                            { return r.skip }
func (r *CategoryRequestFindAll) GetTake() int                            { return r.take }
func (r *CategoryRequestFindAll) SetOrderBy(orderBy string)               { r.orderBy = orderBy }
func (r *CategoryRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}
func (r *CategoryRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}
func (r *CategoryRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}
func (r *CategoryRequestFindAll) SetSearch(search string) { r.search = search }
func (r *CategoryRequestFindAll) GetSearch() string       { return r.search }
