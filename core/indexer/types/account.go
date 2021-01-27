package types

import "github.com/ElrondNetwork/elrond-go/data/state"

// AccountInfo holds (serializable) data about an account
type AccountInfo struct {
	Address         string  `json:"address,omitempty"`
	Nonce           uint64  `json:"nonce,omitempty"`
	Balance         string  `json:"balance"`
	BalanceNum      float64 `json:"balanceNum"`
	TokenIdentifier string  `json:"token,omitempty"`
	Properties      string  `json:"properties,omitempty"`
	IsSender        bool    `json:"-"`
}

// AccountBalanceHistory represents an entry in the user accounts balances history
type AccountBalanceHistory struct {
	Address         string `json:"address"`
	Timestamp       int64  `json:"timestamp"`
	Balance         string `json:"balance"`
	TokenIdentifier string `json:"token,omitempty"`
	IsSender        bool   `json:"isSender,omitempty"`
}

// AccountEGLD is a structure that is needed for EGLD accounts
type AccountEGLD struct {
	Account  state.UserAccountHandler
	IsSender bool
}

// AccountESDT is a structure that is needed for ESDT accounts
type AccountESDT struct {
	Account         state.UserAccountHandler
	TokenIdentifier string
	IsSender        bool
}

// AlteredAccount is a structure that holds information about an altered account
type AlteredAccount struct {
	IsSender        bool
	IsESDTOperation bool
	TokenIdentifier string
}
