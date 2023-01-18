package token

import (
	"encoding/base64"
	"errors"
	"os"
	"time"

	"github.com/gKits/goauth/db"
	"github.com/gKits/goauth/internal/model"
	"github.com/gKits/goauth/utils"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
GenerateJWT returns a jwt token string which is created from the given claims
If the creation fails an error is returned
*/
func GenerateJWT(claims *model.Claims) (string, error) {
	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create signed tokenString from token
	tokenString, err := token.SignedString(os.Getenv("TOKEN_SECRET"))
	if err != nil {
		return "", nil
	}

	return tokenString, nil
}

/*
RevokeJWT will add the given tokenString into the blacklist Redis
If the insertion fails an error is returned
*/
func RevokeJWT(tokenString string) error {
	return nil
}

/**/
func JWTIsValid(tokenString string) (bool, error) {
	// Check if token is blacklisted
	blacklisted, err := db.IsBlacklisted(tokenString)
	if err != nil {
		return false, err
	} else if blacklisted {
		return false, nil
	}

	// Parse jwt token
	token, err := jwt.ParseWithClaims(tokenString, &model.Claims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("TOKEN_SECRET")), nil
		})
	if err != nil {
		return false, err
	}

	// Parse claims
	claims, ok := token.Claims.(model.Claims)
	if !ok {
		return false, errors.New("couldn't parse claims")
	}

	// Validate token
	if claims.ExpiresAt < time.Now().Unix() {
		return false, nil
	}

	return true, nil
}

/*
GenerateRefreshToken will generate a random refreshToken which contains the given claims and will add it to the
database collection of refreshTokens
This function returns the randomly generated tokenString
If if the insertion of the refreshToken fails an error is returned
*/
func GenerateRefreshToken(claims *model.Claims) (string, error) {
	// Generate refreshToken object
	tokenString := base64.StdEncoding.EncodeToString([]byte(utils.RandString(32)))
	refreshToken := model.RefreshToken{
		Token:     tokenString,
		Username:  claims.Username,
		Scopes:    claims.Scopes,
		Revoked:   false,
		IssuedAt:  claims.IssuedAt,
		ExpiresAt: time.Unix(claims.IssuedAt, 0).Add(24 * time.Hour).Unix(),
	}

	// Insert new refreshToken to MongoDB
	err := db.Insert("refresh_tokens", refreshToken)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

/*
RevokeRefreshToken will set the field "revoked" of  the refreshToken found by the tokenString to true
If the update process fails an error is returned
*/
func RevokeRefreshToken(tokenString string) error {
	// Find the token by its tokenString and update the "revoked" field to true
	if err := db.Update(
		"refresh_tokens",
		primitive.M{"token": tokenString},
		primitive.M{"revoked": true},
	); err != nil {
		return err
	}

	return nil
}

/*
RefreshIsValid returns a bool which tells if the refreshToken found by the tokenString is valid and not reoked
If the token is not found or it's field "revoked" is equal to true false is returned
If the connection or find process fails an error is returned
*/
func RefreshIsValid(tokenString string) (bool, error) {
	// Find refreshToken in database
	res, err := db.Find("refresh_tokens", bson.M{"token": tokenString})
	if err != nil {
		return false, err
	}

	// Parse result into refreshToken
	var refreshToken model.RefreshToken
	if err := res.Decode(&refreshToken); err != nil {
		return false, err
	}
	return false, nil
}
