package buildingrestriction

type BuildingRefResponse struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	ProjectName  string `json:"project_name"`
	Subdistrict  string `json:"subdistrict"`
	Citytown     string `json:"citytown"`
	Province     string `json:"province"`
	BuildingType string `json:"building_type"`
}

type BuildingRestrictionResponse struct {
	Id        int                      `json:"id"`
	Name      string                   `json:"name"`
	Buildings []BuildingRefResponse    `json:"buildings"`
	CreatedAt string                   `json:"created_at"`
	UpdatedAt string                   `json:"updated_at"`
}
