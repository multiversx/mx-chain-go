package process

import (
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/guardians"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/storage"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

// TransactionProcessor is the main interface for transaction execution engine
type TransactionProcessor interface {
	ProcessTransaction(transaction *transaction.Transaction) (vmcommon.ReturnCode, error)
	VerifyTransaction(transaction *transaction.Transaction) error
	GetSenderAndReceiverAccounts(transaction *transaction.Transaction) (state.UserAccountHandler, state.UserAccountHandler, error)
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
	ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error)
	IsInterfaceNil() bool
}

// SmartContractProcessorFacade is the main interface for smart contract result execution engine
type SmartContractProcessorFacade interface {
	SmartContractProcessor
	SmartContractResultProcessor
}

// TxTypeHandler is an interface to calculate the transaction type
type TxTypeHandler interface {
	ComputeTransactionType(tx data.TransactionHandler) (TransactionType, TransactionType)
	IsInterfaceNil() bool
}

// TxValidator can determine if a provided transaction handler is valid or not from the process point of view
type TxValidator interface {
	CheckTxValidity(interceptedTx InterceptedTransactionHandler) error
	CheckTxWhiteList(data InterceptedData) error
	IsInterfaceNil() bool
}

// TxValidatorHandler defines the functionality that is needed for a TxValidator to validate a transaction
type TxValidatorHandler interface {
	SenderShardId() uint32
	ReceiverShardId() uint32
	Nonce() uint64
	SenderAddress() []byte
	Fee() *big.Int
}

// InterceptedTransactionHandler defines an intercepted data wrapper over transaction handler that has
// receiver and sender shard getters
type InterceptedTransactionHandler interface {
	SenderShardId() uint32
	ReceiverShardId() uint32
	Nonce() uint64
	SenderAddress() []byte
	Fee() *big.Int
	Transaction() data.TransactionHandler
}

// TxVersionCheckerHandler defines the functionality that is needed for a TxVersionChecker to validate transaction version
type TxVersionCheckerHandler interface {
	IsGuardedTransaction(tx *transaction.Transaction) bool
	IsSignedWithHash(tx *transaction.Transaction) bool
	CheckTxVersion(tx *transaction.Transaction) error
	IsInterfaceNil() bool
}

// HdrValidatorHandler defines the functionality that is needed for a HdrValidator to validate a header
type HdrValidatorHandler interface {
	Hash() []byte
	HeaderHandler() data.HeaderHandler
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
	Identifiers() [][]byte
	String() string
}

// InterceptorProcessor further validates and saves received data
type InterceptorProcessor interface {
	Validate(data InterceptedData, fromConnectedPeer core.PeerID) error
	Save(data InterceptedData, fromConnectedPeer core.PeerID, topic string) error
	RegisterHandler(handler func(topic string, hash []byte, data interface{}))
	IsInterfaceNil() bool
}

// InterceptorThrottler can monitor the number of the currently running interceptor go routines
type InterceptorThrottler interface {
	CanProcess() bool
	StartProcessing()
	EndProcessing()
	IsInterfaceNil() bool
}

// TransactionCoordinator is an interface to coordinate transaction processing using multiple processors
type TransactionCoordinator interface {
	RequestMiniBlocks(header data.HeaderHandler)
	RequestBlockTransactions(body *block.Body)
	IsDataPreparedForProcessing(haveTime func() time.Duration) error

	SaveTxsToStorage(body *block.Body)
	RestoreBlockDataFromStorage(body *block.Body) (int, error)
	RemoveBlockDataFromPool(body *block.Body) error
	RemoveTxsFromPool(body *block.Body) error

	ProcessBlockTransaction(header data.HeaderHandler, body *block.Body, haveTime func() time.Duration) (block.MiniBlockSlice, error)

	CreateBlockStarted()
	CreateMbsAndProcessCrossShardTransactionsDstMe(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo, haveTime func() bool, haveAdditionalTime func() bool, scheduledMode bool) (block.MiniBlockSlice, uint32, bool, error)
	CreateMbsAndProcessTransactionsFromMe(haveTime func() bool, randomness []byte) block.MiniBlockSlice
	CreatePostProcessMiniBlocks() block.MiniBlockSlice
	CreateMarshalizedData(body *block.Body) map[string][][]byte
	GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler
	GetAllCurrentLogs() []*data.LogData

	CreateReceiptsHash() ([]byte, error)
	VerifyCreatedBlockTransactions(hdr data.HeaderHandler, body *block.Body) error
	GetCreatedInShardMiniBlocks() []*block.MiniBlock
	VerifyCreatedMiniBlocks(hdr data.HeaderHandler, body *block.Body) error
	AddIntermediateTransactions(mapSCRs map[block.Type][]data.TransactionHandler) error
	GetAllIntermediateTxs() map[block.Type]map[string]data.TransactionHandler
	AddTxsFromMiniBlocks(miniBlocks block.MiniBlockSlice)
	AddTransactions(txHandlers []data.TransactionHandler, blockType block.Type)
	IsInterfaceNil() bool
}

// SmartContractProcessor is the main interface for the smart contract caller engine
type SmartContractProcessor interface {
	ExecuteSmartContractTransaction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error)
	ExecuteBuiltInFunction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error)
	DeploySmartContract(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error)
	ProcessIfError(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error
	IsPayable(sndAddress []byte, recvAddress []byte) (bool, error)
	CheckBuiltinFunctionIsExecutable(expectedBuiltinFunction string, tx data.TransactionHandler) error
	IsInterfaceNil() bool
}

// IntermediateTransactionHandler handles transactions which are not resolved in only one step
type IntermediateTransactionHandler interface {
	AddIntermediateTransactions(txs []data.TransactionHandler) error
	GetNumOfCrossInterMbsAndTxs() (int, int)
	CreateAllInterMiniBlocks() []*block.MiniBlock
	VerifyInterMiniBlocks(body *block.Body) error
	SaveCurrentIntermediateTxToStorage()
	GetAllCurrentFinishedTxs() map[string]data.TransactionHandler
	CreateBlockStarted()
	GetCreatedInShardMiniBlock() *block.MiniBlock
	RemoveProcessedResults(key []byte) [][]byte
	InitProcessedResults(key []byte)
	IsInterfaceNil() bool
}

