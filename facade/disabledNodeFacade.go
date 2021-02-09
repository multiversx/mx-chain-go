package facade

import (
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/block"
	"github.com/ElrondNetwork/elrond-go/api/hardfork"
	"github.com/ElrondNetwork/elrond-go/api/node"
	transactionApi "github.com/ElrondNetwork/elrond-go/api/transaction"
	"github.com/ElrondNetwork/elrond-go/api/validator"
	"github.com/ElrondNetwork/elrond-go/api/vmValues"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/vm"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ = address.FacadeHandler(&disabledNodeFacade{})
var _ = hardfork.FacadeHandler(&disabledNodeFacade{})
var _ = node.FacadeHandler(&disabledNodeFacade{})
var _ = transactionApi.FacadeHandler(&disabledNodeFacade{})
var _ = validator.FacadeHandler(&disabledNodeFacade{})
var _ = vmValues.FacadeHandler(&disabledNodeFacade{})

var errNodeStarting = errors.New("Node is starting")
var emptStr string = ""

// nodeFacade represents a facade for grouping the functionality for the node
type disabledNodeFacade struct {
}

// NewDisableNodeFacade creates a new Facade with a NodeWrapper
func NewDisableNodeFacade() (*disabledNodeFacade, error) {
	return &disabledNodeFacade{}, nil
}

// SetSyncer sets the current syncer
func (nf *disabledNodeFacade) SetSyncer(_ ntp.SyncTimer) {
}

// SetTpsBenchmark sets the tps benchmark handler
func (nf *disabledNodeFacade) SetTpsBenchmark(_ statistics.TPSBenchmark) {
}

// TpsBenchmark returns the tps benchmark handler
func (nf *disabledNodeFacade) TpsBenchmark() statistics.TPSBenchmark {
	return nil
}

// StartBackgroundServices starts all background services needed for the correct functionality of the node
func (nf *disabledNodeFacade) StartBackgroundServices() {
}

// RestAPIServerDebugMode return true is debug mode for Rest API is enabled
func (nf *disabledNodeFacade) RestAPIServerDebugMode() bool {
	return false
}

// RestApiInterface returns the interface on which the rest API should start on, based on the config file provided.
// The API will start on the DefaultRestInterface value unless a correct value is passed or
//  the value is explicitly set to off, in which case it will not start at all
func (nf *disabledNodeFacade) RestApiInterface() string {
	return emptStr
}

// GetBalance gets the current balance for a specified address
func (nf *disabledNodeFacade) GetBalance(_ string) (*big.Int, error) {
	return nil, errNodeStarting
}

// GetUsername gets the username for a specified address
func (nf *disabledNodeFacade) GetUsername(_ string) (string, error) {
	return emptStr, errNodeStarting
}

// GetValueForKey gets the value for a key in a given address
func (nf *disabledNodeFacade) GetValueForKey(_ string, _ string) (string, error) {
	return emptStr, errNodeStarting
}

// GetESDTBalance returns the ESDT balance and if it is frozen
func (nf *disabledNodeFacade) GetESDTBalance(_ string, _ string) (string, string, error) {
	return emptStr, emptStr, errNodeStarting
}

// GetAllESDTTokens returns all the esdt tokens for a given address
func (nf *disabledNodeFacade) GetAllESDTTokens(_ string) ([]string, error) {
	return nil, errNodeStarting
}

// CreateTransaction creates a transaction from all needed fields
func (nf *disabledNodeFacade) CreateTransaction(
	_ uint64,
	_ string,
	_ string,
	_ string,
	_ uint64,
	_ uint64,
	_ []byte,
	_ string,
	_ string,
	_ uint32,
	_ uint32,
) (*transaction.Transaction, []byte, error) {
	return nil, nil, errNodeStarting
}

// ValidateTransaction will validate a transaction
func (nf *disabledNodeFacade) ValidateTransaction(_ *transaction.Transaction) error {
	return errNodeStarting
}

// ValidateTransactionForSimulation will validate a transaction for the simulation process
func (nf *disabledNodeFacade) ValidateTransactionForSimulation(_ *transaction.Transaction) error {
	return errNodeStarting
}

// ValidatorStatisticsApi will return the statistics for all validators
func (nf *disabledNodeFacade) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return nil, errNodeStarting
}

