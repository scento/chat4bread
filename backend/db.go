package main

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

// ORM is the object relationship mapper for the MongoDB database.
type ORM struct {
	DB *mongo.Database
}

// GeoJSON defines geometric structures such that we can use MongoDB GIS operations.
type GeoJSON struct {
	Type     string    `bson:"type"`
	Coords   []float64 `bson:"coordinates"`
	Distance float64   `bson:"distance"`
}

// MakeGeoJSONPnt creates a new GeoJSON point.
func MakeGeoJSONPnt(lat float64, lon float64) GeoJSON {
	return GeoJSON{Type: "Point", Coords: []float64{lon, lat}}
}

// User object bundles all relevant information about an user.
type User struct {
	ID       primitive.ObjectID `bson:"_id"`
	Phone    int64              `bson:"phone"`
	Name     *string            `bson:"name"`
	Location *GeoJSON           `bson:"location"`
	Kind     *string            `bson:"kind"`
	Action   string             `bson:"action"`
	Reqs     []string           `bson:"requirements"`
}

// Product object bundles all relevant information about a product.
type Product struct {
	ID   primitive.ObjectID `bson:"_id"`
	Name string             `bson:"name"`
}

// Offer object bundles all relevant information about an offer.
type Offer struct {
	ID              primitive.ObjectID `bson:"_id"`
	Product         primitive.ObjectID `bson:"product"`
	Seller          primitive.ObjectID `bson:"seller"`
	Price           float64            `bson:"price"`
	NormalizedPrice float64            `bson:"normalized_price"`
	Mass            float64            `bson:"mass"`
	Units           uint64             `bson:"units"`
}

// NewORM initializes the ORM.
func NewORM(client *mongo.Client, database string) *ORM {
	return &ORM{DB: client.Database(database)}
}

// CreateIndicies initializes the ORM indicies.
func (orm *ORM) CreateIndicies() error {
	index := mongo.IndexModel{Keys: bsonx.Doc{{"location", bsonx.String("2dsphere")}},
		Options: options.Index().SetName("user-loc-2dsphere")}
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	users := orm.DB.Collection("users")
	_, err := users.Indexes().CreateOne(ctx, index)
	return err
}

// UserByPhone looks for a user by its phone number/username.
func (orm *ORM) UserByPhone(phone int64) (*User, error) {
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
func (orm *ORM) NewUser(phone int64) error {
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
	_, err := users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"location": MakeGeoJSONPnt(lat, lng)}})
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

// FindFarmersNear finds farmers near a geo point within a specific range in meters.
func (orm *ORM) FindFarmersNear(lat float64, lng float64, dist float64) ([]User, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	collection := orm.DB.Collection("users")
	cur, err := collection.Aggregate(ctx, []bson.M{bson.M{"$geoNear": bson.M{"near": MakeGeoJSONPnt(lat, lng), "minDistance": 0, "maxDistance": dist, "distanceField": "location.distance", "spherical": true}}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var users []User
	for cur.Next(ctx) {
		var user User
		err := cur.Decode(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// FindOrCreateProduct finds a product or creates a new one.
func (orm *ORM) FindOrCreateProduct(name string) (*Product, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	products := orm.DB.Collection("products")

	// In a real implementation, please use something atomic.
	var product Product
	err := products.FindOne(ctx, bson.M{"name": name}).Decode(&product)
	if err == mongo.ErrNoDocuments {
		res, err := products.InsertOne(ctx, bson.M{"name": name})
		if err != nil {
			return nil, err
		}
		if insertedID, ok := res.InsertedID.(primitive.ObjectID); ok {
			return &Product{ID: insertedID, Name: name}, nil
		}
		return nil, errors.New("Cannot convert ObjectID.")
	} else if err != nil {
		return nil, err
	}

	return &product, err

}

// CreateMassOffer creates a new offer based on a specific mass.
func (orm *ORM) CreateMassOffer(user primitive.ObjectID, product primitive.ObjectID,
	price float64, mass float64) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	offers := orm.DB.Collection("offers")
	_, err := offers.InsertOne(ctx, bson.M{"product": product, "seller": user, "price": price, "mass": mass, "normalized_price": price / mass})
	return err
}

// CreateUnitOffer creates a new offer based on a number of units to sell.
func (orm *ORM) CreateUnitOffer(user primitive.ObjectID, product primitive.ObjectID,
	price float64, units uint64) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	offers := orm.DB.Collection("offers")
	_, err := offers.InsertOne(ctx, bson.M{"product": product, "seller": user, "price": price, "units": units, "normalized_price": price / float64(units)})
	return err
}

// FindMassOffer finds a offer fulfilling pricing criterea.
func (orm *ORM) FindMassOffer(product primitive.ObjectID, price float64, mass float64) (*Offer,
	*User, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	offers := orm.DB.Collection("offers")
	var offer Offer
	err := offers.FindOne(ctx, bson.M{"product": product, "mass": bson.M{"$gt": mass}, "normalized_price": bson.M{"$lt": (price / mass)}}).Decode(&offer)
	if err == mongo.ErrNoDocuments {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}

	users := orm.DB.Collection("users")
	var user User
	err = users.FindOne(ctx, bson.M{"_id": offer.Seller}).Decode(&user)
	if err != nil {
		return nil, nil, err
	}

	return &offer, &user, nil
}

// FindUnitOffer finds a offer fulfilling pricing criterea.
func (orm *ORM) FindUnitOffer(product primitive.ObjectID, price float64, units uint64) (*Offer,
	*User, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	offers := orm.DB.Collection("offers")
	var offer Offer
	err := offers.FindOne(ctx, bson.M{"product": product, "units": bson.M{"$gt": units}, "normalized_price": bson.M{"$lt": price / float64(units)}}).Decode(&offer)
	if err == mongo.ErrNoDocuments {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}

	users := orm.DB.Collection("users")
	var user User
	err = users.FindOne(ctx, bson.M{"_id": offer.Seller}).Decode(&user)
	if err != nil {
		return nil, nil, err
	}

	return &offer, &user, nil
}

// ReduceMassOffer reduces the publicly available offer by a specific mass.
func (orm *ORM) ReduceMassOffer(offer primitive.ObjectID, mass float64) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	users := orm.DB.Collection("offers")
	_, err := users.UpdateOne(ctx, bson.M{"_id": offer}, bson.M{"$inc": bson.M{"mass": (-1 * mass)}})
	return err
}

// ReduceUnitOffer reduces the publicly available offer by a specific amount.
func (orm *ORM) ReduceUnitOffer(offer primitive.ObjectID, units uint64) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	users := orm.DB.Collection("offers")
	_, err := users.UpdateOne(ctx, bson.M{"_id": offer},
		bson.M{"$inc": bson.M{"units": (-1 * int64(units))}})
	return err
}

// GetAveragePrice returns the average price for a product.
func (orm *ORM) GetAveragePrice(product primitive.ObjectID) (*float64, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	collection := orm.DB.Collection("offers")
	cur, err := collection.Aggregate(ctx, []bson.M{bson.M{"$group": bson.M{"_id": "$product", "avgPrice": bson.M{"$avg": "$normalized_price"}}}, bson.M{"$match": bson.M{"_id": product}}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var agg bson.M
		err := cur.Decode(&agg)
		if err != nil {
			return nil, err
		}
		if val, ok := agg["avgPrice"].(float64); ok {
			return &val, nil
		}
	}

	return nil, nil
}
