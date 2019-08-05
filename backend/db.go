package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// ORM is the object relationship mapper for the MongoDB database.
type ORM struct {
	DB *mongo.Database
}

// GeoJSON defines geometric structures such that we can use MongoDB GIS operations.
type GeoJSON struct {
	Type   string `bson:"type"`
	Coords []float64  `bson:"coordinates"`
}

// MakeGeoJSONPnt creates a new GeoJSON point.
func MakeGeoJSONPnt(lat float64, lon float64) GeoJSON {
	return GeoJSON{Type: "Point", Coords: []float64{lon, lat}}
}

// User object bundling all relevant information about an user.
type User struct {
	ID       primitive.ObjectID `bson:"_id"`
	Phone    string             `bson:"phone"`
	Name     *string            `bson:"name"`
	Location *GeoJSON           `bson:"location"`
	Kind 	 *string 			`bson:"kind"`

	Action string   `bson:"action"`
	Reqs   []string `bson:"requirements"`
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
		"action": "onboarding", "requirements": []string{"name", "location", "type"}})
	return err
}

// ResetUserState resets the user state.
func (orm *ORM) ResetUserState(user *User) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	users := orm.DB.Collection("users")
	_, err := users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"action": "",
		"requirements": []string{}}})
	return err
}

// SetUserName sets the name of the user.
func (orm *ORM) SetUserName(user *User, name string) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	users := orm.DB.Collection("users")
	_, err := users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"name": name}})
	return err
}

// SetUserLocation sets the location of the user.
func (orm *ORM) SetUserLocation(user *User, lat float64, lng float64) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	users := orm.DB.Collection("users")
	_, err := users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set":
		bson.M{"location": MakeGeoJSONPnt(lat, lng)}})
	return err
}

// SetUserKind sets the type of the user.
func (orm *ORM) SetUserKind(user *User, kind string) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	users := orm.DB.Collection("users")
	_, err := users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"kind": kind}})
	return err
}

// PopRequirement removes the top requirement of the current action.
func (orm *ORM) PopRequirement(user *User) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	users := orm.DB.Collection("users")
	_, err := users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$pop": bson.M{"requirements": -1}})
	return err
}