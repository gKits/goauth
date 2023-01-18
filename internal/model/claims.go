package model

import "github.com/golang-jwt/jwt"

type Claims struct {
	Username string   `json:"username"`
	Scopes   []string `json:"scopes"`
	jwt.StandardClaims
}
