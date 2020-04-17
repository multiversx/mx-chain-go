package genesis

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
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
func (p *pendingProcessor) ProcessTransactionsDstMe(mapTxs map[string]data.TransactionHandler) error {
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

	for _, info := range sortedTxs {
		address, err := p.pubKeyConv.CreateAddressFromBytes(info.tx.GetRcvAddr())
		if err != nil {
			continue
		}

		dstShardID := p.shardCoordinator.ComputeId(address)
		if dstShardID != p.shardCoordinator.SelfId() {
			continue
		}

		err = p.processSingleTransaction(info)
		if err != nil {
			continue
		}
	}

	_, err := p.accounts.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (p *pendingProcessor) processSingleTransaction(info *txInfo) error {
	rwdTx, ok := info.tx.(*rewardTx.RewardTx)
	if ok {
		return p.rwdTxProcessor.ProcessRewardTransaction(rwdTx)
	}

	scrTx, ok := info.tx.(*smartContractResult.SmartContractResult)
	if ok {
		return p.scrTxProcessor.ProcessSmartContractResult(scrTx)
	}

	tx, ok := info.tx.(*transaction.Transaction)
	if ok {
		return p.txProcessor.ProcessTransaction(tx)
	}

	return nil
}

// RootHash returns the roothash of the accounts
func (p *pendingProcessor) RootHash() ([]byte, error) {
	return p.accounts.RootHash()
}

// IsInterfaceNil returns true if underlying object is nil
func (p *pendingProcessor) IsInterfaceNil() bool {
	return p == nil
}
