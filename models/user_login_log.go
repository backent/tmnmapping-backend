package models

import "database/sql"

type UserLoginLog struct {
	Id          int
	UserId      int
	LoggedInAt  string
	IPAddress   string
}

type NullAbleUserLoginLog struct {
	Id         sql.NullInt64
	UserId     sql.NullInt64
	LoggedInAt sql.NullString
	IPAddress  sql.NullString
}

var UserLoginLogTable string = "user_login_logs"

func NullAbleUserLoginLogToUserLoginLog(n NullAbleUserLoginLog) UserLoginLog {
	return UserLoginLog{
		Id:         int(n.Id.Int64),
		UserId:     int(n.UserId.Int64),
		LoggedInAt: n.LoggedInAt.String,
		IPAddress:  n.IPAddress.String,
	}
}
