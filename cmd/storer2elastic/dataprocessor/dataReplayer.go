package dataprocessor

import (
	"time"

	storer2ElasticData "github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/data"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type dataReplayer struct {
	databaseReader   DatabaseReaderHandler
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	//nodesCoordinators map[uint32]NodesCoordinator
	uint64Converter   typeConverters.Uint64ByteSliceConverter
	headerMarshalizer HeaderMarshalizerHandler
}

type persistersHolder struct {
	miniBlocksPersister           storage.Persister
	transactionPersister          storage.Persister
	unsignedTransactionsPersister storage.Persister
	rewardTransactionsPersister   storage.Persister
	shardHeadersPersister         storage.Persister
}

type metaBlocksPersistersHolder struct {
	hdrHashNoncePersister storage.Persister
	metaBlocksPersister   storage.Persister
}

// DataReplayerArgs holds the arguments needed for creating a new dataReplayer
type DataReplayerArgs struct {
	ElasticIndexer           indexer.Indexer
	DatabaseReader           DatabaseReaderHandler
	ShardCoordinator         sharding.Coordinator
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	HeaderMarshalizer        HeaderMarshalizerHandler
}

// NewDataReplayer returns a new instance of dataReplayer
func NewDataReplayer(args DataReplayerArgs) (*dataReplayer, error) {
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
	if check.IfNil(args.Uint64ByteSliceConverter) {
		return nil, ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(args.HeaderMarshalizer) {
		return nil, ErrNilHeaderMarshalizer
	}

	return &dataReplayer{
		databaseReader:    args.DatabaseReader,
		shardCoordinator:  args.ShardCoordinator,
		marshalizer:       args.Marshalizer,
		hasher:            args.Hasher,
		uint64Converter:   args.Uint64ByteSliceConverter,
		headerMarshalizer: args.HeaderMarshalizer,
	}, nil
}

// Range will range over the data in storage until the handler returns false or the time is out
func (dr *dataReplayer) Range(handler func(persistedData storer2ElasticData.RoundPersistedData) bool) error {
	errChan := make(chan error, 0)
	go func() {
		dr.startIndexing(errChan, handler)
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(time.Hour):
		return ErrTimeIsOut
	}
}

func (dr *dataReplayer) startIndexing(errChan chan error, persistedDataHandler func(persistedData storer2ElasticData.RoundPersistedData) bool) {
	records, err := dr.databaseReader.GetDatabaseInfo()
	if err != nil {
		errChan <- err
		return
	}

	metachainRecords, err := getMetaChainDatabasesInfo(records)
	if err != nil {
		errChan <- err
		return
	}

	for _, metaDB := range metachainRecords {
		err = dr.processMetaChainDatabase(metaDB, records, persistedDataHandler)
		if err != nil {
			errChan <- err
			return
		}
	}

	errChan <- nil
}

func (dr *dataReplayer) processMetaChainDatabase(
	record *databasereader.DatabaseInfo,
	dbsInfo []*databasereader.DatabaseInfo,
	persistedDataHandler func(persistedData storer2ElasticData.RoundPersistedData) bool,
) error {
	metaHeadersPersisters, err := dr.prepareMetaPersistersHolder(record)
	if err != nil {
		return err
	}

	defer func() {
		dr.closeMetaPersisters(metaHeadersPersisters)
	}()

	shardPersistersHolder := make(map[uint32]*persistersHolder)
	for shardID := uint32(0); shardID < dr.shardCoordinator.NumberOfShards(); shardID++ {
		shardDBInfo, err := getShardDatabaseForEpoch(dbsInfo, record.Epoch, shardID)
		if err != nil {
			return err
		}
		shardPersistersHolder[shardID], err = dr.preparePersistersHolder(shardDBInfo)
		if err != nil {
			return err
		}
	}

	defer func() {
		for _, shardPersisters := range shardPersistersHolder {
			dr.closePersisters(shardPersisters)
		}
	}()

	metachainPersisters, err := dr.preparePersistersHolder(record)
	if err != nil {
		return err
	}

	defer func() {
		dr.closePersisters(metachainPersisters)
	}()

	epochStartMetaBlock, err := dr.getEpochStartMetaBlock(record, metaHeadersPersisters)
	if err != nil {
		return err
	}

	//dr.processValidatorsForEpoch(epochStartMetaBlock.Epoch, epochStartMetaBlock, metachainPersisters.miniBlocksPersister)

	startingNonce := epochStartMetaBlock.Nonce
	roundData, err := dr.processMetaBlock(
		epochStartMetaBlock,
		dbsInfo,
		record,
		metaHeadersPersisters,
		metachainPersisters,
		shardPersistersHolder,
	)
	if err != nil {
		return err
	}
	if !persistedDataHandler(*roundData) {
		return ErrRangeIsOver
	}

	for {
		startingNonce++
		metaBlock, err := dr.getMetaBlockForNonce(startingNonce, record, metaHeadersPersisters)
		if err == nil {
			roundData, err = dr.processMetaBlock(
				metaBlock,
				dbsInfo,
				record,
				metaHeadersPersisters,
				metachainPersisters,
				shardPersistersHolder,
			)
			if err != nil {
				log.Warn(err.Error())
				break
			}
			if !persistedDataHandler(*roundData) {
				return ErrRangeIsOver
			}
		} else {
			log.Error(err.Error())
			break
		}
	}

	log.Info("finished indexing all headers from an epoch", "epoch", record.Epoch)
	return nil
}

func (dr *dataReplayer) prepareMetaPersistersHolder(record *databasereader.DatabaseInfo) (*metaBlocksPersistersHolder, error) {
	metaPersHolder := &metaBlocksPersistersHolder{}

	metaBlocksUnit, err := dr.databaseReader.LoadPersister(record, "MetaBlock")
	if err != nil {
		return nil, err
	}
	metaPersHolder.metaBlocksPersister = metaBlocksUnit

	headerHashNonceUnit, err := dr.databaseReader.LoadStaticPersister(record, "MetaHdrHashNonce")
	if err != nil {
		return nil, err
	}
	metaPersHolder.hdrHashNoncePersister = headerHashNonceUnit

	return metaPersHolder, nil
}

func (dr *dataReplayer) closeMetaPersisters(persistersHolder *metaBlocksPersistersHolder) {
	err := persistersHolder.metaBlocksPersister.Close()
	log.LogIfError(err)

	err = persistersHolder.hdrHashNoncePersister.Close()
	log.LogIfError(err)
}

func (dr *dataReplayer) getEpochStartMetaBlock(record *databasereader.DatabaseInfo, metaPersisters *metaBlocksPersistersHolder) (*block.MetaBlock, error) {
	epochStartIdentifier := core.EpochStartIdentifier(record.Epoch)
	metaBlockBytes, err := metaPersisters.metaBlocksPersister.Get([]byte(epochStartIdentifier))
	if err == nil && len(metaBlockBytes) > 0 {
		return dr.headerMarshalizer.UnmarshalMetaBlock(metaBlockBytes)
	}
	log.Warn("epoch start meta block not found", "epoch", record.Epoch)

	// header should be found by the epoch start identifier. if not, try getting the nonce 0 (the genesis block)
	return dr.getMetaBlockForNonce(0, record, metaPersisters)
}

func (dr *dataReplayer) getMetaBlockForNonce(nonce uint64, record *databasereader.DatabaseInfo, metaPersisters *metaBlocksPersistersHolder) (*block.MetaBlock, error) {
	nonceBytes := dr.uint64Converter.ToByteSlice(nonce)
	metaBlockHash, err := metaPersisters.hdrHashNoncePersister.Get(nonceBytes)
	if err != nil {
		return nil, err
	}

	metaBlockBytes, err := metaPersisters.metaBlocksPersister.Get(metaBlockHash)
	if err != nil {
		return nil, err
	}

	return dr.headerMarshalizer.UnmarshalMetaBlock(metaBlockBytes)
}

func (dr *dataReplayer) processMetaBlock(
	metaBlock *block.MetaBlock,
	dbsInfo []*databasereader.DatabaseInfo,
	record *databasereader.DatabaseInfo,
	metaBlocksPersisters *metaBlocksPersistersHolder,
	persisters *persistersHolder,
	shardPersisters map[uint32]*persistersHolder,
) (*storer2ElasticData.RoundPersistedData, error) {
	metaHdrData, err := dr.processHeader(persisters, record, metaBlock, record.Epoch)
	if err != nil {
		return nil, err
	}

	shardsHeaderData := make(map[uint32]*storer2ElasticData.HeaderData)
	for _, shardInfo := range metaBlock.ShardInfo {
		shardHdrData, err := dr.processShardInfo(dbsInfo, &shardInfo, metaBlock.Epoch, shardPersisters[shardInfo.ShardID])
		if err != nil {
			continue
		}

		shardsHeaderData[shardInfo.ShardID] = shardHdrData
	}

	return &storer2ElasticData.RoundPersistedData{
		MetaBlockData: metaHdrData,
		ShardHeaders:  shardsHeaderData,
	}, nil
}

func (dr *dataReplayer) processShardInfo(
	dbsInfos []*databasereader.DatabaseInfo,
	shardInfo *block.ShardData,
	epoch uint32,
	shardPersisters *persistersHolder,
) (*storer2ElasticData.HeaderData, error) {
	shardDBInfo, err := getShardDatabaseForEpoch(dbsInfos, epoch, shardInfo.ShardID)
	if err != nil {
		return nil, err
	}

	shardHeader, err := dr.getShardHeader(shardPersisters, shardInfo.HeaderHash)
	if err != nil {
		// if new epoch, shard headers can be found in the previous epoch's storage
		// TODO: for this case return and use all the persisters from the previous epoch in order to find txs
		shardHeader, err = dr.getFromShardStorer(dbsInfos, shardInfo, epoch-1)
		if err != nil {
			return nil, err
		}
	}

	return dr.processHeader(shardPersisters, shardDBInfo, shardHeader, epoch)
}

func (dr *dataReplayer) getFromShardStorer(dbsInfos []*databasereader.DatabaseInfo, shardInfo *block.ShardData, epoch uint32) (*block.Header, error) {
	shardDBInfo, err := getShardDatabaseForEpoch(dbsInfos, epoch, shardInfo.ShardID)
	if err != nil {
		return nil, err
	}

	shardPersisters, err := dr.preparePersistersHolder(shardDBInfo)
	if err != nil {
		return nil, err
	}

	defer func() {
		dr.closePersisters(shardPersisters)
	}()

	return dr.getShardHeader(shardPersisters, shardInfo.HeaderHash)
}

func (dr *dataReplayer) getShardHeader(
	shardPersisters *persistersHolder,
	hash []byte,
) (*block.Header, error) {
	shardHeaderBytes, err := shardPersisters.shardHeadersPersister.Get(hash)
	if err != nil {
		return nil, err
	}

	return dr.headerMarshalizer.UnmarshalShardHeader(shardHeaderBytes)
}

func (dr *dataReplayer) processHeader(persisters *persistersHolder, dbInfo *databasereader.DatabaseInfo, hdr data.HeaderHandler, epoch uint32) (*storer2ElasticData.HeaderData, error) {
	miniBlocksHashes := make([]string, 0)
	shardIDs := dr.getShardIDs()
	for _, shard := range shardIDs {
		miniBlocksForShard := hdr.GetMiniBlockHeadersWithDst(shard)
		for hash := range miniBlocksForShard {
			miniBlocksHashes = append(miniBlocksHashes, hash)
		}
	}

	body, txPool, err := dr.processBodyAndTransactionsPoolForHeader(persisters, miniBlocksHashes)
	if err != nil {
		return nil, err
	}

	return &storer2ElasticData.HeaderData{
		Header:           hdr,
		Body:             body,
		TransactionsPool: txPool,
	}, nil
}

func (dr *dataReplayer) getShardIDs() []uint32 {
	shardIDs := make([]uint32, 0)
	for shard := uint32(0); shard < dr.shardCoordinator.NumberOfShards(); shard++ {
		shardIDs = append(shardIDs, shard)
	}
	shardIDs = append(shardIDs, core.MetachainShardId)

	return shardIDs
}

func (dr *dataReplayer) processBodyAndTransactionsPoolForHeader(
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
		err = dr.marshalizer.Unmarshal(recoveredMiniBlock, mbBytes)
		if err != nil {
			log.Warn("cannot unmarshal miniblock", "hash", mbHash, "error", err)
			continue
		}

		dr.processTransactionsForMiniBlock(persisters, recoveredMiniBlock.TxHashes, txPool, recoveredMiniBlock.Type)
		blockBody.MiniBlocks = append(blockBody.MiniBlocks, recoveredMiniBlock)
	}

	return blockBody, txPool, nil
}

func (dr *dataReplayer) processTransactionsForMiniBlock(
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
			tx, getTxErr = dr.getRegularTx(persisters, txHash)
		case block.RewardsBlock:
			tx, getTxErr = dr.getRewardTx(persisters, txHash)
		case block.SmartContractResultBlock:
			tx, getTxErr = dr.getUnsignedTx(persisters, txHash)
		}
		if getTxErr != nil || tx == nil {
			continue
		}

		txPool[string(txHash)] = tx
	}
}