// DataMarshalizer defines the behavior of a structure that is able to marshalize containing data
type DataMarshalizer interface {
	CreateMarshalledData(txHashes [][]byte) ([][]byte, error)
}

// TransactionVerifier interface validates if the transaction is good and if it should be processed
type TransactionVerifier interface {
	IsTransactionValid(tx data.TransactionHandler) error
}

// TransactionFeeHandler processes the transaction fee
type TransactionFeeHandler interface {
	CreateBlockStarted(gasAndFees scheduled.GasAndFees)
	GetAccumulatedFees() *big.Int
	GetDeveloperFees() *big.Int
	ProcessTransactionFee(cost *big.Int, devFee *big.Int, txHash []byte)
	ProcessTransactionFeeRelayedUserTx(cost *big.Int, devFee *big.Int, userTxHash []byte, originalTxHash []byte)
	RevertFees(txHashes [][]byte)
	IsInterfaceNil() bool
}

// PreProcessor is an interface used to prepare and process transaction data
type PreProcessor interface {
	CreateBlockStarted()
	IsDataPrepared(requestedTxs int, haveTime func() time.Duration) error

	RemoveBlockDataFromPools(body *block.Body, miniBlockPool storage.Cacher) error
	RemoveTxsFromPools(body *block.Body) error
	RestoreBlockDataIntoPools(body *block.Body, miniBlockPool storage.Cacher) (int, error)
	SaveTxsToStorage(body *block.Body) error

	ProcessBlockTransactions(header data.HeaderHandler, body *block.Body, haveTime func() bool) (block.MiniBlockSlice, error)
	RequestBlockTransactions(body *block.Body) int

	RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int
	ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, haveAdditionalTime func() bool, scheduledMode bool, partialMbExecutionMode bool, indexOfLastTxProcessed int, preProcessorExecutionInfoHandler PreProcessorExecutionInfoHandler) ([][]byte, int, bool, error)
	CreateAndProcessMiniBlocks(haveTime func() bool, randomness []byte) (block.MiniBlockSlice, error)

	GetAllCurrentUsedTxs() map[string]data.TransactionHandler
	AddTxsFromMiniBlocks(miniBlocks block.MiniBlockSlice)
	AddTransactions(txHandlers []data.TransactionHandler)
	IsInterfaceNil() bool
}

// BlockProcessor is the main interface for block execution engine
type BlockProcessor interface {
	ProcessBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error)
	ProcessScheduledBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlock(header data.HeaderHandler, body data.BodyHandler) error
	RevertCurrentBlock()
	PruneStateOnRollback(currHeader data.HeaderHandler, currHeaderHash []byte, prevHeader data.HeaderHandler, prevHeaderHash []byte)
	RevertStateToBlock(header data.HeaderHandler, rootHash []byte) error
	CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error)
	RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error
	CreateBlock(initialHdr data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error)
	MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBody(dta []byte) data.BodyHandler
	DecodeBlockHeader(dta []byte) data.HeaderHandler
	SetNumProcessedObj(numObj uint64)
	RestoreBlockBodyIntoPools(body data.BodyHandler) error
	NonceOfFirstCommittedBlock() core.OptionalUint64
	Close() error
	IsInterfaceNil() bool
}

// SmartContractProcessorFull is the main interface for smart contract result execution engine
type SmartContractProcessorFull interface {
	SmartContractProcessor
	SmartContractResultProcessor
}

// ScheduledBlockProcessor is the interface for the scheduled miniBlocks execution part of the block processor
type ScheduledBlockProcessor interface {
	ProcessScheduledBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	IsInterfaceNil() bool
}

// ValidatorStatisticsProcessor is the main interface for validators' consensus participation statistics
type ValidatorStatisticsProcessor interface {
	UpdatePeerState(header data.CommonHeaderHandler, cache map[string]data.CommonHeaderHandler) ([]byte, error)
	Process(shardValidatorInfo data.ShardValidatorInfoHandler) error
	IsInterfaceNil() bool
	RootHash() ([]byte, error)
	ResetValidatorStatisticsAtNewEpoch(vInfos map[uint32][]*state.ValidatorInfo) error
	GetValidatorInfoForRootHash(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error)
	ProcessRatingsEndOfEpoch(validatorInfos map[uint32][]*state.ValidatorInfo, epoch uint32) error
	Commit() ([]byte, error)
	DisplayRatings(epoch uint32)
	SetLastFinalizedRootHash([]byte)
	LastFinalizedRootHash() []byte
	PeerAccountToValidatorInfo(peerAccount state.PeerAccountHandler) *state.ValidatorInfo
	SaveNodesCoordinatorUpdates(epoch uint32) (bool, error)
}

