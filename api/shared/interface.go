package shared

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	txSimData "github.com/ElrondNetwork/elrond-go/process/txsimulator/data"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/gin-gonic/gin"
)

// HttpServerCloser defines the basic actions of starting and closing that a web server should be able to do
type HttpServerCloser interface {
	Start()
	Close() error
	IsInterfaceNil() bool
}

// MiddlewareProcessor defines a processor used internally by the web server when processing requests
type MiddlewareProcessor interface {
	MiddlewareHandlerFunc() gin.HandlerFunc
	IsInterfaceNil() bool
}

// ApiFacadeHandler interface defines methods that can be used from `elrondFacade` context variable
type ApiFacadeHandler interface {
	RestApiInterface() string
	RestAPIServerDebugMode() bool
	PprofEnabled() bool
	IsInterfaceNil() bool
}

// UpgradeableHttpServerHandler defines the actions that an upgradeable http server need to do
type UpgradeableHttpServerHandler interface {
	StartHttpServer() error
	UpdateFacade(facade FacadeHandler) error
	Close() error
	IsInterfaceNil() bool
}

// GroupHandler defines the actions needed to be performed by an gin API group
type GroupHandler interface {
	UpdateFacade(newFacade interface{}) error
	RegisterRoutes(
		ws *gin.RouterGroup,
		apiConfig config.ApiRoutesConfig,
	)
	IsInterfaceNil() bool
}

// FacadeHandler defines all the methods that a facade should implement
type FacadeHandler interface {
	GetBalance(address string, options api.AccountQueryOptions) (*big.Int, api.BlockInfo, error)
	GetUsername(address string, options api.AccountQueryOptions) (string, api.BlockInfo, error)
	GetValueForKey(address string, key string, options api.AccountQueryOptions) (string, api.BlockInfo, error)
	GetAccount(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error)
	GetESDTData(address string, key string, nonce uint64, options api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error)
	GetESDTsRoles(address string, options api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error)
	GetNFTTokenIDsRegisteredByAddress(address string, options api.AccountQueryOptions) ([]string, api.BlockInfo, error)
	GetESDTsWithRole(address string, role string, options api.AccountQueryOptions) ([]string, api.BlockInfo, error)
	GetAllESDTTokens(address string, options api.AccountQueryOptions) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error)
	GetKeyValuePairs(address string, options api.AccountQueryOptions) (map[string]string, api.BlockInfo, error)
	GetBlockByHash(hash string, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByRound(round uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalShardBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error)
	GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error)
	GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalStartOfEpochMetaBlock(format common.ApiOutputFormat, epoch uint32) (interface{}, error)
	GetInternalMiniBlockByHash(format common.ApiOutputFormat, hash string, epoch uint32) (interface{}, error)
	Trigger(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTrigger() bool
	GetTotalStakedValue() (*api.StakeValues, error)
	GetDirectStakedList() ([]*api.DirectStakedValue, error)
	GetDelegatorsList() ([]*api.Delegator, error)
	StatusMetrics() external.StatusMetricsHandler
	GetTokenSupply(token string) (*api.ESDTSupply, error)
	GetAllIssuedESDTs(tokenType string) ([]string, error)
	GetHeartbeats() ([]data.PubKeyHeartbeat, error)
	GetQueryHandler(name string) (debug.QueryHandler, error)
	GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error)
	GetProof(rootHash string, address string) (*common.GetProofResponse, error)
	GetProofDataTrie(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error)
	GetProofCurrentRootHash(address string) (*common.GetProofResponse, error)
	VerifyProof(rootHash string, address string, proof [][]byte) (bool, error)
	GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool)
	CreateTransaction(nonce uint64, value string, receiver string, receiverUsername []byte, sender string, senderUsername []byte, gasPrice uint64,
		gasLimit uint64, data []byte, signatureHex string, chainID string, version uint32, options uint32) (*transaction.Transaction, []byte, error)
	ValidateTransaction(tx *transaction.Transaction) error
	ValidateTransactionForSimulation(tx *transaction.Transaction, checkSignature bool) error
	SendBulkTransactions([]*transaction.Transaction) (uint64, error)
	SimulateTransactionExecution(tx *transaction.Transaction) (*txSimData.SimulationResults, error)
	GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error)
	EncodeAddressPubkey(pk []byte) (string, error)
	ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error)
	ExecuteSCQuery(*process.SCQuery) (*vm.VMOutputApi, error)
	DecodeAddressPubkey(pk string) ([]byte, error)
	RestApiInterface() string
	RestAPIServerDebugMode() bool
	PprofEnabled() bool
	GetGenesisNodesPubKeys() (map[uint32][]string, map[uint32][]string, error)
	GetGenesisBalances() ([]*common.InitialAccountAPI, error)
	GetGasConfigs() map[string]map[string]uint64
	GetTransactionsPool() (*common.TransactionsPoolAPIResponse, error)
	IsInterfaceNil() bool
}
