package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	serviceAuth "github.com/malikabdulaziz/tmn-backend/services/auth"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newAuthService wires up ServiceAuthImpl for tests.
func newAuthService(
	db *sql.DB,
	repoAuth *mocks.MockRepositoryAuth,
	repoUser *mocks.MockRepositoryUser,
) serviceAuth.ServiceAuthInterface {
	return serviceAuth.NewServiceAuthImpl(db, repoAuth, repoUser)
}

// TestLogin_HappyPath verifies that a valid username/password returns the correct
// LoginResponse and a non-empty JWT token.
func TestLogin_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoAuth := &mocks.MockRepositoryAuth{}
	repoUser := &mocks.MockRepositoryUser{}
	svc := newAuthService(db, repoAuth, repoUser)

	user := testutil.NewUser(42, "admin", "secret123", "admin")

	// CommitOrRollback calls tx.Commit() on success.
	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	repoUser.On("FindByUsername", mock.Anything, mock.AnythingOfType("*sql.Tx"), "admin").
		Return(user, nil)

	// Issue is called with the string representation of the user ID.
	repoAuth.On("Issue", "42").
		Return("jwt-token-value", nil)

	response, token := svc.Login(context.Background(), "admin", "secret123")

	assert.Equal(t, 42, response.User.Id)
	assert.Equal(t, "admin", response.User.Username)
	assert.Equal(t, "Test User", response.User.Name)
	assert.Equal(t, "admin", response.User.Role)
	assert.Equal(t, "jwt-token-value", token)

	repoUser.AssertExpectations(t)
	repoAuth.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestLogin_UserNotFound verifies that a missing user causes a BadRequestError panic.
// CommitOrRollback calls tx.Rollback() when a panic is in flight.
func TestLogin_UserNotFound(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoAuth := &mocks.MockRepositoryAuth{}
	repoUser := &mocks.MockRepositoryUser{}
	svc := newAuthService(db, repoAuth, repoUser)

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoUser.On("FindByUsername", mock.Anything, mock.AnythingOfType("*sql.Tx"), "nobody").
		Return(testutil.NewUser(0, "", "", ""), sql.ErrNoRows)

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "invalid credentials"},
		func() { svc.Login(context.Background(), "nobody", "anypassword") },
	)

	repoUser.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestLogin_WrongPassword verifies that a correct username but wrong password
// also causes a BadRequestError panic (same message for security).
func TestLogin_WrongPassword(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoAuth := &mocks.MockRepositoryAuth{}
	repoUser := &mocks.MockRepositoryUser{}
	svc := newAuthService(db, repoAuth, repoUser)

	user := testutil.NewUser(42, "admin", "correctpassword", "admin")

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoUser.On("FindByUsername", mock.Anything, mock.AnythingOfType("*sql.Tx"), "admin").
		Return(user, nil)

	assert.PanicsWithValue(t,
		exceptions.BadRequestError{Error: "invalid credentials"},
		func() { svc.Login(context.Background(), "admin", "wrongpassword") },
	)

	repoUser.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestLogin_TokenIssueFails verifies that a JWT Issue failure propagates as a panic.
func TestLogin_TokenIssueFails(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoAuth := &mocks.MockRepositoryAuth{}
	repoUser := &mocks.MockRepositoryUser{}
	svc := newAuthService(db, repoAuth, repoUser)

	user := testutil.NewUser(42, "admin", "secret", "admin")

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	repoUser.On("FindByUsername", mock.Anything, mock.AnythingOfType("*sql.Tx"), "admin").
		Return(user, nil)

	repoAuth.On("Issue", "42").
		Return("", errors.New("signing key unavailable"))

	assert.Panics(t, func() {
		svc.Login(context.Background(), "admin", "secret")
	})

	repoUser.AssertExpectations(t)
	repoAuth.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