func (dr *dataReplayer) getRegularTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
	txBytes, err := holder.transactionPersister.Get(txHash)
	if err != nil {
		log.Warn("cannot get tx from storer", "txHash", txHash)
		return nil, err
	}

	recoveredTx := &transaction.Transaction{}
	err = dr.marshalizer.Unmarshal(recoveredTx, txBytes)
	if err != nil {
		log.Warn("cannot unmarshal tx", "txHash", txHash, "error", err)
		return nil, err
	}

	return recoveredTx, nil
}

func (dr *dataReplayer) getRewardTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
	txBytes, err := holder.rewardTransactionsPersister.Get(txHash)
	if err != nil {
		log.Warn("cannot get tx from storer", "txHash", txHash)
		return nil, err
	}

	recoveredTx := &rewardTx.RewardTx{}
	err = dr.marshalizer.Unmarshal(recoveredTx, txBytes)
	if err != nil {
		log.Warn("cannot unmarshal reward tx", "txHash", txHash, "error", err)
		return nil, err
	}

	return recoveredTx, nil
}

func (dr *dataReplayer) getUnsignedTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
	txBytes, err := holder.unsignedTransactionsPersister.Get(txHash)
	if err != nil {
		log.Warn("cannot get unsigned tx from storer", "txHash", txHash)
		return nil, err
	}

	recoveredTx := &smartContractResult.SmartContractResult{}
	err = dr.marshalizer.Unmarshal(recoveredTx, txBytes)
	if err != nil {
		log.Warn("cannot unmarshal unsigned tx", "txHash", txHash, "error", err)
		return nil, err
	}

	return recoveredTx, nil
}

