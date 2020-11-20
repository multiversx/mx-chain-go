package process

import (
	"fmt"
	"sort"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

var log = logger.GetOrCreate("update/process/")

// ArgsPendingTransactionProcessor defines the arguments structure
type ArgsPendingTransactionProcessor struct {
	Accounts         state.AccountsAdapter
	TxProcessor      process.TransactionProcessor
	RwdTxProcessor   process.RewardTransactionProcessor
	ScrTxProcessor   process.SmartContractResultProcessor
	PubKeyConv       core.PubkeyConverter
	ShardCoordinator sharding.Coordinator
}

type pendingProcessor struct {
	accounts         state.AccountsAdapter
	txProcessor      process.TransactionProcessor
	rwdTxProcessor   process.RewardTransactionProcessor
	scrTxProcessor   process.SmartContractResultProcessor
	pubKeyConv       core.PubkeyConverter
	shardCoordinator sharding.Coordinator
}

// NewPendingTransactionProcessor creates a pending transaction processor to be used after hardfork import
func NewPendingTransactionProcessor(args ArgsPendingTransactionProcessor) (*pendingProcessor, error) {
	if check.IfNil(args.Accounts) {
		return nil, update.ErrNilAccounts
	}
	if check.IfNil(args.TxProcessor) {
		return nil, update.ErrNilTxProcessor
	}
	if check.IfNil(args.ScrTxProcessor) {
		return nil, update.ErrNilSCRProcessor
	}
	if check.IfNil(args.RwdTxProcessor) {
		return nil, update.ErrNilRwdTxProcessor
	}
	if check.IfNil(args.PubKeyConv) {
		return nil, update.ErrNilPubKeyConverter
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}

	return &pendingProcessor{
		accounts:         args.Accounts,
		txProcessor:      args.TxProcessor,
		rwdTxProcessor:   args.RwdTxProcessor,
		scrTxProcessor:   args.ScrTxProcessor,
		pubKeyConv:       args.PubKeyConv,
		shardCoordinator: args.ShardCoordinator,
	}, nil
}

// ProcessTransactionsDstMe processes all the transactions in which destination is the current shard
func (p *pendingProcessor) ProcessTransactionsDstMe(txsInfo []*update.TxInfo) (block.MiniBlockSlice, error) {
	mapMiniBlocks := make(map[string]*block.MiniBlock)
	for _, txInfo := range txsInfo {
		sndShardID := core.MetachainShardId
		if len(txInfo.Tx.GetSndAddr()) > 0 {
			sndShardID = p.shardCoordinator.ComputeId(txInfo.Tx.GetSndAddr())
		}

		dstShardID := p.shardCoordinator.ComputeId(txInfo.Tx.GetRcvAddr())
		if dstShardID != p.shardCoordinator.SelfId() {
			continue
		}

		blockType, err := p.processSingleTransaction(txInfo)
		if err != nil {
			log.Debug("could not process transaction",
				"err", err,
				"snd", txInfo.Tx.GetSndAddr(),
				"rcv", txInfo.Tx.GetRcvAddr(),
				"value", txInfo.Tx.GetValue().String(),
				"data", txInfo.Tx.GetData())
			continue
		}

		localID := fmt.Sprintf("%d_%d_%d", sndShardID, dstShardID, blockType)
		mb, ok := mapMiniBlocks[localID]
		if !ok {
			mb = &block.MiniBlock{
				TxHashes:        make([][]byte, 0),
				ReceiverShardID: dstShardID,
				SenderShardID:   sndShardID,
				Type:            blockType,
			}
			mapMiniBlocks[localID] = mb
		}

		mb.TxHashes = append(mb.TxHashes, txInfo.TxHash)
	}

	_, err := p.accounts.Commit()
	if err != nil {
		return nil, err
	}

	miniBlocks := getSortedSliceFromMbsMap(mapMiniBlocks)
	return miniBlocks, nil
}

func getSortedSliceFromMbsMap(mbsMap map[string]*block.MiniBlock) block.MiniBlockSlice {
	miniBlocks := make(block.MiniBlockSlice, 0, len(mbsMap))
	for _, mb := range mbsMap {
		miniBlocks = append(miniBlocks, mb)
	}

	sort.Slice(miniBlocks, func(i, j int) bool {
		if miniBlocks[i].SenderShardID == miniBlocks[j].SenderShardID {
			return miniBlocks[i].Type < miniBlocks[j].Type
		}
		return miniBlocks[i].SenderShardID < miniBlocks[j].SenderShardID
	})

	return miniBlocks
}

func (p *pendingProcessor) processSingleTransaction(txInfo *update.TxInfo) (block.Type, error) {
	rwdTx, ok := txInfo.Tx.(*rewardTx.RewardTx)
	if ok {
		err := p.rwdTxProcessor.ProcessRewardTransaction(rwdTx)
		if err != nil {
			return 0, err
		}
		return block.RewardsBlock, nil
	}

	scrTx, ok := txInfo.Tx.(*smartContractResult.SmartContractResult)
	if ok {
		_, err := p.scrTxProcessor.ProcessSmartContractResult(scrTx)
		if err != nil {
			return 0, err
		}
		return block.SmartContractResultBlock, nil
	}

	tx, ok := txInfo.Tx.(*transaction.Transaction)
	if ok {
		_, err := p.txProcessor.ProcessTransaction(tx)
		if err != nil {
			return 0, err
		}
		return block.TxBlock, nil
	}

	return block.InvalidBlock, nil
}

// RootHash returns the roothash of the accounts
func (p *pendingProcessor) RootHash() ([]byte, error) {
	return p.accounts.RootHash()
}

// IsInterfaceNil returns true if underlying object is nil
func (p *pendingProcessor) IsInterfaceNil() bool {
	return p == nil
}
