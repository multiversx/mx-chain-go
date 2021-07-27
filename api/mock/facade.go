package mock

import (
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	txSimData "github.com/ElrondNetwork/elrond-go/process/txsimulator/data"
	"github.com/ElrondNetwork/elrond-go/state"
)

// Facade is the mock implementation of a node router handler
type Facade struct {
	ShouldErrorStart           bool
	ShouldErrorStop            bool
	GetHeartbeatsHandler       func() ([]data.PubKeyHeartbeat, error)
	BalanceHandler             func(string) (*big.Int, error)
	GetAccountHandler          func(address string) (api.AccountResponse, error)
	GenerateTransactionHandler func(sender string, receiver string, value *big.Int, code string) (*transaction.Transaction, error)
	GetTransactionHandler      func(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	CreateTransactionHandler   func(nonce uint64, value string, receiver string, receiverUsername []byte, sender string, senderUsername []byte, gasPrice uint64,
		gasLimit uint64, data []byte, signatureHex string, chainID string, version uint32, options uint32) (*transaction.Transaction, []byte, error)
	ValidateTransactionHandler              func(tx *transaction.Transaction) error
	ValidateTransactionForSimulationHandler func(tx *transaction.Transaction, bypassSignature bool) error
	SendBulkTransactionsHandler             func(txs []*transaction.Transaction) (uint64, error)
	ExecuteSCQueryHandler                   func(query *process.SCQuery) (*vm.VMOutputApi, error)
	StatusMetricsHandler                    func() external.StatusMetricsHandler
	ValidatorStatisticsHandler              func() (map[string]*state.ValidatorApiResponse, error)
	ComputeTransactionGasLimitHandler       func(tx *transaction.Transaction) (*transaction.CostResponse, error)
	NodeConfigCalled                        func() map[string]interface{}
	GetQueryHandlerCalled                   func(name string) (debug.QueryHandler, error)
	GetValueForKeyCalled                    func(address string, key string) (string, error)
	GetPeerInfoCalled                       func(pid string) ([]core.QueryP2PPeerInfo, error)
	GetThrottlerForEndpointCalled           func(endpoint string) (core.Throttler, bool)
	GetUsernameCalled                       func(address string) (string, error)
	GetKeyValuePairsCalled                  func(address string) (map[string]string, error)
	SimulateTransactionExecutionHandler     func(tx *transaction.Transaction) (*txSimData.SimulationResults, error)
	GetNumCheckpointsFromAccountStateCalled func() uint32
	GetNumCheckpointsFromPeerStateCalled    func() uint32
	GetESDTDataCalled                       func(address string, key string, nonce uint64) (*esdt.ESDigitalToken, error)
	GetAllESDTTokensCalled                  func(address string) (map[string]*esdt.ESDigitalToken, error)
	GetESDTsWithRoleCalled                  func(address string, role string) ([]string, error)
	GetNFTTokenIDsRegisteredByAddressCalled func(address string) ([]string, error)
	GetBlockByHashCalled                    func(hash string, withTxs bool) (*api.Block, error)
	GetBlockByNonceCalled                   func(nonce uint64, withTxs bool) (*api.Block, error)
	GetTotalStakedValueHandler              func() (*api.StakeValues, error)
	GetAllIssuedESDTsCalled                 func(tokenType string) ([]string, error)
	GetDirectStakedListHandler              func() ([]*api.DirectStakedValue, error)
	GetDelegatorsListHandler                func() ([]*api.Delegator, error)
	GetProofCalled                          func(string, string) ([][]byte, error)
	GetProofCurrentRootHashCalled           func(string) ([][]byte, []byte, error)
	VerifyProofCalled                       func(string, string, [][]byte) (bool, error)
}

// GetProof -
func (f *Facade) GetProof(rootHash string, address string) ([][]byte, error) {
	if f.GetProofCalled != nil {
		return f.GetProofCalled(rootHash, address)
	}

	return nil, nil
}

// GetProofCurrentRootHash -
func (f *Facade) GetProofCurrentRootHash(address string) ([][]byte, []byte, error) {
	if f.GetProofCurrentRootHashCalled != nil {
		return f.GetProofCurrentRootHashCalled(address)
	}

	return nil, nil, nil
}

// VerifyProof -
func (f *Facade) VerifyProof(rootHash string, address string, proof [][]byte) (bool, error) {
	if f.VerifyProofCalled != nil {
		return f.VerifyProofCalled(rootHash, address, proof)
	}

	return false, nil
}

// GetUsername -
func (f *Facade) GetUsername(address string) (string, error) {
	if f.GetUsernameCalled != nil {
		return f.GetUsernameCalled(address)
	}

	return "", nil
}

// GetThrottlerForEndpoint -
func (f *Facade) GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool) {
	if f.GetThrottlerForEndpointCalled != nil {
		return f.GetThrottlerForEndpointCalled(endpoint)
	}

	return nil, false
}