// TransactionLogProcessor is the main interface for saving logs generated by smart contract calls
type TransactionLogProcessor interface {
	GetAllCurrentLogs() []*data.LogData
	GetLog(txHash []byte) (data.LogHandler, error)
	SaveLog(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error
	Clean()
	IsInterfaceNil() bool
}

// TransactionLogProcessorDatabase is interface the  for saving logs also in RAM
type TransactionLogProcessorDatabase interface {
	GetLogFromCache(txHash []byte) (*data.LogData, bool)
	EnableLogToBeSavedInCache()
	Clean()
	IsInterfaceNil() bool
}

// ValidatorsProvider is the main interface for validators' provider
type ValidatorsProvider interface {
	GetLatestValidators() map[string]*accounts.ValidatorApiResponse
	IsInterfaceNil() bool
	Close() error
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
	Close() error
	AddSyncStateListener(func(isSyncing bool))
	GetNodeState() common.NodeState
	StartSyncingBlocks() error
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
	SetFinalToLastCheckpoint()
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
	Iterate(handler func(key string, interceptor Interceptor) bool)
	Close() error
	IsInterfaceNil() bool
}

// InterceptorsContainerFactory defines the functionality to create an interceptors container
type InterceptorsContainerFactory interface {
	Create() (InterceptorsContainer, InterceptorsContainer, error)
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
	Close() error
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
	Close() error
	BlockChainHookImpl() BlockChainHookHandler
	IsInterfaceNil() bool
}

// EpochStartTriggerHandler defines that actions which are needed by processor for start of epoch
type EpochStartTriggerHandler interface {
	Update(round uint64, nonce uint64)
	IsEpochStart() bool
	Epoch() uint32
	MetaEpoch() uint32
	EpochStartRound() uint64
	SetProcessed(header data.HeaderHandler, body data.BodyHandler)
	RevertStateToBlock(header data.HeaderHandler) error
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
	AddProcessedHeader(handler data.HeaderHandler) error
	RevertHeader(handler data.HeaderHandler) error
	GetPendingMiniBlocks(shardID uint32) [][]byte
	SetPendingMiniBlocks(shardID uint32, mbHashes [][]byte)
	IsInterfaceNil() bool
}

// BlockChainHookHandler defines the actions which should be performed by implementation
type BlockChainHookHandler interface {
	GetCode(account vmcommon.UserAccountHandler) []byte
	GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error)
	GetStorageData(accountAddress []byte, index []byte) ([]byte, uint32, error)
	GetBlockhash(nonce uint64) ([]byte, error)
	LastNonce() uint64
	LastRound() uint64
	LastTimeStamp() uint64
	LastRandomSeed() []byte
	LastEpoch() uint32
	GetStateRootHash() []byte
	CurrentNonce() uint64
	CurrentRound() uint64
	CurrentTimeStamp() uint64
	CurrentRandomSeed() []byte
	CurrentEpoch() uint32
	NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error)
	ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	SaveNFTMetaDataToSystemAccount(tx data.TransactionHandler) error
	GetShardOfAddress(address []byte) uint32
	IsSmartContract(address []byte) bool
	GetBuiltinFunctionNames() vmcommon.FunctionNames
	GetBuiltinFunctionsContainer() vmcommon.BuiltInFunctionContainer
	GetAllState(_ []byte) (map[string][]byte, error)
	GetESDTToken(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error)
	IsPaused(tokenID []byte) bool
	IsLimitedTransfer(tokenID []byte) bool
	NumberOfShards() uint32
	SetCurrentHeader(hdr data.HeaderHandler)
	SaveCompiledCode(codeHash []byte, code []byte)
	GetCompiledCode(codeHash []byte) (bool, []byte)
	IsPayable(sndAddress []byte, recvAddress []byte) (bool, error)
	DeleteCompiledCode(codeHash []byte)
	ClearCompiledCodes()
	GetSnapshot() int
	RevertToSnapshot(snapshot int) error
	ExecuteSmartContractCallOnOtherVM(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	SetVMContainer(vmContainer VirtualMachinesContainer) error
	Close() error
	FilterCodeMetadataForUpgrade(input []byte) ([]byte, error)
	ApplyFiltersOnSCCodeMetadata(codeMetadata vmcommon.CodeMetadata) vmcommon.CodeMetadata
	ResetCounters()
	GetCounterValues() map[string]uint64
	IsInterfaceNil() bool
	IsBuiltinFunctionName(functionName string) bool
}

// Interceptor defines what a data interceptor should do
// It should also adhere to the p2p.MessageProcessor interface so it can wire to a p2p.Messenger
type Interceptor interface {
	ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error
	SetInterceptedDebugHandler(handler InterceptedDebugger) error
	RegisterHandler(handler func(topic string, hash []byte, data interface{}))
	Close() error
	IsInterfaceNil() bool
}

// TopicHandler defines the functionality needed by structs to manage topics and message processors
type TopicHandler interface {
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error
	ID() core.PeerID
	IsInterfaceNil() bool
}

// SignaturesHandler defines the behavior of a struct able to handle signatures
type SignaturesHandler interface {
	Verify(payload []byte, pid core.PeerID, signature []byte) error
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
	RequestTrieNodes(destShardID uint32, hashes [][]byte, topic string)
	RequestStartOfEpochMetaBlock(epoch uint32)
	RequestInterval() time.Duration
	SetNumPeersToQuery(key string, intra int, cross int) error
	GetNumPeersToQuery(key string) (int, int, error)
	RequestTrieNode(requestHash []byte, topic string, chunkIndex uint32)
	CreateTrieNodeIdentifier(requestHash []byte, chunkIndex uint32) []byte
	RequestPeerAuthenticationsByHashes(destShardID uint32, hashes [][]byte)
	RequestValidatorInfo(hash []byte)
	RequestValidatorsInfo(hashes [][]byte)
	IsInterfaceNil() bool
}

// CallArgumentsParser defines the functionality to parse transaction data into call arguments
type CallArgumentsParser interface {
	ParseData(data string) (string, [][]byte, error)
	ParseArguments(data string) ([][]byte, error)
	IsInterfaceNil() bool
}

// DeployArgumentsParser defines the functionality to parse transaction data into call arguments
type DeployArgumentsParser interface {
	ParseData(data string) (*parsers.DeployArgs, error)
	IsInterfaceNil() bool
}

// StorageArgumentsParser defines the functionality to parse transaction data into call arguments
type StorageArgumentsParser interface {
	CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) string
	GetStorageUpdates(data string) ([]*vmcommon.StorageUpdate, error)
	IsInterfaceNil() bool
}

// ArgumentsParser defines the functionality to parse transaction data into arguments and code for smart contracts
type ArgumentsParser interface {
	ParseCallData(data string) (string, [][]byte, error)
	ParseArguments(data string) ([][]byte, error)
	ParseDeployData(data string) (*parsers.DeployArgs, error)

	CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) string
	GetStorageUpdates(data string) ([]*vmcommon.StorageUpdate, error)
	IsInterfaceNil() bool
}

