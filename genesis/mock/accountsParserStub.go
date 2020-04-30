package mock

import (
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// AccountsParserStub -
type AccountsParserStub struct {
	InitialAccountsSplitOnAddressesShardsCalled           func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error)
	InitialAccountsSplitOnDelegationAddressesShardsCalled func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error)
	InitialAccountsCalled                                 func() []genesis.InitialAccountHandler
}

// InitialAccountsSplitOnAddressesShards -
func (aps *AccountsParserStub) InitialAccountsSplitOnAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error) {
	if aps.InitialAccountsSplitOnAddressesShardsCalled != nil {
		return aps.InitialAccountsSplitOnAddressesShardsCalled(shardCoordinator)
	}

	return make(map[uint32][]genesis.InitialAccountHandler), nil
}

// InitialAccountsSplitOnDelegationAddressesShards -
func (aps *AccountsParserStub) InitialAccountsSplitOnDelegationAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error) {
	if aps.InitialAccountsSplitOnDelegationAddressesShardsCalled != nil {
		return aps.InitialAccountsSplitOnDelegationAddressesShardsCalled(shardCoordinator)
	}

	return make(map[uint32][]genesis.InitialAccountHandler), nil
}

// InitialAccounts -
func (aps *AccountsParserStub) InitialAccounts() []genesis.InitialAccountHandler {
	if aps.InitialAccountsCalled != nil {
		return aps.InitialAccountsCalled()
	}

	return make([]genesis.InitialAccountHandler, 0)
}

// IsInterfaceNil -
func (aps *AccountsParserStub) IsInterfaceNil() bool {
	return aps == nil
}
