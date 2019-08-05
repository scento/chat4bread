package main

import (
	"fmt"
	"log"
)

// Machine is the state machine for messaging actions.
type Machine struct {
	ORM *ORM
	CAI *CAI
}

// NewMachine initializes a new Machine.
func NewMachine(orm *ORM, cai *CAI) *Machine {
	return &Machine{ORM: orm, CAI: cai}
}

// Generate creates a response for a new incoming message.
func (m *Machine) Generate(phone string, message string) (string, error) {
	user, err := m.ORM.UserByPhone(phone)
	if err != nil {
		return "", err
	}

	if user == nil {
		err = m.ORM.NewUser(phone)
		return "Hi, here is your Chat4Bread market platform. Who are you?", err
	} else if user.Action == "onboarding" {
		return m.Onboarding(user, message)
	}

	log.Printf("Error state: action %s, requirements %v", user.Action, user.Reqs)
	return "Sorry, but I don't know what to say.", nil
}

// Onboarding handles the initialization workflow of a new user.
func (m *Machine) Onboarding(user *User, message string) (string, error) {
	intent, err := m.CAI.Intent(message)
	if err != nil {
		return "", err
	}

	if len(user.Reqs) > 0 {
		switch user.Reqs[0] {
		case "name":
			if intent.Slug != "get_name" || intent.FullName == "" {
				return "We didn't understand you. What is your name?", nil
			}
			err = m.ORM.SetUserName(user, intent.FullName)
			if err != nil {
				return "", err
			}
			err = m.ORM.PopRequirement(user)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Hi %s, where do you live?", intent.FullName), nil
		case "location":
			if intent.Slug != "get_location" || intent.Lat == 0.0 || intent.Lng == 0.0 {
				return "We didn't understand you. What is your address?", nil
			}
			err = m.ORM.SetUserLocation(user, intent.Lat, intent.Lng)
			if err != nil {
				return "", err
			}
			err = m.ORM.PopRequirement(user)
			if err != nil {
				return "", err
			}
			return "Great to have you here. Are you a farmer or a consumer?", nil
		case "type":
			if intent.Slug != "get_type_buyer" && intent.Slug != "get_type_farmer" {
				return "We didn't understand you. Are you a farmer or a customer?", nil
			}

			if intent.Slug == "get_type_buyer" {
				err = m.ORM.SetUserKind(user, "consumer")
			} else {
				err = m.ORM.SetUserKind(user, "farmer")
			}
			if err != nil {
				return "", err
			}

			err = m.ORM.PopRequirement(user)
			if err != nil {
				return "", err
			}

			if intent.Slug == "get_type_buyer" {
				return "Welcome to the market. You can now look for organic food or find a" +
					" local farmer.", nil
			} else {
				return "Welcome to the market. You can now sell and buy products or learn about" +
					" the current market prices for your goods.", nil
			}
		default:
			return fmt.Sprintf("Unknown requirement %s", user.Reqs[0]), nil
		}
	}

	err = m.ORM.ResetUserState(user)
	if err != nil {
		return "", err
	}
	return "Welcome to the market. Have fun!", nil
}