// BlockSizeThrottler defines the functionality of adapting the node to the network speed/latency when it should send a
// block to its peers which should be received in a limited time frame
type BlockSizeThrottler interface {
	GetCurrentMaxSize() uint32
	Add(round uint64, size uint32)
	Succeed(round uint64)
	ComputeCurrentMaxSize()
	IsInterfaceNil() bool
}

type rewardsHandler interface {
	LeaderPercentage() float64
	ProtocolSustainabilityPercentage() float64
	ProtocolSustainabilityAddress() string
	MinInflationRate() float64
	MaxInflationRate(year uint32) float64
	RewardsTopUpGradientPoint() *big.Int
	RewardsTopUpFactor() float64
}

// RewardsHandler will return information about rewards
type RewardsHandler interface {
	rewardsHandler
	IsInterfaceNil() bool
}

// EndOfEpochEconomics defines the functionality that is needed to compute end of epoch economics data
type EndOfEpochEconomics interface {
	ComputeEndOfEpochEconomics(metaBlock *block.MetaBlock) (*block.Economics, error)
	VerifyRewardsPerBlock(metaBlock *block.MetaBlock, correctedProtocolSustainability *big.Int, computedEconomics *block.Economics) error
	IsInterfaceNil() bool
}

type feeHandler interface {
	GenesisTotalSupply() *big.Int
	DeveloperPercentage() float64
	GasPerDataByte() uint64
	MaxGasLimitPerBlock(shardID uint32) uint64
	MaxGasLimitPerMiniBlock(shardID uint32) uint64
	MaxGasLimitPerBlockForSafeCrossShard() uint64
	MaxGasLimitPerMiniBlockForSafeCrossShard() uint64
	MaxGasLimitPerTx() uint64
	ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64
	ComputeMoveBalanceFee(tx data.TransactionWithFeeHandler) *big.Int
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValues(tx data.TransactionWithFeeHandler) error
	ComputeFeeForProcessing(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int
	MinGasPrice() uint64
	MaxGasPriceSetGuardian() uint64
	GasPriceModifier() float64
	MinGasLimit() uint64
	ExtraGasLimitGuardedTx() uint64
	SplitTxGasInCategories(tx data.TransactionWithFeeHandler) (uint64, uint64)
	GasPriceForProcessing(tx data.TransactionWithFeeHandler) uint64
	GasPriceForMove(tx data.TransactionWithFeeHandler) uint64
	MinGasPriceForProcessing() uint64
	ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimitBasedOnBalance(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error)
}

// TxGasHandler handles a transaction gas and gas cost
type TxGasHandler interface {
	SplitTxGasInCategories(tx data.TransactionWithFeeHandler) (uint64, uint64)
	GasPriceForProcessing(tx data.TransactionWithFeeHandler) uint64
	GasPriceForMove(tx data.TransactionWithFeeHandler) uint64
	MinGasPrice() uint64
	ComputeFeeForProcessing(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int
	GasPriceModifier() float64
	MinGasLimit() uint64
	MinGasPriceForProcessing() uint64
	IsInterfaceNil() bool
}

// FeeHandler is able to perform some economics calculation on a provided transaction
type FeeHandler interface {
	feeHandler
	IsInterfaceNil() bool
}

// EconomicsDataHandler provides some economics related computation and read access to economics data
type EconomicsDataHandler interface {
	rewardsHandler
	feeHandler
	SetStatusHandler(statusHandler core.AppStatusHandler) error
	IsInterfaceNil() bool
}

// SmartContractToProtocolHandler is able to translate data from smart contract state into protocol changes
type SmartContractToProtocolHandler interface {
	UpdateProtocol(body *block.Body, nonce uint64) error
	IsInterfaceNil() bool
}

// PeerChangesHandler will create the peer changes data for current block and will verify them
type PeerChangesHandler interface {
	PeerChanges() []block.PeerData
	VerifyPeerChanges(peerChanges []block.PeerData) error
	IsInterfaceNil() bool
}

// TimeCacher defines the cache that can keep a record for a bounded time
type TimeCacher interface {
	Add(key string) error
	Upsert(key string, span time.Duration) error
	Has(key string) bool
	Sweep()
	Len() int
	IsInterfaceNil() bool
}

// PeerBlackListCacher can determine if a certain peer id is or not blacklisted
type PeerBlackListCacher interface {
	Upsert(pid core.PeerID, span time.Duration) error
	Has(pid core.PeerID) bool
	Sweep()
	IsInterfaceNil() bool
}

// PeerShardMapper can return the public key of a provided peer ID
type PeerShardMapper interface {
	UpdatePeerIDPublicKeyPair(pid core.PeerID, pk []byte)
	PutPeerIdShardId(pid core.PeerID, shardID uint32)
	PutPeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType)
	GetPeerInfo(pid core.PeerID) core.P2PPeerInfo
	IsInterfaceNil() bool
}

// NetworkShardingCollector defines the updating methods used by the network sharding component
type NetworkShardingCollector interface {
	PeerShardMapper
	UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32)
}

// NetworkConnectionWatcher defines a watchdog functionality used to specify if the current node
// is still connected to the rest of the network
type NetworkConnectionWatcher interface {
	IsConnectedToTheNetwork() bool
	IsInterfaceNil() bool
}

// SCQuery represents a prepared query for executing a function of the smart contract
type SCQuery struct {
	ScAddress      []byte
	FuncName       string
	CallerAddr     []byte
	CallValue      *big.Int
	Arguments      [][]byte
	SameScState    bool
	ShouldBeSynced bool
}

