package model

type User struct {
	Username string   `json:"username omitempty" bson:"username omitempty"`
	Email    string   `json:"email omitempty" bson:"email omitempty"`
	Password string   `json:"password omitempty" bson:"password omitempty"`
	Scopes   []string `json:"scopes omitempty" bson:"scopes omitempty"`
}
