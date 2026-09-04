package auth

type LoginResponse struct {
	User UserResponse `json:"user"`
}

type UserResponse struct {
	Id                  int    `json:"id"`
	Username            string `json:"username"`
	Name                string `json:"name"`
	Role                string `json:"role"`
	CanCreateQuotations bool   `json:"can_create_quotations"`
	SalesGroup          string `json:"sales_group"`
	LastLogin           string `json:"last_login"`
}
