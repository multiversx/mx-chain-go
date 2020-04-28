package genesis

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

// AccountsParser contains the parsed genesis json file and has some functionality regarding processed data
type AccountsParser interface {
	InitialAccountsSplitOnAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]InitialAccountHandler, error)
	InitialAccountsSplitOnDelegationAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]InitialAccountHandler, error)
	InitialAccounts() []InitialAccountHandler
	IsInterfaceNil() bool
}

// InitialNodesHandler contains the initial nodes setup
type InitialNodesHandler interface {
	InitialNodesInfo() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler)
	MinNumberOfNodes() uint32
	IsInterfaceNil() bool
}

// InitialAccountHandler represents the interface that describes the data held by an initial account
type InitialAccountHandler interface {
	Clone() InitialAccountHandler
	GetAddress() string
	AddressBytes() []byte
	GetStakingValue() *big.Int
	GetBalanceValue() *big.Int
	GetSupply() *big.Int
	GetDelegationHandler() DelegationDataHandler
	IsInterfaceNil() bool
}

// DelegationDataHandler represents the interface that describes the data held by a delegation address
type DelegationDataHandler interface {
	GetAddress() string
	AddressBytes() []byte
	GetValue() *big.Int
	IsInterfaceNil() bool
}

// InitialSmartContractHandler represents the interface that describes the smart contract
type InitialSmartContractHandler interface {
	GetOwner() string
	OwnerBytes() []byte
	GetFilename() string
	GetVmType() string
	IsInterfaceNil() bool
}

// InitialSmartContractParser contains the parsed genesis initial smart contracts
//json file and has some functionality regarding processed data
type InitialSmartContractParser interface {
	InitialSmartContractsSplitOnOwnersShards(shardCoordinator sharding.Coordinator) (map[uint32][]InitialSmartContractHandler, error)
	InitialSmartContracts() []InitialSmartContractHandler
	IsInterfaceNil() bool
}
