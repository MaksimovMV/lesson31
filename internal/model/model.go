package model

type User struct {
	ID      string   `json:"id,omitempty" bson:"_id,omitempty"`
	Name    string   `json:"name,omitempty" bson:"name,omitempty"`
	Age     int      `json:"age,omitempty" bson:"age,omitempty"`
	Friends []string `json:"friends,omitempty" bson:"friends,omitempty"`
}
