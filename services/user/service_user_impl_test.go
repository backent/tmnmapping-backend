package user_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	serviceUser "github.com/malikabdulaziz/tmn-backend/services/user"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	webUser "github.com/malikabdulaziz/tmn-backend/web/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func commitExpected(t *testing.T) (serviceUser.ServiceUserInterface, *mocks.MockRepositoryUser, func()) {
	svc, repoUser, _, assertMock := newService(t, true)
	return svc, repoUser, assertMock
}

func rollbackExpected(t *testing.T) (serviceUser.ServiceUserInterface, *mocks.MockRepositoryUser, func()) {
	svc, repoUser, _, assertMock := newService(t, false)
	return svc, repoUser, assertMock
}

// newService wires the service with both repositories it now depends on.
func newService(t *testing.T, commits bool) (serviceUser.ServiceUserInterface, *mocks.MockRepositoryUser, *mocks.MockRepositorySalesAssignment, func()) {
	t.Helper()

	db, sqlMock := testutil.NewMockDB(t)
	repoUser := &mocks.MockRepositoryUser{}
	repoAssignment := &mocks.MockRepositorySalesAssignment{}

	sqlMock.ExpectBegin()
	if commits {
		sqlMock.ExpectCommit()
	} else {
		sqlMock.ExpectRollback()
	}

	return serviceUser.NewServiceUserImpl(db, repoUser, repoAssignment), repoUser, repoAssignment, func() {
		assert.NoError(t, sqlMock.ExpectationsWereMet())
	}
}

func validCreateRequest() webUser.CreateUserRequest {
	return webUser.CreateUserRequest{
		Username:            "newsales",
		Name:                "New Sales",
		Email:               "newsales@test.com",
		Password:            "secret123",
		Role:                models.RoleSales,
		CanCreateQuotations: true,
		SalesGroup:          models.SalesGroupSalesTeam,
	}
}

func validUpdateRequest() webUser.UpdateUserRequest {
	return webUser.UpdateUserRequest{
		Username: "existing",
		Name:     "Existing User",
		Email:    "existing@test.com",
		Role:     models.RoleSales,
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestCreate_HashesPasswordAndNeverReturnsIt(t *testing.T) {
	svc, repoUser, assertMock := commitExpected(t)
	request := validCreateRequest()

	repoUser.On("FindByUsername", mock.Anything, mock.AnythingOfType("*sql.Tx"), "newsales").
		Return(models.User{}, sql.ErrNoRows)

	var stored models.User
	repoUser.On("Create", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.AnythingOfType("models.User")).
		Run(func(args mock.Arguments) { stored = args.Get(2).(models.User) }).
		Return(models.User{Id: 10, Username: "newsales", Name: "New Sales", Role: models.RoleSales}, nil)

	response := svc.Create(context.Background(), request)

	assert.Equal(t, 10, response.Id)
	assert.Equal(t, "newsales", response.Username)

	// The stored password must be a bcrypt hash of the plaintext, not the plaintext.
	assert.NotEqual(t, "secret123", stored.Password)
	assert.True(t, helpers.CheckPassword("secret123", stored.Password))

	repoUser.AssertExpectations(t)
	assertMock()
}

func TestCreate_RejectsDuplicateUsername(t *testing.T) {
	svc, repoUser, assertMock := rollbackExpected(t)

	repoUser.On("FindByUsername", mock.Anything, mock.AnythingOfType("*sql.Tx"), "newsales").
		Return(testutil.NewUser(1, "newsales", "whatever", models.RoleSales), nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("username is already taken"),
		func() { svc.Create(context.Background(), validCreateRequest()) })

	repoUser.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdate_LeavesPasswordAloneWhenBlank(t *testing.T) {
	svc, repoUser, assertMock := commitExpected(t)
	existing := testutil.NewUser(5, "existing", "original", models.RoleSales)

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).Return(existing, nil)
	repoUser.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.AnythingOfType("models.User")).
		Return(existing, nil)

	svc.Update(context.Background(), validUpdateRequest(), 5, 1)

	repoUser.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	repoUser.AssertExpectations(t)
	assertMock()
}

func TestUpdate_HashesNewPasswordWhenProvided(t *testing.T) {
	svc, repoUser, assertMock := commitExpected(t)
	existing := testutil.NewUser(5, "existing", "original", models.RoleSales)

	request := validUpdateRequest()
	request.Password = "brandnew123"

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).Return(existing, nil)
	repoUser.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.AnythingOfType("models.User")).
		Return(existing, nil)

	var hashed string
	repoUser.On("UpdatePassword", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5, mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) { hashed = args.String(3) }).
		Return(nil)

	svc.Update(context.Background(), request, 5, 1)

	assert.True(t, helpers.CheckPassword("brandnew123", hashed))
	repoUser.AssertExpectations(t)
	assertMock()
}

