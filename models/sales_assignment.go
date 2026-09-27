package models

import "database/sql"

// SalesAssignment names the sales PIC for one customer + brand pair.
//
// This is what scopes the quotation wizard's customer list. users.sales_group does
// not: it is a reporting attribute that no rule reads.
type SalesAssignment struct {
	Id               int    `json:"id"`
	CustomerId       int    `json:"customer_id"`
	CustomerCode     string `json:"customer_code"`
	CustomerName     string `json:"customer_name"`
	BrandId          int    `json:"brand_id"`
	BrandCode        string `json:"brand_code"`
	BrandName        string `json:"brand_name"`
	SalesUserId      int    `json:"sales_user_id"`
	SalesUsername    string `json:"sales_username"`
	SalesName        string `json:"sales_name"`
	Status           string `json:"status"`
	RegistrationDate string `json:"registration_date"`
	ExpiryDate       string `json:"expiry_date"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type NullAbleSalesAssignment struct {
	Id               sql.NullInt64
	CustomerId       sql.NullInt64
	CustomerCode     sql.NullString
	CustomerName     sql.NullString
	BrandId          sql.NullInt64
	BrandCode        sql.NullString
	BrandName        sql.NullString
	SalesUserId      sql.NullInt64
	SalesUsername    sql.NullString
	SalesName        sql.NullString
	Status           sql.NullString
	RegistrationDate sql.NullString
	ExpiryDate       sql.NullString
	CreatedAt        sql.NullString
	UpdatedAt        sql.NullString
}

var SalesAssignmentTable string = "sales_assignments"

func NullAbleSalesAssignmentToSalesAssignment(n NullAbleSalesAssignment) SalesAssignment {
	return SalesAssignment{
		Id:               int(n.Id.Int64),
		CustomerId:       int(n.CustomerId.Int64),
		CustomerCode:     n.CustomerCode.String,
		CustomerName:     n.CustomerName.String,
		BrandId:          int(n.BrandId.Int64),
		BrandCode:        n.BrandCode.String,
		BrandName:        n.BrandName.String,
		SalesUserId:      int(n.SalesUserId.Int64),
		SalesUsername:    n.SalesUsername.String,
		SalesName:        n.SalesName.String,
		Status:           n.Status.String,
		RegistrationDate: n.RegistrationDate.String,
		ExpiryDate:       n.ExpiryDate.String,
		CreatedAt:        n.CreatedAt.String,
		UpdatedAt:        n.UpdatedAt.String,
	}
}
