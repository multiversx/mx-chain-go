package factory

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// Type defines account types to save in accounts trie
type Type uint8

const (
	// UserAccount identifies an account holding balance, storage updates, code
	UserAccount Type = 0
	// ShardStatistics identifies a shard, keeps the statistics
	ShardStatistics Type = 1
	// ValidatorAccount identifies an account holding stake, crypto public keys, assigned shard, rating
	ValidatorAccount Type = 2
)

// NewAccountFactoryCreator returns an account factory depending on shard coordinator self id
func NewAccountFactoryCreator(accountType Type) (state.AccountFactory, error) {
	switch accountType {
	case UserAccount:
		return NewAccountCreator(), nil
	case ShardStatistics:
		return NewMetaAccountCreator(), nil
	case ValidatorAccount:
		return NewPeerAccountCreator(), nil
	default:
		return nil, state.ErrUnknownAccountType
	}
}
