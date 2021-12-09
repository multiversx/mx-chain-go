package mock

import (
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	txSimData "github.com/ElrondNetwork/elrond-go/process/txsimulator/data"
	"github.com/ElrondNetwork/elrond-go/state"
)

// FacadeStub is the mock implementation of a node router handler
type FacadeStub struct {
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
	GetESDTsRolesCalled                     func(address string) (map[string][]string, error)
	GetNFTTokenIDsRegisteredByAddressCalled func(address string) ([]string, error)
	GetBlockByHashCalled                    func(hash string, withTxs bool) (*api.Block, error)
	GetBlockByNonceCalled                   func(nonce uint64, withTxs bool) (*api.Block, error)
	GetBlockByRoundCalled                   func(round uint64, withTxs bool) (*api.Block, error)
	GetTotalStakedValueHandler              func() (*api.StakeValues, error)
	GetAllIssuedESDTsCalled                 func(tokenType string) ([]string, error)
	GetDirectStakedListHandler              func() ([]*api.DirectStakedValue, error)
	GetDelegatorsListHandler                func() ([]*api.Delegator, error)
	GetProofCalled                          func(string, string) (*common.GetProofResponse, error)
	GetProofCurrentRootHashCalled           func(string) (*common.GetProofResponse, error)
	GetProofDataTrieCalled                  func(string, string, string) (*common.GetProofResponse, *common.GetProofResponse, error)
	VerifyProofCalled                       func(string, string, [][]byte) (bool, error)
	GetTokenSupplyCalled                    func(token string) (string, error)
}

// GetTokenSupply -
func (f *FacadeStub) GetTokenSupply(token string) (string, error) {
	if f.GetTokenSupplyCalled != nil {
		return f.GetTokenSupplyCalled(token)
	}

	return "", nil
}

// GetProof -
func (f *FacadeStub) GetProof(rootHash string, address string) (*common.GetProofResponse, error) {
	if f.GetProofCalled != nil {
		return f.GetProofCalled(rootHash, address)
	}

	return nil, nil
}

// GetProofCurrentRootHash -
func (f *FacadeStub) GetProofCurrentRootHash(address string) (*common.GetProofResponse, error) {
	if f.GetProofCurrentRootHashCalled != nil {
		return f.GetProofCurrentRootHashCalled(address)
	}

	return nil, nil
}

// GetProofDataTrie -
func (f *FacadeStub) GetProofDataTrie(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error) {
	if f.GetProofDataTrieCalled != nil {
		return f.GetProofDataTrieCalled(rootHash, address, key)
	}

	return nil, nil, nil
}

// VerifyProof -
func (f *FacadeStub) VerifyProof(rootHash string, address string, proof [][]byte) (bool, error) {
	if f.VerifyProofCalled != nil {
		return f.VerifyProofCalled(rootHash, address, proof)
	}

	return false, nil
}

// GetUsername -
func (f *FacadeStub) GetUsername(address string) (string, error) {
	if f.GetUsernameCalled != nil {
		return f.GetUsernameCalled(address)
	}

	return "", nil
}

// GetThrottlerForEndpoint -
func (f *FacadeStub) GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool) {
	if f.GetThrottlerForEndpointCalled != nil {
		return f.GetThrottlerForEndpointCalled(endpoint)
	}

	return nil, false
}

// RestApiInterface -
func (f *FacadeStub) RestApiInterface() string {
	return "localhost:8080"
}

// RestAPIServerDebugMode -
func (f *FacadeStub) RestAPIServerDebugMode() bool {
	return false
}

// PprofEnabled -
func (f *FacadeStub) PprofEnabled() bool {
	return false
}

// GetHeartbeats returns the slice of heartbeat info
func (f *FacadeStub) GetHeartbeats() ([]data.PubKeyHeartbeat, error) {
	return f.GetHeartbeatsHandler()
}

// GetBalance is the mock implementation of a handler's GetBalance method
func (f *FacadeStub) GetBalance(address string) (*big.Int, error) {
	return f.BalanceHandler(address)
}

// GetValueForKey is the mock implementation of a handler's GetValueForKey method
func (f *FacadeStub) GetValueForKey(address string, key string) (string, error) {
	if f.GetValueForKeyCalled != nil {
		return f.GetValueForKeyCalled(address, key)
	}

	return "", nil
}

// GetKeyValuePairs -
func (f *FacadeStub) GetKeyValuePairs(address string) (map[string]string, error) {
	if f.GetKeyValuePairsCalled != nil {
		return f.GetKeyValuePairsCalled(address)
	}

	return nil, nil
}

// GetESDTData -
func (f *FacadeStub) GetESDTData(address string, key string, nonce uint64) (*esdt.ESDigitalToken, error) {
	if f.GetESDTDataCalled != nil {
		return f.GetESDTDataCalled(address, key, nonce)
	}

	return &esdt.ESDigitalToken{Value: big.NewInt(0)}, nil
}

// GetESDTsRoles -
func (f *FacadeStub) GetESDTsRoles(address string) (map[string][]string, error) {
	if f.GetESDTsRolesCalled != nil {
		return f.GetESDTsRolesCalled(address)
	}

	return map[string][]string{}, nil
}

// GetAllESDTTokens -
func (f *FacadeStub) GetAllESDTTokens(address string) (map[string]*esdt.ESDigitalToken, error) {
	if f.GetAllESDTTokensCalled != nil {
		return f.GetAllESDTTokensCalled(address)
	}

	return make(map[string]*esdt.ESDigitalToken), nil
}

