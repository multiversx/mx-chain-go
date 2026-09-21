package bootstrap

import (
	"sync/atomic"

	"github.com/multiversx/mx-chain-go/process"
)

// historicalRequestHandler requests boundary data in both the requested epoch
// and its predecessor. SetEpoch follows the bootstrap traversal and validator sync;
// explicitly supplied epochs take precedence over that current epoch.
type historicalRequestHandler struct {
	process.RequestHandler
	epoch atomic.Uint32
}

func (h *historicalRequestHandler) SetEpoch(epoch uint32) {
	h.RequestHandler.SetEpoch(epoch)
	h.epoch.Store(epoch)
}

func (e *epochStartBootstrap) enableHistoricalRequests() {
	if e.flagsConfig.StartInEpochOffset > 0 {
		e.requestHandler = &historicalRequestHandler{RequestHandler: e.requestHandler}
	}
}

func (h *historicalRequestHandler) RequestMiniBlock(shardID uint32, hash []byte) {
	h.RequestMiniBlockForEpoch(shardID, hash, h.epoch.Load())
}

func (h *historicalRequestHandler) RequestMiniBlockForEpoch(shardID uint32, hash []byte, epoch uint32) {
	h.RequestHandler.RequestMiniBlockForEpoch(shardID, hash, epoch)
	if epoch > 0 {
		h.RequestHandler.RequestMiniBlockForEpoch(shardID, hash, epoch-1)
	}
}

func (h *historicalRequestHandler) RequestMiniBlocks(shardID uint32, hashes [][]byte) {
	h.RequestMiniBlocksForEpoch(shardID, hashes, h.epoch.Load())
}

func (h *historicalRequestHandler) RequestMiniBlocksForEpoch(shardID uint32, hashes [][]byte, epoch uint32) {
	h.RequestHandler.RequestMiniBlocksForEpoch(shardID, hashes, epoch)
	if epoch > 0 {
		h.RequestHandler.RequestMiniBlocksForEpoch(shardID, hashes, epoch-1)
	}
}

func (h *historicalRequestHandler) RequestShardHeader(shardID uint32, hash []byte) {
	h.RequestShardHeaderForEpoch(shardID, hash, h.epoch.Load())
}

func (h *historicalRequestHandler) RequestShardHeaderForEpoch(shardID uint32, hash []byte, epoch uint32) {
	h.RequestHandler.RequestShardHeaderForEpoch(shardID, hash, epoch)
	if epoch > 0 {
		h.RequestHandler.RequestShardHeaderForEpoch(shardID, hash, epoch-1)
	}
}

func (h *historicalRequestHandler) RequestMetaHeader(hash []byte) {
	h.RequestMetaHeaderForEpoch(hash, h.epoch.Load())
}

func (h *historicalRequestHandler) RequestMetaHeaderForEpoch(hash []byte, epoch uint32) {
	h.RequestHandler.RequestMetaHeaderForEpoch(hash, epoch)
	if epoch > 0 {
		h.RequestHandler.RequestMetaHeaderForEpoch(hash, epoch-1)
	}
}

func (h *historicalRequestHandler) RequestShardHeaderByNonce(shardID uint32, nonce uint64) {
	h.RequestShardHeaderByNonceForEpoch(shardID, nonce, h.epoch.Load())
}

func (h *historicalRequestHandler) RequestShardHeaderByNonceForEpoch(shardID uint32, nonce uint64, epoch uint32) {
	h.RequestHandler.RequestShardHeaderByNonceForEpoch(shardID, nonce, epoch)
	if epoch > 0 {
		h.RequestHandler.RequestShardHeaderByNonceForEpoch(shardID, nonce, epoch-1)
	}
}

func (h *historicalRequestHandler) RequestMetaHeaderByNonce(nonce uint64) {
	h.RequestMetaHeaderByNonceForEpoch(nonce, h.epoch.Load())
}

func (h *historicalRequestHandler) RequestMetaHeaderByNonceForEpoch(nonce uint64, epoch uint32) {
	h.RequestHandler.RequestMetaHeaderByNonceForEpoch(nonce, epoch)
	if epoch > 0 {
		h.RequestHandler.RequestMetaHeaderByNonceForEpoch(nonce, epoch-1)
	}
}

func (h *historicalRequestHandler) RequestTransactions(shardID uint32, hashes [][]byte) {
	h.RequestTransactionsForEpoch(shardID, hashes, h.epoch.Load())
}

func (h *historicalRequestHandler) RequestTransactionsForEpoch(shardID uint32, hashes [][]byte, epoch uint32) {
	h.RequestHandler.RequestTransactionsForEpoch(shardID, hashes, epoch)
	if epoch > 0 {
		h.RequestHandler.RequestTransactionsForEpoch(shardID, hashes, epoch-1)
	}
}

func (h *historicalRequestHandler) RequestUnsignedTransactions(shardID uint32, hashes [][]byte) {
	h.RequestUnsignedTransactionsForEpoch(shardID, hashes, h.epoch.Load())
}

func (h *historicalRequestHandler) RequestUnsignedTransactionsForEpoch(shardID uint32, hashes [][]byte, epoch uint32) {
	h.RequestHandler.RequestUnsignedTransactionsForEpoch(shardID, hashes, epoch)
	if epoch > 0 {
		h.RequestHandler.RequestUnsignedTransactionsForEpoch(shardID, hashes, epoch-1)
	}
}

func (h *historicalRequestHandler) RequestRewardTransactions(shardID uint32, hashes [][]byte) {
	h.RequestRewardTransactionsForEpoch(shardID, hashes, h.epoch.Load())
}

func (h *historicalRequestHandler) RequestRewardTransactionsForEpoch(shardID uint32, hashes [][]byte, epoch uint32) {
	h.RequestHandler.RequestRewardTransactionsForEpoch(shardID, hashes, epoch)
	if epoch > 0 {
		h.RequestHandler.RequestRewardTransactionsForEpoch(shardID, hashes, epoch-1)
	}
}
