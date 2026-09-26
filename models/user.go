package models

import "database/sql"

type User struct {
	Id                  int
	Username            string
	Name                string
	Email               string
	Password            string
	Role                string
	CanCreateQuotations bool
	SalesGroup          string
	CreatedAt           string
	UpdatedAt           string
}

// CanRaiseQuotations reports whether this user may create a quotation they own.
//
// The per-user flag is the rule, with one exemption: admin. Administration includes
// entering data on the business's behalf, and an admin can set the flag on their own
// account anyway -- denying them would be theatre, not a control.
//
// This lives on the model so the service that enforces it and the login response
// that advertises it cannot disagree. The frontend reads the answer rather than
// re-deriving the exemption.
func (u User) CanRaiseQuotations() bool {
	return u.CanCreateQuotations || HasRole(NormalizeRole(u.Role), RoleAdmin)
}

type NullAbleUser struct {
	Id                  sql.NullInt32
	Username            sql.NullString
	Name                sql.NullString
	Email               sql.NullString
	Password            sql.NullString
	Role                sql.NullString
	CanCreateQuotations sql.NullBool
	SalesGroup          sql.NullString
	CreatedAt           sql.NullString
	UpdatedAt           sql.NullString
}

var UserTable string = "users"

var UserProxySalesTable string = "user_proxy_sales"

func NullAbleUserToUser(nullAbleUser NullAbleUser) User {
	return User{
		Id:                  int(nullAbleUser.Id.Int32),
		Username:            nullAbleUser.Username.String,
		Name:                nullAbleUser.Name.String,
		Email:               nullAbleUser.Email.String,
		Password:            nullAbleUser.Password.String,
		Role:                nullAbleUser.Role.String,
		CanCreateQuotations: nullAbleUser.CanCreateQuotations.Bool,
		SalesGroup:          nullAbleUser.SalesGroup.String,
		CreatedAt:           nullAbleUser.CreatedAt.String,
		UpdatedAt:           nullAbleUser.UpdatedAt.String,
	}
}
