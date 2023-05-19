package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/sharding"
)

// AccountsParserStub -
type AccountsParserStub struct {
	InitialAccountsSplitOnAddressesShardsCalled           func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error)
	InitialAccountsSplitOnDelegationAddressesShardsCalled func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error)
	InitialAccountsCalled                                 func() []genesis.InitialAccountHandler
	GetTotalStakedForDelegationAddressCalled              func(delegationAddress string) *big.Int
	GetInitialAccountsForDelegatedCalled                  func(addressBytes []byte) []genesis.InitialAccountHandler
	GenerateInitialTransactionsCalled                     func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*block.MiniBlock, map[uint32]*outport.TransactionPool, error)
	GenesisMintingAddressCalled                           func() string
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

// GenesisMintingAddress -
func (aps *AccountsParserStub) GenesisMintingAddress() string {
	if aps.GenesisMintingAddressCalled != nil {
		return aps.GenesisMintingAddressCalled()
	}

	return ""
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

// GenerateInitialTransactions -
func (aps *AccountsParserStub) GenerateInitialTransactions(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*block.MiniBlock, map[uint32]*outport.TransactionPool, error) {
	if aps.GenerateInitialTransactionsCalled != nil {
		return aps.GenerateInitialTransactionsCalled(shardCoordinator, initialIndexingData)
	}

	return make([]*block.MiniBlock, 0), make(map[uint32]*outport.TransactionPool), nil
}

// IsInterfaceNil -
func (aps *AccountsParserStub) IsInterfaceNil() bool {
	return aps == nil
}
