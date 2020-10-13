package facade

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/api/block"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

//NodeHandler contains all functions that a node should contain.
type NodeHandler interface {
	// GetBalance returns the balance for a specific address
	GetBalance(address string) (*big.Int, error)

	// GetUsername returns the username for a specific address
	GetUsername(address string) (string, error)

	// GetValueForKey returns the value of a key from a given account
	GetValueForKey(address string, key string) (string, error)

	//CreateTransaction will return a transaction from all needed fields
	CreateTransaction(nonce uint64, value string, receiverHex string, senderHex string, gasPrice uint64,
		gasLimit uint64, data []byte, signatureHex string, chainID string, version uint32) (*transaction.Transaction, []byte, error)

	//ValidateTransaction will validate a transaction
	ValidateTransaction(tx *transaction.Transaction) error

	//SendBulkTransactions will send a bulk of transactions on the 'send transactions pipe' channel
	SendBulkTransactions(txs []*transaction.Transaction) (uint64, error)

	//GetTransaction will return a transaction based on the hash
	GetTransaction(hash string) (*transaction.ApiTransactionResult, error)

	// GetAccount returns an accountResponse containing information
	//  about the account corelated with provided address
	GetAccount(address string) (state.UserAccountHandler, error)

	// GetHeartbeats returns the heartbeat status for each public key defined in genesis.json
	GetHeartbeats() []data.PubKeyHeartbeat

	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool

	// ValidatorStatisticsApi return the statistics for all the validators
	ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error)
	DirectTrigger(epoch uint32) error
	IsSelfTrigger() bool

	EncodeAddressPubkey(pk []byte) (string, error)
	DecodeAddressPubkey(pk string) ([]byte, error)

	GetQueryHandler(name string) (debug.QueryHandler, error)
	GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error)

	GetBlockByHash(hash string, withTxs bool) (*block.APIBlock, error)
	GetBlockByNonce(nonce uint64, withTxs bool) (*block.APIBlock, error)
}

// TransactionSimulatorProcessor defines the actions which a transaction simulator processor has to implement
type TransactionSimulatorProcessor interface {
	ProcessTx(tx *transaction.Transaction) (*transaction.SimulationResults, error)
	IsInterfaceNil() bool
}

// ApiResolver defines a structure capable of resolving REST API requests
type ApiResolver interface {
	ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error)
	StatusMetrics() external.StatusMetricsHandler
	Close() error
	IsInterfaceNil() bool
}

// HardforkTrigger defines the structure used to trigger hardforks
type HardforkTrigger interface {
	Trigger(epoch uint32) error
	IsSelfTrigger() bool
	IsInterfaceNil() bool
}
