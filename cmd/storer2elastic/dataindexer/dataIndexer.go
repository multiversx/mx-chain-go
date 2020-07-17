package dataindexer

import (
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("dataindexer")

type dataProcessor struct {
	elasticIndexer    indexer.Indexer
	databaseReader    DatabaseReaderHandler
	shardCoordinator  sharding.Coordinator
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	nodesCoordinators map[uint32]NodesCoordinator
	hdrsToIndexLater  []data.HeaderHandler
}

type persistersHolder struct {
	miniBlocksPersister           storage.Persister
	transactionPersister          storage.Persister
	unsignedTransactionsPersister storage.Persister
	rewardTransactionsPersister   storage.Persister
}

// Args holds the arguments needed for creating a new dataProcessor
type Args struct {
	ElasticIndexer    indexer.Indexer
	DatabaseReader    DatabaseReaderHandler
	ShardCoordinator  sharding.Coordinator
	Marshalizer       marshal.Marshalizer
	Hasher            hashing.Hasher
	GenesisNodesSetup sharding.GenesisNodesSetupHandler
}

// New returns a new instance of dataProcessor
func New(args Args) (*dataProcessor, error) {
	if check.IfNil(args.ElasticIndexer) {
		return nil, ErrNilElasticIndexer
	}
	if check.IfNil(args.DatabaseReader) {
		return nil, ErrNilDatabaseReader
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(args.GenesisNodesSetup) {
		return nil, ErrNilGenesisNodesSetup
	}

	dpi := &dataProcessor{
		elasticIndexer:   args.ElasticIndexer,
		databaseReader:   args.DatabaseReader,
		shardCoordinator: args.ShardCoordinator,
		marshalizer:      args.Marshalizer,
		hasher:           args.Hasher,
		hdrsToIndexLater: make([]data.HeaderHandler, 0),
	}

	nodesCoordinator, err := dpi.createNodesCoordinators(args.GenesisNodesSetup)
	if err != nil {
		return nil, err
	}

	dpi.nodesCoordinators = nodesCoordinator
	return dpi, nil
}

// Index will prepare the data for indexing and will try to index all the data before the given timeout
func (dpi *dataProcessor) Index(timeout int) error {
	errChan := make(chan error, 0)
	go func() {
		dpi.startIndexing(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(time.Duration(timeout) * time.Second):
		return ErrTimeIsOut
	}
}

func (dpi *dataProcessor) startIndexing(errChan chan error) {
	records, err := dpi.databaseReader.GetDatabaseInfo()
	if err != nil {
		errChan <- err
		return
	}

	for _, db := range records {
		hdrs, err := dpi.databaseReader.GetHeaders(db)
		if err != nil {
			errChan <- err
			return
		}

		persisters, err := dpi.preparePersistersHolder(db)
		for _, hdr := range hdrs {
			err = dpi.indexHeader(persisters, db, hdr, db.Epoch)
			if err != nil {
				log.Warn("cannot index header", "error", err)
			}
		}
		for _, hdr := range dpi.hdrsToIndexLater {
			err = dpi.indexHeader(persisters, db, hdr, db.Epoch)
			if err != nil {
				log.Warn("cannot index header", "error", err)
			}
		}
		dpi.closePersisters(persisters)

		log.Info("database info", "epoch", db.Epoch, "shard", db.Shard)
	}

	errChan <- nil
}

func (dpi *dataProcessor) indexHeader(persisters *persistersHolder, dbInfo *databasereader.DatabaseInfo, hdr data.HeaderHandler, epoch uint32) error {
	if metaBlock, ok := hdr.(*block.MetaBlock); ok {
		if metaBlock.IsStartOfEpochBlock() {
			dpi.processValidatorsForEpoch(metaBlock.Epoch, metaBlock, persisters.miniBlocksPersister)
		}
	}

	if !dpi.canIndexHeaderNow(hdr) {
		dpi.hdrsToIndexLater = append(dpi.hdrsToIndexLater, hdr)
		return fmt.Errorf("cannot index header now")
	}

	miniBlocksHashes := make([]string, 0)
	shardIDs := dpi.getShardIDs()
	for _, shard := range shardIDs {
		miniBlocksForShard := hdr.GetMiniBlockHeadersWithDst(shard)
		for hash := range miniBlocksForShard {
			miniBlocksHashes = append(miniBlocksHashes, hash)
		}
	}

	body, txPool, err := dpi.processBodyAndTransactionsPoolForHeader(persisters, miniBlocksHashes)
	if err != nil {
		return err
	}

	singersIndexes, err := dpi.computeSignersIndexes(hdr)
	if err != nil {
		return err
	}
	notarizedHdrs := dpi.computeNotarizedHeaders(hdr)
	dpi.elasticIndexer.SaveBlock(body, hdr, txPool, singersIndexes, notarizedHdrs)
	log.Info("indexed header", "epoch", hdr.GetEpoch(), "shard", hdr.GetShardID(), "nonce", hdr.GetNonce())
	return nil
}

func (dpi *dataProcessor) getShardIDs() []uint32 {
	shardIDs := make([]uint32, 0)
	for shard := uint32(0); shard < dpi.shardCoordinator.NumberOfShards(); shard++ {
		shardIDs = append(shardIDs, shard)
	}
	shardIDs = append(shardIDs, core.MetachainShardId)

	return shardIDs
}

func (dpi *dataProcessor) processBodyAndTransactionsPoolForHeader(
	persisters *persistersHolder,
	mbHashes []string,
) (*block.Body, map[string]data.TransactionHandler, error) {
	txPool := make(map[string]data.TransactionHandler)
	mbUnit := persisters.miniBlocksPersister

	blockBody := &block.Body{}
	for _, mbHash := range mbHashes {
		mbBytes, err := mbUnit.Get([]byte(mbHash))
		if err != nil {
			log.Warn("miniblock not found in storage", "hash", []byte(mbHash))
			continue
		}

		recoveredMiniBlock := &block.MiniBlock{}
		err = dpi.marshalizer.Unmarshal(recoveredMiniBlock, mbBytes)
		if err != nil {
			log.Warn("cannot unmarshal miniblock", "hash", mbHash, "error", err)
			continue
		}

		dpi.processTransactionsForMiniBlock(persisters, recoveredMiniBlock.TxHashes, txPool, recoveredMiniBlock.Type)
		blockBody.MiniBlocks = append(blockBody.MiniBlocks, recoveredMiniBlock)
	}

	return blockBody, txPool, nil
}

func (dpi *dataProcessor) processTransactionsForMiniBlock(
	persisters *persistersHolder,
	txHashes [][]byte,
	txPool map[string]data.TransactionHandler,
	mbType block.Type,
) {
	for _, txHash := range txHashes {
		var tx data.TransactionHandler
		var getTxErr error
		switch mbType {
		case block.TxBlock, block.InvalidBlock:
			tx, getTxErr = dpi.getRegularTx(persisters, txHash)
		case block.RewardsBlock:
			tx, getTxErr = dpi.getRewardTx(persisters, txHash)
		case block.SmartContractResultBlock:
			tx, getTxErr = dpi.getUnsignedTx(persisters, txHash)
		}
		if getTxErr != nil || tx == nil {
			continue
		}

		txPool[string(txHash)] = tx
	}
}

func (dpi *dataProcessor) getRegularTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
	txBytes, err := holder.transactionPersister.Get(txHash)
	if err != nil {
		log.Warn("cannot get tx from storer", "txHash", txHash)
		return nil, err
	}

	recoveredTx := &transaction.Transaction{}
	err = dpi.marshalizer.Unmarshal(recoveredTx, txBytes)
	if err != nil {
		log.Warn("cannot unmarshal tx", "txHash", txHash, "error", err)
		return nil, err
	}

	return recoveredTx, nil
}

func (dpi *dataProcessor) getRewardTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
	txBytes, err := holder.rewardTransactionsPersister.Get(txHash)
	if err != nil {
		log.Warn("cannot get tx from storer", "txHash", txHash)
		return nil, err
	}

	recoveredTx := &rewardTx.RewardTx{}
	err = dpi.marshalizer.Unmarshal(recoveredTx, txBytes)
	if err != nil {
		log.Warn("cannot unmarshal reward tx", "txHash", txHash, "error", err)
		return nil, err
	}

	return recoveredTx, nil
}

func (dpi *dataProcessor) getUnsignedTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
	txBytes, err := holder.unsignedTransactionsPersister.Get(txHash)
	if err != nil {
		log.Warn("cannot get unsigned tx from storer", "txHash", txHash)
		return nil, err
	}

	recoveredTx := &smartContractResult.SmartContractResult{}
	err = dpi.marshalizer.Unmarshal(recoveredTx, txBytes)
	if err != nil {
		log.Warn("cannot unmarshal unsigned tx", "txHash", txHash, "error", err)
		return nil, err
	}

	return recoveredTx, nil
}
