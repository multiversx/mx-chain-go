package common

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

// TrieIteratorChannels defines the channels that are being used when iterating the trie nodes
type TrieIteratorChannels struct {
	LeavesChan chan core.KeyValueHolder
	ErrChan    BufferedErrChan
}

// TrieType defines the type of the trie
type TrieType string

const (
	// MainTrie represents the main trie in which all the accounts and SC code are stored
	MainTrie TrieType = "mainTrie"

	// DataTrie represents a data trie in which all the data related to an account is stored
	DataTrie TrieType = "dataTrie"
)

// BufferedErrChan is an interface that defines the methods for a buffered error channel
type BufferedErrChan interface {
	WriteInChanNonBlocking(err error)
	ReadFromChanNonBlocking() error
	Close()
	IsInterfaceNil() bool
}

// Trie is an interface for Merkle Trees implementations
type Trie interface {
	Get(key []byte) ([]byte, uint32, error)
	Update(key, value []byte) error
	Delete(key []byte) error
	RootHash() ([]byte, error)
	Commit() error
	Recreate(root []byte) (Trie, error)
	RecreateFromEpoch(options RootHashHolder) (Trie, error)
	String() string
	GetObsoleteHashes() [][]byte
	GetDirtyHashes() (ModifiedHashes, error)
	GetOldRoot() []byte
	GetSerializedNodes([]byte, uint64) ([][]byte, uint64, error)
	GetSerializedNode([]byte) ([]byte, error)
	GetAllLeavesOnChannel(allLeavesChan *TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder KeyBuilder, trieLeafParser TrieLeafParser) error
	GetAllHashes() ([][]byte, error)
	GetProof(key []byte) ([][]byte, []byte, error)
	VerifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error)
	GetStorageManager() StorageManager
	IsMigratedToLatestVersion() (bool, error)
	Close() error
	IsInterfaceNil() bool
}

// TrieLeafParser is used to parse trie leaves
type TrieLeafParser interface {
	ParseLeaf(key []byte, val []byte, version core.TrieNodeVersion) (core.KeyValueHolder, error)
	IsInterfaceNil() bool
}

// TrieStats is used to collect the trie statistics for the given rootHash
type TrieStats interface {
	GetTrieStats(address string, rootHash []byte) (TrieStatisticsHandler, error)
}

// StorageMarker is used to mark the given storer as synced and active
type StorageMarker interface {
	MarkStorerAsSyncedAndActive(storer StorageManager)
	IsInterfaceNil() bool
}

// KeyBuilder is used for building trie keys as you traverse the trie
type KeyBuilder interface {
	BuildKey(keyPart []byte)
	GetKey() ([]byte, error)
	Clone() KeyBuilder
	IsInterfaceNil() bool
}

// DataTrieHandler is an interface that declares the methods used for dataTries
type DataTrieHandler interface {
	RootHash() ([]byte, error)
	GetAllLeavesOnChannel(leavesChannels *TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder KeyBuilder, trieLeafParser TrieLeafParser) error
	IsMigratedToLatestVersion() (bool, error)
	IsInterfaceNil() bool
}

// StorageManager manages all trie storage operations
type StorageManager interface {
	TrieStorageInteractor
	GetFromCurrentEpoch(key []byte) ([]byte, error)
	PutInEpoch(key []byte, val []byte, epoch uint32) error
	PutInEpochWithoutCache(key []byte, val []byte, epoch uint32) error
	TakeSnapshot(address string, rootHash []byte, mainTrieRootHash []byte, iteratorChannels *TrieIteratorChannels, missingNodesChan chan []byte, stats SnapshotStatisticsHandler, epoch uint32)
	SetCheckpoint(rootHash []byte, mainTrieRootHash []byte, iteratorChannels *TrieIteratorChannels, missingNodesChan chan []byte, stats SnapshotStatisticsHandler)
	GetLatestStorageEpoch() (uint32, error)
	IsPruningEnabled() bool
	IsPruningBlocked() bool
	EnterPruningBufferingMode()
	ExitPruningBufferingMode()
	AddDirtyCheckpointHashes([]byte, ModifiedHashes) bool
	RemoveFromAllActiveEpochs(hash []byte) error
	SetEpochForPutOperation(uint32)
	ShouldTakeSnapshot() bool
	GetBaseTrieStorageManager() StorageManager
	IsClosed() bool
	Close() error
	IsInterfaceNil() bool
}

// TrieStorageInteractor defines the methods used for interacting with the trie storage
type TrieStorageInteractor interface {
	BaseStorer
	GetIdentifier() string
}

