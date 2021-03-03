package integrationTests

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/api"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	dataApi "github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/vm"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// TestBootstrapper extends the Bootstrapper interface with some functions intended to be used only in tests
// as it simplifies the reproduction of edge cases
type TestBootstrapper interface {
	process.Bootstrapper
	RollBack(revertUsingForkNonce bool) error
	SetProbableHighestNonce(nonce uint64)
}

// TestEpochStartTrigger extends the epochStart trigger interface with some functions intended to by used only
// in tests as it simplifies the reproduction of test scenarios
type TestEpochStartTrigger interface {
	epochStart.TriggerHandler
	GetRoundsPerEpoch() uint64
	SetTrigger(triggerHandler epochStart.TriggerHandler)
	SetRoundsPerEpoch(roundsPerEpoch uint64)
	SetMinRoundsBetweenEpochs(minRoundsPerEpoch uint64)
	SetEpoch(epoch uint32)
}

// NodesCoordinatorFactory is used for creating a nodesCoordinator in the integration tests
type NodesCoordinatorFactory interface {
	CreateNodesCoordinator(arg ArgIndexHashedNodesCoordinatorFactory) sharding.NodesCoordinator
}

// NetworkShardingUpdater defines the updating methods used by the network sharding component
type NetworkShardingUpdater interface {
	GetPeerInfo(pid core.PeerID) core.P2PPeerInfo
	UpdatePeerIdPublicKey(pid core.PeerID, pk []byte)
	UpdatePublicKeyShardId(pk []byte, shardId uint32)
	UpdatePeerIdShardId(pid core.PeerID, shardId uint32)
	IsInterfaceNil() bool
}

// Facade is the node facade used to decouple the node implementation with the web server. Used in integration tests
type Facade interface {
	GetBalance(address string) (*big.Int, error)
	GetUsername(address string) (string, error)
	GetValueForKey(address string, key string) (string, error)
	GetAccount(address string) (state.UserAccountHandler, error)
	GetCode(account state.UserAccountHandler) []byte
	GetESDTBalance(address string, key string) (string, string, error)
	GetAllESDTTokens(address string) ([]string, error)
	GetBlockByHash(hash string, withTxs bool) (*dataApi.Block, error)
	GetBlockByNonce(nonce uint64, withTxs bool) (*dataApi.Block, error)
	Trigger(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTrigger() bool
	GetTotalStakedValue() (*big.Int, error)
	GetHeartbeats() ([]data.PubKeyHeartbeat, error)
	TpsBenchmark() statistics.TPSBenchmark
	StatusMetrics() external.StatusMetricsHandler
	GetQueryHandler(name string) (debug.QueryHandler, error)
	GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error)
	GetNumCheckpointsFromAccountState() uint32
	GetNumCheckpointsFromPeerState() uint32
	CreateTransaction(nonce uint64, value string, receiver string, receiverUsername []byte, sender string, senderUsername []byte, gasPrice uint64,
		gasLimit uint64, data []byte, signatureHex string, chainID string, version uint32, options uint32) (*transaction.Transaction, []byte, error)
	ValidateTransaction(tx *transaction.Transaction) error
	ValidateTransactionForSimulation(tx *transaction.Transaction, bypassSignature bool) error
	SendBulkTransactions([]*transaction.Transaction) (uint64, error)
	SimulateTransactionExecution(tx *transaction.Transaction) (*transaction.SimulationResults, error)
	GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error)
	EncodeAddressPubkey(pk []byte) (string, error)
	GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool)
	ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error)
	ExecuteSCQuery(*process.SCQuery) (*vm.VMOutputApi, error)
	DecodeAddressPubkey(pk string) ([]byte, error)
	CreateMiddlewareLimiters() ([]api.MiddlewareProcessor, error)
	IsInterfaceNil() bool
}
