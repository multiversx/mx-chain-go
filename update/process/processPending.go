package process

import (
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
func (p *pendingProcessor) ProcessTransactionsDstMe(mbInfo *update.MbInfo) (*block.MiniBlock, error) {
	miniBlock := &block.MiniBlock{
		TxHashes:        make([][]byte, 0),
		ReceiverShardID: mbInfo.ReceiverShardID,
		SenderShardID:   mbInfo.SenderShardID,
		Type:            mbInfo.Type,
	}

	for _, txInfo := range mbInfo.TxsInfo {
		err := p.processSingleTransaction(txInfo)
		if err != nil {
			log.Debug("could not process transaction",
				"err", err,
				"snd", txInfo.Tx.GetSndAddr(),
				"rcv", txInfo.Tx.GetRcvAddr(),
				"value", txInfo.Tx.GetValue().String(),
				"data", txInfo.Tx.GetData())
			continue
		}

		miniBlock.TxHashes = append(miniBlock.TxHashes, txInfo.TxHash)
	}

	return miniBlock, nil
}

func (p *pendingProcessor) processSingleTransaction(txInfo *update.TxInfo) error {
	rwdTx, ok := txInfo.Tx.(*rewardTx.RewardTx)
	if ok {
		err := p.rwdTxProcessor.ProcessRewardTransaction(rwdTx)
		if err != nil {
			return err
		}
		return nil
	}

	scrTx, ok := txInfo.Tx.(*smartContractResult.SmartContractResult)
	if ok {
		_, err := p.scrTxProcessor.ProcessSmartContractResult(scrTx)
		if err != nil {
			return err
		}
		return nil
	}

	tx, ok := txInfo.Tx.(*transaction.Transaction)
	if ok {
		_, err := p.txProcessor.ProcessTransaction(tx)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

// RootHash returns the roothash of the accounts
func (p *pendingProcessor) RootHash() ([]byte, error) {
	return p.accounts.RootHash()
}

// Commit commits the changes of the accounts
func (p *pendingProcessor) Commit() ([]byte, error) {
	return p.accounts.Commit()
}

// IsInterfaceNil returns true if underlying object is nil
func (p *pendingProcessor) IsInterfaceNil() bool {
	return p == nil
}
