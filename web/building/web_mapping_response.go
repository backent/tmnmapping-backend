package building

type MappingBuildingImageResponse struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type MappingBuildingResponse struct {
	Id             int                          `json:"id"`
	Name           string                       `json:"name"`
	BuildingType   string                       `json:"building_type"`
	GradeResource  string                       `json:"grade_resource"`
	CompletionYear int                          `json:"completion_year"`
	Subdistrict    string                       `json:"subdistrict"`
	Citytown       string                       `json:"citytown"`
	Province       string                       `json:"province"`
	Address        string                       `json:"address"`
	BuildingStatus string                       `json:"building_status"`
	Sellable      string                       `json:"sellable"`
	Connectivity   string                       `json:"connectivity"`
	Images         []MappingBuildingImageResponse `json:"images"`
}

type MappingBuildingsResponse struct {
	Data            []MappingBuildingResponse `json:"data"`
	TotalApartment int                       `json:"total_appartment"`
	TotalHotel      int                       `json:"total_hotel"`
	TotalOffice     int                       `json:"total_office"`
	TotalRetail     int                       `json:"total_retail"`
	TotalOthers     int                       `json:"total_others"`
}
