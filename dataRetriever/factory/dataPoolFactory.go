package factory

import (
	"io/ioutil"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool/headersCache"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache/capacity"
	"github.com/ElrondNetwork/elrond-go/storage/storageCacherAdapter"
	trieNodeFactory "github.com/ElrondNetwork/elrond-go/storage/storageCacherAdapter/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var log = logger.GetOrCreate("dataRetriever/factory")

// ArgsDataPool holds the arguments needed for NewDataPoolFromConfig function
type ArgsDataPool struct {
	Config           *config.Config
	EconomicsData    process.EconomicsDataHandler
	ShardCoordinator sharding.Coordinator
	Marshalizer      marshal.Marshalizer
}

// NewDataPoolFromConfig will return a new instance of a PoolsHolder
func NewDataPoolFromConfig(args ArgsDataPool) (dataRetriever.PoolsHolder, error) {
	log.Debug("creatingDataPool from config")

	if args.Config == nil {
		return nil, dataRetriever.ErrNilConfig
	}
	if check.IfNil(args.EconomicsData) {
		return nil, dataRetriever.ErrNilEconomicsData
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, dataRetriever.ErrNilShardCoordinator
	}

	mainConfig := args.Config

	txPool, err := txpool.NewShardedTxPool(txpool.ArgShardedTxPool{
		Config:         factory.GetCacherFromConfig(mainConfig.TxDataPool),
		NumberOfShards: args.ShardCoordinator.NumberOfShards(),
		SelfShardID:    args.ShardCoordinator.SelfId(),
		TxGasHandler:   args.EconomicsData,
	})
	if err != nil {
		log.Error("error creating txpool")
		return nil, err
	}

	uTxPool, err := shardedData.NewShardedData(dataRetriever.UnsignedTxPoolName, factory.GetCacherFromConfig(mainConfig.UnsignedTransactionDataPool))
	if err != nil {
		log.Error("error creating smart contract result pool")
		return nil, err
	}

	rewardTxPool, err := shardedData.NewShardedData(dataRetriever.RewardTxPoolName, factory.GetCacherFromConfig(mainConfig.RewardTransactionDataPool))
	if err != nil {
		log.Error("error creating reward transaction pool")
		return nil, err
	}

	hdrPool, err := headersCache.NewHeadersPool(mainConfig.HeadersPoolConfig)
	if err != nil {
		log.Error("error creating headers pool")
		return nil, err
	}

	cacherCfg := factory.GetCacherFromConfig(mainConfig.TxBlockBodyDataPool)
	txBlockBody, err := storageUnit.NewCache(cacherCfg)
	if err != nil {
		log.Error("error creating txBlockBody")
		return nil, err
	}

	cacherCfg = factory.GetCacherFromConfig(mainConfig.PeerBlockBodyDataPool)
	peerChangeBlockBody, err := storageUnit.NewCache(cacherCfg)
	if err != nil {
		log.Error("error creating peerChangeBlockBody")
		return nil, err
	}

	cacher, err := capacity.NewCapacityLRU(
		int(mainConfig.TrieSyncStorage.Capacity),
		int64(mainConfig.TrieSyncStorage.SizeInBytes),
	)
	if err != nil {
		return nil, err
	}

	dbCfg := factory.GetDBFromConfig(mainConfig.TrieSyncStorage.DB)
	argDB := storageUnit.ArgDB{
		DBType:            dbCfg.Type,
		Path:              dbCfg.FilePath,
		BatchDelaySeconds: dbCfg.BatchDelaySeconds,
		MaxBatchSize:      dbCfg.MaxBatchSize,
		MaxOpenFiles:      dbCfg.MaxOpenFiles,
	}

	if mainConfig.TrieSyncStorage.DB.UseTmpAsFilePath {
		filePath, err := ioutil.TempDir("", "trieSyncStorage")
		if err != nil {
			return nil, err
		}

		argDB.Path = filePath
	}

	db, err := storageUnit.NewDB(argDB)
	if err != nil {
		return nil, err
	}

	tnf := trieNodeFactory.NewTrieNodeFactory()
	adaptedTrieNodesStorage, err := storageCacherAdapter.NewStorageCacherAdapter(cacher, db, tnf, args.Marshalizer)
	if err != nil {
		return nil, err
	}

	cacherCfg = factory.GetCacherFromConfig(mainConfig.SmartContractDataPool)
	smartContracts, err := storageUnit.NewCache(cacherCfg)
	if err != nil {
		log.Error("error creating smartContracts cache unit")
		return nil, err
	}

	currBlockTxs, err := dataPool.NewCurrentBlockPool()
	if err != nil {
		return nil, err
	}

	return dataPool.NewDataPool(
		txPool,
		uTxPool,
		rewardTxPool,
		hdrPool,
		txBlockBody,
		peerChangeBlockBody,
		adaptedTrieNodesStorage,
		currBlockTxs,
		smartContracts,
	)
}
