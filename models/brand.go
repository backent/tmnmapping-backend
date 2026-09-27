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

	// The person a quotation for this brand is addressed to. The quotation wizard
	// prefills from here; the quotation then keeps its own copy, so editing these
	// never rewrites a quotation already sent.
	AttentionTo  string `json:"attention_to"`
	JobTitle     string `json:"job_title"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
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
	AttentionTo  sql.NullString
	JobTitle     sql.NullString
	ContactPhone sql.NullString
	ContactEmail sql.NullString
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
		AttentionTo:  n.AttentionTo.String,
		JobTitle:     n.JobTitle.String,
		ContactPhone: n.ContactPhone.String,
		ContactEmail: n.ContactEmail.String,
		CreatedAt:    n.CreatedAt.String,
		UpdatedAt:    n.UpdatedAt.String,
	}
}
