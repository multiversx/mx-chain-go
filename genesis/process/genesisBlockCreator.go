package process

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	factoryState "github.com/ElrondNetwork/elrond-go/data/state/factory"
	triesFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/process/intermediate"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/files"
	hardfork "github.com/ElrondNetwork/elrond-go/update/genesis"
)

const accountStartNonce = uint64(0)

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

	if mustDoHardForkImportProcess(gbc.arg) {
		err = gbc.createHardForkImportHandler()
		if err != nil {
			return nil, err
		}
	}

	return gbc, nil
}

func mustDoHardForkImportProcess(arg ArgsGenesisBlockCreator) bool {
	return arg.ImportStartHandler.ShouldStartImport() && arg.StartEpochNum <= arg.HardForkConfig.StartEpoch
}

func getGenesisBlocksRoundNonceEpoch(arg ArgsGenesisBlockCreator) (uint64, uint64, uint32) {
	if arg.HardForkConfig.AfterHardFork {
		return arg.HardForkConfig.StartRound, arg.HardForkConfig.StartNonce, arg.HardForkConfig.StartEpoch
	}
	return 0, 0, 0
}

func (gbc *genesisBlockCreator) createHardForkImportHandler() error {
	importFolder := filepath.Join(gbc.arg.WorkingDir, gbc.arg.HardForkConfig.ImportFolder)

	importConfig := gbc.arg.HardForkConfig.ImportStateStorageConfig
	dbConfig := factory.GetDBFromConfig(importConfig.DB)
	dbConfig.FilePath = path.Join(importFolder, importConfig.DB.FilePath)
	importStore, err := storageUnit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(importConfig.Cache),
		dbConfig,
		factory.GetBloomFromConfig(importConfig.Bloom),
	)
	if err != nil {
		return err
	}

	args := files.ArgsNewMultiFileReader{
		ImportFolder: importFolder,
		ImportStore:  importStore,
	}
	multiFileReader, err := files.NewMultiFileReader(args)
	if err != nil {
		return err
	}

	argsHardForkImport := hardfork.ArgsNewStateImport{
		Reader:              multiFileReader,
		Hasher:              gbc.arg.Hasher,
		Marshalizer:         gbc.arg.Marshalizer,
		ShardID:             gbc.arg.ShardCoordinator.SelfId(),
		StorageConfig:       gbc.arg.HardForkConfig.ImportStateStorageConfig,
		TrieStorageManagers: gbc.arg.TrieStorageManagers,
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
	if arg.TrieStorageManagers == nil {
		return genesis.ErrNilTrieStorageManager
	}
	if check.IfNil(arg.ImportStartHandler) {
		return update.ErrNilImportStartHandler
	}
	if check.IfNil(arg.SignMarshalizer) {
		return process.ErrNilMarshalizer
	}

	return nil
}

// CreateGenesisBlocks will try to create the genesis blocks for all shards
func (gbc *genesisBlockCreator) CreateGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	var err error
	var genesisBlock data.HeaderHandler
	var newArgument ArgsGenesisBlockCreator

	if mustDoHardForkImportProcess(gbc.arg) {
		err = gbc.arg.importHandler.ImportAll()
		if err != nil {
			return nil, err
		}

		err = gbc.computeDNSAddressesIfHardFork()
		if err != nil {
			return nil, err
		}
	}

	nodesListSplitter, err := intermediate.NewNodesListSplitter(gbc.arg.InitialNodesSetup, gbc.arg.AccountsParser)
	if err != nil {
		return nil, err
	}

	for shardID := uint32(0); shardID < gbc.arg.ShardCoordinator.NumberOfShards(); shardID++ {
		log.Debug("genesis block creator",
			"shard ID", shardID,
		)
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

	log.Debug("genesis block creator",
		"shard ID", "meta",
	)

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

	for i := uint32(0); i < gbc.arg.ShardCoordinator.NumberOfShards(); i++ {
		gb := genesisBlocks[i]

		log.Info("genesis block created",
			"shard ID", gb.GetShardID(),
			"nonce", gb.GetNonce(),
			"round", gb.GetRound(),
			"root hash", gb.GetRootHash(),
		)
	}
	log.Info("genesis block created",
		"shard ID", "metachain",
		"nonce", genesisBlock.GetNonce(),
		"round", genesisBlock.GetRound(),
		"root hash", genesisBlock.GetRootHash(),
	)

	//TODO call here trie pruning on all roothashes not from current shard

	return genesisBlocks, nil
}

// in case of hardfork initial smart contracts deployment is not called as they are all imported from previous state
func (gbc *genesisBlockCreator) computeDNSAddressesIfHardFork() error {
	var dnsSC genesis.InitialSmartContractHandler
	for _, sc := range gbc.arg.SmartContractParser.InitialSmartContracts() {
		if sc.GetType() == genesis.DNSType {
			dnsSC = sc
			break
		}
	}

	if dnsSC == nil || check.IfNil(dnsSC) {
		return nil
	}

	builtInFuncs := builtInFunctions.NewBuiltInFunctionContainer()
	argsHook := hooks.ArgBlockChainHook{
		Accounts:         gbc.arg.Accounts,
		PubkeyConv:       gbc.arg.PubkeyConv,
		StorageService:   gbc.arg.Store,
		BlockChain:       gbc.arg.Blkc,
		ShardCoordinator: gbc.arg.ShardCoordinator,
		Marshalizer:      gbc.arg.Marshalizer,
		Uint64Converter:  gbc.arg.Uint64ByteSliceConverter,
		BuiltInFunctions: builtInFuncs,
	}
	blockChainHook, err := hooks.NewBlockChainHookImpl(argsHook)
	if err != nil {
		return err
	}

	isForCurrentShard := func([]byte) bool {
		// after hardfork we are interested only in the smart contract addresses, as they are already deployed
		return true
	}
	initialAddresses := intermediate.GenerateInitialPublicKeys(genesis.InitialDNSAddress, isForCurrentShard)
	for _, address := range initialAddresses {
		scResultingAddress, errNewAddress := blockChainHook.NewAddress(address, accountStartNonce, dnsSC.VmTypeBytes())
		if errNewAddress != nil {
			return errNewAddress
		}

		dnsSC.AddAddressBytes(scResultingAddress)
		dnsSC.AddAddress(gbc.arg.PubkeyConv.Encode(scResultingAddress))
	}

	return nil
}

func (gbc *genesisBlockCreator) getNewArgForShard(shardID uint32) (ArgsGenesisBlockCreator, error) {
	var err error

	isCurrentShard := shardID == gbc.arg.ShardCoordinator.SelfId()
	shouldUseProvidedAccountsDB := isCurrentShard && (gbc.arg.StartEpochNum == 0 || mustDoHardForkImportProcess(gbc.arg))
	if shouldUseProvidedAccountsDB {
		return gbc.arg, nil
	}

	newArgument := gbc.arg //copy the arguments
	newArgument.Accounts, err = createAccountAdapter(
		newArgument.Marshalizer,
		newArgument.Hasher,
		factoryState.NewAccountCreator(),
		gbc.arg.TrieStorageManagers[triesFactory.UserAccountTrie],
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