// GasHandler is able to perform some gas calculation
type GasHandler interface {
	Init()
	Reset(key []byte)
	SetGasProvided(gasProvided uint64, hash []byte)
	SetGasProvidedAsScheduled(gasProvided uint64, hash []byte)
	SetGasRefunded(gasRefunded uint64, hash []byte)
	SetGasPenalized(gasPenalized uint64, hash []byte)
	GasProvided(hash []byte) uint64
	GasProvidedAsScheduled(hash []byte) uint64
	GasRefunded(hash []byte) uint64
	GasPenalized(hash []byte) uint64
	TotalGasProvided() uint64
	TotalGasProvidedAsScheduled() uint64
	TotalGasProvidedWithScheduled() uint64
	TotalGasRefunded() uint64
	TotalGasPenalized() uint64
	RemoveGasProvided(hashes [][]byte)
	RemoveGasProvidedAsScheduled(hashes [][]byte)
	RemoveGasRefunded(hashes [][]byte)
	RemoveGasPenalized(hashes [][]byte)
	RestoreGasSinceLastReset(key []byte)
	ComputeGasProvidedByMiniBlock(*block.MiniBlock, map[string]data.TransactionHandler) (uint64, uint64, error)
	ComputeGasProvidedByTx(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error)
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

// InterceptedHeaderSigVerifier is the interface needed at interceptors level to check that a header's signature is correct
type InterceptedHeaderSigVerifier interface {
	VerifyRandSeedAndLeaderSignature(header data.HeaderHandler) error
	VerifyRandSeed(header data.HeaderHandler) error
	VerifyLeaderSignature(header data.HeaderHandler) error
	VerifySignature(header data.HeaderHandler) error
	IsInterfaceNil() bool
}

// HeaderIntegrityVerifier encapsulates methods useful to check that a header's integrity is correct
type HeaderIntegrityVerifier interface {
	Verify(header data.HeaderHandler) error
	GetVersion(epoch uint32) string
	IsInterfaceNil() bool
}

// BlockTracker defines the functionality for node to track the blocks which are received from network
type BlockTracker interface {
	AddCrossNotarizedHeader(shradID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte)
	AddSelfNotarizedHeader(shardID uint32, selfNotarizedHeader data.HeaderHandler, selfNotarizedHeaderHash []byte)
	AddTrackedHeader(header data.HeaderHandler, hash []byte)
	CheckBlockAgainstFinal(headerHandler data.HeaderHandler) error
	CheckBlockAgainstRoundHandler(headerHandler data.HeaderHandler) error
	CheckBlockAgainstWhitelist(interceptedData InterceptedData) bool
	CleanupHeadersBehindNonce(shardID uint32, selfNotarizedNonce uint64, crossNotarizedNonce uint64)
	CleanupInvalidCrossHeaders(metaNewEpoch uint32, metaRoundAttestingEpoch uint64)
	ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	ComputeLongestMetaChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error)
	ComputeLongestShardsChainsFromLastNotarized() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error)
	DisplayTrackedHeaders()
	GetCrossNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error)
	GetLastCrossNotarizedHeadersForAllShards() (map[uint32]data.HeaderHandler, error)
	GetLastSelfNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error)
	GetSelfNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error)
	GetTrackedHeaders(shardID uint32) ([]data.HeaderHandler, [][]byte)
	GetTrackedHeadersForAllShards() map[uint32][]data.HeaderHandler
	GetTrackedHeadersWithNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
	IsShardStuck(shardID uint32) bool
	ShouldSkipMiniBlocksCreationFromSelf() bool
	RegisterCrossNotarizedHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterSelfNotarizedFromCrossHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterSelfNotarizedHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RegisterFinalMetachainHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	RemoveLastNotarizedHeaders()
	RestoreToGenesis()
	ShouldAddHeader(headerHandler data.HeaderHandler) bool
	IsInterfaceNil() bool
}

// FloodPreventer defines the behavior of a component that is able to signal that too many events occurred
// on a provided identifier between Reset calls
type FloodPreventer interface {
	IncreaseLoad(pid core.PeerID, size uint64) error
	ApplyConsensusSize(size int)
	Reset()
	IsInterfaceNil() bool
}

// TopicFloodPreventer defines the behavior of a component that is able to signal that too many events occurred
// on a provided identifier between Reset calls, on a given topic
type TopicFloodPreventer interface {
	IncreaseLoad(pid core.PeerID, topic string, numMessages uint32) error
	ResetForTopic(topic string)
	ResetForNotRegisteredTopics()
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopic(pid core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
	ApplyConsensusSize(size int)
	SetDebugger(debugger AntifloodDebugger) error
	BlacklistPeer(peer core.PeerID, reason string, duration time.Duration)
	IsOriginatorEligibleForTopic(pid core.PeerID, topic string) error
	IsInterfaceNil() bool
	Close() error
}

// PeerValidatorMapper can determine the peer info from a peer id
type PeerValidatorMapper interface {
	GetPeerInfo(pid core.PeerID) core.P2PPeerInfo
	IsInterfaceNil() bool
}

// SCQueryService defines how data should be get from a SC account
type SCQueryService interface {
	ExecuteQuery(query *SCQuery) (*vmcommon.VMOutput, error)
	ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error)
	Close() error
	IsInterfaceNil() bool
}

// EpochStartDataCreator defines the functionality for node to create epoch start data
type EpochStartDataCreator interface {
	CreateEpochStartData() (*block.EpochStart, error)
	VerifyEpochStartDataForMetablock(metaBlock *block.MetaBlock) error
	IsInterfaceNil() bool
}

// RewardsCreator defines the functionality for the metachain to create rewards at end of epoch
type RewardsCreator interface {
	CreateRewardsMiniBlocks(
		metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
	) (block.MiniBlockSlice, error)
	VerifyRewardsMiniBlocks(
		metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
	) error
	GetProtocolSustainabilityRewards() *big.Int
	GetLocalTxCache() epochStart.TransactionCacher
	CreateMarshalledData(body *block.Body) map[string][][]byte
	GetRewardsTxs(body *block.Body) map[string]data.TransactionHandler
	SaveBlockDataToStorage(metaBlock data.MetaHeaderHandler, body *block.Body)
	DeleteBlockDataFromStorage(metaBlock data.MetaHeaderHandler, body *block.Body)
	RemoveBlockDataFromPools(metaBlock data.MetaHeaderHandler, body *block.Body)
	IsInterfaceNil() bool
}

