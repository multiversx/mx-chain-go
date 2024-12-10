package requestHandlers

import (
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

// HashSliceRequester can request multiple hashes at once
type HashSliceRequester interface {
	RequestDataFromHashArray(hashes [][]byte, epoch uint32) error
	IsInterfaceNil() bool
}

// ChunkRequester can request a chunk of a large data
type ChunkRequester interface {
	RequestDataFromReferenceAndChunk(reference []byte, chunkIndex uint32) error
}

// NonceRequester can request data for a specific nonce
type NonceRequester interface {
	RequestDataFromNonce(nonce uint64, epoch uint32) error
}

// EpochRequester can request data for a specific epoch
type EpochRequester interface {
	RequestDataFromEpoch(identifier []byte) error
}

// HeaderRequester defines what a block header requester can do
type HeaderRequester interface {
	NonceRequester
	EpochRequester
}

// RequestHandlerArgs holds all dependencies required by the process data factory to create components
type RequestHandlerArgs struct {
	RequestersFinder      dataRetriever.RequestersFinder
	RequestedItemsHandler dataRetriever.RequestedItemsHandler
	WhiteListHandler      process.WhiteListHandler
	MaxTxsToRequest       int
	ShardID               uint32
	RequestInterval       time.Duration
}

type baseRequestHandler interface {
	getTrieNodeRequester(topic string) (dataRetriever.Requester, error)
	getTrieNodesRequester(topic string, destShardID uint32) (dataRetriever.Requester, error)
	getStartOfEpochMetaBlockRequester(topic string) (dataRetriever.Requester, error)
	getMetaHeaderRequester() (HeaderRequester, error)
	getShardHeaderRequester(shardID uint32) (dataRetriever.Requester, error)
	getValidatorsInfoRequester() (dataRetriever.Requester, error)
	getMiniBlocksRequester(destShardID uint32) (dataRetriever.Requester, error)
}
