package poi

type POIPointResponse struct {
	Id          int     `json:"id"`
	POIName     string  `json:"poi_name"`
	Address     string  `json:"address"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Category    string  `json:"category"`
	SubCategory string  `json:"sub_category"`
	MotherBrand string  `json:"mother_brand"`
	Branch      string  `json:"branch"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type POIResponse struct {
	Id        int                `json:"id"`
	Brand     string             `json:"brand"`
	Color     string             `json:"color"`
	Points    []POIPointResponse `json:"points"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}