// EpochStartValidatorInfoCreator defines the functionality for the metachain to create validator statistics at end of epoch
type EpochStartValidatorInfoCreator interface {
	CreateValidatorInfoMiniBlocks(validatorInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error)
	VerifyValidatorInfoMiniBlocks(miniBlocks []*block.MiniBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error
	GetLocalValidatorInfoCache() epochStart.ValidatorInfoCacher
	CreateMarshalledData(body *block.Body) map[string][][]byte
	GetValidatorInfoTxs(body *block.Body) map[string]*state.ShardValidatorInfo
	SaveBlockDataToStorage(metaBlock data.HeaderHandler, body *block.Body)
	DeleteBlockDataFromStorage(metaBlock data.HeaderHandler, body *block.Body)
	RemoveBlockDataFromPools(metaBlock data.HeaderHandler, body *block.Body)
	IsInterfaceNil() bool
}

// EpochStartSystemSCProcessor defines the functionality for the metachain to process system smart contract and end of epoch
type EpochStartSystemSCProcessor interface {
	ProcessSystemSmartContract(validatorInfos map[uint32][]*state.ValidatorInfo, nonce uint64, epoch uint32) error
	ProcessDelegationRewards(
		miniBlocks block.MiniBlockSlice,
		rewardTxs epochStart.TransactionCacher,
	) error
	ToggleUnStakeUnBond(value bool) error
	IsInterfaceNil() bool
}

// ValidityAttester is able to manage the valid blocks
type ValidityAttester interface {
	CheckBlockAgainstFinal(headerHandler data.HeaderHandler) error
	CheckBlockAgainstRoundHandler(headerHandler data.HeaderHandler) error
	CheckBlockAgainstWhitelist(interceptedData InterceptedData) bool
	IsInterfaceNil() bool
}

// MiniBlockProvider defines what a miniblock data provider should do
type MiniBlockProvider interface {
	GetMiniBlocks(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
	GetMiniBlocksFromPool(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
	GetMiniBlocksFromStorer(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
	IsInterfaceNil() bool
}

// RoundTimeDurationHandler defines the methods to get the time duration of a round
type RoundTimeDurationHandler interface {
	TimeDuration() time.Duration
	IsInterfaceNil() bool
}

// RoundHandler defines the actions which should be handled by a round implementation
type RoundHandler interface {
	Index() int64
	IsInterfaceNil() bool
}

// SelectionChance defines the actions which should be handled by a round implementation
type SelectionChance interface {
	GetMaxThreshold() uint32
	GetChancePercent() uint32
}

// RatingsInfoHandler defines the information needed for the rating computation
type RatingsInfoHandler interface {
	StartRating() uint32
	MaxRating() uint32
	MinRating() uint32
	SignedBlocksThreshold() float32
	MetaChainRatingsStepHandler() RatingsStepHandler
	ShardChainRatingsStepHandler() RatingsStepHandler
	SelectionChances() []SelectionChance
	IsInterfaceNil() bool
}

// RatingsStepHandler defines the information needed for the rating computation on shards or meta
type RatingsStepHandler interface {
	ProposerIncreaseRatingStep() int32
	ProposerDecreaseRatingStep() int32
	ValidatorIncreaseRatingStep() int32
	ValidatorDecreaseRatingStep() int32
	ConsecutiveMissedBlocksPenalty() float32
}

// ValidatorInfoSyncer defines the method needed for validatorInfoProcessing
type ValidatorInfoSyncer interface {
	SyncMiniBlocks(headerHandler data.HeaderHandler) ([][]byte, data.BodyHandler, error)
	SyncValidatorsInfo(bodyHandler data.BodyHandler) ([][]byte, map[string]*state.ShardValidatorInfo, error)
	IsInterfaceNil() bool
}

// RatingChanceHandler provides the methods needed for the computation of chances from the Rating
type RatingChanceHandler interface {
	// GetMaxThreshold returns the threshold until this ChancePercentage holds
	GetMaxThreshold() uint32
	// GetChancePercentage returns the percentage for the RatingChanceHandler
	GetChancePercentage() uint32
	// IsInterfaceNil verifies if the interface is nil
	IsInterfaceNil() bool
}

// WhiteListHandler is the interface needed to add whitelisted data
type WhiteListHandler interface {
	Remove(keys [][]byte)
	Add(keys [][]byte)
	IsWhiteListed(interceptedData InterceptedData) bool
	IsWhiteListedAtLeastOne(identifiers [][]byte) bool
	IsInterfaceNil() bool
}

// InterceptedDebugger defines an interface for debugging the intercepted data
type InterceptedDebugger interface {
	LogReceivedHashes(topic string, hashes [][]byte)
	LogProcessedHashes(topic string, hashes [][]byte, err error)
	IsInterfaceNil() bool
}

// PreferredPeersHolderHandler defines the behavior of a component able to handle preferred peers operations
type PreferredPeersHolderHandler interface {
	Get() map[uint32][]core.PeerID
	Contains(peerID core.PeerID) bool
	IsInterfaceNil() bool
}

// AntifloodDebugger defines an interface for debugging the antiflood behavior
type AntifloodDebugger interface {
	AddData(pid core.PeerID, topic string, numRejected uint32, sizeRejected uint64, sequence []byte, isBlacklisted bool)
	Close() error
	IsInterfaceNil() bool
}

// PoolsCleaner defines the functionality to clean pools for old records
type PoolsCleaner interface {
	Close() error
	StartCleaning()
	IsInterfaceNil() bool
}

// EpochHandler defines what a component which handles current epoch should be able to do
type EpochHandler interface {
	MetaEpoch() uint32
	IsInterfaceNil() bool
}

// EpochStartEventNotifier provides Register and Unregister functionality for the end of epoch events
type EpochStartEventNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	IsInterfaceNil() bool
}

// NodesCoordinator provides Validator methods needed for the peer processing
type NodesCoordinator interface {
	GetValidatorWithPublicKey(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error)
	GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	GetAllLeavingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	IsInterfaceNil() bool
}

// EpochNotifier can notify upon an epoch change and provide the current epoch
type EpochNotifier interface {
	// TODO RoundSubscriberHandler should be move to elrond-core
	RegisterNotifyHandler(handler vmcommon.EpochSubscriberHandler)
	CurrentEpoch() uint32
	CheckEpoch(header data.HeaderHandler)
	IsInterfaceNil() bool
}

// RoundNotifier can notify upon an epoch change and provide the current epoch
type RoundNotifier interface {
	RegisterNotifyHandler(handler vmcommon.RoundSubscriberHandler)
	CurrentRound() uint64
	CheckRound(header data.HeaderHandler)
	IsInterfaceNil() bool
}

// EnableRoundsHandler is an interface which can be queried to check for round activation features/fixes
type EnableRoundsHandler interface {
	RoundConfirmed(round uint64, timestamp uint64)
	IsDisableAsyncCallV1Enabled() bool
	IsInterfaceNil() bool
}

// ESDTPauseHandler provides IsPaused function for an ESDT token
type ESDTPauseHandler interface {
	IsPaused(token []byte) bool
	IsInterfaceNil() bool
}

// ESDTRoleHandler provides IsAllowedToExecute function for an ESDT
type ESDTRoleHandler interface {
	CheckAllowedToExecute(account state.UserAccountHandler, tokenID []byte, action []byte) error
	IsInterfaceNil() bool
}

// PayableHandler provides IsPayable function which returns if an account is payable or not
type PayableHandler interface {
	IsPayable(sndAddress []byte, recvAddress []byte) (bool, error)
	IsInterfaceNil() bool
}

// FallbackHeaderValidator defines the behaviour of a component able to signal when a fallback header validation could be applied
type FallbackHeaderValidator interface {
	ShouldApplyFallbackValidation(headerHandler data.HeaderHandler) bool
	IsInterfaceNil() bool
}

// CoreComponentsHolder holds the core components needed by the interceptors
type CoreComponentsHolder interface {
	InternalMarshalizer() marshal.Marshalizer
	SetInternalMarshalizer(marshalizer marshal.Marshalizer) error
	TxMarshalizer() marshal.Marshalizer
	Hasher() hashing.Hasher
	TxSignHasher() hashing.Hasher
	Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter
	AddressPubKeyConverter() core.PubkeyConverter
	ValidatorPubKeyConverter() core.PubkeyConverter
	PathHandler() storage.PathManagerHandler
	ChainID() string
	MinTransactionVersion() uint32
	TxVersionChecker() TxVersionCheckerHandler
	GenesisNodesSetup() sharding.GenesisNodesSetupHandler
	EpochNotifier() EpochNotifier
	ChanStopNodeProcess() chan endProcess.ArgEndProcess
	NodeTypeProvider() core.NodeTypeProviderHandler
	ProcessStatusHandler() common.ProcessStatusHandler
	HardforkTriggerPubKey() []byte
	EnableEpochsHandler() common.EnableEpochsHandler
	IsInterfaceNil() bool
}

// CryptoComponentsHolder holds the crypto components needed by the interceptors
type CryptoComponentsHolder interface {
	TxSignKeyGen() crypto.KeyGenerator
	BlockSignKeyGen() crypto.KeyGenerator
	TxSingleSigner() crypto.SingleSigner
	BlockSigner() crypto.SingleSigner
	GetMultiSigner(epoch uint32) (crypto.MultiSigner, error)
	MultiSignerContainer() cryptoCommon.MultiSignerContainer
	SetMultiSignerContainer(ms cryptoCommon.MultiSignerContainer) error
	PeerSignatureHandler() crypto.PeerSignatureHandler
	PublicKey() crypto.PublicKey
	PrivateKey() crypto.PrivateKey
	Clone() interface{}
	ManagedPeersHolder() common.ManagedPeersHolder
	IsInterfaceNil() bool
}

// AdditionalStorageServiceCreator defines the actions needed for a component that can create additional storage services
type AdditionalStorageServiceCreator interface {
	CreateAdditionalStorageService(func(store dataRetriever.StorageService, shardID string) error, dataRetriever.StorageService, string) error
	IsInterfaceNil() bool
}

// RunTypeComponentsHolder holds the runType components needed by the interceptors
type RunTypeComponentsHolder interface {
	AdditionalStorageServiceCreator() AdditionalStorageServiceCreator
	IsInterfaceNil() bool
}

// StatusCoreComponentsHolder holds the status core components
type StatusCoreComponentsHolder interface {
	AppStatusHandler() core.AppStatusHandler
	IsInterfaceNil() bool
}

// NumConnectedPeersProvider defines the actions that a component that provides the number of connected peers should do
type NumConnectedPeersProvider interface {
	ConnectedPeers() []core.PeerID
	IsInterfaceNil() bool
}

// CheckedChunkResult is the DTO used to hold the results after checking a chunk of intercepted data
type CheckedChunkResult struct {
	IsChunk        bool
	HaveAllChunks  bool
	CompleteBuffer []byte
}

// InterceptedChunksProcessor defines the component that is able to process chunks of intercepted data
type InterceptedChunksProcessor interface {
	CheckBatch(b *batch.Batch, whiteListHandler WhiteListHandler) (CheckedChunkResult, error)
	Close() error
	IsInterfaceNil() bool
}

// AccountsDBSyncer defines the methods for the accounts db syncer
type AccountsDBSyncer interface {
	SyncAccounts(rootHash []byte, storageMarker common.StorageMarker) error
	IsInterfaceNil() bool
}

// CurrentNetworkEpochProviderHandler is an interface able to compute if the provided epoch is active on the network or not
type CurrentNetworkEpochProviderHandler interface {
	EpochIsActiveInNetwork(epoch uint32) bool
	IsInterfaceNil() bool
}

// ScheduledTxsExecutionHandler defines the functionality for execution of scheduled transactions
type ScheduledTxsExecutionHandler interface {
	Init()
	AddScheduledTx(txHash []byte, tx data.TransactionHandler) bool
	AddScheduledMiniBlocks(miniBlocks block.MiniBlockSlice)
	Execute(txHash []byte) error
	ExecuteAll(haveTime func() time.Duration) error
	GetScheduledIntermediateTxs() map[block.Type][]data.TransactionHandler
	GetScheduledMiniBlocks() block.MiniBlockSlice
	GetScheduledGasAndFees() scheduled.GasAndFees
	SetScheduledInfo(scheduledInfo *ScheduledInfo)
	GetScheduledRootHashForHeader(headerHash []byte) ([]byte, error)
	GetScheduledRootHashForHeaderWithEpoch(headerHash []byte, epoch uint32) ([]byte, error)
	RollBackToBlock(headerHash []byte) error
	SaveStateIfNeeded(headerHash []byte)
	SaveState(headerHash []byte, scheduledInfo *ScheduledInfo)
	GetScheduledRootHash() []byte
	SetScheduledRootHash(rootHash []byte)
	SetScheduledGasAndFees(gasAndFees scheduled.GasAndFees)
	SetTransactionProcessor(txProcessor TransactionProcessor)
	SetTransactionCoordinator(txCoordinator TransactionCoordinator)
	IsScheduledTx(txHash []byte) bool
	IsMiniBlockExecuted(mbHash []byte) bool
	IsInterfaceNil() bool
}

// ShardedPool is a perspective of the sharded data pool
type ShardedPool interface {
	AddData(key []byte, data interface{}, sizeInBytes int, cacheID string)
}

// InterceptedSignedTransactionHandler provides additional handling for signed transactions
type InterceptedSignedTransactionHandler interface {
	InterceptedTransactionHandler
	GetTxMessageForSignatureVerification() ([]byte, error)
}

// GuardianChecker can check an account guardian
type GuardianChecker interface {
	GetActiveGuardian(handler vmcommon.UserAccountHandler) ([]byte, error)
	HasActiveGuardian(uah state.UserAccountHandler) bool
	HasPendingGuardian(uah state.UserAccountHandler) bool
	IsInterfaceNil() bool
}

// GuardedAccountHandler allows setting and getting the configured account guardian
type GuardedAccountHandler interface {
	GetActiveGuardian(handler vmcommon.UserAccountHandler) ([]byte, error)
	HasActiveGuardian(uah state.UserAccountHandler) bool
	HasPendingGuardian(uah state.UserAccountHandler) bool
	SetGuardian(uah vmcommon.UserAccountHandler, guardianAddress []byte, txGuardianAddress []byte, guardianServiceUID []byte) error
	CleanOtherThanActive(uah vmcommon.UserAccountHandler)
	GetConfiguredGuardians(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error)
	IsInterfaceNil() bool
}

// DoubleTransactionDetector is able to detect if a transaction hash is present more than once in a block body
type DoubleTransactionDetector interface {
	ProcessBlockBody(body *block.Body)
	IsInterfaceNil() bool
}

// TxsSenderHandler handles transactions sending
type TxsSenderHandler interface {
	SendBulkTransactions(txs []*transaction.Transaction) (uint64, error)
	Close() error
	IsInterfaceNil() bool
}

// PreProcessorExecutionInfoHandler handles pre processor execution info needed by the transactions preprocessors
type PreProcessorExecutionInfoHandler interface {
	GetNumOfCrossInterMbsAndTxs() (int, int)
	InitProcessedTxsResults(key []byte)
	RevertProcessedTxsResults(txHashes [][]byte, key []byte)
}

// ProcessedMiniBlocksTracker handles tracking of processed mini blocks
type ProcessedMiniBlocksTracker interface {
	SetProcessedMiniBlockInfo(metaBlockHash []byte, miniBlockHash []byte, processedMbInfo *processedMb.ProcessedMiniBlockInfo)
	RemoveHeaderHash(metaBlockHash []byte)
	RemoveMiniBlockHash(miniBlockHash []byte)
	GetProcessedMiniBlocksInfo(metaBlockHash []byte) map[string]*processedMb.ProcessedMiniBlockInfo
	GetProcessedMiniBlockInfo(miniBlockHash []byte) (*processedMb.ProcessedMiniBlockInfo, []byte)
	IsMiniBlockFullyProcessed(metaBlockHash []byte, miniBlockHash []byte) bool
	ConvertProcessedMiniBlocksMapToSlice() []bootstrapStorage.MiniBlocksInMeta
	ConvertSliceToProcessedMiniBlocksMap(miniBlocksInMetaBlocks []bootstrapStorage.MiniBlocksInMeta)
	DisplayProcessedMiniBlocks()
	IsInterfaceNil() bool
}

// PeerAuthenticationPayloadValidator defines the operations supported by an entity able to validate timestamps
// found in peer authentication messages
type PeerAuthenticationPayloadValidator interface {
	ValidateTimestamp(payloadTimestamp int64) error
	IsInterfaceNil() bool
}

// Debugger defines what a process debugger implementation should do
type Debugger interface {
	SetLastCommittedBlockRound(round uint64)
	Close() error
	IsInterfaceNil() bool
}

// ProcessDebuggerSetter allows setting a debugger on the process component
type ProcessDebuggerSetter interface {
	SetProcessDebugger(debugger Debugger) error
}

// DebuggerBlockProcessor is the interface for block execution engine with debugger
type DebuggerBlockProcessor interface {
	BlockProcessor
	ProcessDebuggerSetter
}

// AdditionalStorageServiceCreator defines the actions needed for a component that can create additional storage services
type AdditionalStorageServiceCreator interface {
	CreateAdditionalStorageUnits(func(store dataRetriever.StorageService, shardID string) error, dataRetriever.StorageService, string) error
	IsInterfaceNil() bool
}
