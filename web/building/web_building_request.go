package building

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/web"
)

type UpdateBuildingRequest struct {
	Sellable     string `json:"sellable" validate:"omitempty,oneof=sell not_sell"`
	Connectivity string `json:"connectivity" validate:"omitempty,oneof=online manual not_yet_checked"`
	ResourceType string `json:"resource_type"`
}

type BuildingRequestFindAll struct {
	take              int
	skip              int
	orderBy           string
	orderDirection    string
	search            string
	buildingStatus    string
	sellable          string
	connectivity      string
	resourceType      string
	competitorLocation *bool
	cbdArea           string
	subdistrict       string
	citytown        string
	province        string
	gradeResource   string
	buildingType    string
}

func (r *BuildingRequestFindAll) SetSkip(skip int) {
	r.skip = skip
}

func (r *BuildingRequestFindAll) SetTake(take int) {
	r.take = take
}

func (r *BuildingRequestFindAll) GetSkip() int {
	return r.skip
}

func (r *BuildingRequestFindAll) GetTake() int {
	return r.take
}

func (r *BuildingRequestFindAll) SetOrderBy(orderBy string) {
	r.orderBy = orderBy
}

func (r *BuildingRequestFindAll) SetOrderDirection(orderDirection string) {
	r.orderDirection = strings.ToUpper(orderDirection)
}

func (r *BuildingRequestFindAll) GetOrderBy() string {
	// set default order by
	if r.orderBy == "" {
		return "created_at"
	}
	return r.orderBy
}

func (r *BuildingRequestFindAll) GetOrderDirection() string {
	// set default order direction
	if r.orderDirection == "" {
		return "DESC"
	}
	return r.orderDirection
}

func (r *BuildingRequestFindAll) SetSearch(search string) {
	r.search = search
}

func (r *BuildingRequestFindAll) GetSearch() string {
	return r.search
}

func (r *BuildingRequestFindAll) SetBuildingStatus(buildingStatus string) {
	r.buildingStatus = buildingStatus
}

func (r *BuildingRequestFindAll) GetBuildingStatus() string {
	return r.buildingStatus
}

func (r *BuildingRequestFindAll) SetSellable(sellable string) {
	r.sellable = sellable
}

func (r *BuildingRequestFindAll) GetSellable() string {
	return r.sellable
}

func (r *BuildingRequestFindAll) SetConnectivity(connectivity string) {
	r.connectivity = connectivity
}

func (r *BuildingRequestFindAll) GetConnectivity() string {
	return r.connectivity
}

func (r *BuildingRequestFindAll) SetResourceType(resourceType string) {
	r.resourceType = resourceType
}

func (r *BuildingRequestFindAll) GetResourceType() string {
	return r.resourceType
}

func (r *BuildingRequestFindAll) SetCompetitorLocation(competitorLocation *bool) {
	r.competitorLocation = competitorLocation
}

func (r *BuildingRequestFindAll) GetCompetitorLocation() *bool {
	return r.competitorLocation
}

func (r *BuildingRequestFindAll) SetCbdArea(cbdArea string) {
	r.cbdArea = cbdArea
}

func (r *BuildingRequestFindAll) GetCbdArea() string {
	return r.cbdArea
}

func (r *BuildingRequestFindAll) SetSubdistrict(subdistrict string) {
	r.subdistrict = subdistrict
}

func (r *BuildingRequestFindAll) GetSubdistrict() string {
	return r.subdistrict
}

func (r *BuildingRequestFindAll) SetCitytown(citytown string) {
	r.citytown = citytown
}

func (r *BuildingRequestFindAll) GetCitytown() string {
	return r.citytown
}

func (r *BuildingRequestFindAll) SetProvince(province string) {
	r.province = province
}

func (r *BuildingRequestFindAll) GetProvince() string {
	return r.province
}

func (r *BuildingRequestFindAll) SetGradeResource(gradeResource string) {
	r.gradeResource = gradeResource
}

func (r *BuildingRequestFindAll) GetGradeResource() string {
	return r.gradeResource
}

func (r *BuildingRequestFindAll) SetBuildingType(buildingType string) {
	r.buildingType = buildingType
}

func (r *BuildingRequestFindAll) GetBuildingType() string {
	return r.buildingType
}

// ExportMappingRequest is the request body for POST /admin/mapping-building/export (legacy, by IDs)
type ExportMappingRequest struct {
	Ids []int `json:"ids"`
}

