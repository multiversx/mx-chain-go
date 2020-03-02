package process

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// TransactionProcessor is the main interface for transaction execution engine
type TransactionProcessor interface {
	ProcessTransaction(transaction *transaction.Transaction) error
	IsInterfaceNil() bool
}

// RewardTransactionProcessor is the interface for reward transaction execution engine
type RewardTransactionProcessor interface {
	ProcessRewardTransaction(rewardTx *rewardTx.RewardTx) error
	IsInterfaceNil() bool
}

// RewardTransactionPreProcessor prepares the processing of reward transactions
type RewardTransactionPreProcessor interface {
	AddComputedRewardMiniBlocks(computedRewardMiniblocks block.MiniBlockSlice)
	IsInterfaceNil() bool
}

// SmartContractResultProcessor is the main interface for smart contract result execution engine
type SmartContractResultProcessor interface {
	ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) error
	IsInterfaceNil() bool
}

// TxTypeHandler is an interface to calculate the transaction type
type TxTypeHandler interface {
	ComputeTransactionType(tx data.TransactionHandler) (TransactionType, error)
	IsInterfaceNil() bool
}

// TxValidator can determine if a provided transaction handler is valid or not from the process point of view
type TxValidator interface {
	CheckTxValidity(txHandler TxValidatorHandler) error
	IsInterfaceNil() bool
}

// TxValidatorHandler defines the functionality that is needed for a TxValidator to validate a transaction
type TxValidatorHandler interface {
	SenderShardId() uint32
	Nonce() uint64
	SenderAddress() state.AddressContainer
	Fee() *big.Int
}

// HdrValidatorHandler defines the functionality that is needed for a HdrValidator to validate a header
type HdrValidatorHandler interface {
	Hash() []byte
	HeaderHandler() data.HeaderHandler
}

// HeaderValidator can determine if a provided header handler is valid or not from the process point of view
type HeaderValidator interface {
	HeaderValidForProcessing(headerHandler HdrValidatorHandler) error
	IsInterfaceNil() bool
}

// InterceptedDataFactory can create new instances of InterceptedData
type InterceptedDataFactory interface {
	Create(buff []byte) (InterceptedData, error)
	IsInterfaceNil() bool
}

// InterceptedData represents the interceptor's view of the received data
type InterceptedData interface {
	CheckValidity() error
	IsForCurrentShard() bool
	IsInterfaceNil() bool
	Hash() []byte
	Type() string
}

// InterceptorProcessor further validates and saves received data
type InterceptorProcessor interface {
	Validate(data InterceptedData) error
	Save(data InterceptedData) error
	SignalEndOfProcessing(data []InterceptedData)
	IsInterfaceNil() bool
}

// InterceptorThrottler can
type InterceptorThrottler interface {
	CanProcess() bool
	StartProcessing()
	EndProcessing()
	IsInterfaceNil() bool
}

// TransactionCoordinator is an interface to coordinate transaction processing using multiple processors
type TransactionCoordinator interface {
	RequestMiniBlocks(header data.HeaderHandler)
	RequestBlockTransactions(body block.Body)
	IsDataPreparedForProcessing(haveTime func() time.Duration) error

	SaveBlockDataToStorage(body block.Body) error
	RestoreBlockDataFromStorage(body block.Body) (int, error)
	RemoveBlockDataFromPool(body block.Body) error

	ProcessBlockTransaction(body block.Body, haveTime func() time.Duration) error

	CreateBlockStarted()
	CreateMbsAndProcessCrossShardTransactionsDstMe(header data.HeaderHandler, processedMiniBlocksHashes map[string]struct{}, maxTxSpaceRemained uint32, maxMbSpaceRemained uint32, haveTime func() bool) (block.MiniBlockSlice, uint32, bool)
	CreateMbsAndProcessTransactionsFromMe(maxTxSpaceRemained uint32, maxMbSpaceRemained uint32, haveTime func() bool) block.MiniBlockSlice

	CreateMarshalizedData(body block.Body) map[string][][]byte
	GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler

	CreateReceiptsHash() ([]byte, error)
	VerifyCreatedBlockTransactions(hdr data.HeaderHandler, body block.Body) error
	IsInterfaceNil() bool
}

