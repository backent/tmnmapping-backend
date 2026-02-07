package savedpolygon

type SavedPolygonPointResponse struct {
	Ord int     `json:"ord"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type SavedPolygonResponse struct {
	Id        int                        `json:"id"`
	Name      string                     `json:"name"`
	Points    []SavedPolygonPointResponse `json:"points"`
	CreatedAt string                     `json:"created_at"`
	UpdatedAt string                     `json:"updated_at"`
}
