package poipoint

type POIRefResponse struct {
	Id    int    `json:"id"`
	Brand string `json:"brand"`
}

type POIPointResponse struct {
	Id              int              `json:"id"`
	POIName         string           `json:"poi_name"`
	Address         string           `json:"address"`
	Latitude        float64          `json:"latitude"`
	Longitude       float64          `json:"longitude"`
	CategoryId      *int             `json:"category_id"`
	SubCategoryId   *int             `json:"sub_category_id"`
	MotherBrandId   *int             `json:"mother_brand_id"`
	BranchId        *int             `json:"branch_id"`
	Category        string           `json:"category"`
	SubCategory     string           `json:"sub_category"`
	MotherBrand     string           `json:"mother_brand"`
	Branch          string           `json:"branch"`
	POIs            []POIRefResponse `json:"pois"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
}

type POIPointUsageResponse struct {
	POIs []POIRefResponse `json:"pois"`
}
