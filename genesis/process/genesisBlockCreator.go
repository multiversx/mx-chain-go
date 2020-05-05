package process

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	factoryState "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/process/intermediate"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/update/files"
	hardfork "github.com/ElrondNetwork/elrond-go/update/genesis"
)

type genesisBlockCreationHandler func(arg ArgsGenesisBlockCreator, nodesListSplitter genesis.NodesListSplitter) (data.HeaderHandler, error)

type genesisBlockCreator struct {
	arg                 ArgsGenesisBlockCreator
	shardCreatorHandler genesisBlockCreationHandler
	metaCreatorHandler  genesisBlockCreationHandler
}

// NewGenesisBlockCreator creates a new genesis block creator instance able to create genesis blocks on all initial shards
func NewGenesisBlockCreator(arg ArgsGenesisBlockCreator) (*genesisBlockCreator, error) {
	err := checkArgumentsForBlockCreator(arg)
	if err != nil {
		return nil, fmt.Errorf("%w while creating NewGenesisBlockCreator", err)
	}

	gbc := &genesisBlockCreator{
		arg:                 arg,
		shardCreatorHandler: CreateShardGenesisBlock,
		metaCreatorHandler:  CreateMetaGenesisBlock,
	}

	if arg.HardForkConfig.MustImport {
		err = gbc.createHardForkImportHandler()
		if err != nil {
			return nil, err
		}
	}

	return gbc, nil
}

func (gbc *genesisBlockCreator) createHardForkImportHandler() error {
	importConfig := gbc.arg.HardForkConfig.ImportStateStorageConfig
	importStore, err := storageUnit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(importConfig.Cache),
		factory.GetDBFromConfig(importConfig.DB),
		factory.GetBloomFromConfig(importConfig.Bloom),
	)
	if err != nil {
		return err
	}

	args := files.ArgsNewMultiFileReader{
		ImportFolder: gbc.arg.HardForkConfig.ImportFolder,
		ImportStore:  importStore,
	}
	multiFileReader, err := files.NewMultiFileReader(args)
	if err != nil {
		return err
	}

	argsHardForkImport := hardfork.ArgsNewStateImport{
		Reader:         multiFileReader,
		Hasher:         gbc.arg.Hasher,
		Marshalizer:    gbc.arg.Marshalizer,
		ShardID:        gbc.arg.ShardCoordinator.SelfId(),
		StorageConfig:  config.StorageConfig{},
		TrieFactory:    gbc.arg.TrieFactory,
		TriesContainer: gbc.arg.TriesContainer,
	}
	importHandler, err := hardfork.NewStateImport(argsHardForkImport)
	if err != nil {
		return err
	}

	gbc.arg.importHandler = importHandler
	return nil
}