// SmartContractProcessor is the main interface for the smart contract caller engine
type SmartContractProcessor interface {
	ExecuteSmartContractTransaction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) error
	DeploySmartContract(tx data.TransactionHandler, acntSrc state.UserAccountHandler) error
	IsInterfaceNil() bool
}

// IntermediateTransactionHandler handles transactions which are not resolved in only one step
type IntermediateTransactionHandler interface {
	AddIntermediateTransactions(txs []data.TransactionHandler) error
	CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock
	VerifyInterMiniBlocks(body block.Body) error
	SaveCurrentIntermediateTxToStorage() error
	GetAllCurrentFinishedTxs() map[string]data.TransactionHandler
	CreateBlockStarted()
	GetCreatedInShardMiniBlock() *block.MiniBlock
	IsInterfaceNil() bool
}

// DataMarshalizer defines the behavior of a structure that is able to marshalize containing data
type DataMarshalizer interface {
	CreateMarshalizedData(txHashes [][]byte) ([][]byte, error)
}

// InternalTransactionProducer creates system transactions (e.g. rewards)
type InternalTransactionProducer interface {
	CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock
	IsInterfaceNil() bool
}

// TransactionVerifier interface validates if the transaction is good and if it should be processed
type TransactionVerifier interface {
	IsTransactionValid(tx data.TransactionHandler) error
}

// TransactionFeeHandler processes the transaction fee
type TransactionFeeHandler interface {
	CreateBlockStarted()
	GetAccumulatedFees() *big.Int
	ProcessTransactionFee(cost *big.Int)
	IsInterfaceNil() bool
}

// PreProcessor is an interface used to prepare and process transaction data
type PreProcessor interface {
	CreateBlockStarted()
	IsDataPrepared(requestedTxs int, haveTime func() time.Duration) error

	RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error
	RestoreTxBlockIntoPools(body block.Body, miniBlockPool storage.Cacher) (int, error)
	SaveTxBlockToStorage(body block.Body) error

	ProcessBlockTransactions(body block.Body, haveTime func() bool) error
	RequestBlockTransactions(body block.Body) int

	RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int
	ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool) error
	CreateAndProcessMiniBlocks(maxTxSpaceRemained uint32, maxMbSpaceRemained uint32, haveTime func() bool) (block.MiniBlockSlice, error)

	GetAllCurrentUsedTxs() map[string]data.TransactionHandler
	IsInterfaceNil() bool
}

