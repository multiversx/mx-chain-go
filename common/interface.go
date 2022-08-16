package common

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// NumNodesDTO represents the DTO structure that will hold the number of nodes split by category and other
// trie structure relevant data such as maximum number of trie levels including the roothash node and all leaves
type NumNodesDTO struct {
	Leaves     int
	Extensions int
	Branches   int
	MaxLevel   int
}

// Trie is an interface for Merkle Trees implementations
type Trie interface {
	Get(key []byte) ([]byte, error)
	Update(key, value []byte) error
	Delete(key []byte) error
	RootHash() ([]byte, error)
	Commit() error
	Recreate(root []byte) (Trie, error)
	String() string
	GetObsoleteHashes() [][]byte
	GetDirtyHashes() (ModifiedHashes, error)
	GetOldRoot() []byte
	GetSerializedNodes([]byte, uint64) ([][]byte, uint64, error)
	GetSerializedNode([]byte) ([]byte, error)
	GetNumNodes() NumNodesDTO
	GetAllLeavesOnChannel(leavesChannel chan core.KeyValueHolder, ctx context.Context, rootHash []byte) error
	GetAllHashes() ([][]byte, error)
	GetProof(key []byte) ([][]byte, []byte, error)
	VerifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error)
	GetStorageManager() StorageManager
	MarkStorerAsSyncedAndActive()
	Close() error
	IsInterfaceNil() bool
}

// StorageManager manages all trie storage operations
type StorageManager interface {
	Get(key []byte) ([]byte, error)
	GetFromCurrentEpoch(key []byte) ([]byte, error)
	Put(key []byte, val []byte) error
	PutInEpoch(key []byte, val []byte, epoch uint32) error
	PutInEpochWithoutCache(key []byte, val []byte, epoch uint32) error
	TakeSnapshot(rootHash []byte, mainTrieRootHash []byte, leavesChan chan core.KeyValueHolder, errChan chan error, stats SnapshotStatisticsHandler, epoch uint32)
	SetCheckpoint(rootHash []byte, mainTrieRootHash []byte, leavesChan chan core.KeyValueHolder, errChan chan error, stats SnapshotStatisticsHandler)
	GetLatestStorageEpoch() (uint32, error)
	IsPruningEnabled() bool
	IsPruningBlocked() bool
	EnterPruningBufferingMode()
	ExitPruningBufferingMode()
	AddDirtyCheckpointHashes([]byte, ModifiedHashes) bool
	Remove(hash []byte) error
	SetEpochForPutOperation(uint32)
	ShouldTakeSnapshot() bool
	IsClosed() bool
	Close() error
	IsInterfaceNil() bool
}

// DBWriteCacher is used to cache changes made to the trie, and only write to the database when it's needed
type DBWriteCacher interface {
	Put(key, val []byte) error
	Get(key []byte) ([]byte, error)
	Remove(key []byte) error
	Close() error
	IsInterfaceNil() bool
}

// SnapshotDbHandler is used to keep track of how many references a snapshot db has
type SnapshotDbHandler interface {
	DBWriteCacher
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
	AddSize(uint64)
	SnapshotFinished()
	NewSnapshotStarted()
	NewDataTrie()
	WaitForSnapshotsToFinish()
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

// GasScheduleNotifierAPI defines the behavior of the gas schedule notifier components that is used for api
type GasScheduleNotifierAPI interface {
	core.GasScheduleNotifier
	LatestGasScheduleCopy() map[string]map[string]uint64
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
	IsHeartbeatDisableFlagEnabled() bool
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
	IsRefactorPeersMiniBlocksFlagEnabled() bool

	IsInterfaceNil() bool
}
