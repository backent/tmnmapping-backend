package user

import (
	"context"

	webUser "github.com/malikabdulaziz/tmn-backend/web/user"
)

// ServiceUserInterface manages user accounts. Every method that can change or remove
// an account takes the acting user's id so the service can refuse the operations that
// would lock the caller — or everyone — out of the admin role.
type ServiceUserInterface interface {
	Create(ctx context.Context, request webUser.CreateUserRequest) webUser.UserResponse
	FindAll(ctx context.Context, request webUser.UserRequestFindAll) ([]webUser.UserResponse, int)
	FindById(ctx context.Context, id int) webUser.UserResponse
	Update(ctx context.Context, request webUser.UpdateUserRequest, id int, actorId int) webUser.UserResponse
	Delete(ctx context.Context, id int, actorId int)
}
