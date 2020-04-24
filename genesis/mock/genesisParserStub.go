package mock

import (
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// GenesisParserStub -
type GenesisParserStub struct {
	InitialAccountsSplitOnAddressesShardsCalled           func(shardCoordinator sharding.Coordinator) (map[uint32][]*genesis.InitialAccount, error)
	InitialAccountsSplitOnDelegationAddressesShardsCalled func(shardCoordinator sharding.Coordinator) (map[uint32][]*genesis.InitialAccount, error)
	InitialAccountsCalled                                 func() []*genesis.InitialAccount
}

// InitialAccountsSplitOnAddressesShards -
func (gps *GenesisParserStub) InitialAccountsSplitOnAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]*genesis.InitialAccount, error) {
	if gps.InitialAccountsSplitOnAddressesShardsCalled != nil {
		return gps.InitialAccountsSplitOnAddressesShardsCalled(shardCoordinator)
	}

	return make(map[uint32][]*genesis.InitialAccount), nil
}

// InitialAccountsSplitOnDelegationAddressesShards -
func (gps *GenesisParserStub) InitialAccountsSplitOnDelegationAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]*genesis.InitialAccount, error) {
	if gps.InitialAccountsSplitOnDelegationAddressesShardsCalled != nil {
		return gps.InitialAccountsSplitOnDelegationAddressesShardsCalled(shardCoordinator)
	}

	return make(map[uint32][]*genesis.InitialAccount), nil
}

// InitialAccounts -
func (gps *GenesisParserStub) InitialAccounts() []*genesis.InitialAccount {
	if gps.InitialAccountsCalled != nil {
		return gps.InitialAccountsCalled()
	}

	return make([]*genesis.InitialAccount, 0)
}

// IsInterfaceNil -
func (gps *GenesisParserStub) IsInterfaceNil() bool {
	return gps == nil
}
