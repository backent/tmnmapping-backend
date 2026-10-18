package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryUserImpl struct {
}

func NewRepositoryUserImpl() RepositoryUserInterface {
	return &RepositoryUserImpl{}
}

// userColumns is the full projection. Password is included because the auth service
// needs it; the web layer never puts it in a response.
const userColumns = "id, username, name, email, password, role, can_create_quotations, sales_group, created_at, updated_at"

var userAllowedOrderBy = map[string]bool{
	"id": true, "username": true, "name": true, "email": true,
	"role": true, "created_at": true, "updated_at": true,
}

var userAllowedOrderDir = map[string]bool{"ASC": true, "DESC": true}

// userSafeOrder keeps the ORDER BY clause off the interpolation path.
func userSafeOrder(orderBy, orderDirection string) (string, string) {
	if !userAllowedOrderBy[orderBy] {
		orderBy = "created_at"
	}
	if !userAllowedOrderDir[orderDirection] {
		orderDirection = "DESC"
	}

	return orderBy, orderDirection
}

func scanUser(rows *sql.Rows) (models.User, error) {
	var n models.NullAbleUser

	err := rows.Scan(&n.Id, &n.Username, &n.Name, &n.Email, &n.Password, &n.Role,
		&n.CanCreateQuotations, &n.SalesGroup, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.User{}, err
	}

	return models.NullAbleUserToUser(n), nil
}

func (repository *RepositoryUserImpl) Create(ctx context.Context, tx *sql.Tx, user models.User) (models.User, error) {
	SQL := "INSERT INTO " + models.UserTable +
		" (username, name, email, password, role, can_create_quotations, sales_group)" +
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at"

	err := tx.QueryRowContext(ctx, SQL,
		user.Username, user.Name, user.Email, user.Password,
		user.Role, user.CanCreateQuotations, nullIfEmpty(user.SalesGroup),
	).Scan(&user.Id, &user.CreatedAt, &user.UpdatedAt)

	return user, err
}

func (repository *RepositoryUserImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.User, error) {
	orderBy, orderDirection = userSafeOrder(orderBy, orderDirection)

	var rows *sql.Rows
	var err error

	if search != "" {
		SQL := "SELECT " + userColumns + " FROM " + models.UserTable +
			" WHERE username ILIKE $1 OR name ILIKE $1 OR email ILIKE $1" +
			" ORDER BY " + orderBy + " " + orderDirection + ", username ASC LIMIT $2 OFFSET $3"
		rows, err = tx.QueryContext(ctx, SQL, "%"+search+"%", take, skip)
	} else {
		SQL := "SELECT " + userColumns + " FROM " + models.UserTable +
			" ORDER BY " + orderBy + " " + orderDirection + ", username ASC LIMIT $1 OFFSET $2"
		rows, err = tx.QueryContext(ctx, SQL, take, skip)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, user)
	}

	return list, rows.Err()
}

func (repository *RepositoryUserImpl) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	var total int
	var err error

	if search != "" {
		SQL := "SELECT COUNT(*) FROM " + models.UserTable +
			" WHERE username ILIKE $1 OR name ILIKE $1 OR email ILIKE $1"
		err = tx.QueryRowContext(ctx, SQL, "%"+search+"%").Scan(&total)
	} else {
		SQL := "SELECT COUNT(*) FROM " + models.UserTable
		err = tx.QueryRowContext(ctx, SQL).Scan(&total)
	}

	return total, err
}

func (repository *RepositoryUserImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.User, error) {
	SQL := "SELECT " + userColumns + " FROM " + models.UserTable + " WHERE id = $1"
	rows, err := tx.QueryContext(ctx, SQL, id)
	if err != nil {
		return models.User{}, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanUser(rows)
	}

	return models.User{}, sql.ErrNoRows
}

func (repository *RepositoryUserImpl) FindByUsername(ctx context.Context, tx *sql.Tx, username string) (models.User, error) {
	SQL := "SELECT " + userColumns + " FROM " + models.UserTable + " WHERE username = $1"
	rows, err := tx.QueryContext(ctx, SQL, username)
	if err != nil {
		return models.User{}, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanUser(rows)
	}

	return models.User{}, sql.ErrNoRows
}

