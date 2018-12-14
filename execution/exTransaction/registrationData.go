package exTransaction

import (
	"math/big"
)

// RegisterAddress is the address used for a node to register / unregister in the Elrond network
const RegisterAddress = "0000000000000000000000000000000000000000000000000000000000000000"

// ActionRequested defines the type used to refer to the action requested in the registration transaction
type ActionRequested int

const (
	// ArRegister defines a requested action of registration
	ArRegister ActionRequested = iota
	// ArUnregister defines a requested action of unregistration
	ArUnregister
)

// RegistrationData holds the data which is sent with a registration transaction
type RegistrationData struct {
	NodeId string
	Stake  big.Int
	Action ActionRequested
}
