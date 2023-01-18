package auth

import (
	"strings"
	"time"

	"github.com/gKits/goauth/db"
	"github.com/gKits/goauth/internal/model"
	"github.com/gKits/goauth/token"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

/*
register takes the user data from the body and inserts the new user into the users collection of the MongoDB
If there is a need field for the user data missing the response will be 400
If the password encryption or database insertion fails it will be 500
*/
func register(c *fiber.Ctx) error {
	// Parse request body to user model
	var userData map[string]string
	if err := c.BodyParser(&userData); err != nil {
		return c.Status(400).JSON(&fiber.Map{"error": "user form "})
	}
	username, ok := userData["username"]
	if !ok {
		return c.Status(400).JSON(&fiber.Map{"error": "username missing"})
	}
	email, ok := userData["email"]
	if !ok {
		return c.Status(400).JSON(&fiber.Map{"error": "email missing"})
	}
	rawPassword, ok := userData["password"]
	if !ok {
		return c.Status(400).JSON(&fiber.Map{"error": "password missing"})
	}

	// Encrypt password
	bytePassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(&fiber.Map{"error": "encryption failed"})
	}
	password := string(bytePassword)

	// Create user model from userData
	user := &model.User{
		Username: username,
		Email:    email,
		Password: password,
	}

	// Insert user into database
	if err := db.Insert("users", user); err != nil {
		return c.Status(500).JSON(&fiber.Map{"error": "database insertion failed"})
	}

	return c.Status(200).JSON(&fiber.Map{"message": "successfully registered user"})
}

/*
login takes a username and password from the context body and validates those credentials with the users from the
database
After successful validation the response contains a token pair of an access and refresh token and attaches them to
the Authorization header and cookie respectively
On failed validation the function will respond with an error status
*/
func login(c *fiber.Ctx) error {
	// Parse credentials from body
	var creds map[string]string
	if err := c.BodyParser(&creds); err != nil {
		return c.Status(400).JSON(&fiber.Map{"error": "credentials missing"})
	}

	// Get username and password from request form
	username, ok := creds["username"]
	if !ok {
		return c.Status(401).JSON(&fiber.Map{"error": "credentials invalid"})
	}
	password, ok := creds["password"]
	if !ok {
		return c.Status(401).JSON(&fiber.Map{"error": "credentials invalid"})
	}

	// Mongo DB search user by username or email
	var user model.User
	res, err := db.Find(
		"users",
		bson.M{
			"$or": bson.M{
				"username": username,
				"email":    username,
			}},
	)
	if err != nil {
		return c.Status(401).JSON(&fiber.Map{"error": "credentials invalid"})
	}

	// Decode MongoDB result to user model
	if err = res.Decode(&user); err != nil {
		return c.Status(500).JSON(&fiber.Map{"error": "something went wrong"})
	}

	// Compare password to hashed password of user
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return c.Status(401).JSON(&fiber.Map{
			"error": "credentials invalid"})
	}

	// Current timestamp for consistent issuing timestamps
	issuedAt := time.Now()

	// Generate claims
	claims := &model.Claims{
		Username: user.Username,
		Scopes:   user.Scopes,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issuedAt.Unix(),
			ExpiresAt: issuedAt.Add(15 * time.Minute).Unix(),
		},
	}

	// Generate access token
	accessToken, err := token.GenerateJWT(claims)
	if err != nil {
		return c.Status(500).JSON(&fiber.Map{"error": "jwt generation failed"})
	}

	// Generate refresh token
	refreshToken, err := token.GenerateRefreshToken(claims)
	if err != nil {
		return c.Status(500).JSON(&fiber.Map{"error": "refresh token generation failed"})
	}

	// Attach access token to Authorization header
	c.Set("Authorization", "Bearer "+accessToken)

	// Attach refresh token to cookie
	c.Cookie(&fiber.Cookie{
		Name:    "refresh_token",
		Value:   refreshToken,
		Expires: time.Now(),
	})

	return c.Status(200).JSON(&fiber.Map{"access_token": accessToken, "refresh_token": refreshToken})
}

/*
logout revokes the access and refresh token that are used for the request and unsets the Authorization header and
clears the cookie
If either the refresh or access token is missing from the header or cookie respectively response is 400
If the revoking of either of the token fails the response is 500
*/
func logout(c *fiber.Ctx) error {
	// Get accessToken from header
	accessTokenString := c.Get("Authorization")
	if accessTokenString == "" {
		return c.Status(400).JSON(&fiber.Map{"error": "access token missing from header"})
	}
	accessTokenString = strings.TrimPrefix(accessTokenString, "Bearer ")
	c.Set("Authorization", "")

	// Get refreshToken from cookie
	refreshTokenString := c.Cookies("refresh_token")
	if refreshTokenString == "" {
		return c.Status(400).JSON(&fiber.Map{"error": "refresh token missing from cookie"})
	}
	c.ClearCookie("refresh_token")

	// Revoke tokens
	if err := token.RevokeJWT(accessTokenString); err != nil {
		return c.Status(500).JSON(&fiber.Map{"error": "jwt revoking failed"})
	}
	if err := token.RevokeRefreshToken(refreshTokenString); err != nil {
		return c.Status(500).JSON(&fiber.Map{"error": "refresh token revoking failed"})
	}

	return c.Status(200).JSON(&fiber.Map{"message": "logout successful"})
}

/*
refresh responds with a new pair of access and refresh tokens if the orginal accessToken is expired and the
original refreshToken is still valid
If the accessToken is not expired or blacklisted response will be 403
If the refreshToken is already expired or revoked response will be 401
If the process fails at any other point the response is 500
*/
func refresh(c *fiber.Ctx) error {
    // Get accessToken from Authorization header
    accessTokenString := c.Get("Authorization")
    if accessTokenString == "" {
        return c.Status(400).JSON(&fiber.Map{"error": "access token missing from header"})
    }
    accessTokenString = strings.TrimPrefix(accessTokenString, "Bearer ")
    c.Set("Authorization", "")

    // Get refreshToken from cookie
    refreshTokenString := c.Cookies("refresh_token")
    if refreshTokenString == "" {
        return c.Status(400).JSON(&fiber.Map{"error": "refresh token missing from cookie"})
    }
    c.ClearCookie("refresh_token")

    // Get claims
    var refreshToken model.RefreshToken
    res, err := db.Find("refresh_token", bson.M{"token": refreshTokenString})
    if err != nil {
        return c.Status(401).JSON(&fiber.Map{"error": "refresh token invalid"})
    }
    if err = res.Decode(&refreshToken); err != nil  {
        return c.Status(500).JSON(&fiber.Map{"error": "something went wrong"})
    }

    return c.Status(200).JSON(&fiber.Map{"message": ""})
}
