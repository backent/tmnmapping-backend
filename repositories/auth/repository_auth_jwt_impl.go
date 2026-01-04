package auth

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/malikabdulaziz/tmn-backend/helpers"
)

type RepositoryAuthJWTImpl struct {
	secretKeys    []byte
	tokenLifeTime int
}

func NewRepositoryAuthJWTImpl() RepositoryAuthInterface {

	tokenLifeTime, err := strconv.Atoi(os.Getenv("APP_TOKEN_EXPIRE_IN_SEC"))
	helpers.PanicIfError(err)

	return &RepositoryAuthJWTImpl{
		secretKeys:    []byte(os.Getenv("APP_SECRET_KEY")),
		tokenLifeTime: tokenLifeTime,
	}
}

func (implementation *RepositoryAuthJWTImpl) Issue(payload string) (string, error) {
	// Create the Claims
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(implementation.tokenLifeTime))),
		Issuer:    payload,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	generatedToken, err := token.SignedString(implementation.secretKeys)
	return generatedToken, err
}

func (implementation *RepositoryAuthJWTImpl) Validate(tokenString string) (int, bool) {

	// Parse takes the token string and a function for looking up the key
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return implementation.secretKeys, nil
	})

	if payload, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if id, ok := payload["iss"].(string); ok {
			intId, err := strconv.Atoi(id)
			if err != nil {
				return 0, false
			}
			return intId, true
		} else {
			return 0, false
		}
	} else {
		return 0, false
	}
}

