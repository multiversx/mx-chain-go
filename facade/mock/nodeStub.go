package mock

import (
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
	DirectTriggerCalled                            func(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTriggerCalled                            func() bool
	GetQueryHandlerCalled                          func(name string) (debug.QueryHandler, error)
	GetValueForKeyCalled                           func(address string, key string) (string, error)
	GetPeerInfoCalled                              func(pid string) ([]core.QueryP2PPeerInfo, error)
	GetUsernameCalled                              func(address string) (string, error)
	GetESDTDataCalled                              func(address string, key string, nonce uint64) (*esdt.ESDigitalToken, error)
	GetAllESDTTokensCalled                         func(address string) (map[string]*esdt.ESDigitalToken, error)
	GetNFTTokenIDsRegisteredByAddressCalled        func(address string) ([]string, error)
	GetESDTsWithRoleCalled                         func(address string, role string) ([]string, error)
	GetESDTsRolesCalled                            func(address string) (map[string][]string, error)
	GetKeyValuePairsCalled                         func(address string) (map[string]string, error)
	GetAllIssuedESDTsCalled                        func(tokenType string) ([]string, error)
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
func (ns *NodeStub) GetKeyValuePairs(address string) (map[string]string, error) {
	if ns.GetKeyValuePairsCalled != nil {
		return ns.GetKeyValuePairsCalled(address)
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
	return ns.GetBalanceHandler(address)
}

// CreateTransaction -
func (ns *NodeStub) CreateTransaction(nonce uint64, value string, receiver string, receiverUsername []byte, sender string, senderUsername []byte, gasPrice uint64,
	gasLimit uint64, data []byte, signatureHex string, chainID string, version uint32, options uint32) (*transaction.Transaction, []byte, error) {

	return ns.CreateTransactionHandler(nonce, value, receiver, receiverUsername, sender, senderUsername, gasPrice, gasLimit, data, signatureHex, chainID, version, options)
}

//ValidateTransaction -
func (ns *NodeStub) ValidateTransaction(tx *transaction.Transaction) error {
	return ns.ValidateTransactionHandler(tx)
}

// ValidateTransactionForSimulation -
func (ns *NodeStub) ValidateTransactionForSimulation(tx *transaction.Transaction, bypassSignature bool) error {
	return ns.ValidateTransactionForSimulationCalled(tx, bypassSignature)
}

// SendBulkTransactions -
func (ns *NodeStub) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return ns.SendBulkTransactionsHandler(txs)
}

// GetAccount -
func (ns *NodeStub) GetAccount(address string) (api.AccountResponse, error) {
	return ns.GetAccountHandler(address)
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
	return ns.GetHeartbeatsHandler()
}

// ValidatorStatisticsApi -
func (ns *NodeStub) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return ns.ValidatorStatisticsApiCalled()
}

// DirectTrigger -
func (ns *NodeStub) DirectTrigger(epoch uint32, withEarlyEndOfEpoch bool) error {
	return ns.DirectTriggerCalled(epoch, withEarlyEndOfEpoch)
}

// IsSelfTrigger -
func (ns *NodeStub) IsSelfTrigger() bool {
	return ns.IsSelfTriggerCalled()
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
func (ns *NodeStub) GetESDTsRoles(address string) (map[string][]string, error) {
	if ns.GetESDTsRolesCalled != nil {
		return ns.GetESDTsRolesCalled(address)
	}

	return map[string][]string{}, nil
}

// GetESDTsWithRole -
func (ns *NodeStub) GetESDTsWithRole(address string, role string) ([]string, error) {
	if ns.GetESDTsWithRoleCalled != nil {
		return ns.GetESDTsWithRoleCalled(address, role)
	}

	return make([]string, 0), nil
}

// GetAllESDTTokens -
func (ns *NodeStub) GetAllESDTTokens(address string) (map[string]*esdt.ESDigitalToken, error) {
	if ns.GetAllESDTTokensCalled != nil {
		return ns.GetAllESDTTokensCalled(address)
	}

	return make(map[string]*esdt.ESDigitalToken), nil
}

// GetTokenSupply -
func (ns *NodeStub) GetTokenSupply(_ string) (*api.ESDTSupply, error) {
	return nil, nil
}

// GetAllIssuedESDTs -
func (ns *NodeStub) GetAllIssuedESDTs(tokenType string) ([]string, error) {
	if ns.GetAllIssuedESDTsCalled != nil {
		return ns.GetAllIssuedESDTsCalled(tokenType)
	}
	return make([]string, 0), nil
}

// GetNFTTokenIDsRegisteredByAddress -
func (ns *NodeStub) GetNFTTokenIDsRegisteredByAddress(address string) ([]string, error) {
	if ns.GetNFTTokenIDsRegisteredByAddressCalled != nil {
		return ns.GetNFTTokenIDsRegisteredByAddressCalled(address)
	}

	return make([]string, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ns *NodeStub) IsInterfaceNil() bool {
	return ns == nil
}
