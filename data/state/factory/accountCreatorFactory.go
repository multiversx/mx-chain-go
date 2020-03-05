package factory

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// NewAccountFactoryCreator returns an account factory depending on shard coordinator self id
func NewAccountFactoryCreator(accountType state.Type) (state.AccountFactory, error) {
	switch accountType {
	case state.UserAccount:
		return NewAccountCreator(), nil
	case state.ShardStatistics:
		return NewMetaAccountCreator(), nil
	case state.ValidatorAccount:
		return NewPeerAccountCreator(), nil
	default:
		return nil, state.ErrUnknownAccountType
	}
}
