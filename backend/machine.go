package main

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Machine struct {
	DB *mongo.Client
}

func NewMachine(db *mongo.Client) *Machine {
	return &Machine{DB: db}
}

func (*Machine) Generate(user string, message string) (string, error) {
	//@TODO implement state machine + database logic
	return message, nil
}