// ExportMappingByFilterRequest is the POST body for export by filters (bounds always null = all matching)
type ExportMappingByFilterRequest struct {
	Filters   ExportMappingFilters `json:"filters"`
	MapCenter *struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"map_center"`
	Bounds interface{} `json:"bounds"` // always null from frontend; do not use for export
}

// ExportMappingFilters mirrors frontend MappingFilters for export
type ExportMappingFilters struct {
	DistrictSubdistrict   []string `json:"district_subdistrict"`
	BuildingType          []string `json:"building_type"`
	BuildingGrade         []string `json:"building_grade"`
	Progress              []string `json:"progress"`
	LcdPresence           []string `json:"lcd_presence"`
	Sellable              []string `json:"sellable"`
	Connectivity          []string `json:"connectivity"`
	Year                  [2]int   `json:"year"` // [min, max]
	SalesPackageIds       []int    `json:"sales_package_ids"`
	BuildingRestrictionIds []int   `json:"building_restriction_ids"`
	Lat                   *float64 `json:"lat"`
	Lng                   *float64 `json:"lng"`
	Radius                *float64 `json:"radius"` // km; backend expects meters
	PoiID                 *int     `json:"poi_id"`
	Polygon []struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"polygon"`
}

// BuildMappingRequestFromExportBody maps ExportMappingByFilterRequest into MappingBuildingRequest.
// Bounds are never set (export is always all buildings matching filters).
func BuildMappingRequestFromExportBody(body *ExportMappingByFilterRequest) MappingBuildingRequest {
	var req MappingBuildingRequest
	f := body.Filters

	if len(f.DistrictSubdistrict) > 0 {
		req.SetSubdistrict(strings.Join(f.DistrictSubdistrict, ","))
	}
	if len(f.BuildingType) > 0 {
		req.SetBuildingType(strings.Join(f.BuildingType, ","))
	}
	if len(f.BuildingGrade) > 0 {
		req.SetBuildingGrade(strings.Join(f.BuildingGrade, ","))
	}
	if len(f.Progress) > 0 {
		req.SetProgress(strings.Join(f.Progress, ","))
	}
	if len(f.Sellable) > 0 {
		req.SetSellable(strings.Join(f.Sellable, ","))
	}
	if len(f.Connectivity) > 0 {
		req.SetConnectivity(strings.Join(f.Connectivity, ","))
	}
	if len(f.LcdPresence) > 0 {
		req.SetLCDPresence(strings.Join(f.LcdPresence, ","))
	}
	if f.Year[0] != 0 || f.Year[1] != 0 {
		req.SetYear(fmt.Sprintf("%d,%d", f.Year[0], f.Year[1]))
	}
	if len(f.SalesPackageIds) > 0 {
		req.SetSalesPackageIds(intSliceToComma(f.SalesPackageIds))
	}
	if len(f.BuildingRestrictionIds) > 0 {
		req.SetBuildingRestrictionIds(intSliceToComma(f.BuildingRestrictionIds))
	}
	if f.Lat != nil {
		req.SetLat(fmt.Sprintf("%v", *f.Lat))
	} else if body.MapCenter != nil {
		req.SetLat(fmt.Sprintf("%v", body.MapCenter.Lat))
	}
	if f.Lng != nil {
		req.SetLng(fmt.Sprintf("%v", *f.Lng))
	} else if body.MapCenter != nil {
		req.SetLng(fmt.Sprintf("%v", body.MapCenter.Lng))
	}
	if f.Radius != nil {
		req.SetRadius(fmt.Sprintf("%.0f", *f.Radius*1000)) // km -> meters
	}
	if f.PoiID != nil {
		req.SetPOIId(strconv.Itoa(*f.PoiID))
	}
	if len(f.Polygon) >= 3 {
		polyJSON, _ := json.Marshal(f.Polygon)
		req.SetPolygon(string(polyJSON))
	}
	// Bounds are never set: export is all buildings matching filters
	return req
}

func intSliceToComma(ids []int) string {
	if len(ids) == 0 {
		return ""
	}
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = strconv.Itoa(id)
	}
	return strings.Join(s, ",")
}

var _ web.RequestPagination = (*BuildingRequestFindAll)(nil)
var _ web.RequestOrder = (*BuildingRequestFindAll)(nil)