// An admin who demotes themselves loses access to this screen and needs SQL to recover.
func TestUpdate_RefusesSelfRoleChange(t *testing.T) {
	svc, repoUser, assertMock := rollbackExpected(t)
	existing := testutil.NewUser(3, "boss", "secret123", models.RoleAdmin)

	request := validUpdateRequest()
	request.Username = "boss"
	request.Role = models.RoleSales

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).Return(existing, nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("you cannot change your own role"),
		func() { svc.Update(context.Background(), request, 3, 3) })

	repoUser.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

func TestUpdate_AllowsEditingYourOwnProfileWithoutChangingRole(t *testing.T) {
	svc, repoUser, assertMock := commitExpected(t)
	existing := testutil.NewUser(3, "boss", "secret123", models.RoleAdmin)

	request := validUpdateRequest()
	request.Username = "boss"
	request.Name = "Renamed Boss"
	request.Role = models.RoleAdmin

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).Return(existing, nil)
	repoUser.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.AnythingOfType("models.User")).
		Return(existing, nil)

	svc.Update(context.Background(), request, 3, 3)

	repoUser.AssertExpectations(t)
	assertMock()
}

func TestUpdate_RejectsUsernameTakenByAnotherUser(t *testing.T) {
	svc, repoUser, assertMock := rollbackExpected(t)
	existing := testutil.NewUser(5, "existing", "secret123", models.RoleSales)

	request := validUpdateRequest()
	request.Username = "taken"

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 5).Return(existing, nil)
	repoUser.On("FindByUsername", mock.Anything, mock.AnythingOfType("*sql.Tx"), "taken").
		Return(testutil.NewUser(9, "taken", "secret123", models.RoleSales), nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("username is already taken"),
		func() { svc.Update(context.Background(), request, 5, 1) })

	assertMock()
}

// Demoting the only admin would leave nobody able to manage users or master data.
func TestUpdate_RefusesToDemoteTheLastAdmin(t *testing.T) {
	svc, repoUser, assertMock := rollbackExpected(t)
	existing := testutil.NewUser(7, "onlyadmin", "secret123", models.RoleAdmin)

	request := validUpdateRequest()
	request.Username = "onlyadmin"
	request.Role = models.RoleSales

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 7).Return(existing, nil)
	repoUser.On("CountByRole", mock.Anything, mock.AnythingOfType("*sql.Tx"), models.RoleAdmin).Return(1, nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("this is the last admin account; promote another user first"),
		func() { svc.Update(context.Background(), request, 7, 1) })

	repoUser.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

func TestUpdate_AllowsDemotingAnAdminWhenOthersRemain(t *testing.T) {
	svc, repoUser, assertMock := commitExpected(t)
	existing := testutil.NewUser(7, "someadmin", "secret123", models.RoleAdmin)

	request := validUpdateRequest()
	request.Username = "someadmin"
	request.Role = models.RoleSales

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 7).Return(existing, nil)
	repoUser.On("CountByRole", mock.Anything, mock.AnythingOfType("*sql.Tx"), models.RoleAdmin).Return(3, nil)
	repoUser.On("Update", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.AnythingOfType("models.User")).
		Return(existing, nil)

	svc.Update(context.Background(), request, 7, 1)

	repoUser.AssertExpectations(t)
	assertMock()
}

