package domain

type User struct {
	ID       string `bson:"_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
	Salt     string `json:"salt,omitempty"`
}