// BlockProcessor is the main interface for block execution engine
type BlockProcessor interface {
	ProcessBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlock(header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountState()
	PruneStateOnRollback(currHeader data.HeaderHandler, prevHeader data.HeaderHandler)
	RevertStateToBlock(header data.HeaderHandler) error
	CreateNewHeader(round uint64) data.HeaderHandler
	RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error
	CreateBlock(initialHdr data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error)
	ApplyProcessedMiniBlocks(processedMiniBlocks *processedMb.ProcessedMiniBlockTracker)
	MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBodyAndHeader(dta []byte) (data.BodyHandler, data.HeaderHandler)
	DecodeBlockBody(dta []byte) data.BodyHandler
	DecodeBlockHeader(dta []byte) data.HeaderHandler
	SetNumProcessedObj(numObj uint64)
	IsInterfaceNil() bool
}

// ValidatorStatisticsProcessor is the main interface for validators' consensus participation statistics
type ValidatorStatisticsProcessor interface {
	UpdatePeerState(header data.HeaderHandler) ([]byte, error)
	RevertPeerState(header data.HeaderHandler) error
	GetPeerAccount(address []byte) (state.PeerAccountHandler, error)
	Process(validatorInfo *state.ValidatorInfo) error
	IsInterfaceNil() bool
	RootHash() ([]byte, error)
	ResetValidatorStatisticsAtNewEpoch(vInfos map[uint32][]*state.ValidatorInfo) error
	GetValidatorInfoForRootHash(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error)
}

// Checker provides functionality to checks the integrity and validity of a data structure
type Checker interface {
	// IntegrityAndValidity does both validity and integrity checks on the data structure
	IntegrityAndValidity(coordinator sharding.Coordinator) error
	// Integrity checks only the integrity of the data
	Integrity(coordinator sharding.Coordinator) error
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// HeaderConstructionValidator provides functionality to verify header construction
type HeaderConstructionValidator interface {
	IsHeaderConstructionValid(currHdr, prevHdr data.HeaderHandler) error
	IsInterfaceNil() bool
}

// SigVerifier provides functionality to verify a signature of a signed data structure that holds also the verifying parameters
type SigVerifier interface {
	VerifySig() error
}

// SignedDataValidator provides functionality to check the validity and signature of a data structure
type SignedDataValidator interface {
	SigVerifier
	Checker
}

// HashAccesser interface provides functionality over hashable objects
type HashAccesser interface {
	SetHash([]byte)
	Hash() []byte
}

// Bootstrapper is an interface that defines the behaviour of a struct that is able
// to synchronize the node
type Bootstrapper interface {
	AddSyncStateListener(func(isSyncing bool))
	ShouldSync() bool
	StopSync()
	StartSync()
	SetStatusHandler(handler core.AppStatusHandler) error
	IsInterfaceNil() bool
}

// ForkDetector is an interface that defines the behaviour of a struct that is able
// to detect forks
type ForkDetector interface {
	AddHeader(header data.HeaderHandler, headerHash []byte, state BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error
	RemoveHeader(nonce uint64, hash []byte)
	CheckFork() *ForkInfo
	GetHighestFinalBlockNonce() uint64
	GetHighestFinalBlockHash() []byte
	ProbableHighestNonce() uint64
	ResetFork()
	SetRollBackNonce(nonce uint64)
	RestoreToGenesis()
	GetNotarizedHeaderHash(nonce uint64) []byte
	ResetProbableHighestNonce()
	IsInterfaceNil() bool
}

// InterceptorsContainer defines an interceptors holder data type with basic functionality
type InterceptorsContainer interface {
	Get(key string) (Interceptor, error)
	Add(key string, val Interceptor) error
	AddMultiple(keys []string, interceptors []Interceptor) error
	Replace(key string, val Interceptor) error
	Remove(key string)
	Len() int
	IsInterfaceNil() bool
}

// InterceptorsContainerFactory defines the functionality to create an interceptors container
type InterceptorsContainerFactory interface {
	Create() (InterceptorsContainer, error)
	IsInterfaceNil() bool
}

// PreProcessorsContainer defines an PreProcessors holder data type with basic functionality
type PreProcessorsContainer interface {
	Get(key block.Type) (PreProcessor, error)
	Add(key block.Type, val PreProcessor) error
	AddMultiple(keys []block.Type, preprocessors []PreProcessor) error
	Replace(key block.Type, val PreProcessor) error
	Remove(key block.Type)
	Len() int
	Keys() []block.Type
	IsInterfaceNil() bool
}

// PreProcessorsContainerFactory defines the functionality to create an PreProcessors container
type PreProcessorsContainerFactory interface {
	Create() (PreProcessorsContainer, error)
	IsInterfaceNil() bool
}

// IntermediateProcessorContainer defines an IntermediateProcessor holder data type with basic functionality
type IntermediateProcessorContainer interface {
	Get(key block.Type) (IntermediateTransactionHandler, error)
	Add(key block.Type, val IntermediateTransactionHandler) error
	AddMultiple(keys []block.Type, preprocessors []IntermediateTransactionHandler) error
	Replace(key block.Type, val IntermediateTransactionHandler) error
	Remove(key block.Type)
	Len() int
	Keys() []block.Type
	IsInterfaceNil() bool
}

// IntermediateProcessorsContainerFactory defines the functionality to create an IntermediateProcessors container
type IntermediateProcessorsContainerFactory interface {
	Create() (IntermediateProcessorContainer, error)
	IsInterfaceNil() bool
}

// VirtualMachinesContainer defines a virtual machine holder data type with basic functionality
type VirtualMachinesContainer interface {
	Get(key []byte) (vmcommon.VMExecutionHandler, error)
	Add(key []byte, val vmcommon.VMExecutionHandler) error
	AddMultiple(keys [][]byte, vms []vmcommon.VMExecutionHandler) error
	Replace(key []byte, val vmcommon.VMExecutionHandler) error
	Remove(key []byte)
	Len() int
	Keys() [][]byte
	IsInterfaceNil() bool
}

// VirtualMachinesContainerFactory defines the functionality to create a virtual machine container
type VirtualMachinesContainerFactory interface {
	Create() (VirtualMachinesContainer, error)
	BlockChainHookImpl() BlockChainHookHandler
	IsInterfaceNil() bool
}

// EpochStartTriggerHandler defines that actions which are needed by processor for start of epoch
type EpochStartTriggerHandler interface {
	Update(round uint64)
	ReceivedHeader(header data.HeaderHandler)
	IsEpochStart() bool
	Epoch() uint32
	EpochStartRound() uint64
	SetProcessed(header data.HeaderHandler)
	Revert(round uint64)
	EpochStartMetaHdrHash() []byte
	GetSavedStateKey() []byte
	LoadState(key []byte) error
	IsInterfaceNil() bool
	SetFinalityAttestingRound(round uint64)
	EpochFinalityAttestingRound() uint64
	RequestEpochStartIfNeeded(interceptedHeader data.HeaderHandler)
}

// EpochBootstrapper defines the actions needed by bootstrapper
type EpochBootstrapper interface {
	SetCurrentEpochStartRound(round uint64)
	IsInterfaceNil() bool
}

// PendingMiniBlocksHandler is an interface to keep unfinalized miniblocks
type PendingMiniBlocksHandler interface {
	PendingMiniBlockHeaders(lastNotarizedHeaders []data.HeaderHandler) ([]block.ShardMiniBlockHeader, error)
	AddProcessedHeader(handler data.HeaderHandler) error
	RevertHeader(handler data.HeaderHandler) error
	GetNumPendingMiniBlocks(shardID uint32) uint32
	SetNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32)
	IsInterfaceNil() bool
}

// BlockChainHookHandler defines the actions which should be performed by implementation
type BlockChainHookHandler interface {
	TemporaryAccountsHandler
	SetCurrentHeader(hdr data.HeaderHandler)
}

// Interceptor defines what a data interceptor should do
// It should also adhere to the p2p.MessageProcessor interface so it can wire to a p2p.Messenger
type Interceptor interface {
	ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error
	IsInterfaceNil() bool
}

// TopicHandler defines the functionality needed by structs to manage topics and message processors
type TopicHandler interface {
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error
	IsInterfaceNil() bool
}

// DataPacker can split a large slice of byte slices in smaller packets
type DataPacker interface {
	PackDataInChunks(data [][]byte, limit int) ([][]byte, error)
	IsInterfaceNil() bool
}

// RequestHandler defines the methods through which request to data can be made
type RequestHandler interface {
	SetEpoch(epoch uint32)
	RequestShardHeader(shardID uint32, hash []byte)
	RequestMetaHeader(hash []byte)
	RequestMetaHeaderByNonce(nonce uint64)
	RequestShardHeaderByNonce(shardID uint32, nonce uint64)
	RequestTransaction(destShardID uint32, txHashes [][]byte)
	RequestUnsignedTransactions(destShardID uint32, scrHashes [][]byte)
	RequestRewardTransactions(destShardID uint32, txHashes [][]byte)
	RequestMiniBlock(destShardID uint32, miniblockHash []byte)
	RequestMiniBlocks(destShardID uint32, miniblocksHashes [][]byte)
	RequestTrieNodes(destShardID uint32, hash []byte, topic string)
	RequestStartOfEpochMetaBlock(epoch uint32)
	IsInterfaceNil() bool
}

// ArgumentsParser defines the functionality to parse transaction data into arguments and code for smart contracts
type ArgumentsParser interface {
	GetArguments() ([][]byte, error)
	GetCode() ([]byte, error)
	GetFunction() (string, error)
	ParseData(data string) error

	CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) string
	GetStorageUpdates(data string) ([]*vmcommon.StorageUpdate, error)
	IsInterfaceNil() bool
}

// TemporaryAccountsHandler defines the functionality to create temporary accounts and pass to VM.
// This holder will contain usually one account from shard X that calls a SC in shard Y
// so when executing the code in shard Y, this impl will hold an ephemeral copy of the sender account from shard X
type TemporaryAccountsHandler interface {
	AddTempAccount(address []byte, balance *big.Int, nonce uint64)
	CleanTempAccounts()
	TempAccount(address []byte) state.AccountHandler
	IsInterfaceNil() bool
}

// BlockSizeThrottler defines the functionality of adapting the node to the network speed/latency when it should send a
// block to its peers which should be received in a limited time frame
type BlockSizeThrottler interface {
	MaxItemsToAdd() uint32
	Add(round uint64, items uint32)
	Succeed(round uint64)
	ComputeMaxItems()
	IsInterfaceNil() bool
}

// PoolsCleaner define the functionality that is needed for a pools cleaner
type PoolsCleaner interface {
	Clean(duration time.Duration) (bool, error)
	NumRemovedTxs() uint64
	IsInterfaceNil() bool
}

// RewardsHandler will return information about rewards
type RewardsHandler interface {
	LeaderPercentage() float64
	MinInflationRate() float64
	MaxInflationRate() float64
	IsInterfaceNil() bool
}

// EndOfEpochEconomics defines the functionality that is needed to compute end of epoch economics data
type EndOfEpochEconomics interface {
	ComputeEndOfEpochEconomics(metaBlock *block.MetaBlock) (*block.Economics, error)
	VerifyRewardsPerBlock(metaBlock *block.MetaBlock) error
	IsInterfaceNil() bool
}

// ValidatorSettingsHandler defines the functionality which is needed for validators' settings
type ValidatorSettingsHandler interface {
	UnBondPeriod() uint64
	GenesisNodePrice() *big.Int
	IsInterfaceNil() bool
}

// FeeHandler is able to perform some economics calculation on a provided transaction
type FeeHandler interface {
	DeveloperPercentage() float64
	MaxGasLimitPerBlock() uint64
	ComputeGasLimit(tx TransactionWithFeeHandler) uint64
	ComputeFee(tx TransactionWithFeeHandler) *big.Int
	CheckValidityTxValues(tx TransactionWithFeeHandler) error
	MinGasPrice() uint64
	IsInterfaceNil() bool
}

// TransactionWithFeeHandler represents a transaction structure that has economics variables defined
type TransactionWithFeeHandler interface {
	GetGasLimit() uint64
	GetGasPrice() uint64
	GetData() []byte
	GetRecvAddress() []byte
}

// EconomicsAddressesHandler will return information about economics addresses
type EconomicsAddressesHandler interface {
	CommunityAddress() string
	BurnAddress() string
	IsInterfaceNil() bool
}

// SmartContractToProtocolHandler is able to translate data from smart contract state into protocol changes
type SmartContractToProtocolHandler interface {
	UpdateProtocol(body block.Body, nonce uint64) error
	IsInterfaceNil() bool
}

// PeerChangesHandler will create the peer changes data for current block and will verify them
type PeerChangesHandler interface {
	PeerChanges() []block.PeerData
	VerifyPeerChanges(peerChanges []block.PeerData) error
	IsInterfaceNil() bool
}

// MiniBlocksCompacter defines the functionality that is needed for mini blocks compaction and expansion
type MiniBlocksCompacter interface {
	Compact(block.MiniBlockSlice, map[string]data.TransactionHandler) block.MiniBlockSlice
	Expand(block.MiniBlockSlice, map[string]data.TransactionHandler) (block.MiniBlockSlice, error)
	IsInterfaceNil() bool
}

// BlackListHandler can determine if a certain key is or not blacklisted
type BlackListHandler interface {
	Add(key string) error
	Has(key string) bool
	Sweep()
	IsInterfaceNil() bool
}

// NetworkConnectionWatcher defines a watchdog functionality used to specify if the current node
// is still connected to the rest of the network
type NetworkConnectionWatcher interface {
	IsConnectedToTheNetwork() bool
	IsInterfaceNil() bool
}

// SCQuery represents a prepared query for executing a function of the smart contract
type SCQuery struct {
	ScAddress []byte
	FuncName  string
	Arguments [][]byte
}

// GasHandler is able to perform some gas calculation
type GasHandler interface {
	Init()
	SetGasConsumed(gasConsumed uint64, hash []byte)
	SetGasRefunded(gasRefunded uint64, hash []byte)
	GasConsumed(hash []byte) uint64
	GasRefunded(hash []byte) uint64
	TotalGasConsumed() uint64
	TotalGasRefunded() uint64
	RemoveGasConsumed(hashes [][]byte)
	RemoveGasRefunded(hashes [][]byte)
	ComputeGasConsumedByMiniBlock(*block.MiniBlock, map[string]data.TransactionHandler) (uint64, uint64, error)
	ComputeGasConsumedByTx(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error)
	IsInterfaceNil() bool
}

// BootStorer is the interface needed by bootstrapper to read/write data in storage
type BootStorer interface {
	SaveLastRound(round int64) error
	Put(round int64, bootData bootstrapStorage.BootstrapData) error
	Get(round int64) (bootstrapStorage.BootstrapData, error)
	GetHighestRound() int64
	IsInterfaceNil() bool
}

// BootstrapperFromStorage is the interface needed by boot component to load data from storage
type BootstrapperFromStorage interface {
	LoadFromStorage() error
	GetHighestBlockNonce() uint64
	IsInterfaceNil() bool
}

// RequestBlockBodyHandler is the interface needed by process block
type RequestBlockBodyHandler interface {
	GetBlockBodyFromPool(headerHandler data.HeaderHandler) (data.BodyHandler, error)
}

// InterceptedHeaderSigVerifier is the interface needed at interceptors level to check a header if is correct
type InterceptedHeaderSigVerifier interface {
	VerifyRandSeedAndLeaderSignature(header data.HeaderHandler) error
	VerifySignature(header data.HeaderHandler) error
	IsInterfaceNil() bool
}

// BlockTracker defines the functionality for node to track the blocks which are received from network
type BlockTracker interface {
	AddCrossNotarizedHeader(shradID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte)
	AddSelfNotarizedHeader(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte)
	AddTrackedHeader(header data.HeaderHandler, hash []byte)
	CheckBlockAgainstFinal(headerHandler data.HeaderHandler) error
	CheckBlockAgainstRounder(headerHandler data.HeaderHandler) error
	CleanupHeadersBehindNonce(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64)
	CleanupInvalidCrossHeaders(metaNewEpoch uint32, metaRoundAttestingEpoch uint64)
	ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	ComputeLongestMetaChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error)
	ComputeLongestShardsChainsFromLastNotarized() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error)
	DisplayTrackedHeaders()
	GetCrossNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	GetFinalHeader(shardID uint32) (data.HeaderHandler, []byte, error)
	GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error)
	GetLastCrossNotarizedHeadersForAllShards() (map[uint32]data.HeaderHandler, error)
	GetLastSelfNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error)
	GetTrackedHeaders(shardID uint32) ([]data.HeaderHandler, [][]byte)
	GetTrackedHeadersForAllShards() map[uint32][]data.HeaderHandler
	GetTrackedHeadersWithNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
	IsShardStuck(shardID uint32) bool
	RegisterCrossNotarizedHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterSelfNotarizedHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RemoveLastNotarizedHeaders()
	RestoreToGenesis()
	IsInterfaceNil() bool
}