// Update writes the profile fields. The password has its own method so that an
// update request without a password cannot blank it out.
func (repository *RepositoryUserImpl) Update(ctx context.Context, tx *sql.Tx, user models.User) (models.User, error) {
	SQL := "UPDATE " + models.UserTable +
		" SET username = $1, name = $2, email = $3, role = $4," +
		" can_create_quotations = $5, sales_group = $6, updated_at = $7" +
		" WHERE id = $8 RETURNING updated_at"

	err := tx.QueryRowContext(ctx, SQL,
		user.Username, user.Name, user.Email, user.Role,
		user.CanCreateQuotations, nullIfEmpty(user.SalesGroup), time.Now(), user.Id,
	).Scan(&user.UpdatedAt)

	return user, err
}

func (repository *RepositoryUserImpl) UpdatePassword(ctx context.Context, tx *sql.Tx, id int, hashedPassword string) error {
	SQL := "UPDATE " + models.UserTable + " SET password = $1, updated_at = $2 WHERE id = $3"
	_, err := tx.ExecContext(ctx, SQL, hashedPassword, time.Now(), id)

	return err
}

func (repository *RepositoryUserImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := "DELETE FROM " + models.UserTable + " WHERE id = $1"
	_, err := tx.ExecContext(ctx, SQL, id)

	return err
}

// CountByRole backs the "do not remove the last admin" guard.
func (repository *RepositoryUserImpl) CountByRole(ctx context.Context, tx *sql.Tx, role string) (int, error) {
	var total int
	SQL := "SELECT COUNT(*) FROM " + models.UserTable + " WHERE role = $1"
	err := tx.QueryRowContext(ctx, SQL, role).Scan(&total)

	return total, err
}

func (repository *RepositoryUserImpl) CreateLoginLog(ctx context.Context, tx *sql.Tx, userId int, ipAddress string) error {
	SQL := "INSERT INTO " + models.UserLoginLogTable + " (user_id, ip_address) VALUES ($1, $2)"
	_, err := tx.ExecContext(ctx, SQL, userId, ipAddress)
	return err
}

func (repository *RepositoryUserImpl) FindLastLoginByUserId(ctx context.Context, tx *sql.Tx, userId int) (models.UserLoginLog, error) {
	SQL := "SELECT id, user_id, logged_in_at, ip_address FROM " + models.UserLoginLogTable + " WHERE user_id = $1 ORDER BY logged_in_at DESC LIMIT 1 OFFSET 1"
	rows, err := tx.QueryContext(ctx, SQL, userId)
	if err != nil {
		return models.UserLoginLog{}, err
	}
	defer rows.Close()

	log := models.NullAbleUserLoginLog{}
	if rows.Next() {
		err := rows.Scan(&log.Id, &log.UserId, &log.LoggedInAt, &log.IPAddress)
		if err != nil {
			return models.UserLoginLog{}, err
		}
		return models.NullAbleUserLoginLogToUserLoginLog(log), nil
	}
	return models.UserLoginLog{}, sql.ErrNoRows
}

// nullIfEmpty keeps sales_group NULL rather than "", which the CHECK constraint
// added in migration 015 would reject.
func nullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}

	return value
}

// FindByRole lists everyone holding a role. Approval routing needs exactly one
// holder; returning all of them lets the caller refuse an ambiguous directory rather
// than silently picking the first.
func (repository *RepositoryUserImpl) FindByRole(ctx context.Context, tx *sql.Tx, role string) ([]models.User, error) {
	SQL := "SELECT " + userColumns + " FROM " + models.UserTable + " WHERE role = $1 ORDER BY id"
	rows, err := tx.QueryContext(ctx, SQL, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, user)
	}

	return list, rows.Err()
}

// CanCreateOnBehalfOf reads the proxy-entry allow-list added in migration 015.
// Proxy entry is explicit per spec 2026-07-23: holding a role is not enough.
func (repository *RepositoryUserImpl) CanCreateOnBehalfOf(ctx context.Context, tx *sql.Tx, actorUserId int, ownerUserId int) (bool, error) {
	var count int
	SQL := "SELECT COUNT(*) FROM " + models.UserProxySalesTable +
		" WHERE actor_user_id = $1 AND owner_user_id = $2"
	err := tx.QueryRowContext(ctx, SQL, actorUserId, ownerUserId).Scan(&count)

	return count > 0, err
}
