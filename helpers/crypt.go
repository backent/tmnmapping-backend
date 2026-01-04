package helpers

import "golang.org/x/crypto/bcrypt"

func HashPassword(plainPassword string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), 10)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func CheckPassword(plainPassword string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return false
	}
	return true
}

