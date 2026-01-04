package web

type WebResponse struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Data   interface{} `json:"data"`
	Extras interface{} `json:"extras"`
}

type Pagination struct {
	Take  int `json:"take"`
	Skip  int `json:"skip"`
	Total int `json:"total"`
}

