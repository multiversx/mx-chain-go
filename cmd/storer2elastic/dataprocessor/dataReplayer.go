package dataprocessor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	storer2ElasticData "github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/data"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
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
	databaseReader    DatabaseReaderHandler
	generalConfig     config.Config
	shardCoordinator  sharding.Coordinator
	marshalizer       marshal.Marshalizer
	uint64Converter   typeConverters.Uint64ByteSliceConverter
	headerMarshalizer HeaderMarshalizerHandler
	emptyReceiptHash  []byte
	startingEpoch     uint32
}

type persistersHolder struct {
	miniBlocksPersister           storage.Persister
	transactionPersister          storage.Persister
	unsignedTransactionsPersister storage.Persister
	rewardTransactionsPersister   storage.Persister
	shardHeadersPersister         storage.Persister
	receiptsPersister             storage.Persister
}

type metaBlocksPersistersHolder struct {
	hdrHashNoncePersister storage.Persister
	metaBlocksPersister   storage.Persister
}

// DataReplayerArgs holds the arguments needed for creating a new dataReplayer
type DataReplayerArgs struct {
	ElasticIndexer           indexer.Indexer
	DatabaseReader           DatabaseReaderHandler
	GeneralConfig            config.Config
	ShardCoordinator         sharding.Coordinator
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	HeaderMarshalizer        HeaderMarshalizerHandler
	StartingEpoch            uint32
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

	emptyReceiptHash, err := core.CalculateHash(args.Marshalizer, args.Hasher, &batch.Batch{Data: [][]byte{}})
	if err != nil {
		return nil, err
	}

	return &dataReplayer{
		databaseReader:    args.DatabaseReader,
		generalConfig:     args.GeneralConfig,
		shardCoordinator:  args.ShardCoordinator,
		marshalizer:       args.Marshalizer,
		uint64Converter:   args.Uint64ByteSliceConverter,
		headerMarshalizer: args.HeaderMarshalizer,
		emptyReceiptHash:  emptyReceiptHash,
		startingEpoch:     args.StartingEpoch,
	}, nil
}

// Range will range over the data in storage until the handler returns false or the time is out
func (dr *dataReplayer) Range(handler func(persistedData storer2ElasticData.RoundPersistedData) bool) error {
	errChan := make(chan error, 0)
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		dr.startIndexing(errChan, handler)
	}()

	select {
	case err := <-errChan:
		return err
	case <-signalChannel:
		log.Info("the application will stop after an OS signal")
		return ErrOsSignalIntercepted
	}
}

