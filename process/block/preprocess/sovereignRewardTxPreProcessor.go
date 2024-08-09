package preprocess

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

type sovereignRewardsTxPreProcessor struct {
	*rewardTxPreprocessor
}

// ArgsRewardTxPreProcessor is a structure placeholder for arguments needed to create a rewards tx pre-processor
type ArgsRewardTxPreProcessor struct {
	RewardTxDataPool           dataRetriever.ShardedDataCacherNotifier
	Store                      dataRetriever.StorageService
	Hasher                     hashing.Hasher
	Marshalizer                marshal.Marshalizer
	RewardProcessor            process.RewardTransactionProcessor
	ShardCoordinator           sharding.Coordinator
	Accounts                   state.AccountsAdapter
	OnRequestRewardTransaction func(shardID uint32, txHashes [][]byte)
	GasHandler                 process.GasHandler
	PubkeyConverter            core.PubkeyConverter
	BlockSizeComputation       BlockSizeComputationHandler
	BalanceComputation         BalanceComputationHandler
	ProcessedMiniBlocksTracker process.ProcessedMiniBlocksTracker
	TxExecutionOrderHandler    common.TxExecutionOrderHandler
}

// NewSovereignRewardsTxPreProcessor creates a sovereign rewards tx pre-processor
func NewSovereignRewardsTxPreProcessor(args ArgsRewardTxPreProcessor) (*sovereignRewardsTxPreProcessor, error) {
	rtp, err := baseCreateRewardTxPreProc(args)
	if err != nil {
		return nil, err
	}

	srtp := &sovereignRewardsTxPreProcessor{
		rtp,
	}
	srtp.rewardTxPool.RegisterOnAdded(srtp.receivedRewardTransaction)
	return srtp, err
}

// receivedRewardTransaction is a callback function called when a new reward transaction
// is added in the reward transactions pool. We do not need to wait for broadcasted txs on a channel, since
// each node is processing these rewards.
func (srtp *sovereignRewardsTxPreProcessor) receivedRewardTransaction(key []byte, value interface{}) {
	tx, ok := value.(data.TransactionHandler)
	if !ok {
		log.Warn("sovereignRewardsTxPreProcessor.receivedRewardTransaction", "error", process.ErrWrongTypeAssertion)
		return
	}

	err := srtp.receivedRewardTx(key, tx, &srtp.rewardTxsForBlock)
	if err != nil {
		log.Error("sovereignRewardsTxPreProcessor.receivedRewardTx", "error", err)
	}
}

func (srtp *sovereignRewardsTxPreProcessor) receivedRewardTx(
	txHash []byte,
	tx data.TransactionHandler,
	forBlock *txsForBlock,
) error {
	if check.IfNil(tx) {
		log.Warn("nil rewardsTransaction in sovereignRewardsTxPreProcessor.receivedRewardTx", "hash", txHash)
		return process.ErrMissingTransaction
	}

	forBlock.mutTxsForBlock.Lock()
	forBlock.txHashAndInfo[string(txHash)] = &txInfo{
		tx: tx,
	}
	forBlock.mutTxsForBlock.Unlock()

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (srtp *sovereignRewardsTxPreProcessor) IsInterfaceNil() bool {
	return srtp == nil
}
