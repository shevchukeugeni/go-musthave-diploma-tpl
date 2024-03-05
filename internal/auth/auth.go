package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth"
)

var TokenAuth *jwtauth.JWTAuth

const Secret = "secret-gofermart"

func init() {
	TokenAuth = jwtauth.New("HS256", []byte(Secret), nil)
}

func GetUserID(r *http.Request) (string, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return "", err
	}

	userID, ok := claims["user_id"]
	if !ok || fmt.Sprint(userID) == "" {
		return "", errors.New("invalid token")
	}

	return fmt.Sprint(userID), nil
}

func GenerateToken(userID string) (string, error) {
	_, tokenString, err := TokenAuth.Encode(map[string]interface{}{"user_id": userID, "exp": time.Now().Unix() + int64(time.Hour)})
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