func checkArgumentsForBlockCreator(arg ArgsGenesisBlockCreator) error {
	if check.IfNil(arg.Accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(arg.PubkeyConv) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(arg.InitialNodesSetup) {
		return process.ErrNilNodesSetup
	}
	if check.IfNil(arg.Economics) {
		return process.ErrNilEconomicsData
	}
	if check.IfNil(arg.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(arg.Store) {
		return process.ErrNilStore
	}
	if check.IfNil(arg.Blkc) {
		return process.ErrNilBlockChain
	}
	if check.IfNil(arg.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arg.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(arg.Uint64ByteSliceConverter) {
		return process.ErrNilUint64Converter
	}
	if check.IfNil(arg.DataPool) {
		return process.ErrNilPoolsHolder
	}
	if check.IfNil(arg.AccountsParser) {
		return genesis.ErrNilAccountsParser
	}
	if arg.GasMap == nil {
		return process.ErrNilGasSchedule
	}
	if check.IfNil(arg.TxLogsProcessor) {
		return process.ErrNilTxLogsProcessor
	}
	if check.IfNil(arg.SmartContractParser) {
		return genesis.ErrNilSmartContractParser
	}

	return nil
}

// CreateGenesisBlocks will try to create the genesis blocks for all shards
func (gbc *genesisBlockCreator) CreateGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	var err error
	var genesisBlock data.HeaderHandler
	var newArgument ArgsGenesisBlockCreator

	if gbc.arg.HardForkConfig.MustImport {
		err = gbc.arg.importHandler.ImportAll()
		if err != nil {
			return nil, err
		}
	}

	nodesListSplitter, err := intermediate.NewNodesListSplitter(gbc.arg.InitialNodesSetup, gbc.arg.AccountsParser)
	if err != nil {
		return nil, err
	}

	for shardID := uint32(0); shardID < gbc.arg.ShardCoordinator.NumberOfShards(); shardID++ {
		newArgument, err = gbc.getNewArgForShard(shardID)
		if err != nil {
			return nil, fmt.Errorf("'%w' while creating new argument for shard %d",
				err, shardID)
		}

		genesisBlock, err = gbc.shardCreatorHandler(newArgument, nodesListSplitter)
		if err != nil {
			return nil, fmt.Errorf("'%w' while generating genesis block for shard %d",
				err, shardID)
		}

		genesisBlocks[shardID] = genesisBlock
		err = gbc.saveGenesisBlock(genesisBlock)
		if err != nil {
			return nil, fmt.Errorf("'%w' while saving genesis block for shard %d",
				err, shardID)
		}
	}

	newArgument, err = gbc.getNewArgForShard(core.MetachainShardId)
	if err != nil {
		return nil, fmt.Errorf("'%w' while creating new argument for metachain", err)
	}

	newArgument.Blkc = blockchain.NewMetaChain()
	genesisBlock, err = gbc.metaCreatorHandler(newArgument, nodesListSplitter)
	if err != nil {
		return nil, fmt.Errorf("'%w' while generating genesis block for metachain", err)
	}

	genesisBlocks[core.MetachainShardId] = genesisBlock
	err = gbc.saveGenesisBlock(genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("'%w' while saving genesis block for metachain", err)
	}

	return genesisBlocks, nil
}

func (gbc *genesisBlockCreator) getNewArgForShard(shardID uint32) (ArgsGenesisBlockCreator, error) {
	var err error

	isCurrentShard := shardID == gbc.arg.ShardCoordinator.SelfId()
	shouldRecreate := !isCurrentShard || gbc.arg.StartEpochNum != 0
	if !shouldRecreate {
		return gbc.arg, nil
	}

	newArgument := gbc.arg //copy the arguments
	newArgument.Accounts, err = createInMemoryAccountAdapter(
		newArgument.Marshalizer,
		newArgument.Hasher,
		factoryState.NewAccountCreator(),
	)
	if err != nil {
		return ArgsGenesisBlockCreator{}, fmt.Errorf("'%w' while generating an in-memory accounts adapter for shard %d",
			err, shardID)
	}

	newArgument.ShardCoordinator, err = sharding.NewMultiShardCoordinator(
		newArgument.ShardCoordinator.NumberOfShards(),
		shardID,
	)
	if err != nil {
		return ArgsGenesisBlockCreator{}, fmt.Errorf("'%w' while generating an temporary shard coordinator for shard %d",
			err, shardID)
	}

	return newArgument, err
}

func (gbc *genesisBlockCreator) saveGenesisBlock(header data.HeaderHandler) error {
	blockBuff, err := gbc.arg.Marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	hash := gbc.arg.Hasher.Compute(string(blockBuff))
	unitType := dataRetriever.BlockHeaderUnit
	if header.GetShardID() == core.MetachainShardId {
		unitType = dataRetriever.MetaBlockUnit
	}

	return gbc.arg.Store.Put(unitType, hash, blockBuff)
}

func saveGenesisBodyToStorage(txCoordinator process.TransactionCoordinator, bodyHandler data.BodyHandler) {
	blockBody, ok := bodyHandler.(*block.Body)
	if !ok {
		log.Warn("wrong type assertion when saving genesis body to storage")
		return
	}

	errNotCritical := txCoordinator.SaveBlockDataToStorage(blockBody)
	if errNotCritical != nil {
		log.Warn("could not save genesis block body to storage", "error", errNotCritical)
	}
}
