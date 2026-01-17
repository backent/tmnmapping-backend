package building

type MappingBuildingImageResponse struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type MappingBuildingResponse struct {
	Id                int                            `json:"id"`
	Name              string                         `json:"name"`
	BuildingType      string                         `json:"building_type"`
	GradeResource     string                         `json:"grade_resource"`
	CompletionYear    int                            `json:"completion_year"`
	Subdistrict       string                         `json:"subdistrict"`
	Citytown          string                         `json:"citytown"`
	Province          string                         `json:"province"`
	Address           string                         `json:"address"`
	BuildingStatus    string                         `json:"building_status"`
	Sellable          string                         `json:"sellable"`
	Connectivity      string                         `json:"connectivity"`
	Latitude          float64                        `json:"latitude"`
	Longitude         float64                        `json:"longitude"`
	LcdPresenceStatus string                         `json:"lcd_presence_status"`
	Images            []MappingBuildingImageResponse `json:"images"`
}

type MappingBuildingsResponse struct {
	Data   []MappingBuildingResponse `json:"data"`
	Totals map[string]int            `json:"totals"` // Dynamic totals for all building types
}
