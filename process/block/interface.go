package block

import (
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
	AddHeaderInfo(hash string, headerInfo headerForBlock.HeaderInfo)
	GetMissingData() (uint32, uint32, uint32)
	GetHeadersInfoMap() map[string]headerForBlock.HeaderInfo
	GetHeadersMap() map[string]data.HeaderHandler
	GetHeaderInfo(hash string) (headerForBlock.HeaderInfo, bool)
	GetHighestHeaderNonceForShard(shardID uint32) uint64
	SetHighestHeaderNonceForShard(shardID uint32, nonce uint64)
	GetLastNotarizedHeaderForShard(shardID uint32) (headerForBlock.LastNotarizedHeaderInfoHandler, bool)
	SetLastNotarizedHeaderForShard(shardID uint32, lastNotarizedHeader headerForBlock.LastNotarizedHeaderInfoHandler)
	SetHasProof(hash string)
	SetHeader(hash string, header data.HeaderHandler)
	HasProofRequested(hash string) bool
	SetHasProofRequested(hash string)
	IncreaseMissingProofs()
	DecreaseMissingProofs()
	IncreaseMissingHeaders()
	DecreaseMissingHeaders()
	IncreaseMissingFinalityAttestingHeaders()
	DecreaseMissingFinalityAttestingHeaders()
	SetMissingFinalityAttestingHeaders(missing uint32)
	Reset()
	ResetMissingHeaders()
}
