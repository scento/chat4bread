package main

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
	} else if user.Action == "welcome" {
		return m.Welcome(user, message)
	}

	return "Sorry, but I don't know what to say.", nil
}

// Welcome handles the initialization workflow of a new user.
func (m *Machine) Welcome(user *User, message string) (string, error) {
	return "Be welcomed!", nil
}
