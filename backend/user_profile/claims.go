package user_profile

import "github.com/golang-jwt/jwt"

type Claims struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	jwt.StandardClaims
}
