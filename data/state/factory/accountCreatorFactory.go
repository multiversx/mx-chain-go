package factory

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type Type uint8

const (
	UserAccount      Type = 0
	ShardStatistics  Type = 1
	ValidatorAccount Type = 2
	InvalidType      Type = 3
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
	case InvalidType:
		return nil, state.ErrUnknownAccountType
	default:
		return nil, state.ErrUnknownAccountType
	}
}