// BaseStorer define the base methods needed for a storer
type BaseStorer interface {
	Put(key, val []byte) error
	Get(key []byte) ([]byte, error)
	Remove(key []byte) error
	Close() error
	IsInterfaceNil() bool
}

// SnapshotDbHandler is used to keep track of how many references a snapshot db has
type SnapshotDbHandler interface {
	BaseStorer
	IsInUse() bool
	DecreaseNumReferences()
	IncreaseNumReferences()
	MarkForRemoval()
	MarkForDisconnection()
	SetPath(string)
}

// TriesHolder is used to store multiple tries
type TriesHolder interface {
	Put([]byte, Trie)
	Replace(key []byte, tr Trie)
	Get([]byte) Trie
	GetAll() []Trie
	Reset()
	IsInterfaceNil() bool
}

// Locker defines the operations used to lock different critical areas. Implemented by the RWMutex.
type Locker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

// MerkleProofVerifier is used to verify merkle proofs
type MerkleProofVerifier interface {
	VerifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error)
}

// SizeSyncStatisticsHandler extends the SyncStatisticsHandler interface by allowing setting up the trie node size
type SizeSyncStatisticsHandler interface {
	data.SyncStatisticsHandler
	AddNumBytesReceived(bytes uint64)
	NumBytesReceived() uint64
	NumTries() int
	AddProcessingTime(duration time.Duration)
	IncrementIteration()
	ProcessingTime() time.Duration
	NumIterations() int
}

// SnapshotStatisticsHandler is used to measure different statistics for the trie snapshot
type SnapshotStatisticsHandler interface {
	SnapshotFinished()
	NewSnapshotStarted()
	WaitForSnapshotsToFinish()
	AddTrieStats(handler TrieStatisticsHandler, trieType TrieType)
	IsInterfaceNil() bool
}

// TrieStatisticsHandler is used to collect different statistics about a single trie
type TrieStatisticsHandler interface {
	AddBranchNode(level int, size uint64)
	AddExtensionNode(level int, size uint64)
	AddLeafNode(level int, size uint64, version core.TrieNodeVersion)
	AddAccountInfo(address string, rootHash []byte)

	GetTotalNodesSize() uint64
	GetTotalNumNodes() uint64
	GetMaxTrieDepth() uint32
	GetBranchNodesSize() uint64
	GetNumBranchNodes() uint64
	GetExtensionNodesSize() uint64
	GetNumExtensionNodes() uint64
	GetLeafNodesSize() uint64
	GetNumLeafNodes() uint64
	GetLeavesMigrationStats() map[core.TrieNodeVersion]uint64

	MergeTriesStatistics(statsToBeMerged TrieStatisticsHandler)
	ToString() []string
	IsInterfaceNil() bool
}

// TriesStatisticsCollector is used to merge the statistics for multiple tries
type TriesStatisticsCollector interface {
	Add(trieStats TrieStatisticsHandler, trieType TrieType)
	Print()
	GetNumNodes() uint64
}

// ProcessStatusHandler defines the behavior of a component able to hold the current status of the node and
// able to tell if the node is idle or processing/committing a block
type ProcessStatusHandler interface {
	SetBusy(reason string)
	SetIdle()
	IsIdle() bool
	IsInterfaceNil() bool
}

// BlockInfo provides a block information such as nonce, hash, roothash and so on
type BlockInfo interface {
	GetNonce() uint64
	GetHash() []byte
	GetRootHash() []byte
	Equal(blockInfo BlockInfo) bool
	IsInterfaceNil() bool
}

// ReceiptsHolder holds receipts content (e.g. miniblocks)
type ReceiptsHolder interface {
	GetMiniblocks() []*block.MiniBlock
	IsInterfaceNil() bool
}

// RootHashHolder holds a rootHash and the corresponding epoch
type RootHashHolder interface {
	GetRootHash() []byte
	GetEpoch() core.OptionalUint32
	String() string
	IsInterfaceNil() bool
}

// GasScheduleNotifierAPI defines the behavior of the gas schedule notifier components that is used for api
type GasScheduleNotifierAPI interface {
	core.GasScheduleNotifier
	LatestGasScheduleCopy() map[string]map[string]uint64
}

// PidQueueHandler defines the behavior of a queue of pids
type PidQueueHandler interface {
	Push(pid core.PeerID)
	Pop() core.PeerID
	IndexOf(pid core.PeerID) int
	Promote(idx int)
	Remove(pid core.PeerID)
	DataSizeInBytes() int
	Get(idx int) core.PeerID
	Len() int
	IsInterfaceNil() bool
}