func (dr *dataReplayer) startIndexing(errChan chan error, persistedDataHandler func(persistedData storer2ElasticData.RoundPersistedData) bool) {
	records, err := dr.databaseReader.GetDatabaseInfo()
	if err != nil {
		errChan <- err
		return
	}

	err = checkDatabaseInfo(records)
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
		if metaDB.Epoch < dr.startingEpoch {
			continue
		}

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
		shardDBInfo, errGetShardHDr := getShardDatabaseForEpoch(dbsInfo, record.Epoch, shardID)
		if errGetShardHDr != nil {
			return errGetShardHDr
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

	startingNonce := epochStartMetaBlock.Nonce
	roundData, err := dr.processMetaBlock(
		epochStartMetaBlock,
		dbsInfo,
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
		metaBlock, errGetMb := dr.getMetaBlockForNonce(startingNonce, metaHeadersPersisters)
		if errGetMb != nil {
			log.Error(errGetMb.Error())
			break
		}

		roundData, err = dr.processMetaBlock(
			metaBlock,
			dbsInfo,
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
	}

	log.Info("finished indexing all headers from an epoch", "epoch", record.Epoch)
	return nil
}

func (dr *dataReplayer) prepareMetaPersistersHolder(record *databasereader.DatabaseInfo) (*metaBlocksPersistersHolder, error) {
	metaPersHolder := &metaBlocksPersistersHolder{}

	metaBlocksUnit, err := dr.databaseReader.LoadPersister(record, dr.generalConfig.MetaBlockStorage.DB.FilePath)
	if err != nil {
		return nil, err
	}
	metaPersHolder.metaBlocksPersister = metaBlocksUnit

	headerHashNonceUnit, err := dr.databaseReader.LoadStaticPersister(record, dr.generalConfig.MetaHdrNonceHashStorage.DB.FilePath)
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

	if record.Epoch == 0 {
		return dr.getMetaBlockForNonce(0, metaPersisters)
	}

	return nil, fmt.Errorf("epoch start meta not found for epoch %d", record.Epoch)
}

func (dr *dataReplayer) getMetaBlockForNonce(nonce uint64, metaPersisters *metaBlocksPersistersHolder) (*block.MetaBlock, error) {
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
	persisters *persistersHolder,
	shardPersisters map[uint32]*persistersHolder,
) (*storer2ElasticData.RoundPersistedData, error) {
	metaHdrData, err := dr.processHeader(persisters, metaBlock)
	if err != nil {
		return nil, err
	}

	shardsHeaderData := make(map[uint32]*storer2ElasticData.HeaderData)
	for _, shardInfo := range metaBlock.ShardInfo {
		shardHdrData, errProcessShardInfo := dr.processShardInfo(dbsInfo, &shardInfo, metaBlock.Epoch, shardPersisters[shardInfo.ShardID])
		if errProcessShardInfo != nil {
			log.Warn("cannot process shard info", "error", errProcessShardInfo)
			return nil, errProcessShardInfo
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
	shardHeader, err := dr.getShardHeader(shardPersisters, shardInfo.HeaderHash)
	if err != nil {
		// if new epoch, shard headers can be found in the previous epoch's storage
		// TODO: for this case return and use all the persisters from the previous epoch in order to find txs
		shardHeader, err = dr.getFromShardStorer(dbsInfos, shardInfo, epoch-1)
		if err != nil {
			return nil, err
		}
	}

	return dr.processHeader(shardPersisters, shardHeader)
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

func (dr *dataReplayer) processHeader(persisters *persistersHolder, hdr data.HeaderHandler) (*storer2ElasticData.HeaderData, error) {
	miniBlocksHashes := make([]string, 0)
	shardIDs := dr.getShardIDs()
	for _, shard := range shardIDs {
		miniBlocksForShard := hdr.GetMiniBlockHeadersWithDst(shard)
		for hash := range miniBlocksForShard {
			miniBlocksHashes = append(miniBlocksHashes, hash)
		}
	}

	body, txPool, err := dr.processBodyAndTransactionsPoolForHeader(hdr, persisters, miniBlocksHashes)
	if err != nil {
		return nil, err
	}

	return &storer2ElasticData.HeaderData{
		Header:           hdr,
		Body:             body,
		BodyTransactions: txPool,
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
	header data.HeaderHandler,
	persisters *persistersHolder,
	mbHashes []string,
) (*block.Body, map[string]data.TransactionHandler, error) {
	txPool := make(map[string]data.TransactionHandler)
	mbUnit := persisters.miniBlocksPersister

	blockBody := &block.Body{}
	for _, mbHash := range mbHashes {
		recoveredMiniBlock, err := dr.getMiniBlockFromStorage(mbUnit, []byte(mbHash))
		if err != nil {
			if header.GetShardID() == core.MetachainShardId {
				// could be miniblocks headers for shard, so we can continue
				continue
			}

			return nil, nil, err
		}

		err = dr.processTransactionsForMiniBlock(persisters, recoveredMiniBlock.TxHashes, txPool, recoveredMiniBlock.Type)
		if err != nil {
			return nil, nil, err
		}

		blockBody.MiniBlocks = append(blockBody.MiniBlocks, recoveredMiniBlock)
	}

	receiptsMiniBlocks, err := dr.getReceiptsIfNeeded(persisters, header)
	if err != nil {
		return nil, nil, err
	}
	if receiptsMiniBlocks != nil {
		blockBody.MiniBlocks = append(blockBody.MiniBlocks, receiptsMiniBlocks...)
		for _, miniBlock := range receiptsMiniBlocks {
			err = dr.processTransactionsForMiniBlock(persisters, miniBlock.TxHashes, txPool, miniBlock.Type)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return blockBody, txPool, nil
}

func (dr *dataReplayer) getReceiptsIfNeeded(holder *persistersHolder, hdr data.HeaderHandler) ([]*block.MiniBlock, error) {
	if len(hdr.GetReceiptsHash()) == 0 {
		return nil, nil
	}
	if bytes.Equal(hdr.GetReceiptsHash(), dr.emptyReceiptHash) {
		return nil, nil
	}

	batchBytes, err := holder.receiptsPersister.Get(hdr.GetReceiptsHash())
	if err != nil {
		log.Warn("receipts hash not found", "hash", hdr.GetReceiptsHash())
		return nil, nil
	}

	var batchObj batch.Batch
	err = dr.marshalizer.Unmarshal(&batchObj, batchBytes)
	if err != nil {
		return nil, err
	}

	miniBlocksToReturn := make([]*block.MiniBlock, 0)
	for _, mbBytes := range batchObj.Data {
		recoveredMiniBlock := &block.MiniBlock{}
		errUnmarshalMb := dr.marshalizer.Unmarshal(recoveredMiniBlock, mbBytes)
		if errUnmarshalMb != nil {
			log.Warn("cannot unmarshal receipts miniblock")
			return nil, errUnmarshalMb
		}

		miniBlocksToReturn = append(miniBlocksToReturn, recoveredMiniBlock)
	}

	return miniBlocksToReturn, nil
}

func (dr *dataReplayer) getMiniBlockFromStorage(mbUnit storage.Persister, mbHash []byte) (*block.MiniBlock, error) {
	mbBytes, err := mbUnit.Get(mbHash)
	if err != nil {
		return nil, fmt.Errorf("miniblock with hash %s not found in storage", hex.EncodeToString(mbHash))
	}

	recoveredMiniBlock := &block.MiniBlock{}
	err = dr.marshalizer.Unmarshal(recoveredMiniBlock, mbBytes)
	if err != nil {
		return nil, fmt.Errorf("%w when unmarshaling miniblock with hash %s", err, mbHash)
	}

	return recoveredMiniBlock, nil
}

func (dr *dataReplayer) processTransactionsForMiniBlock(
	persisters *persistersHolder,
	txHashes [][]byte,
	txPool map[string]data.TransactionHandler,
	mbType block.Type,
) error {
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
		case block.ReceiptBlock:
			tx, getTxErr = dr.getReceipt(persisters, txHash)
		case block.PeerBlock:
			// hashes stored inside peer blocks do not stand for a specific transaction, so skip
			continue
		}
		if getTxErr != nil {
			return getTxErr
		}

		if tx == nil {
			return fmt.Errorf("transaction recovered from storage with hash %s is nil", hex.EncodeToString(txHash))
		}

		txPool[string(txHash)] = tx
	}

	return nil
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

func (dr *dataReplayer) getReceipt(holder *persistersHolder, txHash []byte) (data.TransactionHandler, error) {
	txBytes, err := holder.unsignedTransactionsPersister.Get(txHash)
	if err != nil {
		log.Warn("cannot get unsigned tx from storer", "txHash", txHash)
		return nil, err
	}

	recoveredReceipt := &receipt.Receipt{}
	err = dr.marshalizer.Unmarshal(recoveredReceipt, txBytes)
	if err != nil {
		log.Warn("cannot unmarshal receipt", "hash", txHash, "error", err)
		return nil, err
	}

	return recoveredReceipt, nil
}

func (dr *dataReplayer) preparePersistersHolder(dbInfo *databasereader.DatabaseInfo) (*persistersHolder, error) {
	persHold := &persistersHolder{}

	shardHeadersPersister, err := dr.databaseReader.LoadPersister(dbInfo, dr.generalConfig.BlockHeaderStorage.DB.FilePath)
	if err != nil {
		return nil, err
	}
	persHold.shardHeadersPersister = shardHeadersPersister

	miniBlocksPersister, err := dr.databaseReader.LoadPersister(dbInfo, dr.generalConfig.MiniBlocksStorage.DB.FilePath)
	if err != nil {
		return nil, err
	}
	persHold.miniBlocksPersister = miniBlocksPersister

	txsPersister, err := dr.databaseReader.LoadPersister(dbInfo, dr.generalConfig.TxStorage.DB.FilePath)
	if err != nil {
		return nil, err
	}
	persHold.transactionPersister = txsPersister

	uTxsPersister, err := dr.databaseReader.LoadPersister(dbInfo, dr.generalConfig.UnsignedTransactionStorage.DB.FilePath)
	if err != nil {
		return nil, err
	}
	persHold.unsignedTransactionsPersister = uTxsPersister

	rTxsPersister, err := dr.databaseReader.LoadPersister(dbInfo, dr.generalConfig.RewardTxStorage.DB.FilePath)
	if err != nil {
		return nil, err
	}
	persHold.rewardTransactionsPersister = rTxsPersister

	receiptsPersister, err := dr.databaseReader.LoadPersister(dbInfo, dr.generalConfig.ReceiptsStorage.DB.FilePath)
	if err != nil {
		return nil, err
	}
	persHold.receiptsPersister = receiptsPersister

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

	err = persisters.receiptsPersister.Close()
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

func checkDatabaseInfo(dbsInfo []*databasereader.DatabaseInfo) error {
	sliceCopy := make([]databasereader.DatabaseInfo, len(dbsInfo))
	for i := 0; i < len(dbsInfo); i++ {
		sliceCopy[i] = *dbsInfo[i]
	}

	sort.Slice(sliceCopy, func(i int, j int) bool {
		return sliceCopy[i].Epoch < sliceCopy[j].Epoch
	})

	previousEpoch := uint32(0)
	for _, dbInfo := range sliceCopy {
		if dbInfo.Epoch != previousEpoch && dbInfo.Epoch-1 != previousEpoch {
			return fmt.Errorf("configuration for epoch %d is missing", dbInfo.Epoch-1)
		}

		previousEpoch = dbInfo.Epoch
	}

	return nil
}