// GetNFTTokenIDsRegisteredByAddress -
func (f *FacadeStub) GetNFTTokenIDsRegisteredByAddress(address string) ([]string, error) {
	if f.GetNFTTokenIDsRegisteredByAddressCalled != nil {
		return f.GetNFTTokenIDsRegisteredByAddressCalled(address)
	}

	return make([]string, 0), nil
}

// GetESDTsWithRole -
func (f *FacadeStub) GetESDTsWithRole(address string, role string) ([]string, error) {
	if f.GetESDTsWithRoleCalled != nil {
		return f.GetESDTsWithRoleCalled(address, role)
	}

	return make([]string, 0), nil
}

// GetAllIssuedESDTs -
func (f *FacadeStub) GetAllIssuedESDTs(tokenType string) ([]string, error) {
	if f.GetAllIssuedESDTsCalled != nil {
		return f.GetAllIssuedESDTsCalled(tokenType)
	}

	return make([]string, 0), nil
}

// GetAccount -
func (f *FacadeStub) GetAccount(address string) (api.AccountResponse, error) {
	return f.GetAccountHandler(address)
}

// CreateTransaction is  mock implementation of a handler's CreateTransaction method
func (f *FacadeStub) CreateTransaction(
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
func (f *FacadeStub) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	return f.GetTransactionHandler(hash, withResults)
}

// SimulateTransactionExecution is the mock implementation of a handler's SimulateTransactionExecution method
func (f *FacadeStub) SimulateTransactionExecution(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
	return f.SimulateTransactionExecutionHandler(tx)
}

// SendBulkTransactions is the mock implementation of a handler's SendBulkTransactions method
func (f *FacadeStub) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return f.SendBulkTransactionsHandler(txs)
}

// ValidateTransaction -
func (f *FacadeStub) ValidateTransaction(tx *transaction.Transaction) error {
	return f.ValidateTransactionHandler(tx)
}

// ValidateTransactionForSimulation -
func (f *FacadeStub) ValidateTransactionForSimulation(tx *transaction.Transaction, bypassSignature bool) error {
	return f.ValidateTransactionForSimulationHandler(tx, bypassSignature)
}

// ValidatorStatisticsApi is the mock implementation of a handler's ValidatorStatisticsApi method
func (f *FacadeStub) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return f.ValidatorStatisticsHandler()
}

// ExecuteSCQuery is a mock implementation.
func (f *FacadeStub) ExecuteSCQuery(query *process.SCQuery) (*vm.VMOutputApi, error) {
	return f.ExecuteSCQueryHandler(query)
}

// StatusMetrics is the mock implementation for the StatusMetrics
func (f *FacadeStub) StatusMetrics() external.StatusMetricsHandler {
	return f.StatusMetricsHandler()
}

// GetTotalStakedValue -
func (f *FacadeStub) GetTotalStakedValue() (*api.StakeValues, error) {
	return f.GetTotalStakedValueHandler()
}

// GetDirectStakedList -
func (f *FacadeStub) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	return f.GetDirectStakedListHandler()
}

// GetDelegatorsList -
func (f *FacadeStub) GetDelegatorsList() ([]*api.Delegator, error) {
	return f.GetDelegatorsListHandler()
}

// ComputeTransactionGasLimit -
func (f *FacadeStub) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	return f.ComputeTransactionGasLimitHandler(tx)
}

// NodeConfig -
func (f *FacadeStub) NodeConfig() map[string]interface{} {
	return f.NodeConfigCalled()
}

// EncodeAddressPubkey -
func (f *FacadeStub) EncodeAddressPubkey(pk []byte) (string, error) {
	return hex.EncodeToString(pk), nil
}

// DecodeAddressPubkey -
func (f *FacadeStub) DecodeAddressPubkey(pk string) ([]byte, error) {
	return hex.DecodeString(pk)
}

// GetQueryHandler -
func (f *FacadeStub) GetQueryHandler(name string) (debug.QueryHandler, error) {
	return f.GetQueryHandlerCalled(name)
}

// GetPeerInfo -
func (f *FacadeStub) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	return f.GetPeerInfoCalled(pid)
}

// GetNumCheckpointsFromAccountState -
func (f *FacadeStub) GetNumCheckpointsFromAccountState() uint32 {
	if f.GetNumCheckpointsFromAccountStateCalled != nil {
		return f.GetNumCheckpointsFromAccountStateCalled()
	}

	return 0
}

// GetNumCheckpointsFromPeerState -
func (f *FacadeStub) GetNumCheckpointsFromPeerState() uint32 {
	if f.GetNumCheckpointsFromPeerStateCalled != nil {
		return f.GetNumCheckpointsFromPeerStateCalled()
	}

	return 0
}

// GetBlockByNonce -
func (f *FacadeStub) GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error) {
	return f.GetBlockByNonceCalled(nonce, withTxs)
}

// GetBlockByHash -
func (f *FacadeStub) GetBlockByHash(hash string, withTxs bool) (*api.Block, error) {
	return f.GetBlockByHashCalled(hash, withTxs)
}

// GetBlockByRound -
func (f *FacadeStub) GetBlockByRound(round uint64, withTxs bool) (*api.Block, error) {
	if f.GetBlockByRoundCalled != nil {
		return f.GetBlockByRoundCalled(round, withTxs)
	}
	return nil, nil
}

// Trigger -
func (f *FacadeStub) Trigger(_ uint32, _ bool) error {
	return nil
}

// IsSelfTrigger -
func (f *FacadeStub) IsSelfTrigger() bool {
	return false
}

// Close -
func (f *FacadeStub) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *FacadeStub) IsInterfaceNil() bool {
	return f == nil
}

// WrongFacade is a struct that can be used as a wrong implementation of the node router handler
type WrongFacade struct {
}