// EnableEpochsHandler is used to verify the which flags are set in the current epoch based on EnableEpochs config
type EnableEpochsHandler interface {
	BlockGasAndFeesReCheckEnableEpoch() uint32
	StakingV2EnableEpoch() uint32
	ScheduledMiniBlocksEnableEpoch() uint32
	SwitchJailWaitingEnableEpoch() uint32
	BalanceWaitingListsEnableEpoch() uint32
	WaitingListFixEnableEpoch() uint32
	MultiESDTTransferAsyncCallBackEnableEpoch() uint32
	FixOOGReturnCodeEnableEpoch() uint32
	RemoveNonUpdatedStorageEnableEpoch() uint32
	CreateNFTThroughExecByCallerEnableEpoch() uint32
	FixFailExecutionOnErrorEnableEpoch() uint32
	ManagedCryptoAPIEnableEpoch() uint32
	DisableExecByCallerEnableEpoch() uint32
	RefactorContextEnableEpoch() uint32
	CheckExecuteReadOnlyEnableEpoch() uint32
	StorageAPICostOptimizationEnableEpoch() uint32
	MiniBlockPartialExecutionEnableEpoch() uint32
	RefactorPeersMiniBlocksEnableEpoch() uint32
	IsSCDeployFlagEnabled() bool
	IsBuiltInFunctionsFlagEnabled() bool
	IsRelayedTransactionsFlagEnabled() bool
	IsPenalizedTooMuchGasFlagEnabled() bool
	ResetPenalizedTooMuchGasFlag()
	IsSwitchJailWaitingFlagEnabled() bool
	IsBelowSignedThresholdFlagEnabled() bool
	IsSwitchHysteresisForMinNodesFlagEnabled() bool
	IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch() bool
	IsTransactionSignedWithTxHashFlagEnabled() bool
	IsMetaProtectionFlagEnabled() bool
	IsAheadOfTimeGasUsageFlagEnabled() bool
	IsGasPriceModifierFlagEnabled() bool
	IsRepairCallbackFlagEnabled() bool
	IsBalanceWaitingListsFlagEnabled() bool
	IsReturnDataToLastTransferFlagEnabled() bool
	IsSenderInOutTransferFlagEnabled() bool
	IsStakeFlagEnabled() bool
	IsStakingV2FlagEnabled() bool
	IsStakingV2OwnerFlagEnabled() bool
	IsStakingV2FlagEnabledForActivationEpochCompleted() bool
	IsDoubleKeyProtectionFlagEnabled() bool
	IsESDTFlagEnabled() bool
	IsESDTFlagEnabledForCurrentEpoch() bool
	IsGovernanceFlagEnabled() bool
	IsGovernanceFlagEnabledForCurrentEpoch() bool
	IsDelegationManagerFlagEnabled() bool
	IsDelegationSmartContractFlagEnabled() bool
	IsDelegationSmartContractFlagEnabledForCurrentEpoch() bool
	IsCorrectLastUnJailedFlagEnabled() bool
	IsCorrectLastUnJailedFlagEnabledForCurrentEpoch() bool
	IsRelayedTransactionsV2FlagEnabled() bool
	IsUnBondTokensV2FlagEnabled() bool
	IsSaveJailedAlwaysFlagEnabled() bool
	IsReDelegateBelowMinCheckFlagEnabled() bool
	IsValidatorToDelegationFlagEnabled() bool
	IsWaitingListFixFlagEnabled() bool
	IsIncrementSCRNonceInMultiTransferFlagEnabled() bool
	IsESDTMultiTransferFlagEnabled() bool
	IsGlobalMintBurnFlagEnabled() bool
	IsESDTTransferRoleFlagEnabled() bool
	IsBuiltInFunctionOnMetaFlagEnabled() bool
	IsComputeRewardCheckpointFlagEnabled() bool
	IsSCRSizeInvariantCheckFlagEnabled() bool
	IsBackwardCompSaveKeyValueFlagEnabled() bool
	IsESDTNFTCreateOnMultiShardFlagEnabled() bool
	IsMetaESDTSetFlagEnabled() bool
	IsAddTokensToDelegationFlagEnabled() bool
	IsMultiESDTTransferFixOnCallBackFlagEnabled() bool
	IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() bool
	IsCorrectFirstQueuedFlagEnabled() bool
	IsDeleteDelegatorAfterClaimRewardsFlagEnabled() bool
	IsFixOOGReturnCodeFlagEnabled() bool
	IsRemoveNonUpdatedStorageFlagEnabled() bool
	IsOptimizeNFTStoreFlagEnabled() bool
	IsCreateNFTThroughExecByCallerFlagEnabled() bool
	IsStopDecreasingValidatorRatingWhenStuckFlagEnabled() bool
	IsFrontRunningProtectionFlagEnabled() bool
	IsPayableBySCFlagEnabled() bool
	IsCleanUpInformativeSCRsFlagEnabled() bool
	IsStorageAPICostOptimizationFlagEnabled() bool
	IsESDTRegisterAndSetAllRolesFlagEnabled() bool
	IsScheduledMiniBlocksFlagEnabled() bool
	IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled() bool
	IsDoNotReturnOldBlockInBlockchainHookFlagEnabled() bool
	IsAddFailedRelayedTxToInvalidMBsFlag() bool
	IsSCRSizeInvariantOnBuiltInResultFlagEnabled() bool
	IsCheckCorrectTokenIDForTransferRoleFlagEnabled() bool
	IsFailExecutionOnEveryAPIErrorFlagEnabled() bool
	IsMiniBlockPartialExecutionFlagEnabled() bool
	IsManagedCryptoAPIsFlagEnabled() bool
	IsESDTMetadataContinuousCleanupFlagEnabled() bool
	IsDisableExecByCallerFlagEnabled() bool
	IsRefactorContextFlagEnabled() bool
	IsCheckFunctionArgumentFlagEnabled() bool
	IsCheckExecuteOnReadOnlyFlagEnabled() bool
	IsFixAsyncCallbackCheckFlagEnabled() bool
	IsSaveToSystemAccountFlagEnabled() bool
	IsCheckFrozenCollectionFlagEnabled() bool
	IsSendAlwaysFlagEnabled() bool
	IsValueLengthCheckFlagEnabled() bool
	IsCheckTransferFlagEnabled() bool
	IsTransferToMetaFlagEnabled() bool
	IsESDTNFTImprovementV1FlagEnabled() bool
	IsSetSenderInEeiOutputTransferFlagEnabled() bool
	IsChangeDelegationOwnerFlagEnabled() bool
	IsRefactorPeersMiniBlocksFlagEnabled() bool
	IsSCProcessorV2FlagEnabled() bool
	IsFixAsyncCallBackArgsListFlagEnabled() bool
	IsFixOldTokenLiquidityEnabled() bool
	IsRuntimeMemStoreLimitEnabled() bool
	IsRuntimeCodeSizeFixEnabled() bool
	IsMaxBlockchainHookCountersFlagEnabled() bool
	IsWipeSingleNFTLiquidityDecreaseEnabled() bool
	IsAlwaysSaveTokenMetaDataEnabled() bool
	IsSetGuardianEnabled() bool
	IsScToScEventLogEnabled() bool
	IsRelayedNonceFixEnabled() bool
	IsKeepExecOrderOnCreatedSCRsEnabled() bool
	IsMultiClaimOnDelegationEnabled() bool
	IsChangeUsernameEnabled() bool
	IsConsistentTokensValuesLengthCheckEnabled() bool
	IsAutoBalanceDataTriesEnabled() bool
	FixDelegationChangeOwnerOnAccountEnabled() bool

	IsInterfaceNil() bool
}

