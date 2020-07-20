package dataprocessor

import (
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
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("dataprocessor")

type dataProcessor struct {
	elasticIndexer    indexer.Indexer
	databaseReader    DatabaseReaderHandler
	shardCoordinator  sharding.Coordinator
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	nodesCoordinators map[uint32]NodesCoordinator
	uint64Converter   typeConverters.Uint64ByteSliceConverter
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

// Args holds the arguments needed for creating a new dataProcessor
type Args struct {
	ElasticIndexer           indexer.Indexer
	DatabaseReader           DatabaseReaderHandler
	ShardCoordinator         sharding.Coordinator
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	GenesisNodesSetup        sharding.GenesisNodesSetupHandler
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
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
	if check.IfNil(args.Uint64ByteSliceConverter) {
		return nil, ErrNilUint64ByteSliceConverter
	}

	dpi := &dataProcessor{
		elasticIndexer:   args.ElasticIndexer,
		databaseReader:   args.DatabaseReader,
		shardCoordinator: args.ShardCoordinator,
		marshalizer:      args.Marshalizer,
		hasher:           args.Hasher,
		uint64Converter:  args.Uint64ByteSliceConverter,
	}

	nodesCoordinator, err := dpi.createNodesCoordinators(args.GenesisNodesSetup)
	if err != nil {
		return nil, err
	}

	dpi.nodesCoordinators = nodesCoordinator
	return dpi, nil
}

// Index will prepare the data for indexing and will try to index all the data before the given timeout
func (dp *dataProcessor) Index(timeout int) error {
	errChan := make(chan error, 0)
	go func() {
		dp.startIndexing(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(time.Duration(timeout) * time.Second):
		return ErrTimeIsOut
	}
}

func (dp *dataProcessor) startIndexing(errChan chan error) {
	records, err := dp.databaseReader.GetDatabaseInfo()
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
		err = dp.processMetaChainDatabase(metaDB, records)
		if err != nil {
			errChan <- err
			return
		}
	}

	errChan <- nil
}

func (dp *dataProcessor) processMetaChainDatabase(record *databasereader.DatabaseInfo, dbsInfo []*databasereader.DatabaseInfo) error {
	metaHeadersPersisters, err := dp.prepareMetaPersistersHolder(record)
	if err != nil {
		return err
	}

	defer func() {
		dp.closeMetaPersisters(metaHeadersPersisters)
	}()

	shardPersistersHolder := make(map[uint32]*persistersHolder)
	for shardID := uint32(0); shardID < dp.shardCoordinator.NumberOfShards(); shardID++ {
		shardDBInfo, err := getShardDatabaseForEpoch(dbsInfo, record.Epoch, shardID)
		if err != nil {
			return err
		}
		shardPersistersHolder[shardID], err = dp.preparePersistersHolder(shardDBInfo)
		if err != nil {
			return err
		}
	}

	defer func() {
		for _, shardPersisters := range shardPersistersHolder {
			dp.closePersisters(shardPersisters)
		}
	}()

	metachainPersisters, err := dp.preparePersistersHolder(record)
	if err != nil {
		return err
	}

	defer func() {
		dp.closePersisters(metachainPersisters)
	}()

	epochStartMetaBlock, err := dp.getEpochStartMetaBlock(record, metaHeadersPersisters)
	if err != nil {
		return err
	}

	dp.processValidatorsForEpoch(epochStartMetaBlock.Epoch, epochStartMetaBlock, metachainPersisters.miniBlocksPersister)

	startingNonce := epochStartMetaBlock.Nonce
	err = dp.processMetaBlock(epochStartMetaBlock, dbsInfo, record, metaHeadersPersisters, metachainPersisters, shardPersistersHolder)
	if err != nil {
		return err
	}

	for {
		startingNonce++
		metaBlock, err := dp.getMetaBlockForNonce(startingNonce, record, metaHeadersPersisters)
		if err == nil {
			err = dp.processMetaBlock(metaBlock, dbsInfo, record, metaHeadersPersisters, metachainPersisters, shardPersistersHolder)
			if err != nil {
				log.Warn(err.Error())
				break
			}
		} else {
			break
		}
	}

	log.Info("finished indexing all headers from an epoch", "epoch", record.Epoch)
	return nil
}

func (dp *dataProcessor) prepareMetaPersistersHolder(record *databasereader.DatabaseInfo) (*metaBlocksPersistersHolder, error) {
	metaPersHolder := &metaBlocksPersistersHolder{}

	metaBlocksUnit, err := dp.databaseReader.LoadPersister(record, "MetaBlock")
	if err != nil {
		return nil, err
	}
	metaPersHolder.metaBlocksPersister = metaBlocksUnit

	headerHashNonceUnit, err := dp.databaseReader.LoadPersister(record, "MetaHdrHashNonce")
	if err != nil {
		return nil, err
	}
	metaPersHolder.hdrHashNoncePersister = headerHashNonceUnit

	return metaPersHolder, nil
}

func (dp *dataProcessor) closeMetaPersisters(persistersHolder *metaBlocksPersistersHolder) {
	err := persistersHolder.metaBlocksPersister.Close()
	log.LogIfError(err)

	err = persistersHolder.hdrHashNoncePersister.Close()
	log.LogIfError(err)
}

func (dp *dataProcessor) getEpochStartMetaBlock(record *databasereader.DatabaseInfo, metaPersisters *metaBlocksPersistersHolder) (*block.MetaBlock, error) {
	epochStartIdentifier := core.EpochStartIdentifier(record.Epoch)
	metaBlockBytes, err := metaPersisters.metaBlocksPersister.Get([]byte(epochStartIdentifier))
	if err == nil && len(metaBlockBytes) > 0 {
		return dp.unmarshalMetaBlock(metaBlockBytes)
	}
	log.Warn("epoch start meta block not found", "epoch", record.Epoch)

	// header should be found by the epoch start identifier. if not, try getting the nonce 0 (the genesis block)
	return dp.getMetaBlockForNonce(0, record, metaPersisters)
}

func (dp *dataProcessor) getMetaBlockForNonce(nonce uint64, record *databasereader.DatabaseInfo, metaPersisters *metaBlocksPersistersHolder) (*block.MetaBlock, error) {
	nonceBytes := dp.uint64Converter.ToByteSlice(nonce)
	metaBlockHash, err := metaPersisters.hdrHashNoncePersister.Get(nonceBytes)
	if err != nil {
		return nil, err
	}

	metaBlockBytes, err := metaPersisters.metaBlocksPersister.Get(metaBlockHash)
	if err != nil {
		return nil, err
	}

	return dp.unmarshalMetaBlock(metaBlockBytes)
}

func (dp *dataProcessor) processMetaBlock(
	metaBlock *block.MetaBlock,
	dbsInfo []*databasereader.DatabaseInfo,
	record *databasereader.DatabaseInfo,
	metaBlocksPersisters *metaBlocksPersistersHolder,
	persisters *persistersHolder,
	shardPersisters map[uint32]*persistersHolder,
) error {
	err := dp.indexHeader(persisters, record, metaBlock, record.Epoch)
	if err != nil {
		return err
	}

	for _, shardInfo := range metaBlock.ShardInfo {
		err = dp.processShardInfo(dbsInfo, &shardInfo, metaBlock.Epoch, shardPersisters[shardInfo.ShardID])
		log.LogIfError(err)
	}

	return nil
}

func (dp *dataProcessor) processShardInfo(
	dbsInfos []*databasereader.DatabaseInfo,
	shardInfo *block.ShardData,
	epoch uint32,
	shardPersisters *persistersHolder,
) error {
	shardDBInfo, err := getShardDatabaseForEpoch(dbsInfos, epoch, shardInfo.ShardID)
	if err != nil {
		return err
	}

	shardHeader, err := dp.getShardHeader(shardPersisters, shardInfo.HeaderHash)
	if err != nil {
		// if new epoch, shard headers can be found in the previous epoch's storage
		// TODO: for this case return and use all the persisters from the previous epoch in order to find txs
		shardHeader, err = dp.getFromShardStorer(dbsInfos, shardInfo, epoch-1)
		if err != nil {
			return err
		}
	}

	return dp.indexHeader(shardPersisters, shardDBInfo, shardHeader, epoch)
}

func (dp *dataProcessor) getFromShardStorer(dbsInfos []*databasereader.DatabaseInfo, shardInfo *block.ShardData, epoch uint32) (*block.Header, error) {
	shardDBInfo, err := getShardDatabaseForEpoch(dbsInfos, epoch, shardInfo.ShardID)
	if err != nil {
		return nil, err
	}

	shardPersisters, err := dp.preparePersistersHolder(shardDBInfo)
	if err != nil {
		return nil, err
	}

	defer func() {
		dp.closePersisters(shardPersisters)
	}()

	return dp.getShardHeader(shardPersisters, shardInfo.HeaderHash)
}

func (dp *dataProcessor) getShardHeader(
	shardPersisters *persistersHolder,
	hash []byte,
) (*block.Header, error) {
	shardHeaderBytes, err := shardPersisters.shardHeadersPersister.Get(hash)
	if err != nil {
		return nil, err
	}

	return dp.unmarshalShardHeader(shardHeaderBytes)
}

func (dp *dataProcessor) unmarshalMetaBlock(metaBlockBytes []byte) (*block.MetaBlock, error) {
	var metaBlock block.MetaBlock
	err := dp.marshalizer.Unmarshal(&metaBlock, metaBlockBytes)
	if err != nil {
		return nil, err
	}

	return &metaBlock, nil
}

func (dp *dataProcessor) unmarshalShardHeader(shardHeaderBytes []byte) (*block.Header, error) {
	var shardHeader block.Header
	err := dp.marshalizer.Unmarshal(&shardHeader, shardHeaderBytes)
	if err != nil {
		return nil, err
	}

	return &shardHeader, nil
}

func (dp *dataProcessor) indexHeader(persisters *persistersHolder, dbInfo *databasereader.DatabaseInfo, hdr data.HeaderHandler, epoch uint32) error {
	if metaBlock, ok := hdr.(*block.MetaBlock); ok {
		if metaBlock.IsStartOfEpochBlock() {
			dp.processValidatorsForEpoch(metaBlock.Epoch, metaBlock, persisters.miniBlocksPersister)
		}
	}

	miniBlocksHashes := make([]string, 0)
	shardIDs := dp.getShardIDs()
	for _, shard := range shardIDs {
		miniBlocksForShard := hdr.GetMiniBlockHeadersWithDst(shard)
		for hash := range miniBlocksForShard {
			miniBlocksHashes = append(miniBlocksHashes, hash)
		}
	}

	body, txPool, err := dp.processBodyAndTransactionsPoolForHeader(persisters, miniBlocksHashes)
	if err != nil {
		return err
	}

	singersIndexes, err := dp.computeSignersIndexes(hdr)
	if err != nil {
		return err
	}
	notarizedHdrs := dp.computeNotarizedHeaders(hdr)
	dp.elasticIndexer.SaveBlock(body, hdr, txPool, singersIndexes, notarizedHdrs)
	log.Info("indexed header", "epoch", hdr.GetEpoch(), "shard", hdr.GetShardID(), "nonce", hdr.GetNonce())
	return nil
}

func (dp *dataProcessor) getShardIDs() []uint32 {
	shardIDs := make([]uint32, 0)
	for shard := uint32(0); shard < dp.shardCoordinator.NumberOfShards(); shard++ {
		shardIDs = append(shardIDs, shard)
	}
	shardIDs = append(shardIDs, core.MetachainShardId)

	return shardIDs
}

func (dp *dataProcessor) processBodyAndTransactionsPoolForHeader(
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
		err = dp.marshalizer.Unmarshal(recoveredMiniBlock, mbBytes)
		if err != nil {
			log.Warn("cannot unmarshal miniblock", "hash", mbHash, "error", err)
			continue
		}

		dp.processTransactionsForMiniBlock(persisters, recoveredMiniBlock.TxHashes, txPool, recoveredMiniBlock.Type)
		blockBody.MiniBlocks = append(blockBody.MiniBlocks, recoveredMiniBlock)
	}

	return blockBody, txPool, nil
}

func (dp *dataProcessor) processTransactionsForMiniBlock(
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
			tx, getTxErr = dp.getRegularTx(persisters, txHash)
		case block.RewardsBlock:
			tx, getTxErr = dp.getRewardTx(persisters, txHash)
		case block.SmartContractResultBlock:
			tx, getTxErr = dp.getUnsignedTx(persisters, txHash)
		}
		if getTxErr != nil || tx == nil {
			continue
		}

		txPool[string(txHash)] = tx
	}
}

