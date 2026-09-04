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
	}
}
