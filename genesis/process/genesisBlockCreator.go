package process

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"

	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	epochStart "github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	factoryBlock "github.com/multiversx/mx-chain-go/factory/block"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/genesis/process/intermediate"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks/counters"
	"github.com/multiversx/mx-chain-go/sharding"
	disabledState "github.com/multiversx/mx-chain-go/state/disabled"
	factoryState "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/statusHandler"
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

	return gbc, nil
}

func getGenesisBlocksRoundNonceEpoch(arg ArgsGenesisBlockCreator) (uint64, uint64, uint32) {
	return arg.GenesisRound, arg.GenesisNonce, arg.GenesisEpoch
}

func checkArgumentsForBlockCreator(arg ArgsGenesisBlockCreator) error {
	if check.IfNil(arg.Accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(arg.AccountsProposal) {
		return fmt.Errorf("%w for proposal", process.ErrNilAccountsAdapter)
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
	if check.IfNil(arg.Core.InternalMarshalizer()) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arg.Core.Hasher()) {
		return process.ErrNilHasher
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
	if check.IfNil(arg.SmartContractParser) {
		return genesis.ErrNilSmartContractParser
	}
	if arg.TrieStorageManagers == nil {
		return genesis.ErrNilTrieStorageManager
	}
	if check.IfNil(arg.HistoryRepository) {
		return process.ErrNilHistoryRepository
	}
	if check.IfNil(arg.TxExecutionOrderHandler) {
		return process.ErrNilTxExecutionOrderHandler
	}

	return nil
}

func mustDoGenesisProcess(arg ArgsGenesisBlockCreator) bool {
	return arg.StartEpochNum == arg.GenesisEpoch
}

func (gbc *genesisBlockCreator) createEmptyGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	err := gbc.computeDNSAddresses(createGenesisConfig(gbc.arg.EpochConfig.EnableEpochs))
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
	if !mustDoGenesisProcess(gbc.arg) {
		return gbc.createEmptyGenesisBlocks()
	}

	shardIDs := make([]uint32, gbc.arg.ShardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < gbc.arg.ShardCoordinator.NumberOfShards(); i++ {
		shardIDs[i] = i
	}
	shardIDs[gbc.arg.ShardCoordinator.NumberOfShards()] = core.MetachainShardId

	mapArgsGenesisBlockCreator := make(map[uint32]ArgsGenesisBlockCreator)
	mapBodies := make(map[uint32]*block.Body)

	err := gbc.createArgsGenesisBlockCreator(shardIDs, mapArgsGenesisBlockCreator)
	if err != nil {
		return nil, err
	}

	genesisBlocks := make(map[uint32]data.HeaderHandler)
	err = gbc.createHeaders(mapArgsGenesisBlockCreator, mapBodies, shardIDs, genesisBlocks)
	if err != nil {
		return nil, err
	}

	// TODO call here trie pruning on all roothashes not from current shard

	return genesisBlocks, nil
}

func (gbc *genesisBlockCreator) createHeaders(
	mapArgsGenesisBlockCreator map[uint32]ArgsGenesisBlockCreator,
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
			)
		} else {
			genesisBlock, scResults, gbc.initialIndexingData[shardID], err = CreateShardGenesisBlock(
				mapArgsGenesisBlockCreator[shardID],
				mapBodies[shardID],
				nodesListSplitter,
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

// computeDNSAddresses computes DNS addresses for initial smart contracts
func (gbc *genesisBlockCreator) computeDNSAddresses(
	enableEpochsConfig config.EnableEpochs,
) error {
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
		Accounts:                 gbc.arg.Accounts,
		PubkeyConv:               gbc.arg.Core.AddressPubKeyConverter(),
		StorageService:           gbc.arg.Data.StorageService(),
		BlockChain:               gbc.arg.Data.Blockchain(),
		ShardCoordinator:         gbc.arg.ShardCoordinator,
		Marshalizer:              gbc.arg.Core.InternalMarshalizer(),
		Uint64Converter:          gbc.arg.Core.Uint64ByteSliceConverter(),
		BuiltInFunctions:         builtInFuncs,
		NFTStorageHandler:        &disabled.SimpleNFTStorage{},
		GlobalSettingsHandler:    &disabled.ESDTGlobalSettingsHandler{},
		DataPool:                 gbc.arg.Data.Datapool(),
		CompiledSCPool:           gbc.arg.Data.Datapool().SmartContracts(),
		EpochNotifier:            epochNotifier,
		EnableEpochsHandler:      enableEpochsHandler,
		NilCompiledSCStore:       true,
		GasSchedule:              gbc.arg.GasSchedule,
		Counter:                  counters.NewDisabledCounter(),
		MissingTrieNodesNotifier: syncer.NewMissingTrieNodesNotifier(),
		EpochStartTrigger:        epochStart.NewEpochStartTrigger(),
		RoundHandler:             &disabled.RoundHandler{},
	}
	blockChainHook, err := hooks.NewBlockChainHookImpl(argsHook)
	if err != nil {
		return err
	}

	isForCurrentShard := func([]byte) bool {
		// we are interested only in the smart contract addresses, as they are already deployed
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
	newArgument := gbc.arg // copy the arguments
	newArgument.versionedHeaderFactory, err = gbc.createVersionedHeaderFactory()
	if err != nil {
		return ArgsGenesisBlockCreator{}, fmt.Errorf("'%w' while generating a VersionedHeaderFactory instance for shard %d",
			err, shardID)
	}

	if isCurrentShard {
		newArgument.Data = newArgument.Data.Clone().(dataComponentsHandler)
		return newArgument, nil
	}

	argsAccCreator := factoryState.ArgsAccountCreator{
		Hasher:                 newArgument.Core.Hasher(),
		Marshaller:             newArgument.Core.InternalMarshalizer(),
		EnableEpochsHandler:    newArgument.Core.EnableEpochsHandler(),
		StateAccessesCollector: disabledState.NewDisabledStateAccessesCollector(),
	}
	accCreator, err := factoryState.NewAccountCreator(argsAccCreator)
	if err != nil {
		return ArgsGenesisBlockCreator{}, err
	}

	newArgument.Accounts, err = createAccountAdapter(
		newArgument.Core.InternalMarshalizer(),
		newArgument.Core.Hasher(),
		accCreator,
		gbc.arg.TrieStorageManagers[dataRetriever.UserAccountsUnit.String()],
		gbc.arg.Core.AddressPubKeyConverter(),
		newArgument.Core.EnableEpochsHandler(),
	)
	if err != nil {
		return ArgsGenesisBlockCreator{}, fmt.Errorf("'%w' while generating an in-memory accounts adapter for shard %d",
			err, shardID)
	}
	// for genesis we can reuse the same account adapter for proposal as first proposal needs to happen after genesis block execution
	// and the proposal won't use the genesis block creator
	newArgument.AccountsProposal = newArgument.Accounts

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

func (gbc *genesisBlockCreator) createVersionedHeaderFactory() (genesis.VersionedHeaderFactory, error) {
	headerVersionHandler, err := factoryBlock.NewHeaderVersionHandler(
		gbc.arg.HeaderVersionConfigs.VersionsByEpochs,
		gbc.arg.HeaderVersionConfigs.DefaultVersion,
	)
	if err != nil {
		return nil, err
	}

	return factoryBlock.NewShardHeaderFactory(headerVersionHandler)
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