// RestApiInterface -
func (f *Facade) RestApiInterface() string {
	return "localhost:8080"
}

// RestAPIServerDebugMode -
func (f *Facade) RestAPIServerDebugMode() bool {
	return false
}

// PprofEnabled -
func (f *Facade) PprofEnabled() bool {
	return false
}

// GetHeartbeats returns the slice of heartbeat info
func (f *Facade) GetHeartbeats() ([]data.PubKeyHeartbeat, error) {
	return f.GetHeartbeatsHandler()
}

// GetBalance is the mock implementation of a handler's GetBalance method
func (f *Facade) GetBalance(address string) (*big.Int, error) {
	return f.BalanceHandler(address)
}

// GetValueForKey is the mock implementation of a handler's GetValueForKey method
func (f *Facade) GetValueForKey(address string, key string) (string, error) {
	if f.GetValueForKeyCalled != nil {
		return f.GetValueForKeyCalled(address, key)
	}

	return "", nil
}

// GetKeyValuePairs -
func (f *Facade) GetKeyValuePairs(address string) (map[string]string, error) {
	if f.GetKeyValuePairsCalled != nil {
		return f.GetKeyValuePairsCalled(address)
	}

	return nil, nil
}

// GetESDTData -
func (f *Facade) GetESDTData(address string, key string, nonce uint64) (*esdt.ESDigitalToken, error) {
	if f.GetESDTDataCalled != nil {
		return f.GetESDTDataCalled(address, key, nonce)
	}

	return &esdt.ESDigitalToken{Value: big.NewInt(0)}, nil
}

// GetAllESDTTokens -
func (f *Facade) GetAllESDTTokens(address string) (map[string]*esdt.ESDigitalToken, error) {
	if f.GetAllESDTTokensCalled != nil {
		return f.GetAllESDTTokensCalled(address)
	}

	return make(map[string]*esdt.ESDigitalToken), nil
}

// GetNFTTokenIDsRegisteredByAddress -
func (f *Facade) GetNFTTokenIDsRegisteredByAddress(address string) ([]string, error) {
	if f.GetNFTTokenIDsRegisteredByAddressCalled != nil {
		return f.GetNFTTokenIDsRegisteredByAddressCalled(address)
	}

	return make([]string, 0), nil
}

// GetESDTsWithRole -
func (f *Facade) GetESDTsWithRole(address string, role string) ([]string, error) {
	if f.GetESDTsWithRoleCalled != nil {
		return f.GetESDTsWithRoleCalled(address, role)
	}

	return make([]string, 0), nil
}

// GetAllIssuedESDTs -
func (f *Facade) GetAllIssuedESDTs(tokenType string) ([]string, error) {
	if f.GetAllIssuedESDTsCalled != nil {
		return f.GetAllIssuedESDTsCalled(tokenType)
	}

	return make([]string, 0), nil
}

// GetAccount -
func (f *Facade) GetAccount(address string) (api.AccountResponse, error) {
	return f.GetAccountHandler(address)
}

// CreateTransaction is  mock implementation of a handler's CreateTransaction method
func (f *Facade) CreateTransaction(
	nonce uint64,
	value string,
	receiver string,
	receiverUsername []byte,
	sender string,
	senderUsername []byte,
	gasPrice uint64,
	gasLimit uint64,
	data []byte,
	signatureHex string,
	chainID string,
	version uint32,
	options uint32,
) (*transaction.Transaction, []byte, error) {
	return f.CreateTransactionHandler(nonce, value, receiver, receiverUsername, sender, senderUsername, gasPrice, gasLimit, data, signatureHex, chainID, version, options)
}

