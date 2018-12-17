package state

import (
	"math/big"
)

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
	OriginatorPubKey []byte
	NodePubKey       []byte
	Stake            big.Int
	Action           ActionRequested
	RoundIndex       int32
}

// Account is the struct used in serialization/deserialization
type Account struct {
	Nonce            uint64
	Balance          big.Int
	CodeHash         []byte
	RootHash         []byte
	RegistrationData []RegistrationData
}

// NewAccount creates a new account object
func NewAccount() *Account {
	return &Account{}
}

//TODO add Cap'N'Proto converter funcs
