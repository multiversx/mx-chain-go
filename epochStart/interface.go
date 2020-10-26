package epochStart

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// TriggerHandler defines the functionalities for an start of epoch trigger
type TriggerHandler interface {
	Close() error
	ForceEpochStart()
	IsEpochStart() bool
	Epoch() uint32
	MetaEpoch() uint32
	Update(round uint64, nonce uint64)
	EpochStartRound() uint64
	EpochStartMetaHdrHash() []byte
	GetSavedStateKey() []byte
	LoadState(key []byte) error
	SetProcessed(header data.HeaderHandler, body data.BodyHandler)
	SetFinalityAttestingRound(round uint64)
	EpochFinalityAttestingRound() uint64
	RevertStateToBlock(header data.HeaderHandler) error
	SetCurrentEpochStartRound(round uint64)
	RequestEpochStartIfNeeded(interceptedHeader data.HeaderHandler)
	SetAppStatusHandler(handler core.AppStatusHandler) error
	IsInterfaceNil() bool
}

// Rounder defines the actions which should be handled by a round implementation
type Rounder interface {
	// Index returns the current round
	Index() int64
	// TimeStamp returns the time stamp of the round
	TimeStamp() time.Time
	IsInterfaceNil() bool
}

// HeaderValidator defines the actions needed to validate a header
type HeaderValidator interface {
	IsHeaderConstructionValid(currHdr, prevHdr data.HeaderHandler) error
	IsInterfaceNil() bool
}

// RequestHandler defines the methods through which request to data can be made
type RequestHandler interface {
	RequestShardHeader(shardId uint32, hash []byte)
	RequestMetaHeader(hash []byte)
	RequestMetaHeaderByNonce(nonce uint64)
	RequestShardHeaderByNonce(shardId uint32, nonce uint64)
	RequestStartOfEpochMetaBlock(epoch uint32)
	RequestMiniBlocks(destShardID uint32, miniblocksHashes [][]byte)
	RequestInterval() time.Duration
	SetNumPeersToQuery(key string, intra int, cross int) error
	GetNumPeersToQuery(key string) (int, int, error)
	IsInterfaceNil() bool
}

// ActionHandler defines the action taken on epoch start event
type ActionHandler interface {
	EpochStartAction(hdr data.HeaderHandler)
	EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NotifyOrder() uint32
}

// RegistrationHandler provides Register and Unregister functionality for the end of epoch events
type RegistrationHandler interface {
	RegisterHandler(handler ActionHandler)
	UnregisterHandler(handler ActionHandler)
	IsInterfaceNil() bool
}

// Notifier defines which actions should be done for handling new epoch's events
type Notifier interface {
	NotifyAll(hdr data.HeaderHandler)
	NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NotifyEpochChangeConfirmed(epoch uint32)
	IsInterfaceNil() bool
}

// ValidatorStatisticsProcessorHandler defines the actions for processing validator statistics
// needed in the epoch events
type ValidatorStatisticsProcessorHandler interface {
	Process(info data.ShardValidatorInfoHandler) error
	Commit() ([]byte, error)
	IsInterfaceNil() bool
}

// ValidatorInfoCreator defines the methods to create a validator info
type ValidatorInfoCreator interface {
	PeerAccountToValidatorInfo(peerAccount state.PeerAccountHandler) *state.ValidatorInfo
	IsInterfaceNil() bool
}

// HeadersByHashSyncer defines the methods to sync all missing headers by hash
type HeadersByHashSyncer interface {
	SyncMissingHeadersByHash(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error
	GetHeaders() (map[string]data.HeaderHandler, error)
	ClearFields()
	IsInterfaceNil() bool
}

// PendingMiniBlocksSyncHandler defines the methods to sync all pending miniblocks
type PendingMiniBlocksSyncHandler interface {
	SyncPendingMiniBlocks(miniBlockHeaders []block.MiniBlockHeader, ctx context.Context) error
	GetMiniBlocks() (map[string]*block.MiniBlock, error)
	ClearFields()
	IsInterfaceNil() bool
}

// AccountsDBSyncer defines the methods for the accounts db syncer
type AccountsDBSyncer interface {
	GetSyncedTries() map[string]data.Trie
	SyncAccounts(rootHash []byte) error
	IsInterfaceNil() bool
}

// StartOfEpochMetaSyncer defines the methods to synchronize epoch start meta block from the network when nothing is known
type StartOfEpochMetaSyncer interface {
	SyncEpochStartMeta(waitTime time.Duration) (*block.MetaBlock, error)
	IsInterfaceNil() bool
}

// NodesConfigProvider will provide the necessary information for start in epoch economics block creation
type NodesConfigProvider interface {
	ConsensusGroupSize(shardID uint32) int
	IsInterfaceNil() bool
}

// ImportStartHandler can manage the process of starting the import after the hardfork event
type ImportStartHandler interface {
	ShouldStartImport() bool
	IsAfterExportBeforeImport() bool
	IsInterfaceNil() bool
}

// TransactionCacher defines the methods for the local cacher, info for current round
type TransactionCacher interface {
	GetTx(txHash []byte) (data.TransactionHandler, error)
	IsInterfaceNil() bool
}

// StakingDataProvider is able to provide staking data from the system smart contracts
type StakingDataProvider interface {
	GetStakingDataForBlsKey(blsKey []byte) error
	Clean()
	IsInterfaceNil() bool
}
