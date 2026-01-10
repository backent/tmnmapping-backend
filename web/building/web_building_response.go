package building

import "github.com/malikabdulaziz/tmn-backend/models"

type BuildingImageResponse struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type BuildingResponse struct {
	Id                 int                     `json:"id"`
	ExternalBuildingId string                  `json:"external_building_id"`
	IrisCode           string                  `json:"iris_code"`
	Name               string                  `json:"name"`
	ProjectName        string                  `json:"project_name"`
	Audience           int                     `json:"audience"`
	Impression         int                     `json:"impression"`
	CbdArea            string                  `json:"cbd_area"`
	BuildingStatus     string                  `json:"building_status"`
	CompetitorLocation bool                    `json:"competitor_location"`
	Sellable           string                  `json:"sellable"`
	Connectivity       string                  `json:"connectivity"`
	ResourceType       string                  `json:"resource_type"`
	Subdistrict        string                  `json:"subdistrict"`
	Citytown           string                  `json:"citytown"`
	Province           string                  `json:"province"`
	GradeResource      string                  `json:"grade_resource"`
	BuildingType       string                  `json:"building_type"`
	CompletionYear     int                     `json:"completion_year"`
	Images             []BuildingImageResponse `json:"images"`
	SyncedAt           string                  `json:"synced_at"`
	CreatedAt          string                  `json:"created_at"`
	UpdatedAt          string                  `json:"updated_at"`
}

func BuildingModelToBuildingResponse(building models.Building) BuildingResponse {
	// Convert BuildingImage to BuildingImageResponse
	images := make([]BuildingImageResponse, 0, len(building.Images))
	for _, img := range building.Images {
		images = append(images, BuildingImageResponse{
			Name: img.Name,
			Path: img.Path,
		})
	}

	return BuildingResponse{
		Id:                 building.Id,
		ExternalBuildingId: building.ExternalBuildingId,
		IrisCode:           building.IrisCode,
		Name:               building.Name,
		ProjectName:        building.ProjectName,
		Audience:           building.Audience,
		Impression:         building.Impression,
		CbdArea:            building.CbdArea,
		BuildingStatus:     building.BuildingStatus,
		CompetitorLocation: building.CompetitorLocation,
		Sellable:           building.Sellable,
		Connectivity:       building.Connectivity,
		ResourceType:       building.ResourceType,
		Subdistrict:        building.Subdistrict,
		Citytown:           building.Citytown,
		Province:           building.Province,
		GradeResource:      building.GradeResource,
		BuildingType:       building.BuildingType,
		CompletionYear:     building.CompletionYear,
		Images:             images,
		SyncedAt:           building.SyncedAt,
		CreatedAt:          building.CreatedAt,
		UpdatedAt:          building.UpdatedAt,
	}
}

func BuildingModelsToListBuildingResponse(buildings []models.Building) []BuildingResponse {
	if buildings == nil {
		return []BuildingResponse{}
	}

	responses := make([]BuildingResponse, 0, len(buildings))
	for _, building := range buildings {
		responses = append(responses, BuildingModelToBuildingResponse(building))
	}
	return responses
}
