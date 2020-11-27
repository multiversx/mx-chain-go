package process

import (
	"bytes"
	"fmt"
	"math/big"
	"path"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/config"
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
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/update"
	hardForkProcess "github.com/ElrondNetwork/elrond-go/update/process"
	"github.com/ElrondNetwork/elrond-go/update/storing"

	hardfork "github.com/ElrondNetwork/elrond-go/update/genesis"
)

const accountStartNonce = uint64(0)

type genesisBlockCreator struct {
	arg ArgsGenesisBlockCreator
}

// NewGenesisBlockCreator creates a new genesis block creator instance able to create genesis blocks on all initial shards
func NewGenesisBlockCreator(arg ArgsGenesisBlockCreator) (*genesisBlockCreator, error) {
	err := checkArgumentsForBlockCreator(arg)
	if err != nil {
		return nil, fmt.Errorf("%w while creating NewGenesisBlockCreator", err)
	}

	gbc := &genesisBlockCreator{
		arg: arg,
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

	//TODO remove duplicate code found in update/factory/exportHandlerFactory.go
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
		Marshalizer: gbc.arg.Marshalizer,
	}
	hs, err := storing.NewHardforkStorer(arg)

	argsHardForkImport := hardfork.ArgsNewStateImport{
		HardforkStorer:      hs,
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

func createStorer(storageConfig config.StorageConfig, folder string) (storage.Storer, error) {
	dbConfig := factory.GetDBFromConfig(storageConfig.DB)
	dbConfig.FilePath = path.Join(folder, storageConfig.DB.FilePath)
	store, err := storageUnit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(storageConfig.Cache),
		dbConfig,
		factory.GetBloomFromConfig(storageConfig.Bloom),
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
	if check.IfNil(arg.SignMarshalizer) {
		return process.ErrNilMarshalizer
	}
	if arg.GeneralConfig == nil {
		return genesis.ErrNilGeneralSettingsConfig
	}

	return nil
}

func mustDoGenesisProcess(arg ArgsGenesisBlockCreator) bool {
	genesisEpoch := uint32(0)
	if arg.HardForkConfig.AfterHardFork == true {
		genesisEpoch = arg.HardForkConfig.StartEpoch
	}

	if arg.StartEpochNum != genesisEpoch {
		return false
	}

	return true
}

func (gbc *genesisBlockCreator) createEmptyGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	err := gbc.computeDNSAddresses()
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
		}
	}

	return mapEmptyGenesisBlocks, nil
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

		err = gbc.computeDNSAddresses()
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

	err = createArgsGenesisBlockCreator(shardIDs, mapArgsGenesisBlockCreator, gbc.arg)
	if err != nil {
		return nil, err
	}

	if mustDoHardForkImportProcess(gbc.arg) {
		selfShardID := gbc.arg.ShardCoordinator.SelfId()
		err = createHardForkBlockProcessors(selfShardID, shardIDs, mapArgsGenesisBlockCreator, mapHardForkBlockProcessor)
		if err != nil {
			return nil, err
		}

		lastPostMbs, err = update.CreateBody(
			gbc.arg.Hasher,
			gbc.arg.Marshalizer,
			shardIDs,
			mapBodies,
			mapHardForkBlockProcessor,
		)
		if err != nil {
			return nil, err
		}

		err = update.CreatePostMiniBlocks(
			gbc.arg.Hasher,
			gbc.arg.Marshalizer,
			shardIDs,
			lastPostMbs,
			mapBodies,
			mapHardForkBlockProcessor,
		)
		if err != nil {
			return nil, err
		}
	}

	genesisBlocks := make(map[uint32]data.HeaderHandler)
	err = gbc.createHeaders(mapArgsGenesisBlockCreator, mapHardForkBlockProcessor, mapBodies, shardIDs, genesisBlocks)
	if err != nil {
		return nil, err
	}

	//TODO call here trie pruning on all roothashes not from current shard

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

		if shardID == core.MetachainShardId {
			metaArgsGenesisBlockCreator := mapArgsGenesisBlockCreator[core.MetachainShardId]
			metaArgsGenesisBlockCreator.Blkc = blockchain.NewMetaChain()
			genesisBlock, scResults, err = CreateMetaGenesisBlock(
				metaArgsGenesisBlockCreator,
				mapBodies[core.MetachainShardId],
				nodesListSplitter,
				mapHardForkBlockProcessor[core.MetachainShardId],
			)
		} else {
			genesisBlock, scResults, err = CreateShardGenesisBlock(
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
func (gbc *genesisBlockCreator) computeDNSAddresses() error {
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
		Accounts:           gbc.arg.Accounts,
		PubkeyConv:         gbc.arg.PubkeyConv,
		StorageService:     gbc.arg.Store,
		BlockChain:         gbc.arg.Blkc,
		ShardCoordinator:   gbc.arg.ShardCoordinator,
		Marshalizer:        gbc.arg.Marshalizer,
		Uint64Converter:    gbc.arg.Uint64ByteSliceConverter,
		BuiltInFunctions:   builtInFuncs,
		DataPool:           gbc.arg.DataPool,
		CompiledSCPool:     gbc.arg.DataPool.SmartContracts(),
		NilCompiledSCStore: true,
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

func getNewArgForShard(shardID uint32, arg ArgsGenesisBlockCreator) (ArgsGenesisBlockCreator, error) {
	var err error

	isCurrentShard := shardID == arg.ShardCoordinator.SelfId()
	if isCurrentShard {
		return arg, nil
	}

	newArgument := arg //copy the arguments
	newArgument.Accounts, err = createAccountAdapter(
		newArgument.Marshalizer,
		newArgument.Hasher,
		factoryState.NewAccountCreator(),
		arg.TrieStorageManagers[triesFactory.UserAccountTrie],
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

func createArgsGenesisBlockCreator(
	shardIDs []uint32,
	mapArgsGenesisBlockCreator map[uint32]ArgsGenesisBlockCreator,
	arg ArgsGenesisBlockCreator,
) error {
	for _, shardID := range shardIDs {
		log.Debug("createArgsGenesisBlockCreator", "shard ID", shardID)
		newArgument, err := getNewArgForShard(shardID, arg)
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
		log.Debug("createHarForkBlockProcessor", "shard ID", shardID)
		if shardID == core.MetachainShardId {
			argsMetaBlockCreatorAfterHardFork, errCreate := createArgsMetaBlockCreatorAfterHardFork(mapArgsGenesisBlockCreator[shardID])
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
