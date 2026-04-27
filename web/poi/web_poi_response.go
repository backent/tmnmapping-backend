package poi

type POIPointResponse struct {
	Id        int     `json:"id"`
	POIName   string  `json:"poi_name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Branch    string  `json:"branch"`
	BranchId  *int    `json:"branch_id,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type POIResponse struct {
	Id            int                `json:"id"`
	Brand         string             `json:"brand"`
	Color         string             `json:"color"`
	Category      string             `json:"category"`
	SubCategory   string             `json:"sub_category"`
	MotherBrand   string             `json:"mother_brand"`
	CategoryId    *int               `json:"category_id,omitempty"`
	SubCategoryId *int               `json:"sub_category_id,omitempty"`
	MotherBrandId *int               `json:"mother_brand_id,omitempty"`
	Points        []POIPointResponse `json:"points"`
	CreatedAt     string             `json:"created_at"`
	UpdatedAt     string             `json:"updated_at"`
}
