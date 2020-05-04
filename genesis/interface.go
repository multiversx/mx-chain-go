package genesis

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

// DelegationType defines the constant used when checking if a smart contract is of delegation type
const DelegationType = "delegation"

// AccountsParser contains the parsed genesis json file and has some functionality regarding processed data
type AccountsParser interface {
	InitialAccountsSplitOnAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]InitialAccountHandler, error)
	InitialAccountsSplitOnDelegationAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]InitialAccountHandler, error)
	InitialAccounts() []InitialAccountHandler
	GetTotalStakedForDelegationAddress(delegationAddress string) *big.Int
	GetInitialAccountsForDelegated(addressBytes []byte) []InitialAccountHandler
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
	IncrementNonceOffset()
	NonceOffset() uint64
	IsInterfaceNil() bool
}

// DelegationDataHandler represents the interface that describes the data held by a delegation address
type DelegationDataHandler interface {
	GetAddress() string
	AddressBytes() []byte
	GetValue() *big.Int
	IsInterfaceNil() bool
}

// InitialSmartContractHandler represents the interface that describes the the initial smart contract
type InitialSmartContractHandler interface {
	GetOwner() string
	OwnerBytes() []byte
	GetFilename() string
	GetVmType() string
	GetInitParameters() string
	GetType() string
	VmTypeBytes() []byte
	SetAddressBytes(addressBytes []byte)
	AddressBytes() []byte
	IsInterfaceNil() bool
}

// InitialSmartContractParser contains the parsed genesis initial smart contracts
//json file and has some functionality regarding processed data
type InitialSmartContractParser interface {
	InitialSmartContractsSplitOnOwnersShards(shardCoordinator sharding.Coordinator) (map[uint32][]InitialSmartContractHandler, error)
	InitialSmartContracts() []InitialSmartContractHandler
	IsInterfaceNil() bool
}

type TxExecutionProcessor interface {
	ExecuteTransaction(nonce uint64, sndAddr []byte, rcvAddress []byte, value *big.Int, data []byte) error
	GetNonce(senderBytes []byte) (uint64, error)
	AddBalance(senderBytes []byte, value *big.Int) error
	AddNonce(senderBytes []byte, nonce uint64) error
	IsInterfaceNil() bool
}

type NodesHandler interface {
	GetAllStakedNodes() []sharding.GenesisNodeInfoHandler
	GetDelegatedNodes(delegationScAddress []byte) []sharding.GenesisNodeInfoHandler
	IsInterfaceNil() bool
}
