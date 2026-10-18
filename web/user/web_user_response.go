package user

// UserResponse deliberately has no password field. The repository projection
// includes the hash because the auth service needs it, so the omission here is
// the thing that keeps it out of API output.
type UserResponse struct {
	Id                  int    `json:"id"`
	Username            string `json:"username"`
	Name                string `json:"name"`
	Email               string `json:"email"`
	Role                string `json:"role"`
	CanCreateQuotations bool   `json:"can_create_quotations"`
	SalesGroup          string `json:"sales_group"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}
