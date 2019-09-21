package process

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// TransactionProcessor is the main interface for transaction execution engine
type TransactionProcessor interface {
	ProcessTransaction(transaction *transaction.Transaction, round uint64) error
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
	IsTxValidForProcessing(txHandler TxValidatorHandler) bool
	NumRejectedTxs() uint64
	IsInterfaceNil() bool
}

// HeaderValidator can determine if a provided header handler is valid or not from the process point of view
type HeaderValidator interface {
	IsHeaderValidForProcessing(headerHandler data.HeaderHandler) bool
	IsInterfaceNil() bool
}

// TransactionCoordinator is an interface to coordinate transaction processing using multiple processors
type TransactionCoordinator interface {
	RequestMiniBlocks(header data.HeaderHandler)
	RequestBlockTransactions(body block.Body)
	IsDataPreparedForProcessing(haveTime func() time.Duration) error

	SaveBlockDataToStorage(body block.Body) error
	RestoreBlockDataFromStorage(body block.Body) (int, map[int][][]byte, error)
	RemoveBlockDataFromPool(body block.Body) error

	ProcessBlockTransaction(body block.Body, round uint64, haveTime func() time.Duration) error

	CreateBlockStarted()
	CreateMbsAndProcessCrossShardTransactionsDstMe(header data.HeaderHandler, maxTxSpaceRemained uint32, maxMbSpaceRemained uint32, round uint64, haveTime func() bool) (block.MiniBlockSlice, uint32, bool)
	CreateMbsAndProcessTransactionsFromMe(maxTxSpaceRemained uint32, maxMbSpaceRemained uint32, round uint64, haveTime func() bool) block.MiniBlockSlice

	CreateMarshalizedData(body block.Body) (map[uint32]block.MiniBlockSlice, map[string][][]byte)

	GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler

	VerifyCreatedBlockTransactions(body block.Body) error
	IsInterfaceNil() bool
}

// SmartContractProcessor is the main interface for the smart contract caller engine
type SmartContractProcessor interface {
	ComputeTransactionType(tx *transaction.Transaction) (TransactionType, error)
	ExecuteSmartContractTransaction(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler, round uint64) error
	DeploySmartContract(tx *transaction.Transaction, acntSrc state.AccountHandler, round uint64) error
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

// PreProcessor is an interface used to prepare and process transaction data
type PreProcessor interface {
	CreateBlockStarted()
	IsDataPrepared(requestedTxs int, haveTime func() time.Duration) error

	RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error
	RestoreTxBlockIntoPools(body block.Body, miniBlockPool storage.Cacher) (int, map[int][]byte, error)
	SaveTxBlockToStorage(body block.Body) error

	ProcessBlockTransactions(body block.Body, round uint64, haveTime func() time.Duration) error
	RequestBlockTransactions(body block.Body) int

	CreateMarshalizedData(txHashes [][]byte) ([][]byte, error)

	RequestTransactionsForMiniBlock(mb block.MiniBlock) int
	ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, round uint64) error
	CreateAndProcessMiniBlock(sndShardId, dstShardId uint32, spaceRemained int, haveTime func() bool, round uint64) (*block.MiniBlock, error)

	GetAllCurrentUsedTxs() map[string]data.TransactionHandler
	IsInterfaceNil() bool
}

// BlockProcessor is the main interface for block execution engine
type BlockProcessor interface {
	ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountState()
	CreateBlockBody(round uint64, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error
	CreateBlockHeader(body data.BodyHandler, round uint64, haveTime func() bool) (data.HeaderHandler, error)
	MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBody(dta []byte) data.BodyHandler
	DecodeBlockHeader(dta []byte) data.HeaderHandler
	AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler)
	IsInterfaceNil() bool
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

// InterceptedBlockBody interface provides functionality over intercepted blocks
type InterceptedBlockBody interface {
	Checker
	HashAccesser
	GetUnderlyingObject() interface{}
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
	AddHeader(header data.HeaderHandler, headerHash []byte, state BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error
	RemoveHeaders(nonce uint64, hash []byte)
	CheckFork() (forkDetected bool, nonce uint64, hash []byte)
	GetHighestFinalBlockNonce() uint64
	ProbableHighestNonce() uint64
	ResetProbableHighestNonceIfNeeded()
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
	VMAccountsDB() *hooks.VMAccountsDB
	IsInterfaceNil() bool
}

// Interceptor defines what a data interceptor should do
// It should also adhere to the p2p.MessageProcessor interface so it can wire to a p2p.Messenger
type Interceptor interface {
	ProcessReceivedMessage(message p2p.MessageP2P) error
	IsInterfaceNil() bool
}

// MessageHandler defines the functionality needed by structs to send data to other peers
type MessageHandler interface {
	ConnectedPeersOnTopic(topic string) []p2p.PeerID
	SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error
	IsInterfaceNil() bool
}

// TopicHandler defines the functionality needed by structs to manage topics and message processors
type TopicHandler interface {
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error
}

// TopicMessageHandler defines the functionality needed by structs to manage topics, message processors and to send data
// to other peers
type TopicMessageHandler interface {
	MessageHandler
	TopicHandler
}

// ChronologyValidator defines the functionality needed to validate a received header block (shard or metachain)
// from chronology point of view
type ChronologyValidator interface {
	ValidateReceivedBlock(shardID uint32, epoch uint32, nonce uint64, round uint64) error
	IsInterfaceNil() bool
}

// DataPacker can split a large slice of byte slices in smaller packets
type DataPacker interface {
	PackDataInChunks(data [][]byte, limit int) ([][]byte, error)
	IsInterfaceNil() bool
}

// BlocksTracker defines the functionality to track all the notarised blocks
type BlocksTracker interface {
	UnnotarisedBlocks() []data.HeaderHandler
	RemoveNotarisedBlocks(headerHandler data.HeaderHandler) error
	AddBlock(headerHandler data.HeaderHandler)
	SetBlockBroadcastRound(nonce uint64, round int64)
	BlockBroadcastRound(nonce uint64) int64
	IsInterfaceNil() bool
}

// RequestHandler defines the methods through which request to data can be made
type RequestHandler interface {
	RequestHeaderByNonce(shardId uint32, nonce uint64)
	RequestTransaction(shardId uint32, txHashes [][]byte)
	RequestUnsignedTransactions(destShardID uint32, scrHashes [][]byte)
	RequestMiniBlock(shardId uint32, miniblockHash []byte)
	RequestHeader(shardId uint32, hash []byte)
	IsInterfaceNil() bool
}

// ArgumentsParser defines the functionality to parse transaction data into arguments and code for smart contracts
type ArgumentsParser interface {
	GetArguments() ([]*big.Int, error)
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

// TxValidatorHandler defines the functionality that is needed for a TxValidator to validate a transaction
type TxValidatorHandler interface {
	SenderShardId() uint32
	Nonce() uint64
	SenderAddress() state.AddressContainer
	TotalValue() *big.Int
}
