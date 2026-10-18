package models

import "database/sql"

// Customer is an advertiser that buys campaigns.
//
// Not to be confused with MotherBrand, which is a retail/competitor concept used by
// the POI map layer. See docs/QUOTATION_FEATURE_ANALYSIS.md §2.3.
type Customer struct {
	Id        int    `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Industry  string `json:"industry"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type NullAbleCustomer struct {
	Id        sql.NullInt64
	Code      sql.NullString
	Name      sql.NullString
	Industry  sql.NullString
	Status    sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

var CustomerTable string = "customers"

// Master data status values, shared by customers, brands and sales assignments.
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

var MasterDataStatuses = []string{StatusActive, StatusInactive}

func IsValidMasterDataStatus(status string) bool {
	for _, known := range MasterDataStatuses {
		if known == status {
			return true
		}
	}

	return false
}

func NullAbleCustomerToCustomer(n NullAbleCustomer) Customer {
	return Customer{
		Id:        int(n.Id.Int64),
		Code:      n.Code.String,
		Name:      n.Name.String,
		Industry:  n.Industry.String,
		Status:    n.Status.String,
		CreatedAt: n.CreatedAt.String,
		UpdatedAt: n.UpdatedAt.String,
	}
}
