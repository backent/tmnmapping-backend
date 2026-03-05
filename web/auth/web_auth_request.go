package auth

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Remember bool   `json:"remember"`

	UserId int
}