func (dp *dataProcessor) getRegularTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
	txBytes, err := holder.transactionPersister.Get(txHash)
	if err != nil {
		log.Warn("cannot get tx from storer", "txHash", txHash)
		return nil, err
	}

	recoveredTx := &transaction.Transaction{}
	err = dp.marshalizer.Unmarshal(recoveredTx, txBytes)
	if err != nil {
		log.Warn("cannot unmarshal tx", "txHash", txHash, "error", err)
		return nil, err
	}

	return recoveredTx, nil
}

func (dp *dataProcessor) getRewardTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
	txBytes, err := holder.rewardTransactionsPersister.Get(txHash)
	if err != nil {
		log.Warn("cannot get tx from storer", "txHash", txHash)
		return nil, err
	}

	recoveredTx := &rewardTx.RewardTx{}
	err = dp.marshalizer.Unmarshal(recoveredTx, txBytes)
	if err != nil {
		log.Warn("cannot unmarshal reward tx", "txHash", txHash, "error", err)
		return nil, err
	}

	return recoveredTx, nil
}

func (dp *dataProcessor) getUnsignedTx(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
	txBytes, err := holder.unsignedTransactionsPersister.Get(txHash)
	if err != nil {
		log.Warn("cannot get unsigned tx from storer", "txHash", txHash)
		return nil, err
	}

	recoveredTx := &smartContractResult.SmartContractResult{}
	err = dp.marshalizer.Unmarshal(recoveredTx, txBytes)
	if err != nil {
		log.Warn("cannot unmarshal unsigned tx", "txHash", txHash, "error", err)
		return nil, err
	}

	return recoveredTx, nil
}
