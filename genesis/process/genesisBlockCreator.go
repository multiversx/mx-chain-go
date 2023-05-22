package process

import (
	"bytes"
	"fmt"
	"math/big"
	"path"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/genesis/process/intermediate"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks/counters"
	"github.com/multiversx/mx-chain-go/sharding"
	factoryState "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	triesFactory "github.com/multiversx/mx-chain-go/trie/factory"
	"github.com/multiversx/mx-chain-go/update"
	hardfork "github.com/multiversx/mx-chain-go/update/genesis"
	hardForkProcess "github.com/multiversx/mx-chain-go/update/process"
	"github.com/multiversx/mx-chain-go/update/storing"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
)

const accountStartNonce = uint64(0)

type genesisBlockCreator struct {
	arg                 ArgsGenesisBlockCreator
	initialIndexingData map[uint32]*genesis.IndexingData
}

// NewGenesisBlockCreator creates a new genesis block creator instance able to create genesis blocks on all initial shards
func NewGenesisBlockCreator(arg ArgsGenesisBlockCreator) (*genesisBlockCreator, error) {
	err := checkArgumentsForBlockCreator(arg)
	if err != nil {
		return nil, fmt.Errorf("%w while creating NewGenesisBlockCreator", err)
	}

	indexingData := make(map[uint32]*genesis.IndexingData)

	gbc := &genesisBlockCreator{
		arg:                 arg,
		initialIndexingData: indexingData,
	}

	conversionBase := 10
	nodePrice, ok := big.NewInt(0).SetString(arg.SystemSCConfig.StakingSystemSCConfig.GenesisNodePrice, conversionBase)
	if !ok || nodePrice.Cmp(zero) <= 0 {
		return nil, genesis.ErrInvalidInitialNodePrice
	}
	gbc.arg.GenesisNodePrice = big.NewInt(0).Set(nodePrice)

	if mustDoHardForkImportProcess(gbc.arg) {
		err = gbc.createHardForkImportHandler()
		if err != nil {
			return nil, err
		}
	}

	return gbc, nil
}

func mustDoHardForkImportProcess(arg ArgsGenesisBlockCreator) bool {
	return arg.HardForkConfig.AfterHardFork && arg.StartEpochNum <= arg.HardForkConfig.StartEpoch
}

func getGenesisBlocksRoundNonceEpoch(arg ArgsGenesisBlockCreator) (uint64, uint64, uint32) {
	if arg.HardForkConfig.AfterHardFork {
		return arg.HardForkConfig.StartRound, arg.HardForkConfig.StartNonce, arg.HardForkConfig.StartEpoch
	}
	return 0, 0, 0
}

func (gbc *genesisBlockCreator) createHardForkImportHandler() error {
	importFolder := filepath.Join(gbc.arg.WorkingDir, gbc.arg.HardForkConfig.ImportFolder)

	// TODO remove duplicate code found in update/factory/exportHandlerFactory.go
	keysStorer, err := createStorer(gbc.arg.HardForkConfig.ImportKeysStorageConfig, importFolder)
	if err != nil {
		return fmt.Errorf("%w while creating keys storer", err)
	}
	keysVals, err := createStorer(gbc.arg.HardForkConfig.ImportStateStorageConfig, importFolder)
	if err != nil {
		return fmt.Errorf("%w while creating keys-values storer", err)
	}

	arg := storing.ArgHardforkStorer{
		KeysStore:   keysStorer,
		KeyValue:    keysVals,
		Marshalizer: gbc.arg.Core.InternalMarshalizer(),
	}
	hs, err := storing.NewHardforkStorer(arg)
	if err != nil {
		return fmt.Errorf("%w while creating hardfork storer", err)
	}

	argsHardForkImport := hardfork.ArgsNewStateImport{
		HardforkStorer:      hs,
		Hasher:              gbc.arg.Core.Hasher(),
		Marshalizer:         gbc.arg.Core.InternalMarshalizer(),
		ShardID:             gbc.arg.ShardCoordinator.SelfId(),
		StorageConfig:       gbc.arg.HardForkConfig.ImportStateStorageConfig,
		TrieStorageManagers: gbc.arg.TrieStorageManagers,
		AddressConverter:    gbc.arg.Core.AddressPubKeyConverter(),
	}
	importHandler, err := hardfork.NewStateImport(argsHardForkImport)
	if err != nil {
		return err
	}

	gbc.arg.importHandler = importHandler
	return nil
}

