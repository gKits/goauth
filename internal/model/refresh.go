package model

type RefreshToken struct {
	Username  string   `json:"username" bson:"username"`
	Scopes    []string `json:"scopes" bson:"scopes"`
	Token     string   `json:"token" bson:"token"`
	Revoked   bool     `json:"revoked" bson:"revoked"`
	IssuedAt  int64    `json:"issuedAt" bson:"issuedAt"`
	ExpiresAt int64    `json:"expiresAt" bson:"expiresAt"`
}
