package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// AccountsParserStub -
type AccountsParserStub struct {
	InitialAccountsSplitOnAddressesShardsCalled           func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error)
	InitialAccountsSplitOnDelegationAddressesShardsCalled func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error)
	InitialAccountsCalled                                 func() []genesis.InitialAccountHandler
	GetTotalStakedForDelegationAddressCalled              func(delegationAddress string) *big.Int
	GetInitialAccountsForDelegatedCalled                  func(addressBytes []byte) []genesis.InitialAccountHandler
}

// GetTotalStakedForDelegationAddress -
func (aps *AccountsParserStub) GetTotalStakedForDelegationAddress(delegationAddress string) *big.Int {
	if aps.GetTotalStakedForDelegationAddressCalled != nil {
		return aps.GetTotalStakedForDelegationAddressCalled(delegationAddress)
	}

	return big.NewInt(0)
}

// GetInitialAccountsForDelegated -
func (aps *AccountsParserStub) GetInitialAccountsForDelegated(addressBytes []byte) []genesis.InitialAccountHandler {
	if aps.GetInitialAccountsForDelegatedCalled != nil {
		return aps.GetInitialAccountsForDelegatedCalled(addressBytes)
	}

	return make([]genesis.InitialAccountHandler, 0)
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