// EpochStartDataCreator defines the functionality for node to create epoch start data
type EpochStartDataCreator interface {
	CreateEpochStartData() (*block.EpochStart, error)
	VerifyEpochStartDataForMetablock(metaBlock *block.MetaBlock) error
	IsInterfaceNil() bool
}

// EpochStartRewardsCreator defines the functionality for the metachain to create rewards at end of epoch
type EpochStartRewardsCreator interface {
	CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error)
	VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfo) error
	CreateMarshalizedData(body block.Body) map[string][][]byte
	SaveTxBlockToStorage(metaBlock *block.MetaBlock, body block.Body)
	DeleteTxsFromStorage(metaBlock *block.MetaBlock, body block.Body)
	IsInterfaceNil() bool
}

// EpochValidatorInfoCreator defines the functionality for the metachain to create validator statistics at end of epoch
type EpochStartValidatorInfoCreator interface {
	CreateValidatorInfoMiniBlocks(validatorInfos map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error)
	VerifyValidatorInfoMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfo) error
	SaveValidatorInfoBlocksToStorage(metaBlock *block.MetaBlock, body block.Body)
	DeleteValidatorInfoBlocksFromStorage(metaBlock *block.MetaBlock)
	IsInterfaceNil() bool
}

// ValidityAttester is able to manage the valid blocks
type ValidityAttester interface {
	CheckBlockAgainstFinal(headerHandler data.HeaderHandler) error
	CheckBlockAgainstRounder(headerHandler data.HeaderHandler) error
	IsInterfaceNil() bool
}

