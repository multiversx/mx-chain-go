package block

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process/block/headerForBlock"
)

type blockProcessor interface {
	removeStartOfEpochBlockDataFromPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error
}

type gasConsumedProvider interface {
	TotalGasProvided() uint64
	TotalGasProvidedWithScheduled() uint64
	TotalGasRefunded() uint64
	TotalGasPenalized() uint64
	IsInterfaceNil() bool
}

type peerAccountsDBHandler interface {
	MarkSnapshotDone()
}

type receiptsRepository interface {
	SaveReceipts(holder common.ReceiptsHolder, header data.HeaderHandler, headerHash []byte) error
	IsInterfaceNil() bool
}

// HeadersForBlock defines a component able to hold headers for a block
type HeadersForBlock interface {
	AddHeaderUsedInBlock(hash string, header data.HeaderHandler)
	AddHeaderNotUsedInBlock(hash string, header data.HeaderHandler)
	RequestShardHeaders(metaBlock data.MetaHeaderHandler)
	RequestMetaHeaders(shardHeader data.ShardHeaderHandler)
	WaitForHeadersIfNeeded(haveTime func() time.Duration) error
	GetHeaderInfo(hash string) (headerForBlock.HeaderInfo, bool)
	GetHeadersInfoMap() map[string]headerForBlock.HeaderInfo
	GetHeadersMap() map[string]data.HeaderHandler
	ComputeHeadersForCurrentBlock(usedInBlock bool) (map[uint32][]data.HeaderHandler, error)
	ComputeHeadersForCurrentBlockInfo(usedInBlock bool) (map[uint32][]headerForBlock.NonceAndHashInfo, error)
	GetMissingData() (uint32, uint32, uint32)
	Reset()
	IsInterfaceNil() bool
}

// ExecutionResultsVerifier is the interface that defines the methods for verifying execution results
type ExecutionResultsVerifier interface {
	VerifyHeaderExecutionResults(headerHash []byte, header data.HeaderHandler) error
	IsInterfaceNil() bool
}

// MiniBlocksSelectionSession defines a session for selecting mini blocks
type MiniBlocksSelectionSession interface {
	ResetSelectionSession()
	GetMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler
	GetMiniBlocks() block.MiniBlockSlice
	GetMiniBlockHashes() [][]byte
	AddReferencedMetaBlock(metaBlock data.HeaderHandler, metaBlockHash []byte)
	GetReferencedMetaBlockHashes() [][]byte
	GetReferencedMetaBlocks() []data.HeaderHandler
	GetLastMetaBlock() data.HeaderHandler
	GetGasProvided() uint64
	GetNumTxsAdded() uint32
	AddMiniBlocksAndHashes(miniBlocksAndHashes []block.MiniblockAndHash) error
	CreateAndAddMiniBlockFromTransactions(txHashes [][]byte) error
	IsInterfaceNil() bool
}

// MissingDataResolver is the interface that defines the methods for resolving missing data
type MissingDataResolver interface {
	RequestMissingMetaHeadersBlocking(shardHeader data.ShardHeaderHandler, timeout time.Duration) error
	RequestMissingMetaHeaders(shardHeader data.ShardHeaderHandler) error
	WaitForMissingData(timeout time.Duration) error
	IsInterfaceNil() bool
}
