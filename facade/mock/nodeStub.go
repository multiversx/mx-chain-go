package mock

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/state/accounts"
)

// NodeStub -
type NodeStub struct {
	ConnectToAddressesHandler                      func([]string) error
	GetBalanceCalled                               func(address string, options api.AccountQueryOptions) (*big.Int, api.BlockInfo, error)
	GenerateTransactionHandler                     func(sender string, receiver string, amount string, code string) (*transaction.Transaction, error)
	CreateTransactionHandler                       func(txArgs *external.ArgsCreateTransaction) (*transaction.Transaction, []byte, error)
	ValidateTransactionHandler                     func(tx *transaction.Transaction) error
	ValidateTransactionForSimulationCalled         func(tx *transaction.Transaction, bypassSignature bool) error
	SendBulkTransactionsHandler                    func(txs []*transaction.Transaction) (uint64, error)
	GetAccountCalled                               func(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error)
	GetCodeCalled                                  func(codeHash []byte, options api.AccountQueryOptions) ([]byte, api.BlockInfo)
	GetCurrentPublicKeyHandler                     func() string
	GenerateAndSendBulkTransactionsHandler         func(destination string, value *big.Int, nrTransactions uint64) error
	GenerateAndSendBulkTransactionsOneByOneHandler func(destination string, value *big.Int, nrTransactions uint64) error
	GetHeartbeatsHandler                           func() []data.PubKeyHeartbeat
	ValidatorStatisticsApiCalled                   func() (map[string]*accounts.ValidatorApiResponse, error)
	DirectTriggerCalled                            func(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTriggerCalled                            func() bool
	GetQueryHandlerCalled                          func(name string) (debug.QueryHandler, error)
	GetValueForKeyCalled                           func(address string, key string, options api.AccountQueryOptions) (string, api.BlockInfo, error)
	GetGuardianDataCalled                          func(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error)
	GetPeerInfoCalled                              func(pid string) ([]core.QueryP2PPeerInfo, error)
	GetConnectedPeersRatingsOnMainNetworkCalled    func() (string, error)
	GetEpochStartDataAPICalled                     func(epoch uint32) (*common.EpochStartDataAPI, error)
	GetUsernameCalled                              func(address string, options api.AccountQueryOptions) (string, api.BlockInfo, error)
	GetCodeHashCalled                              func(address string, options api.AccountQueryOptions) ([]byte, api.BlockInfo, error)
	GetESDTDataCalled                              func(address string, key string, nonce uint64, options api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error)
	GetAllESDTTokensCalled                         func(address string, options api.AccountQueryOptions, ctx context.Context) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error)
	GetNFTTokenIDsRegisteredByAddressCalled        func(address string, options api.AccountQueryOptions, ctx context.Context) ([]string, api.BlockInfo, error)
	GetESDTsWithRoleCalled                         func(address string, role string, options api.AccountQueryOptions, ctx context.Context) ([]string, api.BlockInfo, error)
	GetESDTsRolesCalled                            func(address string, options api.AccountQueryOptions, ctx context.Context) (map[string][]string, api.BlockInfo, error)
	GetKeyValuePairsCalled                         func(address string, options api.AccountQueryOptions, ctx context.Context) (map[string]string, api.BlockInfo, error)
	GetAllIssuedESDTsCalled                        func(tokenType string, ctx context.Context) ([]string, error)
	GetProofCalled                                 func(rootHash string, key string) (*common.GetProofResponse, error)
	GetProofDataTrieCalled                         func(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error)
	VerifyProofCalled                              func(rootHash string, address string, proof [][]byte) (bool, error)
	GetTokenSupplyCalled                           func(token string) (*api.ESDTSupply, error)
	IsDataTrieMigratedCalled                       func(address string, options api.AccountQueryOptions) (bool, error)
	SendBlockTransactionsHandler                   func([]*api.Block) (uint64, error)
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
func (ns *NodeStub) GetUsername(address string, options api.AccountQueryOptions) (string, api.BlockInfo, error) {
	if ns.GetUsernameCalled != nil {
		return ns.GetUsernameCalled(address, options)
	}

	return "", api.BlockInfo{}, nil
}

// GetCodeHash -
func (ns *NodeStub) GetCodeHash(address string, options api.AccountQueryOptions) ([]byte, api.BlockInfo, error) {
	if ns.GetCodeHashCalled != nil {
		return ns.GetCodeHashCalled(address, options)
	}

	return nil, api.BlockInfo{}, nil
}

// GetKeyValuePairs -
func (ns *NodeStub) GetKeyValuePairs(address string, options api.AccountQueryOptions, ctx context.Context) (map[string]string, api.BlockInfo, error) {
	if ns.GetKeyValuePairsCalled != nil {
		return ns.GetKeyValuePairsCalled(address, options, ctx)
	}

	return nil, api.BlockInfo{}, nil
}

// GetValueForKey -
func (ns *NodeStub) GetValueForKey(address string, key string, options api.AccountQueryOptions) (string, api.BlockInfo, error) {
	if ns.GetValueForKeyCalled != nil {
		return ns.GetValueForKeyCalled(address, key, options)
	}

	return "", api.BlockInfo{}, nil
}

// GetGuardianData -
func (ns *NodeStub) GetGuardianData(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error) {
	if ns.GetGuardianDataCalled != nil {
		return ns.GetGuardianDataCalled(address, options)
	}
	return api.GuardianData{}, api.BlockInfo{}, nil
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
func (ns *NodeStub) GetBalance(address string, options api.AccountQueryOptions) (*big.Int, api.BlockInfo, error) {
	return ns.GetBalanceCalled(address, options)
}

// CreateTransaction -
func (ns *NodeStub) CreateTransaction(txArgs *external.ArgsCreateTransaction) (*transaction.Transaction, []byte, error) {

	return ns.CreateTransactionHandler(txArgs)
}

// ValidateTransaction -
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

// SendBlockTransactions -
func (ns *NodeStub) SendBlockTransactions(apiBlocks []*api.Block) (uint64, error) {
	return ns.SendBlockTransactionsHandler(apiBlocks)
}

// GetAccount -
func (ns *NodeStub) GetAccount(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
	return ns.GetAccountCalled(address, options)
}

// GetCode -
func (ns *NodeStub) GetCode(codeHash []byte, options api.AccountQueryOptions) ([]byte, api.BlockInfo) {
	if ns.GetCodeCalled != nil {
		return ns.GetCodeCalled(codeHash, options)
	}

	return nil, api.BlockInfo{}
}

// GetHeartbeats -
func (ns *NodeStub) GetHeartbeats() []data.PubKeyHeartbeat {
	return ns.GetHeartbeatsHandler()
}

// ValidatorStatisticsApi -
func (ns *NodeStub) ValidatorStatisticsApi() (map[string]*accounts.ValidatorApiResponse, error) {
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

// GetConnectedPeersRatingsOnMainNetwork -
func (ns *NodeStub) GetConnectedPeersRatingsOnMainNetwork() (string, error) {
	if ns.GetConnectedPeersRatingsOnMainNetworkCalled != nil {
		return ns.GetConnectedPeersRatingsOnMainNetworkCalled()
	}

	return "", nil
}

// GetEpochStartDataAPI -
func (ns *NodeStub) GetEpochStartDataAPI(epoch uint32) (*common.EpochStartDataAPI, error) {
	if ns.GetEpochStartDataAPICalled != nil {
		return ns.GetEpochStartDataAPICalled(epoch)
	}

	return &common.EpochStartDataAPI{}, nil
}

// GetESDTData -
func (ns *NodeStub) GetESDTData(address, tokenID string, nonce uint64, options api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
	if ns.GetESDTDataCalled != nil {
		return ns.GetESDTDataCalled(address, tokenID, nonce, options)
	}

	return &esdt.ESDigitalToken{Value: big.NewInt(0)}, api.BlockInfo{}, nil
}

// GetESDTsRoles -
func (ns *NodeStub) GetESDTsRoles(address string, options api.AccountQueryOptions, ctx context.Context) (map[string][]string, api.BlockInfo, error) {
	if ns.GetESDTsRolesCalled != nil {
		return ns.GetESDTsRolesCalled(address, options, ctx)
	}

	return map[string][]string{}, api.BlockInfo{}, nil
}

// GetESDTsWithRole -
func (ns *NodeStub) GetESDTsWithRole(address string, role string, options api.AccountQueryOptions, ctx context.Context) ([]string, api.BlockInfo, error) {
	if ns.GetESDTsWithRoleCalled != nil {
		return ns.GetESDTsWithRoleCalled(address, role, options, ctx)
	}

	return make([]string, 0), api.BlockInfo{}, nil
}

// GetAllESDTTokens -
func (ns *NodeStub) GetAllESDTTokens(address string, options api.AccountQueryOptions, ctx context.Context) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error) {
	if ns.GetAllESDTTokensCalled != nil {
		return ns.GetAllESDTTokensCalled(address, options, ctx)
	}

	return make(map[string]*esdt.ESDigitalToken), api.BlockInfo{}, nil
}

// GetTokenSupply -
func (ns *NodeStub) GetTokenSupply(token string) (*api.ESDTSupply, error) {
	if ns.GetTokenSupplyCalled != nil {
		return ns.GetTokenSupplyCalled(token)
	}
	return nil, nil
}

// GetAllIssuedESDTs -
func (ns *NodeStub) GetAllIssuedESDTs(tokenType string, ctx context.Context) ([]string, error) {
	if ns.GetAllIssuedESDTsCalled != nil {
		return ns.GetAllIssuedESDTsCalled(tokenType, ctx)
	}
	return make([]string, 0), nil
}

// IsDataTrieMigrated -
func (ns *NodeStub) IsDataTrieMigrated(address string, options api.AccountQueryOptions) (bool, error) {
	if ns.IsDataTrieMigratedCalled != nil {
		return ns.IsDataTrieMigratedCalled(address, options)
	}
	return false, nil
}

// GetNFTTokenIDsRegisteredByAddress -
func (ns *NodeStub) GetNFTTokenIDsRegisteredByAddress(address string, options api.AccountQueryOptions, ctx context.Context) ([]string, api.BlockInfo, error) {
	if ns.GetNFTTokenIDsRegisteredByAddressCalled != nil {
		return ns.GetNFTTokenIDsRegisteredByAddressCalled(address, options, ctx)
	}

	return make([]string, 0), api.BlockInfo{}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ns *NodeStub) IsInterfaceNil() bool {
	return ns == nil
}
