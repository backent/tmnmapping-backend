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

