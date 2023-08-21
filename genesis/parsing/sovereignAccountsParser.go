package parsing

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/sharding"
)

type sovereignAccountsParser struct {
	*accountsParser
}

// NewSovereignAccountsParser creates an account parser for sovereign shard
func NewSovereignAccountsParser(accParser *accountsParser) (*sovereignAccountsParser, error) {
	if check.IfNil(accParser) {
		return nil, errNilAccountsParser
	}

	return &sovereignAccountsParser{
		accountsParser: accParser,
	}, nil
}

// GenerateInitialTransactions will generate initial transactions pool and the miniblocks for the generated transactions
func (ap *sovereignAccountsParser) GenerateInitialTransactions(
	shardCoordinator sharding.Coordinator,
	indexingData map[uint32]*genesis.IndexingData,
) ([]*block.MiniBlock, map[uint32]*outportcore.TransactionPool, error) {
	if check.IfNil(shardCoordinator) {
		return nil, nil, genesis.ErrNilShardCoordinator
	}

	shardIDs := make([]uint32, 1)
	shardIDs[0] = core.SovereignChainShardId
	txsPoolPerShard := ap.createIndexerPools(shardIDs)

	mintTxs := ap.createMintTransactions()

	allTxs := ap.getAllTxs(indexingData)
	allTxs = append(allTxs, mintTxs...)
	miniBlocks := createMiniBlocks(shardIDs, block.TxBlock)

	err := ap.setTxsPoolAndMiniBlocks(shardCoordinator, allTxs, txsPoolPerShard, miniBlocks)
	if err != nil {
		return nil, nil, err
	}

	ap.setScrsTxsPool(shardCoordinator, indexingData, txsPoolPerShard)

	miniBlocks = getNonEmptyMiniBlocks(miniBlocks)

	return miniBlocks, txsPoolPerShard, nil
}

func (ap *sovereignAccountsParser) setTxsPoolAndMiniBlocks(
	shardCoordinator sharding.Coordinator,
	allTxs []coreData.TransactionHandler,
	txsPoolPerShard map[uint32]*outportcore.TransactionPool,
	miniBlocks []*block.MiniBlock,
) error {
	receiverShardID, senderShardID := shardCoordinator.SelfId(), shardCoordinator.SelfId()

	for _, txHandler := range allTxs {
		err := ap.setTxPoolAndMiniBlock(txsPoolPerShard, miniBlocks, txHandler, senderShardID, receiverShardID)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if the underlying object is nil
func (ap *sovereignAccountsParser) IsInterfaceNil() bool {
	return ap == nil
}