// GetTransaction is the mock implementation of a handler's GetTransaction method
func (f *Facade) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	return f.GetTransactionHandler(hash, withResults)
}

// SimulateTransactionExecution is the mock implementation of a handler's SimulateTransactionExecution method
func (f *Facade) SimulateTransactionExecution(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
	return f.SimulateTransactionExecutionHandler(tx)
}

// SendBulkTransactions is the mock implementation of a handler's SendBulkTransactions method
func (f *Facade) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return f.SendBulkTransactionsHandler(txs)
}

// ValidateTransaction -
func (f *Facade) ValidateTransaction(tx *transaction.Transaction) error {
	return f.ValidateTransactionHandler(tx)
}

// ValidateTransactionForSimulation -
func (f *Facade) ValidateTransactionForSimulation(tx *transaction.Transaction, bypassSignature bool) error {
	return f.ValidateTransactionForSimulationHandler(tx, bypassSignature)
}

// ValidatorStatisticsApi is the mock implementation of a handler's ValidatorStatisticsApi method
func (f *Facade) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return f.ValidatorStatisticsHandler()
}

// ExecuteSCQuery is a mock implementation.
func (f *Facade) ExecuteSCQuery(query *process.SCQuery) (*vm.VMOutputApi, error) {
	return f.ExecuteSCQueryHandler(query)
}

// StatusMetrics is the mock implementation for the StatusMetrics
func (f *Facade) StatusMetrics() external.StatusMetricsHandler {
	return f.StatusMetricsHandler()
}

// GetTotalStakedValue -
func (f *Facade) GetTotalStakedValue() (*api.StakeValues, error) {
	return f.GetTotalStakedValueHandler()
}

// GetDirectStakedList -
func (f *Facade) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	return f.GetDirectStakedListHandler()
}

// GetDelegatorsList -
func (f *Facade) GetDelegatorsList() ([]*api.Delegator, error) {
	return f.GetDelegatorsListHandler()
}

// ComputeTransactionGasLimit -
func (f *Facade) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	return f.ComputeTransactionGasLimitHandler(tx)
}

// NodeConfig -
func (f *Facade) NodeConfig() map[string]interface{} {
	return f.NodeConfigCalled()
}

// EncodeAddressPubkey -
func (f *Facade) EncodeAddressPubkey(pk []byte) (string, error) {
	return hex.EncodeToString(pk), nil
}

// DecodeAddressPubkey -
func (f *Facade) DecodeAddressPubkey(pk string) ([]byte, error) {
	return hex.DecodeString(pk)
}

// GetQueryHandler -
func (f *Facade) GetQueryHandler(name string) (debug.QueryHandler, error) {
	return f.GetQueryHandlerCalled(name)
}

// GetPeerInfo -
func (f *Facade) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	return f.GetPeerInfoCalled(pid)
}

// GetNumCheckpointsFromAccountState -
func (f *Facade) GetNumCheckpointsFromAccountState() uint32 {
	if f.GetNumCheckpointsFromAccountStateCalled != nil {
		return f.GetNumCheckpointsFromAccountStateCalled()
	}

	return 0
}

// GetNumCheckpointsFromPeerState -
func (f *Facade) GetNumCheckpointsFromPeerState() uint32 {
	if f.GetNumCheckpointsFromPeerStateCalled != nil {
		return f.GetNumCheckpointsFromPeerStateCalled()
	}

	return 0
}

// GetBlockByNonce -
func (f *Facade) GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error) {
	return f.GetBlockByNonceCalled(nonce, withTxs)
}

// GetBlockByHash -
func (f *Facade) GetBlockByHash(hash string, withTxs bool) (*api.Block, error) {
	return f.GetBlockByHashCalled(hash, withTxs)
}

// Close -
func (f *Facade) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *Facade) IsInterfaceNil() bool {
	return f == nil
}

// WrongFacade is a struct that can be used as a wrong implementation of the node router handler
type WrongFacade struct {
}
