package mock

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/state"
)

// NodeStub -
type NodeStub struct {
	AddressHandler             func() (string, error)
	ConnectToAddressesHandler  func([]string) error
	GetBalanceHandler          func(address string) (*big.Int, error)
	GenerateTransactionHandler func(sender string, receiver string, amount string, code string) (*transaction.Transaction, error)
	CreateTransactionHandler   func(nonce uint64, value string, receiver string, receiverUsername []byte, sender string, senderUsername []byte, gasPrice uint64,
		gasLimit uint64, data []byte, signatureHex string, chainID string, version, options uint32) (*transaction.Transaction, []byte, error)
	ValidateTransactionHandler                     func(tx *transaction.Transaction) error
	ValidateTransactionForSimulationCalled         func(tx *transaction.Transaction, bypassSignature bool) error
	SendBulkTransactionsHandler                    func(txs []*transaction.Transaction) (uint64, error)
	GetAccountHandler                              func(address string) (api.AccountResponse, error)
	GetCodeCalled                                  func(codeHash []byte) []byte
	GetCurrentPublicKeyHandler                     func() string
	GenerateAndSendBulkTransactionsHandler         func(destination string, value *big.Int, nrTransactions uint64) error
	GenerateAndSendBulkTransactionsOneByOneHandler func(destination string, value *big.Int, nrTransactions uint64) error
	GetHeartbeatsHandler                           func() []data.PubKeyHeartbeat
	ValidatorStatisticsApiCalled                   func() (map[string]*state.ValidatorApiResponse, error)
	AuctionListApiCalled                           func() ([]*common.AuctionListValidatorAPIResponse, error)
	DirectTriggerCalled                            func(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTriggerCalled                            func() bool
	GetQueryHandlerCalled                          func(name string) (debug.QueryHandler, error)
	GetValueForKeyCalled                           func(address string, key string) (string, error)
	GetPeerInfoCalled                              func(pid string) ([]core.QueryP2PPeerInfo, error)
	GetUsernameCalled                              func(address string) (string, error)
	GetESDTDataCalled                              func(address string, key string, nonce uint64) (*esdt.ESDigitalToken, error)
	GetAllESDTTokensCalled                         func(address string, ctx context.Context) (map[string]*esdt.ESDigitalToken, error)
	GetNFTTokenIDsRegisteredByAddressCalled        func(address string, ctx context.Context) ([]string, error)
	GetESDTsWithRoleCalled                         func(address string, role string, ctx context.Context) ([]string, error)
	GetESDTsRolesCalled                            func(address string, ctx context.Context) (map[string][]string, error)
	GetKeyValuePairsCalled                         func(address string, ctx context.Context) (map[string]string, error)
	GetAllIssuedESDTsCalled                        func(tokenType string, ctx context.Context) ([]string, error)
	GetProofCalled                                 func(rootHash string, key string) (*common.GetProofResponse, error)
	GetProofDataTrieCalled                         func(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error)
	VerifyProofCalled                              func(rootHash string, address string, proof [][]byte) (bool, error)
}

// GetProof -
func (ns *NodeStub) GetProof(rootHash string, key string) (*common.GetProofResponse, error) {
	if ns.GetProofCalled != nil {
		return ns.GetProofCalled(rootHash, key)
	}

	return nil, nil
}

// GetProofDataTrie -
func (ns *NodeStub) GetProofDataTrie(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error) {
	if ns.GetProofDataTrieCalled != nil {
		return ns.GetProofDataTrieCalled(rootHash, address, key)
	}

	return nil, nil, nil
}

// VerifyProof -
func (ns *NodeStub) VerifyProof(rootHash string, address string, proof [][]byte) (bool, error) {
	if ns.VerifyProofCalled != nil {
		return ns.VerifyProofCalled(rootHash, address, proof)
	}

	return false, nil
}

// GetUsername -
func (ns *NodeStub) GetUsername(address string) (string, error) {
	if ns.GetUsernameCalled != nil {
		return ns.GetUsernameCalled(address)
	}

	return "", nil
}

// GetKeyValuePairs -
func (ns *NodeStub) GetKeyValuePairs(address string, ctx context.Context) (map[string]string, error) {
	if ns.GetKeyValuePairsCalled != nil {
		return ns.GetKeyValuePairsCalled(address, ctx)
	}

	return nil, nil
}

// GetValueForKey -
func (ns *NodeStub) GetValueForKey(address string, key string) (string, error) {
	if ns.GetValueForKeyCalled != nil {
		return ns.GetValueForKeyCalled(address, key)
	}

	return "", nil
}

// EncodeAddressPubkey -
func (ns *NodeStub) EncodeAddressPubkey(pk []byte) (string, error) {
	return hex.EncodeToString(pk), nil
}

// DecodeAddressPubkey -
func (ns *NodeStub) DecodeAddressPubkey(pk string) ([]byte, error) {
	return hex.DecodeString(pk)
}

// GetBalance -
func (ns *NodeStub) GetBalance(address string) (*big.Int, error) {
	if ns.GetBalanceHandler != nil {
		return ns.GetBalanceHandler(address)
	}

	return nil, nil
}

// CreateTransaction -
func (ns *NodeStub) CreateTransaction(nonce uint64, value string, receiver string, receiverUsername []byte, sender string, senderUsername []byte, gasPrice uint64,
	gasLimit uint64, data []byte, signatureHex string, chainID string, version uint32, options uint32) (*transaction.Transaction, []byte, error) {

	return ns.CreateTransactionHandler(nonce, value, receiver, receiverUsername, sender, senderUsername, gasPrice, gasLimit, data, signatureHex, chainID, version, options)
}

//ValidateTransaction -
func (ns *NodeStub) ValidateTransaction(tx *transaction.Transaction) error {
	if ns.ValidateTransactionHandler != nil {
		return ns.ValidateTransactionHandler(tx)
	}

	return nil
}

// ValidateTransactionForSimulation -
func (ns *NodeStub) ValidateTransactionForSimulation(tx *transaction.Transaction, bypassSignature bool) error {
	if ns.ValidateTransactionForSimulationCalled != nil {
		return ns.ValidateTransactionForSimulationCalled(tx, bypassSignature)
	}

	return nil
}

// SendBulkTransactions -
func (ns *NodeStub) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	if ns.SendBulkTransactionsHandler != nil {
		return ns.SendBulkTransactionsHandler(txs)
	}

	return 0, nil
}

