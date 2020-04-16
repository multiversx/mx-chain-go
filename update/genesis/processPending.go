package genesis

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgsPendingTransactionProcessor defines the arguments structure
type ArgsPendingTransactionProcessor struct {
	Accounts         state.AccountsDB
	TxProcessor      process.TransactionProcessor
	RwdTxProcessor   process.RewardTransactionProcessor
	ScrTxProcessor   process.SmartContractResultProcessor
	PubKeyConv       state.PubkeyConverter
	ShardCoordinator sharding.Coordinator
}

type pendingProcessor struct {
	accounts         state.AccountsDB
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

	for _, txInfo := range sortedTxs {
		address, err := p.pubKeyConv.CreateAddressFromBytes(txInfo.tx.GetRcvAddr())
		if err != nil {
			continue
		}

		dstShardID := p.shardCoordinator.ComputeId(address)
		if dstShardID != p.shardCoordinator.SelfId() {
			continue
		}

	}

	_, err := p.accounts.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (p *pendingProcessor) RootHash() ([]byte, error) {
	return p.accounts.RootHash()
}

// IsInterfaceNil returns true if underlying object is nil
func (p *pendingProcessor) IsInterfaceNil() bool {
	return p == nil
}