// ManagedPeersHolder defines the operations of an entity that holds managed identities for a node
type ManagedPeersHolder interface {
	AddManagedPeer(privateKeyBytes []byte) error
	GetPrivateKey(pkBytes []byte) (crypto.PrivateKey, error)
	GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error)
	GetMachineID(pkBytes []byte) (string, error)
	GetNameAndIdentity(pkBytes []byte) (string, string, error)
	IncrementRoundsWithoutReceivedMessages(pkBytes []byte)
	ResetRoundsWithoutReceivedMessages(pkBytes []byte)
	GetManagedKeysByCurrentNode() map[string]crypto.PrivateKey
	IsKeyManagedByCurrentNode(pkBytes []byte) bool
	IsKeyRegistered(pkBytes []byte) bool
	IsPidManagedByCurrentNode(pid core.PeerID) bool
	IsKeyValidator(pkBytes []byte) bool
	SetValidatorState(pkBytes []byte, state bool)
	GetNextPeerAuthenticationTime(pkBytes []byte) (time.Time, error)
	SetNextPeerAuthenticationTime(pkBytes []byte, nextTime time.Time)
	IsMultiKeyMode() bool
	IsInterfaceNil() bool
}

// MissingTrieNodesNotifier defines the operations of an entity that notifies about missing trie nodes
type MissingTrieNodesNotifier interface {
	RegisterHandler(handler StateSyncNotifierSubscriber) error
	AsyncNotifyMissingTrieNode(hash []byte)
	IsInterfaceNil() bool
}

// StateSyncNotifierSubscriber defines the operations of an entity that subscribes to a missing trie nodes notifier
type StateSyncNotifierSubscriber interface {
	MissingDataTrieNodeFound(hash []byte)
	IsInterfaceNil() bool
}
