package process

import (
	"fmt"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

// ArgsPendingTransactionProcessor defines the arguments structure
type ArgsPendingTransactionProcessor struct {
	Accounts         state.AccountsAdapter
	TxProcessor      process.TransactionProcessor
	RwdTxProcessor   process.RewardTransactionProcessor
	ScrTxProcessor   process.SmartContractResultProcessor
	PubKeyConv       state.PubkeyConverter
	ShardCoordinator sharding.Coordinator
	TxCoordinator    process.TransactionCoordinator
}

type pendingProcessor struct {
	accounts         state.AccountsAdapter
	txProcessor      process.TransactionProcessor
	rwdTxProcessor   process.RewardTransactionProcessor
	scrTxProcessor   process.SmartContractResultProcessor
	pubKeyConv       state.PubkeyConverter
	shardCoordinator sharding.Coordinator
}

type txInfo struct {
	hash string
	tx   data.TransactionHandler
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
		return nil, update.ErrNilPubkeyConverter
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
func (p *pendingProcessor) ProcessTransactionsDstMe(mapTxs map[string]data.TransactionHandler) (block.MiniBlockSlice, error) {
	sortedTxs := make([]*txInfo, 0, len(mapTxs))
	for hash, tx := range mapTxs {
		sortedTxs = append(sortedTxs, &txInfo{
			hash: hash,
			tx:   tx,
		})
	}

	sort.Slice(sortedTxs, func(i, j int) bool {
		return sortedTxs[i].hash < sortedTxs[j].hash
	})

	mapMiniBlocks := make(map[string]*block.MiniBlock)
	for _, info := range sortedTxs {
		rcvAddress, err := p.pubKeyConv.CreateAddressFromBytes(info.tx.GetRcvAddr())
		if err != nil {
			continue
		}

		sndAddress, err := p.pubKeyConv.CreateAddressFromBytes(info.tx.GetSndAddr())
		if err != nil {
			continue
		}

		sndShardId := p.shardCoordinator.ComputeId(sndAddress)

		dstShardID := p.shardCoordinator.ComputeId(rcvAddress)
		if dstShardID != p.shardCoordinator.SelfId() {
			continue
		}

		blockType, err := p.processSingleTransaction(info)
		if err != nil {
			continue
		}

		localID := fmt.Sprint("%d%d%d", sndShardId, dstShardID, blockType)
		mb, ok := mapMiniBlocks[localID]
		if !ok {
			mb = &block.MiniBlock{
				TxHashes:        make([][]byte, 0),
				ReceiverShardID: dstShardID,
				SenderShardID:   sndShardId,
				Type:            blockType,
			}
		}

		mb.TxHashes = append(mb.TxHashes, []byte(info.hash))
	}

	_, err := p.accounts.Commit()
	if err != nil {
		return nil, err
	}

	miniBlocks := make(block.MiniBlockSlice, 0, len(mapMiniBlocks))
	for _, mb := range mapMiniBlocks {
		miniBlocks = append(miniBlocks, mb)
	}

	sort.Slice(miniBlocks, func(i, j int) bool {
		if miniBlocks[i].SenderShardID == miniBlocks[j].SenderShardID {
			return miniBlocks[i].Type < miniBlocks[j].Type
		}
		return miniBlocks[i].SenderShardID < miniBlocks[j].SenderShardID
	})

	return miniBlocks, nil
}

func (p *pendingProcessor) processSingleTransaction(info *txInfo) (block.Type, error) {
	rwdTx, ok := info.tx.(*rewardTx.RewardTx)
	if ok {
		err := p.rwdTxProcessor.ProcessRewardTransaction(rwdTx)
		if err != nil {
			return 0, err
		}
		return block.RewardsBlock, nil
	}

	scrTx, ok := info.tx.(*smartContractResult.SmartContractResult)
	if ok {
		err := p.scrTxProcessor.ProcessSmartContractResult(scrTx)
		if err != nil {
			return 0, err
		}
		return block.SmartContractResultBlock, nil
	}

	tx, ok := info.tx.(*transaction.Transaction)
	if ok {
		err := p.txProcessor.ProcessTransaction(tx)
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