func TestUpdate_NotFound(t *testing.T) {
	svc, repoUser, assertMock := rollbackExpected(t)

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.User{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NewNotFoundError("user not found"),
		func() { svc.Update(context.Background(), validUpdateRequest(), 404, 1) })

	assertMock()
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete_RefusesSelfDeletion(t *testing.T) {
	svc, repoUser, assertMock := rollbackExpected(t)
	existing := testutil.NewUser(3, "boss", "secret123", models.RoleAdmin)

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 3).Return(existing, nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("you cannot delete your own account"),
		func() { svc.Delete(context.Background(), 3, 3) })

	repoUser.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

func TestDelete_RefusesToRemoveTheLastAdmin(t *testing.T) {
	svc, repoUser, assertMock := rollbackExpected(t)
	existing := testutil.NewUser(7, "onlyadmin", "secret123", models.RoleAdmin)

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 7).Return(existing, nil)
	repoUser.On("CountByRole", mock.Anything, mock.AnythingOfType("*sql.Tx"), models.RoleAdmin).Return(1, nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("this is the last admin account; promote another user first"),
		func() { svc.Delete(context.Background(), 7, 1) })

	repoUser.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

func TestDelete_HappyPath(t *testing.T) {
	svc, repoUser, repoAssignment, assertMock := newService(t, true)
	existing := testutil.NewUser(8, "somesales", "secret123", models.RoleSales)

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 8).Return(existing, nil)
	repoAssignment.On("CountBySalesUser", mock.Anything, mock.AnythingOfType("*sql.Tx"), 8).Return(0, nil)
	repoUser.On("Delete", mock.Anything, mock.AnythingOfType("*sql.Tx"), 8).Return(nil)

	svc.Delete(context.Background(), 8, 1)

	repoUser.AssertExpectations(t)
	repoAssignment.AssertExpectations(t)
	assertMock()
}

// sales_assignments.sales_user_id is ON DELETE RESTRICT, so the service checks first
// rather than letting a foreign-key violation surface as a 500.
func TestDelete_RefusesWhileUserHoldsAssignments(t *testing.T) {
	svc, repoUser, repoAssignment, assertMock := newService(t, false)
	existing := testutil.NewUser(8, "somesales", "secret123", models.RoleSales)

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 8).Return(existing, nil)
	repoAssignment.On("CountBySalesUser", mock.Anything, mock.AnythingOfType("*sql.Tx"), 8).Return(3, nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("this user is the sales PIC for 3 customer/brand assignment(s); reassign them first"),
		func() { svc.Delete(context.Background(), 8, 1) })

	repoUser.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

// ---------------------------------------------------------------------------
// Reads
// ---------------------------------------------------------------------------

func TestFindAll_NormalizesLegacyRolesInResponses(t *testing.T) {
	svc, repoUser, assertMock := commitExpected(t)

	repoUser.On("FindAll", mock.Anything, mock.AnythingOfType("*sql.Tx"), 0, 0, "created_at", "DESC", "").
		Return([]models.User{
			testutil.NewUser(1, "legacy", "secret123", "approver"),
			testutil.NewUser(2, "current", "secret123", models.RoleCEO),
		}, nil)
	repoUser.On("CountAll", mock.Anything, mock.AnythingOfType("*sql.Tx"), "").Return(2, nil)

	var request webUser.UserRequestFindAll
	responses, total := svc.FindAll(context.Background(), request)

	assert.Equal(t, 2, total)
	assert.Equal(t, models.RoleHeadOfSales, responses[0].Role)
	assert.Equal(t, models.RoleCEO, responses[1].Role)

	assertMock()
}

func TestFindById_NotFound(t *testing.T) {
	svc, repoUser, assertMock := rollbackExpected(t)

	repoUser.On("FindById", mock.Anything, mock.AnythingOfType("*sql.Tx"), 404).
		Return(models.User{}, sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.NewNotFoundError("user not found"),
		func() { svc.FindById(context.Background(), 404) })

	assertMock()
}
