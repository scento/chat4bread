package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// ORM is the object relationship mapper for the MongoDB database.
type ORM struct {
	DB *mongo.Database
}

// GeoJSON defines geometric structures such that we can use MongoDB GIS operations.
type GeoJSON struct {
	Type   string `json:"type"`
	Coords []int  `json:"coordinates"`
}

// User object bundling all relevant information about an user.
type User struct {
	Phone    string   `json:"phone"`
	Name     *string  `json:"name"`
	Location *GeoJSON `json:"location"`

	Action string   `json:"action"`
	Reqs   []string `json:"requirements"`
}

// NewORM initializes the ORM.
func NewORM(client *mongo.Client, database string) *ORM {
	return &ORM{DB: client.Database(database)}
}

// UserByPhone looks for a user by its phone number/username.
func (orm *ORM) UserByPhone(phone string) (*User, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	users := orm.DB.Collection("users")
	var user User
	err := users.FindOne(ctx, bson.M{"phone": phone}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

// NewUser adds a new user to the system.
func (orm *ORM) NewUser(phone string) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	users := orm.DB.Collection("users")
	_, err := users.InsertOne(ctx, bson.M{"phone": phone,
		"action": "welcome", "requirements": []string{"name", "location", "type"}})
	return err
}
