package block

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"

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

// ExecutionResultsTracker is the interface that defines the methods for tracking execution results
type ExecutionResultsTracker interface {
	AddExecutionResult(executionResult data.ExecutionResultHandler) error
	GetPendingExecutionResults() ([]data.ExecutionResultHandler, error)
	GetPendingExecutionResultByHash(hash []byte) (data.ExecutionResultHandler, error)
	GetPendingExecutionResultByNonce(nonce uint64) (data.ExecutionResultHandler, error)
	GetLastNotarizedExecutionResult() (data.ExecutionResultHandler, error)
	SetLastNotarizedResult(executionResult data.ExecutionResultHandler) error
	IsInterfaceNil() bool
}

// GasComputation is the interface that defines the methods for gas tracking and computation for the proposed transactions
type GasComputation interface {
	CheckIncomingMiniBlocks(
		miniBlocks []data.MiniBlockHeaderHandler,
		transactions map[string][]data.TransactionHandler,
	) (uint32, uint32)
	CheckOutgoingTransactions(transactions []data.TransactionHandler) uint32
	TotalGasConsumed() uint64
	GetLastMiniBlockIndexIncluded() (bool, uint32)
	GetLasTransactionIndexIncluded() (bool, uint32)
	DecreaseMiniBlockLimit()
	ResetMiniBlockLimit()
	DecreaseBlockLimit()
	ResetBlockLimit()
	Reset()
}