func createStorer(storageConfig config.StorageConfig, folder string) (storage.Storer, error) {
	dbConfig := factory.GetDBFromConfig(storageConfig.DB)
	dbConfig.FilePath = path.Join(folder, storageConfig.DB.FilePath)
	store, err := storageunit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(storageConfig.Cache),
		dbConfig,
	)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func checkArgumentsForBlockCreator(arg ArgsGenesisBlockCreator) error {
	if check.IfNil(arg.Accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(arg.Core) {
		return process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(arg.Data) {
		return process.ErrNilDataComponentsHolder
	}
	if check.IfNil(arg.Core.AddressPubKeyConverter()) {
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
	if check.IfNil(arg.Data.StorageService()) {
		return process.ErrNilStore
	}
	if check.IfNil(arg.Data.Blockchain()) {
		return process.ErrNilBlockChain
	}
	if check.IfNil(arg.Core.InternalMarshalizer()) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arg.Core.Hasher()) {
		return process.ErrNilHasher
	}
	if check.IfNil(arg.Core.Uint64ByteSliceConverter()) {
		return process.ErrNilUint64Converter
	}
	if check.IfNil(arg.Data.Datapool()) {
		return process.ErrNilPoolsHolder
	}
	if check.IfNil(arg.AccountsParser) {
		return genesis.ErrNilAccountsParser
	}
	if check.IfNil(arg.GasSchedule) {
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
	if check.IfNil(arg.Core.TxMarshalizer()) {
		return process.ErrNilMarshalizer
	}
	if arg.EpochConfig == nil {
		return genesis.ErrNilEpochConfig
	}
	if arg.GasSchedule == nil {
		return genesis.ErrNilGasSchedule
	}

	return nil
}

func mustDoGenesisProcess(arg ArgsGenesisBlockCreator) bool {
	genesisEpoch := uint32(0)
	if arg.HardForkConfig.AfterHardFork {
		genesisEpoch = arg.HardForkConfig.StartEpoch
	}

	if arg.StartEpochNum != genesisEpoch {
		return false
	}

	return true
}

func (gbc *genesisBlockCreator) createEmptyGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	err := gbc.computeDNSAddresses(createGenesisConfig())
	if err != nil {
		return nil, err
	}

	round, nonce, epoch := getGenesisBlocksRoundNonceEpoch(gbc.arg)

	mapEmptyGenesisBlocks := make(map[uint32]data.HeaderHandler)
	mapEmptyGenesisBlocks[core.MetachainShardId] = &block.MetaBlock{
		Round:     round,
		Nonce:     nonce,
		Epoch:     epoch,
		TimeStamp: gbc.arg.GenesisTime,
	}
	for i := uint32(0); i < gbc.arg.ShardCoordinator.NumberOfShards(); i++ {
		mapEmptyGenesisBlocks[i] = &block.Header{
			Round:     round,
			Nonce:     nonce,
			Epoch:     epoch,
			TimeStamp: gbc.arg.GenesisTime,
			ShardID:   i,
		}
	}

	return mapEmptyGenesisBlocks, nil
}

// GetIndexingData will return the initial data used for indexing
func (gbc *genesisBlockCreator) GetIndexingData() map[uint32]*genesis.IndexingData {
	return gbc.initialIndexingData
}

// CreateGenesisBlocks will try to create the genesis blocks for all shards
func (gbc *genesisBlockCreator) CreateGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	var err error
	var lastPostMbs []*update.MbInfo

	if !mustDoGenesisProcess(gbc.arg) {
		return gbc.createEmptyGenesisBlocks()
	}

	if mustDoHardForkImportProcess(gbc.arg) {
		err = gbc.arg.importHandler.ImportAll()
		if err != nil {
			return nil, err
		}

		err = gbc.computeDNSAddresses(gbc.arg.EpochConfig.EnableEpochs)
		if err != nil {
			return nil, err
		}
	}

	shardIDs := make([]uint32, gbc.arg.ShardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < gbc.arg.ShardCoordinator.NumberOfShards(); i++ {
		shardIDs[i] = i
	}
	shardIDs[gbc.arg.ShardCoordinator.NumberOfShards()] = core.MetachainShardId

	mapArgsGenesisBlockCreator := make(map[uint32]ArgsGenesisBlockCreator)
	mapHardForkBlockProcessor := make(map[uint32]update.HardForkBlockProcessor)
	mapBodies := make(map[uint32]*block.Body)

	err = gbc.createArgsGenesisBlockCreator(shardIDs, mapArgsGenesisBlockCreator)
	if err != nil {
		return nil, err
	}

	if mustDoHardForkImportProcess(gbc.arg) {
		selfShardID := gbc.arg.ShardCoordinator.SelfId()
		err = createHardForkBlockProcessors(selfShardID, shardIDs, mapArgsGenesisBlockCreator, mapHardForkBlockProcessor)
		if err != nil {
			return nil, err
		}

		args := update.ArgsHardForkProcessor{
			Hasher:                    gbc.arg.Core.Hasher(),
			Marshalizer:               gbc.arg.Core.InternalMarshalizer(),
			ShardIDs:                  shardIDs,
			MapBodies:                 mapBodies,
			MapHardForkBlockProcessor: mapHardForkBlockProcessor,
		}

		lastPostMbs, err = update.CreateBody(args)
		if err != nil {
			return nil, err
		}

		args.PostMbs = lastPostMbs
		err = update.CreatePostMiniBlocks(args)
		if err != nil {
			return nil, err
		}
	}

	genesisBlocks := make(map[uint32]data.HeaderHandler)
	err = gbc.createHeaders(mapArgsGenesisBlockCreator, mapHardForkBlockProcessor, mapBodies, shardIDs, genesisBlocks)
	if err != nil {
		return nil, err
	}

	// TODO call here trie pruning on all roothashes not from current shard

	return genesisBlocks, nil
}

func (gbc *genesisBlockCreator) createHeaders(
	mapArgsGenesisBlockCreator map[uint32]ArgsGenesisBlockCreator,
	mapHardForkBlockProcessor map[uint32]update.HardForkBlockProcessor,
	mapBodies map[uint32]*block.Body,
	shardIDs []uint32,
	genesisBlocks map[uint32]data.HeaderHandler,
) error {
	var nodesListSplitter genesis.NodesListSplitter
	var err error

	nodesListSplitter, err = intermediate.NewNodesListSplitter(gbc.arg.InitialNodesSetup, gbc.arg.AccountsParser)
	if err != nil {
		return err
	}

	allScAddresses := make([][]byte, 0)
	for _, shardID := range shardIDs {
		log.Debug("genesisBlockCreator.createHeaders", "shard", shardID)
		var genesisBlock data.HeaderHandler
		var scResults [][]byte
		var chain data.ChainHandler

		if shardID == core.MetachainShardId {
			metaArgsGenesisBlockCreator := mapArgsGenesisBlockCreator[core.MetachainShardId]
			chain, err = blockchain.NewMetaChain(&statusHandler.NilStatusHandler{})
			if err != nil {
				return fmt.Errorf("'%w' while generating genesis block for metachain", err)
			}

			err = metaArgsGenesisBlockCreator.Data.SetBlockchain(chain)
			if err != nil {
				return fmt.Errorf("'%w' while setting blockchain for metachain", err)
			}
			genesisBlock, scResults, gbc.initialIndexingData[shardID], err = CreateMetaGenesisBlock(
				metaArgsGenesisBlockCreator,
				mapBodies[core.MetachainShardId],
				nodesListSplitter,
				mapHardForkBlockProcessor[core.MetachainShardId],
			)
		} else {
			genesisBlock, scResults, gbc.initialIndexingData[shardID], err = CreateShardGenesisBlock(
				mapArgsGenesisBlockCreator[shardID],
				mapBodies[shardID],
				nodesListSplitter,
				mapHardForkBlockProcessor[shardID],
			)
		}
		if err != nil {
			return fmt.Errorf("'%w' while generating genesis block for shard %d", err, shardID)
		}

		allScAddresses = append(allScAddresses, scResults...)
		genesisBlocks[shardID] = genesisBlock
		err = gbc.saveGenesisBlock(genesisBlock)
		if err != nil {
			return fmt.Errorf("'%w' while saving genesis block for shard %d", err, shardID)
		}
	}

	err = gbc.checkDelegationsAgainstDeployedSC(allScAddresses, gbc.arg)
	if err != nil {
		return err
	}

	for _, shardID := range shardIDs {
		gb := genesisBlocks[shardID]

		log.Info("genesisBlockCreator.createHeaders",
			"shard", gb.GetShardID(),
			"nonce", gb.GetNonce(),
			"round", gb.GetRound(),
			"root hash", gb.GetRootHash(),
		)
	}

	return nil
}

// in case of hardfork initial smart contracts deployment is not called as they are all imported from previous state
func (gbc *genesisBlockCreator) computeDNSAddresses(enableEpochsConfig config.EnableEpochs) error {
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
	epochNotifier := forking.NewGenericEpochNotifier()
	temporaryMetaHeader := &block.MetaBlock{
		Epoch:     gbc.arg.StartEpochNum,
		TimeStamp: gbc.arg.GenesisTime,
	}
	enableEpochsHandler, err := enablers.NewEnableEpochsHandler(enableEpochsConfig, epochNotifier)
	if err != nil {
		return err
	}
	epochNotifier.CheckEpoch(temporaryMetaHeader)

	builtInFuncs := vmcommonBuiltInFunctions.NewBuiltInFunctionContainer()
	argsHook := hooks.ArgBlockChainHook{
		Accounts:              gbc.arg.Accounts,
		PubkeyConv:            gbc.arg.Core.AddressPubKeyConverter(),
		StorageService:        gbc.arg.Data.StorageService(),
		BlockChain:            gbc.arg.Data.Blockchain(),
		ShardCoordinator:      gbc.arg.ShardCoordinator,
		Marshalizer:           gbc.arg.Core.InternalMarshalizer(),
		Uint64Converter:       gbc.arg.Core.Uint64ByteSliceConverter(),
		BuiltInFunctions:      builtInFuncs,
		NFTStorageHandler:     &disabled.SimpleNFTStorage{},
		GlobalSettingsHandler: &disabled.ESDTGlobalSettingsHandler{},
		DataPool:              gbc.arg.Data.Datapool(),
		CompiledSCPool:        gbc.arg.Data.Datapool().SmartContracts(),
		EpochNotifier:         epochNotifier,
		EnableEpochsHandler:   enableEpochsHandler,
		NilCompiledSCStore:    true,
		GasSchedule:           gbc.arg.GasSchedule,
		Counter:               counters.NewDisabledCounter(),
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

		encodedSCResultingAddress, err := gbc.arg.Core.AddressPubKeyConverter().Encode(scResultingAddress)
		if err != nil {
			return err
		}
		dnsSC.AddAddress(encodedSCResultingAddress)
	}

	return nil
}

func (gbc *genesisBlockCreator) getNewArgForShard(shardID uint32) (ArgsGenesisBlockCreator, error) {
	var err error

	isCurrentShard := shardID == gbc.arg.ShardCoordinator.SelfId()
	if isCurrentShard {
		newArgument := gbc.arg // copy the arguments
		newArgument.Data = newArgument.Data.Clone().(dataComponentsHandler)
		return newArgument, nil
	}

	newArgument := gbc.arg // copy the arguments
	newArgument.Accounts, err = createAccountAdapter(
		newArgument.Core.InternalMarshalizer(),
		newArgument.Core.Hasher(),
		factoryState.NewAccountCreator(),
		gbc.arg.TrieStorageManagers[triesFactory.UserAccountTrie],
		gbc.arg.Core.AddressPubKeyConverter(),
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

	// create copy of components handlers we need to change temporarily
	newArgument.Data = newArgument.Data.Clone().(dataComponentsHandler)
	return newArgument, err
}

func (gbc *genesisBlockCreator) saveGenesisBlock(header data.HeaderHandler) error {
	blockBuff, err := gbc.arg.Core.InternalMarshalizer().Marshal(header)
	if err != nil {
		return err
	}

	hash := gbc.arg.Core.Hasher().Compute(string(blockBuff))
	unitType := dataRetriever.BlockHeaderUnit
	if header.GetShardID() == core.MetachainShardId {
		unitType = dataRetriever.MetaBlockUnit
	}

	return gbc.arg.Data.StorageService().Put(unitType, hash, blockBuff)
}

func (gbc *genesisBlockCreator) checkDelegationsAgainstDeployedSC(
	allScAddresses [][]byte,
	arg ArgsGenesisBlockCreator,
) error {
	if mustDoHardForkImportProcess(arg) {
		return nil
	}

	initialAccounts := arg.AccountsParser.InitialAccounts()
	for _, ia := range initialAccounts {
		dh := ia.GetDelegationHandler()
		if check.IfNil(dh) {
			continue
		}
		if len(dh.AddressBytes()) == 0 {
			continue
		}

		found := gbc.searchDeployedContract(allScAddresses, dh.AddressBytes())
		if !found {
			return fmt.Errorf("%w for SC address %s, address %s",
				genesis.ErrMissingDeployedSC, dh.GetAddress(), ia.GetAddress())
		}
	}

	return nil
}

func (gbc *genesisBlockCreator) searchDeployedContract(allScAddresses [][]byte, address []byte) bool {
	for _, addr := range allScAddresses {
		if bytes.Equal(addr, address) {
			return true
		}
	}

	return false
}

// ImportHandler returns the ImportHandler object
func (gbc *genesisBlockCreator) ImportHandler() update.ImportHandler {
	return gbc.arg.importHandler
}

func (gbc *genesisBlockCreator) createArgsGenesisBlockCreator(
	shardIDs []uint32,
	mapArgsGenesisBlockCreator map[uint32]ArgsGenesisBlockCreator,
) error {
	for _, shardID := range shardIDs {
		log.Debug("createArgsGenesisBlockCreator", "shard", shardID)
		newArgument, err := gbc.getNewArgForShard(shardID)
		if err != nil {
			return fmt.Errorf("'%w' while creating new argument for shard %d", err, shardID)
		}

		mapArgsGenesisBlockCreator[shardID] = newArgument
	}

	return nil
}

func createHardForkBlockProcessors(
	selfShardID uint32,
	shardIDs []uint32,
	mapArgsGenesisBlockCreator map[uint32]ArgsGenesisBlockCreator,
	mapHardForkBlockProcessor map[uint32]update.HardForkBlockProcessor,
) error {
	var hardForkBlockProcessor update.HardForkBlockProcessor
	var err error
	for _, shardID := range shardIDs {
		log.Debug("createHarForkBlockProcessor", "shard", shardID)
		if shardID == core.MetachainShardId {
			argsMetaBlockCreatorAfterHardFork, errCreate := createArgsMetaBlockCreatorAfterHardFork(mapArgsGenesisBlockCreator[shardID], selfShardID)
			if errCreate != nil {
				return errCreate
			}
			hardForkBlockProcessor, err = hardForkProcess.NewMetaBlockCreatorAfterHardfork(argsMetaBlockCreatorAfterHardFork)
			if err != nil {
				return err
			}

		} else {
			argsShardBlockAfterHardFork, errCreate := createArgsShardBlockCreatorAfterHardFork(mapArgsGenesisBlockCreator[shardID], selfShardID)
			if errCreate != nil {
				return errCreate
			}

			hardForkBlockProcessor, err = hardForkProcess.NewShardBlockCreatorAfterHardFork(argsShardBlockAfterHardFork)
			if err != nil {
				return err
			}
		}

		mapHardForkBlockProcessor[shardID] = hardForkBlockProcessor
	}

	return nil
}
