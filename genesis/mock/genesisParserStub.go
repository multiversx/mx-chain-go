package mock

import (
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// GenesisParserStub -
type GenesisParserStub struct {
	InitialAccountsSplitOnAddressesShardsCalled           func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error)
	InitialAccountsSplitOnDelegationAddressesShardsCalled func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error)
	InitialAccountsCalled                                 func() []genesis.InitialAccountHandler
}

// InitialAccountsSplitOnAddressesShards -
func (gps *GenesisParserStub) InitialAccountsSplitOnAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error) {
	if gps.InitialAccountsSplitOnAddressesShardsCalled != nil {
		return gps.InitialAccountsSplitOnAddressesShardsCalled(shardCoordinator)
	}

	return make(map[uint32][]genesis.InitialAccountHandler), nil
}

// InitialAccountsSplitOnDelegationAddressesShards -
func (gps *GenesisParserStub) InitialAccountsSplitOnDelegationAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error) {
	if gps.InitialAccountsSplitOnDelegationAddressesShardsCalled != nil {
		return gps.InitialAccountsSplitOnDelegationAddressesShardsCalled(shardCoordinator)
	}

	return make(map[uint32][]genesis.InitialAccountHandler), nil
}

// InitialAccounts -
func (gps *GenesisParserStub) InitialAccounts() []genesis.InitialAccountHandler {
	if gps.InitialAccountsCalled != nil {
		return gps.InitialAccountsCalled()
	}

	return make([]genesis.InitialAccountHandler, 0)
}

// IsInterfaceNil -
func (gps *GenesisParserStub) IsInterfaceNil() bool {
	return gps == nil
}
