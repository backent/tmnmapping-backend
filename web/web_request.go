package web

import (
	"net/http"
	"strconv"
)

type RequestPagination interface {
	SetSkip(skip int)
	SetTake(take int)
	GetSkip() int
	GetTake() int
}

func SetPagination(request RequestPagination, r *http.Request) {
	if r.URL.Query().Has("take") {
		take, err := strconv.Atoi(r.URL.Query().Get("take"))
		if err != nil {
			panic(err)
		}
		request.SetTake(take)
	} else {
		request.SetTake(10)
	}

	if r.URL.Query().Has("skip") {
		skip, err := strconv.Atoi(r.URL.Query().Get("skip"))
		if err != nil {
			panic(err)
		}
		request.SetSkip(skip)
	} else {
		request.SetSkip(0)
	}
}

type RequestOrder interface {
	SetOrderBy(orderBy string)
	SetOrderDirection(orderDirection string)
	GetOrderBy() string
	GetOrderDirection() string
}

func SetOrder(request RequestOrder, r *http.Request) {
	if r.URL.Query().Has("orderBy") {
		orderBy := r.URL.Query().Get("orderBy")
		request.SetOrderBy(orderBy)
	}

	if r.URL.Query().Has("orderDirection") {
		orderDirection := r.URL.Query().Get("orderDirection")
		request.SetOrderDirection(orderDirection)
	}
}

type RequestSearch interface {
	SetSearch(search string)
}

func SetSearch(request RequestSearch, r *http.Request) {
	if r.URL.Query().Has("search") {
		search := r.URL.Query().Get("search")
		if search != "" {
			request.SetSearch(search)
		}
	}
}

type RequestFilter interface {
	SetBuildingStatus(buildingStatus string)
	SetSellable(sellable string)
	SetConnectivity(connectivity string)
	SetResourceType(resourceType string)
	SetCompetitorLocation(competitorLocation *bool)
	SetCbdArea(cbdArea string)
	SetSubdistrict(subdistrict string)
	SetCitytown(citytown string)
	SetProvince(province string)
	SetGradeResource(gradeResource string)
	SetBuildingType(buildingType string)
}

func SetFilters(request RequestFilter, r *http.Request) {
	if r.URL.Query().Has("building_status") {
		buildingStatus := r.URL.Query().Get("building_status")
		if buildingStatus != "" {
			request.SetBuildingStatus(buildingStatus)
		}
	}

	if r.URL.Query().Has("sellable") {
		sellable := r.URL.Query().Get("sellable")
		if sellable != "" {
			request.SetSellable(sellable)
		}
	}

	if r.URL.Query().Has("connectivity") {
		connectivity := r.URL.Query().Get("connectivity")
		if connectivity != "" {
			request.SetConnectivity(connectivity)
		}
	}

	if r.URL.Query().Has("resource_type") {
		resourceType := r.URL.Query().Get("resource_type")
		if resourceType != "" {
			request.SetResourceType(resourceType)
		}
	}

	if r.URL.Query().Has("competitor_location") {
		competitorLocationStr := r.URL.Query().Get("competitor_location")
		if competitorLocationStr != "" {
			competitorLocation, err := strconv.ParseBool(competitorLocationStr)
			if err == nil {
				request.SetCompetitorLocation(&competitorLocation)
			}
		}
	}

	if r.URL.Query().Has("cbd_area") {
		cbdArea := r.URL.Query().Get("cbd_area")
		if cbdArea != "" {
			request.SetCbdArea(cbdArea)
		}
	}

	if r.URL.Query().Has("subdistrict") {
		subdistrict := r.URL.Query().Get("subdistrict")
		if subdistrict != "" {
			request.SetSubdistrict(subdistrict)
		}
	}

	if r.URL.Query().Has("citytown") {
		citytown := r.URL.Query().Get("citytown")
		if citytown != "" {
			request.SetCitytown(citytown)
		}
	}

	if r.URL.Query().Has("province") {
		province := r.URL.Query().Get("province")
		if province != "" {
			request.SetProvince(province)
		}
	}

	if r.URL.Query().Has("grade_resource") {
		gradeResource := r.URL.Query().Get("grade_resource")
		if gradeResource != "" {
			request.SetGradeResource(gradeResource)
		}
	}

	if r.URL.Query().Has("building_type") {
		buildingType := r.URL.Query().Get("building_type")
		if buildingType != "" {
			request.SetBuildingType(buildingType)
		}
	}
}

// MappingFilter interface for mapping-specific filters
type MappingFilter interface {
	SetBuildingType(buildingType string)
	SetBuildingGrade(buildingGrade string)
	SetYear(year string)
	SetSubdistrict(subdistrict string)
	SetProgress(progress string)
	SetSellable(sellable string)
	SetConnectivity(connectivity string)
	SetLCDPresence(lcdPresence string)
}

// SetMappingFilters parses mapping-specific filter parameters from query string
// Supports both filter[key] format (from frontend) and flat key format
func SetMappingFilters(request MappingFilter, r *http.Request) {
	// Helper function to get value from either filter[key] or key format
	getFilterValue := func(key string) string {
		// Try filter[key] format first (frontend format)
		if val := r.URL.Query().Get("filter[" + key + "]"); val != "" {
			return val
		}
		// Fall back to flat key format
		return r.URL.Query().Get(key)
	}

	// Building type - handle comma-separated values
	if buildingType := getFilterValue("building_type"); buildingType != "" {
		request.SetBuildingType(buildingType)
	}

	// Building grade - handle comma-separated values
	if buildingGrade := getFilterValue("building_grade"); buildingGrade != "" {
		request.SetBuildingGrade(buildingGrade)
	}

	// Year
	if year := getFilterValue("year"); year != "" {
		request.SetYear(year)
	}

	// Region (district/subdistrict)
	if region := getFilterValue("district_subdistrict"); region != "" {
		request.SetSubdistrict(region)
	}
	// Also support direct subdistrict parameter
	if subdistrict := getFilterValue("subdistrict"); subdistrict != "" {
		request.SetSubdistrict(subdistrict)
	}

	// Progress
	if progress := getFilterValue("progress"); progress != "" {
		request.SetProgress(progress)
	}

	// Sellable
	if sellable := getFilterValue("sellable"); sellable != "" {
		request.SetSellable(sellable)
	}

	// Connectivity
	if connectivity := getFilterValue("connectivity"); connectivity != "" {
		request.SetConnectivity(connectivity)
	}

	// LCD Presence
	if lcdPresence := getFilterValue("lcd_presence"); lcdPresence != "" {
		request.SetLCDPresence(lcdPresence)
	}
}
