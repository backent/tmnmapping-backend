package testutil

import (
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// NewMockDB creates a sqlmock-backed *sql.DB for unit tests.
// All calls to db.Begin() are intercepted and return a mock *sql.Tx.
// Call sqlMock.ExpectBegin(), sqlMock.ExpectCommit() or sqlMock.ExpectRollback()
// before invoking any service method that uses a transaction.
func NewMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}
