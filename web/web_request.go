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
}

