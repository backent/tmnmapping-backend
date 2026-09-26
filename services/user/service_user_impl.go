package user

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesSalesAssignment "github.com/malikabdulaziz/tmn-backend/repositories/salesassignment"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	webUser "github.com/malikabdulaziz/tmn-backend/web/user"
)

type ServiceUserImpl struct {
	DB *sql.DB
	repositoriesUser.RepositoryUserInterface
	RepositorySalesAssignment repositoriesSalesAssignment.RepositorySalesAssignmentInterface
}

func NewServiceUserImpl(
	db *sql.DB,
	repoUser repositoriesUser.RepositoryUserInterface,
	repoSalesAssignment repositoriesSalesAssignment.RepositorySalesAssignmentInterface,
) ServiceUserInterface {
	return &ServiceUserImpl{
		DB:                        db,
		RepositoryUserInterface:   repoUser,
		RepositorySalesAssignment: repoSalesAssignment,
	}
}

func (s *ServiceUserImpl) Create(ctx context.Context, request webUser.CreateUserRequest) webUser.UserResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	// users.username is UNIQUE; check first so the caller gets a readable message
	// rather than a driver constraint error.
	_, err = s.RepositoryUserInterface.FindByUsername(ctx, tx, request.Username)
	if err == nil {
		panic(exceptions.NewBadRequestError("username is already taken"))
	}
	if err != sql.ErrNoRows {
		helpers.PanicIfError(err)
	}

	hashed, err := helpers.HashPassword(request.Password)
	helpers.PanicIfError(err)

	created, err := s.RepositoryUserInterface.Create(ctx, tx, models.User{
		Username:            request.Username,
		Name:                request.Name,
		Email:               request.Email,
		Password:            hashed,
		Role:                request.Role,
		CanCreateQuotations: request.CanCreateQuotations,
		SalesGroup:          request.SalesGroup,
	})
	helpers.PanicIfError(err)

	return userModelToResponse(created)
}

func (s *ServiceUserImpl) FindAll(ctx context.Context, request webUser.UserRequestFindAll) ([]webUser.UserResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryUserInterface.FindAll(ctx, tx,
		request.GetTake(), request.GetSkip(),
		request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch())
	helpers.PanicIfError(err)

	total, err := s.RepositoryUserInterface.CountAll(ctx, tx, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webUser.UserResponse, len(list))
	for i, user := range list {
		responses[i] = userModelToResponse(user)
	}

	return responses, total
}

func (s *ServiceUserImpl) FindById(ctx context.Context, id int) webUser.UserResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	user, err := s.RepositoryUserInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("user not found"))
	}
	helpers.PanicIfError(err)

	return userModelToResponse(user)
}

func (s *ServiceUserImpl) Update(ctx context.Context, request webUser.UpdateUserRequest, id int, actorId int) webUser.UserResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryUserInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("user not found"))
	}
	helpers.PanicIfError(err)

	// There is no way back: an admin who demotes themselves loses access to this
	// screen and would need SQL to recover.
	if id == actorId && request.Role != models.NormalizeRole(existing.Role) {
		panic(exceptions.NewBadRequestError("you cannot change your own role"))
	}

	if request.Username != existing.Username {
		other, err := s.RepositoryUserInterface.FindByUsername(ctx, tx, request.Username)
		if err == nil && other.Id != id {
			panic(exceptions.NewBadRequestError("username is already taken"))
		}
		if err != nil && err != sql.ErrNoRows {
			helpers.PanicIfError(err)
		}
	}

	s.guardLastAdmin(ctx, tx, existing, request.Role)

	existing.Username = request.Username
	existing.Name = request.Name
	existing.Email = request.Email
	existing.Role = request.Role
	existing.CanCreateQuotations = request.CanCreateQuotations
	existing.SalesGroup = request.SalesGroup

	updated, err := s.RepositoryUserInterface.Update(ctx, tx, existing)
	helpers.PanicIfError(err)

	// An empty password means "leave it alone".
	if request.Password != "" {
		hashed, err := helpers.HashPassword(request.Password)
		helpers.PanicIfError(err)
		helpers.PanicIfError(s.RepositoryUserInterface.UpdatePassword(ctx, tx, id, hashed))
	}

	return userModelToResponse(updated)
}

func (s *ServiceUserImpl) Delete(ctx context.Context, id int, actorId int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryUserInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("user not found"))
	}
	helpers.PanicIfError(err)

	if id == actorId {
		panic(exceptions.NewBadRequestError("you cannot delete your own account"))
	}

	// Deleting the last admin would leave nobody able to manage users or master data.
	s.guardLastAdmin(ctx, tx, existing, "")

	// sales_assignments.sales_user_id is ON DELETE RESTRICT, so without this check
	// the operator would see a foreign-key violation surface as a 500.
	assignments, err := s.RepositorySalesAssignment.CountBySalesUser(ctx, tx, id)
	helpers.PanicIfError(err)
	if assignments > 0 {
		panic(exceptions.NewBadRequestError(
			"this user is the sales PIC for " + strconv.Itoa(assignments) +
				" customer/brand assignment(s); reassign them first"))
	}

	helpers.PanicIfError(s.RepositoryUserInterface.Delete(ctx, tx, id))
}

// guardLastAdmin refuses a change that would remove the final admin account.
// newRole is the role the user is moving to, or "" when they are being deleted.
func (s *ServiceUserImpl) guardLastAdmin(ctx context.Context, tx *sql.Tx, existing models.User, newRole string) {
	if models.NormalizeRole(existing.Role) != models.RoleAdmin || newRole == models.RoleAdmin {
		return
	}

	admins, err := s.RepositoryUserInterface.CountByRole(ctx, tx, models.RoleAdmin)
	helpers.PanicIfError(err)

	if admins <= 1 {
		panic(exceptions.NewBadRequestError("this is the last admin account; promote another user first"))
	}
}

func userModelToResponse(user models.User) webUser.UserResponse {
	return webUser.UserResponse{
		Id:                  user.Id,
		Username:            user.Username,
		Name:                user.Name,
		Email:               user.Email,
		Role:                models.NormalizeRole(user.Role),
		CanCreateQuotations: user.CanCreateQuotations,
		SalesGroup:          user.SalesGroup,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
	}
}
