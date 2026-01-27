package poi

type POIPointResponse struct {
	Id        int     `json:"id"`
	PlaceName string  `json:"place_name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	CreatedAt string  `json:"created_at"`
}

type POIResponse struct {
	Id        int              `json:"id"`
	Name      string           `json:"name"`
	Color     string           `json:"color"`
	Points    []POIPointResponse `json:"points"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}