// MiniBlocksResolver defines what a mini blocks resolver should do
type MiniBlocksResolver interface {
	GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
	GetMiniBlocksFromPool(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
	IsInterfaceNil() bool
}

// BuiltinFunction defines the methods for the built-in protocol smart contract functions
type BuiltinFunction interface {
	ProcessBuiltinFunction(
		tx data.TransactionHandler,
		acntSnd, acntDst state.UserAccountHandler,
		vmInput *vmcommon.ContractCallInput,
	) (*big.Int, error)
	GasUsed() uint64
	IsInterfaceNil() bool
}

// RoundTimeDurationHandler defines the methods to get the time duration of a round
type RoundTimeDurationHandler interface {
	TimeDuration() time.Duration
	IsInterfaceNil() bool
}

// Rounder defines the actions which should be handled by a round implementation
type Rounder interface {
	Index() int64
	IsInterfaceNil() bool
}

// Rounder defines the actions which should be handled by a round implementation
type SelectionChance interface {
	GetMaxThreshold() uint32
	GetChancePercent() uint32
}

// Ratingsinfo defines the information needed for the rating computation
type RatingsInfo interface {
	StartRating() uint32
	MaxRating() uint32
	MinRating() uint32
	ProposerIncreaseRatingStep() uint32
	ProposerDecreaseRatingStep() uint32
	ValidatorIncreaseRatingStep() uint32
	ValidatorDecreaseRatingStep() uint32
	SelectionChances() []SelectionChance
}
