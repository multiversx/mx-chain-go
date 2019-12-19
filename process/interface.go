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
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-vm-common"
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
	TotalValue() *big.Int
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
}

// InterceptorProcessor further validates and saves received data
type InterceptorProcessor interface {
	Validate(data InterceptedData) error
	Save(data InterceptedData) error
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

	CreateMarshalizedData(body block.Body) (map[uint32]block.MiniBlockSlice, map[string][][]byte)

	GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler

	VerifyCreatedBlockTransactions(body block.Body) error
	IsInterfaceNil() bool
}

// SmartContractProcessor is the main interface for the smart contract caller engine
type SmartContractProcessor interface {
	ExecuteSmartContractTransaction(tx data.TransactionHandler, acntSrc, acntDst state.AccountHandler) error
	DeploySmartContract(tx data.TransactionHandler, acntSrc state.AccountHandler) error
	IsInterfaceNil() bool
}

// IntermediateTransactionHandler handles transactions which are not resolved in only one step
type IntermediateTransactionHandler interface {
	AddIntermediateTransactions(txs []data.TransactionHandler) error
	CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock
	VerifyInterMiniBlocks(body block.Body) error
	CreateMarshalizedData(txHashes [][]byte) ([][]byte, error)
	SaveCurrentIntermediateTxToStorage() error
	GetAllCurrentFinishedTxs() map[string]data.TransactionHandler
	CreateBlockStarted()
	IsInterfaceNil() bool
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
	ProcessTransactionFee(cost *big.Int)
	IsInterfaceNil() bool
}

// SpecialAddressHandler responds with needed special addresses
type SpecialAddressHandler interface {
	SetShardConsensusData(randomness []byte, round uint64, epoch uint32, shardID uint32) error
	SetMetaConsensusData(randomness []byte, round uint64, epoch uint32) error
	ConsensusShardRewardData() *data.ConsensusRewardData
	ConsensusMetaRewardData() []*data.ConsensusRewardData
	ClearMetaConsensusData()
	ElrondCommunityAddress() []byte
	LeaderAddress() []byte
	BurnAddress() []byte
	SetElrondCommunityAddress(elrond []byte)
	ShardIdForAddress([]byte) (uint32, error)
	Epoch() uint32
	Round() uint64
	IsCurrentNodeInConsensus() bool
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

	CreateMarshalizedData(txHashes [][]byte) ([][]byte, error)

	RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int
	ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool) error
	CreateAndProcessMiniBlock(sndShardId, dstShardId uint32, spaceRemained int, haveTime func() bool) (*block.MiniBlock, error)
	CreateAndProcessMiniBlocks(maxTxSpaceRemained uint32, maxMbSpaceRemained uint32, haveTime func() bool) (block.MiniBlockSlice, error)

	GetAllCurrentUsedTxs() map[string]data.TransactionHandler
	IsInterfaceNil() bool
}

// BlockProcessor is the main interface for block execution engine
type BlockProcessor interface {
	ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountState()
	RevertStateToBlock(header data.HeaderHandler) error
	CreateNewHeader() data.HeaderHandler
	CreateBlockBody(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error
	ApplyBodyToHeader(hdr data.HeaderHandler, body data.BodyHandler) error
	ApplyProcessedMiniBlocks(processedMiniBlocks map[string]map[string]struct{})
	MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBody(dta []byte) data.BodyHandler
	DecodeBlockHeader(dta []byte) data.HeaderHandler
	AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler)
	RestoreLastNotarizedHrdsToGenesis()
	SetNumProcessedObj(numObj uint64)
	IsInterfaceNil() bool
}

// ValidatorStatisticsProcessor is the main interface for validators' consensus participation statistics
type ValidatorStatisticsProcessor interface {
	UpdatePeerState(header data.HeaderHandler) ([]byte, error)
	RevertPeerState(header data.HeaderHandler) error
	RevertPeerStateToSnapshot(snapshot int) error
	GetPeerAccount(address []byte) (state.PeerAccountHandler, error)
	IsInterfaceNil() bool
	Commit() ([]byte, error)
	RootHash() ([]byte, error)
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
	AddHeader(header data.HeaderHandler, headerHash []byte, state BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error
	RemoveHeaders(nonce uint64, hash []byte)
	CheckFork() *ForkInfo
	GetHighestFinalBlockNonce() uint64
	ProbableHighestNonce() uint64
	ResetProbableHighestNonce()
	ResetFork()
	SetForkNonce(nonce uint64)
	RestoreFinalCheckPointToGenesis()
	GetNotarizedHeaderHash(nonce uint64) []byte
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
	Revert()
	EpochStartMetaHdrHash() []byte
	IsInterfaceNil() bool
	SetFinalityAttestingRound(round uint64)
	EpochFinalityAttestingRound() uint64
}

// PendingMiniBlocksHandler is an interface to keep unfinalized miniblocks
type PendingMiniBlocksHandler interface {
	PendingMiniBlockHeaders(lastNotarizedHeaders []data.HeaderHandler) ([]block.ShardMiniBlockHeader, error)
	AddProcessedHeader(handler data.HeaderHandler) error
	RevertHeader(handler data.HeaderHandler) error
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
	RequestHeaderByNonce(shardId uint32, nonce uint64)
	RequestTransaction(shardId uint32, txHashes [][]byte)
	RequestUnsignedTransactions(destShardID uint32, scrHashes [][]byte)
	RequestRewardTransactions(destShardID uint32, txHashes [][]byte)
	RequestMiniBlock(shardId uint32, miniblockHash []byte)
	RequestHeader(shardId uint32, hash []byte)
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
	RewardsValue() *big.Int
	CommunityPercentage() float64
	LeaderPercentage() float64
	BurnPercentage() float64
	IsInterfaceNil() bool
}

// ValidatorSettingsHandler defines the functionality which is needed for validators' settings
type ValidatorSettingsHandler interface {
	UnBoundPeriod() uint64
	StakeValue() *big.Int
	IsInterfaceNil() bool
}

// FeeHandler is able to perform some economics calculation on a provided transaction
type FeeHandler interface {
	MaxGasLimitPerBlock() uint64
	ComputeGasLimit(tx TransactionWithFeeHandler) uint64
	ComputeFee(tx TransactionWithFeeHandler) *big.Int
	CheckValidityTxValues(tx TransactionWithFeeHandler) error
	IsInterfaceNil() bool
}

// TransactionWithFeeHandler represents a transaction structure that has economics variables defined
type TransactionWithFeeHandler interface {
	GetGasLimit() uint64
	GetGasPrice() uint64
	GetData() string
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
