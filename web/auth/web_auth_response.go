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

	// Permissions is everything this caller's role holds, resolved from
	// models.Permissions. The frontend uses it to hide what the API would reject,
	// so there is only ever one copy of the policy.
	Permissions []string `json:"permissions"`
	LastLogin   string   `json:"last_login"`
}
