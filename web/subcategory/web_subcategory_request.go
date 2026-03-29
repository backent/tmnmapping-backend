package subcategory

import "strings"

type CreateSubCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateSubCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type SubCategoryRequestFindAll struct {
	take           int
	skip           int
	orderBy        string
	orderDirection string
	search         string
}

func (r *SubCategoryRequestFindAll) SetSkip(skip int)                        { r.skip = skip }
func (r *SubCategoryRequestFindAll) SetTake(take int)                        { r.take = take }
func (r *SubCategoryRequestFindAll) GetSkip() int                            { return r.skip }
func (r *SubCategoryRequestFindAll) GetTake() int                            { return r.take }
func (r *SubCategoryRequestFindAll) SetOrderBy(orderBy string)               { r.orderBy = orderBy }
func (r *SubCategoryRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}
func (r *SubCategoryRequestFindAll) GetOrderBy() string {
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}
func (r *SubCategoryRequestFindAll) GetOrderDirection() string {
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}
func (r *SubCategoryRequestFindAll) SetSearch(search string) { r.search = search }
func (r *SubCategoryRequestFindAll) GetSearch() string       { return r.search }
