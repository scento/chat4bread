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
	} else {
		intent, err := m.CAI.Intent(message)
		if err != nil {
			return "", err
		}

		switch intent.Slug {
		case "get_type_farmer":
			fallthrough // Commonly misconception and not possible in this state.
		case "pos_list":
			return m.FarmersNearby(user, intent)
		case "sell":
			return m.SellProduct(user, intent)
		default:
			return fmt.Sprintf("Got intent %s", intent.Slug), nil
		}
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

			err = m.ORM.ResetUserState(user)
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

// FarmersNearby returns a list of farmers near the users location.
func (m *Machine) FarmersNearby(user *User, intent *Intent) (string, error) {
	users, err := m.ORM.FindFarmersNear(user.Location.Coords[1],
		user.Location.Coords[0], 2000)
	if err != nil {
		return "", err
	}

	msg := "We found the following farmers nearby:\n"
	index := 1
	for index, farmer := range users {
		if *farmer.Kind == "farmer" && user.ID != farmer.ID {
			msg += fmt.Sprintf("%d. **%s** (%.2f m)\n", index, *farmer.Name, farmer.Location.Distance)
			index++
		}
	}

	if index == 1 {
		msg = "We could not find any farmers nearby. In the future we might notify you if something changed, but for now, please check from time to time if something changes."
	}

	return msg, nil
}

// SellProduct returns a workflow to sell a product as a farmer.
func (m *Machine) SellProduct(user *User, intent *Intent) (string, error) {
	if user.Kind == nil || *user.Kind != "farmer" {
		return "You registered as a consumer. It is currently not possible to switch the account type without resetting it.", nil
	}

	if intent.Product == "" || intent.Dollars == 0.0 || (intent.Mass == 0.0 && intent.Number == 0) {
		return "It seems like you want to sell something, but we either didn't get the product, price or the amount you want to sell. Please retry with all required information.", nil
	}

	product, err := m.ORM.FindOrCreateProduct(intent.Product)
	if err != nil {
		return "", err
	}

	var msg string
	if intent.Mass > 0.0 {
		err = m.ORM.CreateMassOffer(user.ID, product.ID, intent.Dollars, intent.Mass)
		msg = fmt.Sprintf("We created a new offer. You are selling %dg of %s for %.2f$.", uint(intent.Mass), intent.Product, intent.Dollars)
	} else if intent.Number > 0 {
		err = m.ORM.CreateUnitOffer(user.ID, product.ID, intent.Dollars, uint64(intent.Number))
		msg = fmt.Sprintf("We created a new offer. You are selling %d %s for %.2f$.", uint(intent.Number), intent.Product, intent.Dollars)
	} else {
		msg = "Please retry while specifying a mass or unit number greater than zero."
	}

	return msg, err
}