func (dr *dataReplayer) preparePersistersHolder(dbInfo *databasereader.DatabaseInfo) (*persistersHolder, error) {
	persHold := &persistersHolder{}

	shardHeadersPersister, err := dr.databaseReader.LoadPersister(dbInfo, "BlockHeaders")
	if err != nil {
		return nil, err
	}
	persHold.shardHeadersPersister = shardHeadersPersister

	miniBlocksPersister, err := dr.databaseReader.LoadPersister(dbInfo, "MiniBlocks")
	if err != nil {
		return nil, err
	}
	persHold.miniBlocksPersister = miniBlocksPersister

	txsPersister, err := dr.databaseReader.LoadPersister(dbInfo, "Transactions")
	if err != nil {
		return nil, err
	}
	persHold.transactionPersister = txsPersister

	uTxsPersister, err := dr.databaseReader.LoadPersister(dbInfo, "UnsignedTransactions")
	if err != nil {
		return nil, err
	}
	persHold.unsignedTransactionsPersister = uTxsPersister

	rTxsPersister, err := dr.databaseReader.LoadPersister(dbInfo, "RewardTransactions")
	if err != nil {
		return nil, err
	}
	persHold.rewardTransactionsPersister = rTxsPersister

	return persHold, nil
}

func (dr *dataReplayer) closePersisters(persisters *persistersHolder) {
	err := persisters.shardHeadersPersister.Close()
	log.LogIfError(err)

	err = persisters.miniBlocksPersister.Close()
	log.LogIfError(err)

	err = persisters.transactionPersister.Close()
	log.LogIfError(err)

	err = persisters.unsignedTransactionsPersister.Close()
	log.LogIfError(err)

	err = persisters.rewardTransactionsPersister.Close()
	log.LogIfError(err)
}

// IsInterfaceNil returns true if there is no value under the interface
func (dr *dataReplayer) IsInterfaceNil() bool {
	return dr == nil
}

func getMetaChainDatabasesInfo(records []*databasereader.DatabaseInfo) ([]*databasereader.DatabaseInfo, error) {
	metaChainDBsInfo := make([]*databasereader.DatabaseInfo, 0)
	for _, record := range records {
		if record.Shard == core.MetachainShardId {
			metaChainDBsInfo = append(metaChainDBsInfo, record)
		}
	}

	if len(metaChainDBsInfo) == 0 {
		return nil, ErrNoMetachainDatabase
	}

	return metaChainDBsInfo, nil
}

func getShardDatabaseForEpoch(records []*databasereader.DatabaseInfo, epoch uint32, shard uint32) (*databasereader.DatabaseInfo, error) {
	for _, record := range records {
		if record.Epoch == epoch && record.Shard == shard {
			return record, nil
		}
	}

	return nil, ErrDatabaseInfoNotFound
}
