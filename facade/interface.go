package facade

import (
	"context"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/process"
	txSimData "github.com/multiversx/mx-chain-go/process/txsimulator/data"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// NodeHandler contains all functions that a node should contain.
type NodeHandler interface {
	// GetBalance returns the balance for a specific address
	GetBalance(address string, options api.AccountQueryOptions) (*big.Int, api.BlockInfo, error)

	// GetUsername returns the username for a specific address
	GetUsername(address string, options api.AccountQueryOptions) (string, api.BlockInfo, error)

	// GetCodeHash returns the username for a specific address
	GetCodeHash(address string, options api.AccountQueryOptions) ([]byte, api.BlockInfo, error)

	// GetValueForKey returns the value of a key from a given account
	GetValueForKey(address string, key string, options api.AccountQueryOptions) (string, api.BlockInfo, error)

	// GetGuardianData returns the guardian data for given account
	GetGuardianData(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error)

	// GetKeyValuePairs returns the key-value pairs under a given address
	GetKeyValuePairs(address string, options api.AccountQueryOptions, ctx context.Context) (map[string]string, api.BlockInfo, error)

	// GetAllIssuedESDTs returns all the issued esdt tokens from esdt system smart contract
	GetAllIssuedESDTs(tokenType string, ctx context.Context) ([]string, error)

	// GetESDTData returns the esdt data from a given account, given key and given nonce
	GetESDTData(address, tokenID string, nonce uint64, options api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error)

	// GetESDTsRoles returns the token identifiers and the roles for a given address
	GetESDTsRoles(address string, options api.AccountQueryOptions, ctx context.Context) (map[string][]string, api.BlockInfo, error)

	// GetNFTTokenIDsRegisteredByAddress returns all the token identifiers for semi or non fungible tokens registered by the address
	GetNFTTokenIDsRegisteredByAddress(address string, options api.AccountQueryOptions, ctx context.Context) ([]string, api.BlockInfo, error)

	// GetESDTsWithRole returns the token identifiers where the specified address has the given role
	GetESDTsWithRole(address string, role string, options api.AccountQueryOptions, ctx context.Context) ([]string, api.BlockInfo, error)

	// GetAllESDTTokens returns the value of a key from a given account
	GetAllESDTTokens(address string, options api.AccountQueryOptions, ctx context.Context) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error)

	// GetTokenSupply returns the provided token supply from current shard
	GetTokenSupply(token string) (*api.ESDTSupply, error)

	// CreateTransaction will return a transaction from all needed fields
	CreateTransaction(txArgs *external.ArgsCreateTransaction) (*transaction.Transaction, []byte, error)

	// ValidateTransaction will validate a transaction
	ValidateTransaction(tx *transaction.Transaction) error
	ValidateTransactionForSimulation(tx *transaction.Transaction, checkSignature bool) error

	// SendBulkTransactions will send a bulk of transactions on the 'send transactions pipe' channel
	SendBulkTransactions(txs []*transaction.Transaction) (uint64, error)

	// GetAccount returns an accountResponse containing information
	//  about the account correlated with provided address
	GetAccount(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error)

	// GetCode returns the code for the given code hash
	GetCode(codeHash []byte, options api.AccountQueryOptions) ([]byte, api.BlockInfo)

	// GetHeartbeats returns the heartbeat status for each public key defined in genesis.json
	GetHeartbeats() []data.PubKeyHeartbeat

	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool

	// ValidatorStatisticsApi return the statistics for all the validators
	ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error)
	DirectTrigger(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTrigger() bool

	EncodeAddressPubkey(pk []byte) (string, error)
	DecodeAddressPubkey(pk string) ([]byte, error)

	GetQueryHandler(name string) (debug.QueryHandler, error)
	GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error)
	GetConnectedPeersRatings() string

	GetEpochStartDataAPI(epoch uint32) (*common.EpochStartDataAPI, error)

	GetProof(rootHash string, key string) (*common.GetProofResponse, error)
	GetProofDataTrie(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error)
	VerifyProof(rootHash string, address string, proof [][]byte) (bool, error)
	IsDataTrieMigrated(address string, options api.AccountQueryOptions) (bool, error)
}

// TransactionSimulatorProcessor defines the actions which a transaction simulator processor has to implement
type TransactionSimulatorProcessor interface {
	ProcessTx(tx *transaction.Transaction) (*txSimData.SimulationResults, error)
	IsInterfaceNil() bool
}

// ApiResolver defines a structure capable of resolving REST API requests
type ApiResolver interface {
	ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error)
	StatusMetrics() external.StatusMetricsHandler
	GetTotalStakedValue(ctx context.Context) (*api.StakeValues, error)
	GetDirectStakedList(ctx context.Context) ([]*api.DirectStakedValue, error)
	GetDelegatorsList(ctx context.Context) ([]*api.Delegator, error)
	GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	GetTransactionsPool(fields string) (*common.TransactionsPoolAPIResponse, error)
	GetTransactionsPoolForSender(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error)
	GetLastPoolNonceForSender(sender string) (uint64, error)
	GetTransactionsPoolNonceGapsForSender(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error)
	GetBlockByHash(hash string, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByRound(round uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetAlteredAccountsForBlock(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error)
	GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalShardBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error)
	GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error)
	GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalStartOfEpochMetaBlock(format common.ApiOutputFormat, epoch uint32) (interface{}, error)
	GetInternalStartOfEpochValidatorsInfo(epoch uint32) ([]*state.ShardValidatorInfo, error)
	GetInternalMiniBlock(format common.ApiOutputFormat, txHash string, epoch uint32) (interface{}, error)
	GetGenesisNodesPubKeys() (map[uint32][]string, map[uint32][]string)
	GetGenesisBalances() ([]*common.InitialAccountAPI, error)
	GetGasConfigs() map[string]map[string]uint64
	GetManagedKeysCount() int
	GetEligibleManagedKeys(epoch uint32) ([]string, error)
	GetWaitingManagedKeys(epoch uint32) ([]string, error)
	Close() error
	IsInterfaceNil() bool
}

// HardforkTrigger defines the structure used to trigger hardforks
type HardforkTrigger interface {
	Trigger(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTrigger() bool
	IsInterfaceNil() bool
}