// SendBulkTransactions will send a bulk of transactions on the topic channel
func (nf *disabledNodeFacade) SendBulkTransactions(_ []*transaction.Transaction) (uint64, error) {
	return uint64(0), errNodeStarting
}

// SimulateTransactionExecution will simulate a transaction's execution and will return the results
func (nf *disabledNodeFacade) SimulateTransactionExecution(_ *transaction.Transaction) (*transaction.SimulationResults, error) {
	return nil, errNodeStarting
}

// GetTransaction gets the transaction with a specified hash
func (nf *disabledNodeFacade) GetTransaction(_ string, _ bool) (*transaction.ApiTransactionResult, error) {
	return nil, errNodeStarting
}

// ComputeTransactionGasLimit will estimate how many gas a transaction will consume
func (nf *disabledNodeFacade) ComputeTransactionGasLimit(_ *transaction.Transaction) (uint64, error) {
	return uint64(0), errNodeStarting
}

// GetAccount returns an accountResponse containing information
// about the account correlated with provided address
func (nf *disabledNodeFacade) GetAccount(_ string) (state.UserAccountHandler, error) {
	return nil, errNodeStarting
}

// GetHeartbeats returns the heartbeat status for each public key from initial list or later joined to the network
func (nf *disabledNodeFacade) GetHeartbeats() ([]data.PubKeyHeartbeat, error) {
	return nil, errNodeStarting
}

// StatusMetrics will return the node's status metrics
func (nf *disabledNodeFacade) StatusMetrics() external.StatusMetricsHandler {
	return nil
}

// ExecuteSCQuery retrieves data from existing SC trie
func (nf *disabledNodeFacade) ExecuteSCQuery(_ *process.SCQuery) (*vm.VMOutputApi, error) {
	return nil, errNodeStarting
}

// PprofEnabled returns if profiling mode should be active or not on the application
func (nf *disabledNodeFacade) PprofEnabled() bool {
	return false
}

// Trigger will trigger a hardfork event
func (nf *disabledNodeFacade) Trigger(_ uint32, _ bool) error {
	return errNodeStarting
}

// IsSelfTrigger returns true if the self public key is the same with the registered public key
func (nf *disabledNodeFacade) IsSelfTrigger() bool {
	return false
}

// EncodeAddressPubkey will encode the provided address public key bytes to string
func (nf *disabledNodeFacade) EncodeAddressPubkey(_ []byte) (string, error) {
	return emptStr, errNodeStarting
}

// DecodeAddressPubkey will try to decode the provided address public key string
func (nf *disabledNodeFacade) DecodeAddressPubkey(_ string) ([]byte, error) {
	return nil, errNodeStarting
}

// GetQueryHandler returns the query handler if existing
func (nf *disabledNodeFacade) GetQueryHandler(_ string) (debug.QueryHandler, error) {
	return nil, errNodeStarting
}

// GetPeerInfo returns the peer info of a provided pid
func (nf *disabledNodeFacade) GetPeerInfo(_ string) ([]core.QueryP2PPeerInfo, error) {
	return nil, errNodeStarting
}

// GetThrottlerForEndpoint returns the throttler for a given endpoint if found
func (nf *disabledNodeFacade) GetThrottlerForEndpoint(_ string) (core.Throttler, bool) {
	return nil, false
}

// GetBlockByHash return the block for a given hash
func (nf *disabledNodeFacade) GetBlockByHash(_ string, _ bool) (*block.APIBlock, error) {
	return nil, errNodeStarting
}

// GetBlockByNonce returns the block for a given nonce
func (nf *disabledNodeFacade) GetBlockByNonce(_ uint64, _ bool) (*block.APIBlock, error) {
	return nil, errNodeStarting
}

// Close will cleanup started go routines
func (nf *disabledNodeFacade) Close() error {
	return errNodeStarting
}

// GetNumCheckpointsFromAccountState returns the number of checkpoints of the account state
func (nf *disabledNodeFacade) GetNumCheckpointsFromAccountState() uint32 {
	return uint32(0)
}

// GetNumCheckpointsFromPeerState returns the number of checkpoints of the peer state
func (nf *disabledNodeFacade) GetNumCheckpointsFromPeerState() uint32 {
	return uint32(0)
}

func (nf *disabledNodeFacade) convertVmOutputToApiResponse(_ *vmcommon.VMOutput) *vm.VMOutputApi {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nf *disabledNodeFacade) IsInterfaceNil() bool {
	return nf == nil
}
