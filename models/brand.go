package models

import "database/sql"

// Brand is one advertiser brand belonging to a Customer. A quotation is raised for
// a customer + brand pair.
type Brand struct {
	Id           int    `json:"id"`
	Code         string `json:"code"`
	CustomerId   int    `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	CustomerCode string `json:"customer_code"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type NullAbleBrand struct {
	Id           sql.NullInt64
	Code         sql.NullString
	CustomerId   sql.NullInt64
	CustomerName sql.NullString
	CustomerCode sql.NullString
	Name         sql.NullString
	Category     sql.NullString
	Status       sql.NullString
	CreatedAt    sql.NullString
	UpdatedAt    sql.NullString
}

var BrandTable string = "brands"

func NullAbleBrandToBrand(n NullAbleBrand) Brand {
	return Brand{
		Id:           int(n.Id.Int64),
		Code:         n.Code.String,
		CustomerId:   int(n.CustomerId.Int64),
		CustomerName: n.CustomerName.String,
		CustomerCode: n.CustomerCode.String,
		Name:         n.Name.String,
		Category:     n.Category.String,
		Status:       n.Status.String,
		CreatedAt:    n.CreatedAt.String,
		UpdatedAt:    n.UpdatedAt.String,
	}
}
