package genesis

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// DelegationType defines the constant used when checking if a smart contract is of delegation type
const DelegationType = "delegation"

// DNSType defines the constant used when checking if a smart contract is of dns type
const DNSType = "dns"

// InitialDNSAddress defines the initial address from where the DNS contracts are deployed
var InitialDNSAddress = bytes.Repeat([]byte{1}, 32)

// DelegationResult represents the DTO that contains the delegation results metrics
type DelegationResult struct {
	NumTotalStaked    int
	NumTotalDelegated int
}

// AccountsParser contains the parsed genesis json file and has some functionality regarding processed data
type AccountsParser interface {
	InitialAccountsSplitOnAddressesShards(shardCoordinator sharding.Coordinator) (map[uint32][]InitialAccountHandler, error)
	InitialAccounts() []InitialAccountHandler
	GetTotalStakedForDelegationAddress(delegationAddress string) *big.Int
	GetInitialAccountsForDelegated(addressBytes []byte) []InitialAccountHandler
	IsInterfaceNil() bool
}

// InitialNodesHandler contains the initial nodes setup
type InitialNodesHandler interface {
	InitialNodesInfo() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler)
	MinNumberOfNodes() uint32
	MinShardHysteresisNodes() uint32
	MinMetaHysteresisNodes() uint32
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

// InitialSmartContractHandler represents the interface that describes the the initial smart contract
type InitialSmartContractHandler interface {
	GetOwner() string
	OwnerBytes() []byte
	GetFilename() string
	GetVmType() string
	GetInitParameters() string
	GetType() string
	VmTypeBytes() []byte
	AddAddressBytes(addressBytes []byte)
	AddressesBytes() [][]byte
	AddAddress(address string)
	Addresses() []string
	GetVersion() string
	IsInterfaceNil() bool
}

// InitialSmartContractParser contains the parsed genesis initial smart contracts
//json file and has some functionality regarding processed data
type InitialSmartContractParser interface {
	InitialSmartContractsSplitOnOwnersShards(shardCoordinator sharding.Coordinator) (map[uint32][]InitialSmartContractHandler, error)
	GetDeployedSCAddresses(scType string) (map[string]struct{}, error)
	InitialSmartContracts() []InitialSmartContractHandler
	IsInterfaceNil() bool
}

// TxExecutionProcessor represents a transaction builder and executor containing also related helper functions
type TxExecutionProcessor interface {
	ExecuteTransaction(nonce uint64, sndAddr []byte, rcvAddress []byte, value *big.Int, data []byte) error
	GetAccount(address []byte) (state.UserAccountHandler, bool)
	GetNonce(senderBytes []byte) (uint64, error)
	AddBalance(senderBytes []byte, value *big.Int) error
	AddNonce(senderBytes []byte, nonce uint64) error
	IsInterfaceNil() bool
}

// NodesListSplitter is able to split de initial nodes based on some criteria
type NodesListSplitter interface {
	GetAllNodes() []sharding.GenesisNodeInfoHandler
	GetDelegatedNodes(delegationScAddress []byte) []sharding.GenesisNodeInfoHandler
	IsInterfaceNil() bool
}

// DeployProcessor is able to deploy a smart contract
type DeployProcessor interface {
	Deploy(sc InitialSmartContractHandler) ([][]byte, error)
	IsInterfaceNil() bool
}
