package preprocess

import (
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

// ComputeSortedTxs exported function for compute sorted txs
func (txs *transactions) ComputeSortedTxs(
	sndShardId uint32,
	dstShardId uint32,
	gasBandwidth uint64,
	randomness []byte,
) ([]*txcache.WrappedTransaction, []*txcache.WrappedTransaction, error) {
	return txs.computeSortedTxs(sndShardId, dstShardId, gasBandwidth, randomness)
}

// PrefilterTransactions exported function for prefilter transactions
func (txs *transactions) PrefilterTransactions(
	initialTxs []*txcache.WrappedTransaction,
	additionalTxs []*txcache.WrappedTransaction,
	initialTxsGasEstimation uint64,
	gasBandwidth uint64,
) ([]*txcache.WrappedTransaction, []*txcache.WrappedTransaction) {
	return txs.prefilterTransactions(initialTxs, additionalTxs, initialTxsGasEstimation, gasBandwidth)
}

// CreateAndProcessScheduledMiniBlocksFromMeAsProposer exported function for create process scheduled from me
func (txs *transactions) CreateAndProcessScheduledMiniBlocksFromMeAsProposer(
	haveTime func() bool,
	haveAdditionalTime func() bool,
	sortedTxs []*txcache.WrappedTransaction,
	mapSCTxs map[string]struct{},
) (block.MiniBlockSlice, error) {
	return txs.createAndProcessScheduledMiniBlocksFromMeAsProposer(haveTime, haveAdditionalTime, sortedTxs, mapSCTxs)
}

// SetScheduledTXContinueFunc sets a new scheduled tx verifier function
func (txs *transactions) SetScheduledTXContinueFunc(newFunc func(isShardStuck func(uint32) bool, wrappedTx *txcache.WrappedTransaction, mapSCTxs map[string]struct{}, mbInfo *createScheduledMiniBlocksInfo) (*transaction.Transaction, *block.MiniBlock, bool)) {
	if newFunc != nil {
		txs.scheduledTXContinueFunc = newFunc
	}
}

// ProcessTxsFromMe exported function
func (txs *transactions) ProcessTxsFromMe(
	body *block.Body,
	haveTime func() bool,
	randomness []byte,
) (block.MiniBlockSlice, map[string]struct{}, error) {
	return txs.processTxsFromMe(body, haveTime, randomness)
}