// GetAccount -
func (ns *NodeStub) GetAccount(address string) (api.AccountResponse, error) {
	if ns.GetAccountHandler != nil {
		return ns.GetAccountHandler(address)
	}

	return api.AccountResponse{}, nil
}

// GetCode -
func (ns *NodeStub) GetCode(codeHash []byte) []byte {
	if ns.GetCodeCalled != nil {
		return ns.GetCodeCalled(codeHash)
	}

	return nil
}

// GetHeartbeats -
func (ns *NodeStub) GetHeartbeats() []data.PubKeyHeartbeat {
	if ns.GetHeartbeatsHandler != nil {
		return ns.GetHeartbeatsHandler()
	}

	return nil
}

// ValidatorStatisticsApi -
func (ns *NodeStub) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	if ns.ValidatorStatisticsApiCalled != nil {
		return ns.ValidatorStatisticsApiCalled()
	}

	return nil, nil
}

// AuctionListApi -
func (ns *NodeStub) AuctionListApi() ([]*common.AuctionListValidatorAPIResponse, error) {
	if ns.AuctionListApiCalled != nil {
		return ns.AuctionListApiCalled()
	}

	return nil, nil
}

// DirectTrigger -
func (ns *NodeStub) DirectTrigger(epoch uint32, withEarlyEndOfEpoch bool) error {
	if ns.DirectTriggerCalled != nil {
		return ns.DirectTriggerCalled(epoch, withEarlyEndOfEpoch)
	}

	return nil
}

// IsSelfTrigger -
func (ns *NodeStub) IsSelfTrigger() bool {
	if ns.IsSelfTriggerCalled != nil {
		return ns.IsSelfTriggerCalled()
	}

	return false
}

// GetQueryHandler -
func (ns *NodeStub) GetQueryHandler(name string) (debug.QueryHandler, error) {
	if ns.GetQueryHandlerCalled != nil {
		return ns.GetQueryHandlerCalled(name)
	}

	return nil, nil
}

// GetPeerInfo -
func (ns *NodeStub) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	if ns.GetPeerInfoCalled != nil {
		return ns.GetPeerInfoCalled(pid)
	}

	return make([]core.QueryP2PPeerInfo, 0), nil
}

// GetESDTData -
func (ns *NodeStub) GetESDTData(address, tokenID string, nonce uint64) (*esdt.ESDigitalToken, error) {
	if ns.GetESDTDataCalled != nil {
		return ns.GetESDTDataCalled(address, tokenID, nonce)
	}

	return &esdt.ESDigitalToken{Value: big.NewInt(0)}, nil
}

// GetESDTsRoles -
func (ns *NodeStub) GetESDTsRoles(address string, ctx context.Context) (map[string][]string, error) {
	if ns.GetESDTsRolesCalled != nil {
		return ns.GetESDTsRolesCalled(address, ctx)
	}

	return map[string][]string{}, nil
}

// GetESDTsWithRole -
func (ns *NodeStub) GetESDTsWithRole(address string, role string, ctx context.Context) ([]string, error) {
	if ns.GetESDTsWithRoleCalled != nil {
		return ns.GetESDTsWithRoleCalled(address, role, ctx)
	}

	return make([]string, 0), nil
}

// GetAllESDTTokens -
func (ns *NodeStub) GetAllESDTTokens(address string, ctx context.Context) (map[string]*esdt.ESDigitalToken, error) {
	if ns.GetAllESDTTokensCalled != nil {
		return ns.GetAllESDTTokensCalled(address, ctx)
	}

	return make(map[string]*esdt.ESDigitalToken), nil
}

// GetTokenSupply -
func (ns *NodeStub) GetTokenSupply(_ string) (*api.ESDTSupply, error) {
	return nil, nil
}

// GetAllIssuedESDTs -
func (ns *NodeStub) GetAllIssuedESDTs(tokenType string, ctx context.Context) ([]string, error) {
	if ns.GetAllIssuedESDTsCalled != nil {
		return ns.GetAllIssuedESDTsCalled(tokenType, ctx)
	}
	return make([]string, 0), nil
}

// GetNFTTokenIDsRegisteredByAddress -
func (ns *NodeStub) GetNFTTokenIDsRegisteredByAddress(address string, ctx context.Context) ([]string, error) {
	if ns.GetNFTTokenIDsRegisteredByAddressCalled != nil {
		return ns.GetNFTTokenIDsRegisteredByAddressCalled(address, ctx)
	}

	return make([]string, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ns *NodeStub) IsInterfaceNil() bool {
	return ns == nil
}
