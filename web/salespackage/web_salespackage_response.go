package salespackage

type BuildingRefResponse struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	ProjectName  string `json:"project_name"`
	Subdistrict  string `json:"subdistrict"`
	Citytown     string `json:"citytown"`
	Province     string `json:"province"`
	BuildingType string `json:"building_type"`
}

type SalesPackageResponse struct {
	Id          int                   `json:"id"`
	PackageCode string                `json:"package_code"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Status      string                `json:"status"`
	ScreenCount int                   `json:"screen_count"`
	Traffic     int                   `json:"traffic"`
	Impressions int                   `json:"impressions"`
	Buildings   []BuildingRefResponse `json:"buildings"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
}
