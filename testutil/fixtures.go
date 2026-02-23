package testutil

import (
	"github.com/malikabdulaziz/tmn-backend/models"
	"golang.org/x/crypto/bcrypt"
)

// NewUser creates a test User with a bcrypt-hashed password.
// Uses bcrypt.MinCost (4 rounds) so tests run fast.
func NewUser(id int, username, plainPassword, role string) models.User {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)
	return models.User{
		Id:       id,
		Username: username,
		Name:     "Test User",
		Email:    username + "@test.com",
		Password: string(hashed),
		Role:     role,
	}
}

// NewBuilding creates a minimal valid Building for tests.
func NewBuilding(id int, name string) models.Building {
	return models.Building{
		Id:            id,
		Name:          name,
		BuildingType:  "Office",
		GradeResource: "A",
		Latitude:      -6.2,
		Longitude:     106.8,
	}
}
