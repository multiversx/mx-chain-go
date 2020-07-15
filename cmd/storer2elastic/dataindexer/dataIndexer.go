package dataindexer

import (
	"encoding/hex"
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
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("dataindexer")

type dataPreparerAndIndexer struct {
	elasticIndexer   indexer.Indexer
	databaseReader   DatabaseReaderHandler
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
}

type persistersHolder struct {
	miniBlocksPersister           storage.Persister
	transactionPersister          storage.Persister
	unsignedTransactionsPersister storage.Persister
	rewardTransactionsPersister   storage.Persister
}

// New returns a new instance of dataPreparerAndIndexer
func New(
	elasticIndexer indexer.Indexer,
	databaseReader DatabaseReaderHandler,
	shardCoordinator sharding.Coordinator,
	marshalizer marshal.Marshalizer,
) (*dataPreparerAndIndexer, error) {
	if check.IfNil(elasticIndexer) {
		return nil, ErrNilElasticIndexer
	}
	if check.IfNil(databaseReader) {
		return nil, ErrNilDatabaseReader
	}
	if check.IfNil(shardCoordinator) {
		return nil, ErrNilShardCoordinator
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}

	return &dataPreparerAndIndexer{
		elasticIndexer:   elasticIndexer,
		databaseReader:   databaseReader,
		shardCoordinator: shardCoordinator,
		marshalizer:      marshalizer,
	}, nil
}

// Index will prepare the data for indexing and will try to index all the data before the given timeout
func (dpi *dataPreparerAndIndexer) Index(timeout int) error {
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

func (dpi *dataPreparerAndIndexer) startIndexing(errChan chan error) {
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
		dpi.closePersisters(persisters)

		log.Info("database info", "epoch", db.Epoch, "shard", db.Shard)
	}

	errChan <- nil
}

func (dpi *dataPreparerAndIndexer) indexHeader(persisters *persistersHolder, dbInfo *databasereader.DatabaseInfo, hdr data.HeaderHandler, epoch uint32) error {
	miniBlocksHashes := make([]string, 0)
	shardIDs := dpi.getShardIDs()
	for _, shard := range shardIDs {
		miniBlocksForShard := hdr.GetMiniBlockHeadersWithDst(shard)
		for hash := range miniBlocksForShard {
			miniBlocksHashes = append(miniBlocksHashes, hash)
		}
	}

	body, txPool, err := dpi.processBodyAndTxPoolForHeader(persisters, miniBlocksHashes)
	if err != nil {
		return err
	}

	// TODO: investigate how to fill these 2 fields with real values
	singersIndexes := []uint64{1, 0}
	notarizedHdrs := dpi.computeNotarizedHeaders(hdr)
	dpi.elasticIndexer.SaveBlock(body, hdr, txPool, singersIndexes, notarizedHdrs)
	log.Info("indexed header", "epoch", hdr.GetEpoch(), "shard", hdr.GetShardID(), "nonce", hdr.GetNonce())
	return nil
}

func (dpi *dataPreparerAndIndexer) getShardIDs() []uint32 {
	shardIDs := make([]uint32, 0)
	for shard := uint32(0); shard < dpi.shardCoordinator.NumberOfShards(); shard++ {
		shardIDs = append(shardIDs, shard)
	}
	shardIDs = append(shardIDs, core.MetachainShardId)

	return shardIDs
}

func (dpi *dataPreparerAndIndexer) processBodyAndTxPoolForHeader(
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

func (dpi *dataPreparerAndIndexer) processTransactionsForMiniBlock(
	persisters *persistersHolder,
	txHashes [][]byte,
	txPool map[string]data.TransactionHandler,
	mbType block.Type,
) {
	for _, txHash := range txHashes {
		var tx data.TransactionHandler
		var getTxErr error
		switch mbType {
		case block.TxBlock:
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

func (dpi *dataPreparerAndIndexer) getRegularTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
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

func (dpi *dataPreparerAndIndexer) getRewardTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
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

func (dpi *dataPreparerAndIndexer) getUnsignedTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
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

func (dpi *dataPreparerAndIndexer) preparePersistersHolder(dbInfo *databasereader.DatabaseInfo) (*persistersHolder, error) {
	persHold := &persistersHolder{}

	miniBlocksPersister, err := dpi.databaseReader.LoadPersister(dbInfo, "MiniBlocks")
	if err != nil {
		return nil, err
	}
	persHold.miniBlocksPersister = miniBlocksPersister

	txsPersister, err := dpi.databaseReader.LoadPersister(dbInfo, "Transactions")
	if err != nil {
		return nil, err
	}
	persHold.transactionPersister = txsPersister

	uTxsPersister, err := dpi.databaseReader.LoadPersister(dbInfo, "UnsignedTransactions")
	if err != nil {
		return nil, err
	}
	persHold.unsignedTransactionsPersister = uTxsPersister

	rTxsPersister, err := dpi.databaseReader.LoadPersister(dbInfo, "RewardTransactions")
	if err != nil {
		return nil, err
	}
	persHold.rewardTransactionsPersister = rTxsPersister

	return persHold, nil
}

func (dpi *dataPreparerAndIndexer) closePersisters(persisters *persistersHolder) {
	err := persisters.miniBlocksPersister.Close()
	log.LogIfError(err)

	err = persisters.transactionPersister.Close()
	log.LogIfError(err)

	err = persisters.unsignedTransactionsPersister.Close()
	log.LogIfError(err)

	err = persisters.rewardTransactionsPersister.Close()
	log.LogIfError(err)
}

func (dpi *dataPreparerAndIndexer) computeNotarizedHeaders(hdr data.HeaderHandler) []string {
	metaBlock, ok := hdr.(*block.MetaBlock)
	if !ok {
		return []string{}
	}

	numShardInfo := len(metaBlock.ShardInfo)
	notarizedHdrs := make([]string, 0, numShardInfo)
	for i := 0; i < numShardInfo; i++ {
		notarizedHdrs[i] = hex.EncodeToString(metaBlock.ShardInfo[i].HeaderHash)
	}

	return notarizedHdrs
}
