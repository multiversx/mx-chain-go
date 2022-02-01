package process

import (
	"bytes"
	"encoding/hex"
	"math"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	syncDisabled "github.com/ElrondNetwork/elrond-go/process/sync/disabled"
	processTransaction "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	"github.com/ElrondNetwork/elrond-go/update"
	hardForkProcess "github.com/ElrondNetwork/elrond-go/update/process"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	vmcommonBuiltInFunctions "github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

const unreachableEpoch = uint32(1000000)

// CreateMetaGenesisBlock will create a metachain genesis block
func CreateMetaGenesisBlock(
	arg ArgsGenesisBlockCreator,
	body *block.Body,
	nodesListSplitter genesis.NodesListSplitter,
	hardForkBlockProcessor update.HardForkBlockProcessor,
) (data.MetaHeaderHandler, [][]byte, *genesis.IndexingData, error) {
	if mustDoHardForkImportProcess(arg) {
		return createMetaGenesisBlockAfterHardFork(arg, body, hardForkBlockProcessor)
	}

	indexingData := &genesis.IndexingData{
		DelegationTxs:      make([]data.TransactionHandler, 0),
		ScrsTxs:            make(map[string]data.TransactionHandler),
		StakingTxs:         make([]data.TransactionHandler, 0),
		DeploySystemScTxs:  make([]data.TransactionHandler, 0),
		DeployInitialScTxs: make([]data.TransactionHandler, 0),
	}

	processors, err := createProcessorsForMetaGenesisBlock(arg, createGenesisConfig())
	if err != nil {
		return nil, nil, nil, err
	}

	deploySystemSCTxs, err := deploySystemSmartContracts(arg, processors.txProcessor, processors.systemSCs)
	if err != nil {
		return nil, nil, nil, err
	}
	indexingData.DeploySystemScTxs = deploySystemSCTxs

	stakingTxs, err := setStakedData(arg, processors, nodesListSplitter)
	if err != nil {
		return nil, nil, nil, err
	}
	indexingData.StakingTxs = stakingTxs

	rootHash, err := arg.Accounts.Commit()
	if err != nil {
		return nil, nil, nil, err
	}

	scrsTxs := processors.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
	indexingData.ScrsTxs = scrsTxs

	round, nonce, epoch := getGenesisBlocksRoundNonceEpoch(arg)

	magicDecoded, err := hex.DecodeString(arg.GenesisString)
	if err != nil {
		return nil, nil, nil, err
	}
	prevHash := arg.Core.Hasher().Compute(arg.GenesisString)

	header := &block.MetaBlock{
		RootHash:               rootHash,
		PrevHash:               prevHash,
		RandSeed:               rootHash,
		PrevRandSeed:           rootHash,
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		PubKeysBitmap:          []byte{1},
		ChainID:                []byte(arg.Core.ChainID()),
		SoftwareVersion:        []byte(""),
		TimeStamp:              arg.GenesisTime,
		Round:                  round,
		Nonce:                  nonce,
		Epoch:                  epoch,
		Reserved:               magicDecoded,
	}

	header.EpochStart.Economics = block.Economics{
		TotalSupply:       big.NewInt(0).Set(arg.Economics.GenesisTotalSupply()),
		TotalToDistribute: big.NewInt(0),
		TotalNewlyMinted:  big.NewInt(0),
		RewardsPerBlock:   big.NewInt(0),
		NodePrice:         big.NewInt(0).Set(arg.GenesisNodePrice),
	}

	validatorRootHash, err := arg.ValidatorAccounts.RootHash()
	if err != nil {
		return nil, nil, nil, err
	}

	err = header.SetValidatorStatsRootHash(validatorRootHash)
	if err != nil {
		return nil, nil, nil, err
	}

	err = saveGenesisMetaToStorage(arg.Data.StorageService(), arg.Core.InternalMarshalizer(), header)
	if err != nil {
		return nil, nil, nil, err
	}

	err = processors.vmContainer.Close()
	if err != nil {
		return nil, nil, nil, err
	}

	return header, make([][]byte, 0), indexingData, nil
}

// TODO: index the resulted transactions after a hardfork
func createMetaGenesisBlockAfterHardFork(
	arg ArgsGenesisBlockCreator,
	body *block.Body,
	hardForkBlockProcessor update.HardForkBlockProcessor,
) (data.MetaHeaderHandler, [][]byte, *genesis.IndexingData, error) {
	if check.IfNil(hardForkBlockProcessor) {
		return nil, nil, nil, update.ErrNilHardForkBlockProcessor
	}

	hdrHandler, err := hardForkBlockProcessor.CreateBlock(
		body,
		arg.Core.ChainID(),
		arg.HardForkConfig.StartRound,
		arg.HardForkConfig.StartNonce,
		arg.HardForkConfig.StartEpoch,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	err = hdrHandler.SetTimeStamp(arg.GenesisTime)
	if err != nil {
		return nil, nil, nil, err
	}

	metaHdr, ok := hdrHandler.(*block.MetaBlock)
	if !ok {
		return nil, nil, nil, process.ErrWrongTypeAssertion
	}

	err = arg.Accounts.RecreateTrie(hdrHandler.GetRootHash())
	if err != nil {
		return nil, nil, nil, err
	}

	err = saveGenesisMetaToStorage(arg.Data.StorageService(), arg.Core.InternalMarshalizer(), metaHdr)
	if err != nil {
		return nil, nil, nil, err
	}

	indexingData := &genesis.IndexingData{
		DelegationTxs:      make([]data.TransactionHandler, 0),
		ScrsTxs:            make(map[string]data.TransactionHandler),
		StakingTxs:         make([]data.TransactionHandler, 0),
		DeploySystemScTxs:  make([]data.TransactionHandler, 0),
		DeployInitialScTxs: make([]data.TransactionHandler, 0),
	}

	return metaHdr, make([][]byte, 0), indexingData, nil
}

func createArgsMetaBlockCreatorAfterHardFork(
	arg ArgsGenesisBlockCreator,
	selfShardID uint32,
) (hardForkProcess.ArgsNewMetaBlockCreatorAfterHardFork, error) {
	tmpArg := arg
	tmpArg.Accounts = arg.importHandler.GetAccountsDBForShard(core.MetachainShardId)
	processors, err := createProcessorsForMetaGenesisBlock(tmpArg, arg.EpochConfig.EnableEpochs)
	if err != nil {
		return hardForkProcess.ArgsNewMetaBlockCreatorAfterHardFork{}, err
	}

	argsPendingTxProcessor := hardForkProcess.ArgsPendingTransactionProcessor{
		Accounts:         tmpArg.Accounts,
		TxProcessor:      processors.txProcessor,
		RwdTxProcessor:   &disabled.RewardTxProcessor{},
		ScrTxProcessor:   processors.scrProcessor,
		PubKeyConv:       arg.Core.AddressPubKeyConverter(),
		ShardCoordinator: arg.ShardCoordinator,
	}
	pendingTxProcessor, err := hardForkProcess.NewPendingTransactionProcessor(argsPendingTxProcessor)
	if err != nil {
		return hardForkProcess.ArgsNewMetaBlockCreatorAfterHardFork{}, err
	}

	argsMetaBlockCreatorAfterHardFork := hardForkProcess.ArgsNewMetaBlockCreatorAfterHardFork{
		Hasher:             arg.Core.Hasher(),
		ImportHandler:      arg.importHandler,
		Marshalizer:        arg.Core.InternalMarshalizer(),
		PendingTxProcessor: pendingTxProcessor,
		ShardCoordinator:   arg.ShardCoordinator,
		Storage:            arg.Data.StorageService(),
		TxCoordinator:      processors.txCoordinator,
		ValidatorAccounts:  tmpArg.ValidatorAccounts,
		SelfShardID:        selfShardID,
	}

	return argsMetaBlockCreatorAfterHardFork, nil
}

func saveGenesisMetaToStorage(
	storageService dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	genesisBlock data.HeaderHandler,
) error {

	epochStartID := core.EpochStartIdentifier(genesisBlock.GetEpoch())

	metaHdrStorage := storageService.GetStorer(dataRetriever.MetaBlockUnit)
	if check.IfNil(metaHdrStorage) {
		return process.ErrNilStorage
	}

	triggerStorage := storageService.GetStorer(dataRetriever.BootstrapUnit)
	if check.IfNil(triggerStorage) {
		return process.ErrNilStorage
	}

	marshaledData, err := marshalizer.Marshal(genesisBlock)
	if err != nil {
		return err
	}

	err = metaHdrStorage.Put([]byte(epochStartID), marshaledData)
	if err != nil {
		return err
	}

	err = triggerStorage.Put([]byte(epochStartID), marshaledData)
	if err != nil {
		return err
	}

	return nil
}

func createProcessorsForMetaGenesisBlock(arg ArgsGenesisBlockCreator, enableEpochs config.EnableEpochs) (*genesisProcessors, error) {
	epochNotifier := forking.NewGenericEpochNotifier()
	temporaryMetaHeader := &block.MetaBlock{
		Epoch:     arg.StartEpochNum,
		TimeStamp: arg.GenesisTime,
	}
	epochNotifier.CheckEpoch(temporaryMetaHeader)

	builtInFuncs := vmcommonBuiltInFunctions.NewBuiltInFunctionContainer()
	argsHook := hooks.ArgBlockChainHook{
		Accounts:           arg.Accounts,
		PubkeyConv:         arg.Core.AddressPubKeyConverter(),
		StorageService:     arg.Data.StorageService(),
		BlockChain:         arg.Data.Blockchain(),
		ShardCoordinator:   arg.ShardCoordinator,
		Marshalizer:        arg.Core.InternalMarshalizer(),
		Uint64Converter:    arg.Core.Uint64ByteSliceConverter(),
		BuiltInFunctions:   builtInFuncs,
		NFTStorageHandler:  &disabled.SimpleNFTStorage{},
		DataPool:           arg.Data.Datapool(),
		CompiledSCPool:     arg.Data.Datapool().SmartContracts(),
		EpochNotifier:      epochNotifier,
		NilCompiledSCStore: true,
		EnableEpochs:       enableEpochs,
	}

	pubKeyVerifier, err := disabled.NewMessageSignVerifier(arg.BlockSignKeyGen)
	if err != nil {
		return nil, err
	}
	argsNewVMContainerFactory := metachain.ArgsNewVMContainerFactory{
		ArgBlockChainHook:   argsHook,
		Economics:           arg.Economics,
		MessageSignVerifier: pubKeyVerifier,
		GasSchedule:         arg.GasSchedule,
		NodesConfigProvider: arg.InitialNodesSetup,
		Hasher:              arg.Core.Hasher(),
		Marshalizer:         arg.Core.InternalMarshalizer(),
		SystemSCConfig:      &arg.SystemSCConfig,
		ValidatorAccountsDB: arg.ValidatorAccounts,
		ChanceComputer:      &disabled.Rater{},
		EpochNotifier:       epochNotifier,
		EpochConfig:         arg.EpochConfig,
		ShardCoordinator:    arg.ShardCoordinator,
	}
	virtualMachineFactory, err := metachain.NewVMContainerFactory(argsNewVMContainerFactory)
	if err != nil {
		return nil, err
	}

	vmContainer, err := virtualMachineFactory.CreateForGenesis()
	if err != nil {
		return nil, err
	}

	genesisFeeHandler := &disabled.FeeHandler{}
	interimProcFactory, err := metachain.NewIntermediateProcessorsContainerFactory(
		arg.ShardCoordinator,
		arg.Core.InternalMarshalizer(),
		arg.Core.Hasher(),
		arg.Core.AddressPubKeyConverter(),
		arg.Data.StorageService(),
		arg.Data.Datapool(),
		genesisFeeHandler,
	)
	if err != nil {
		return nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, err
	}

	scForwarder, err := interimProcContainer.Get(block.SmartContractResultBlock)
	if err != nil {
		return nil, err
	}

	badTxForwarder, err := interimProcContainer.Get(block.InvalidBlock)
	if err != nil {
		return nil, err
	}

	esdtTransferParser, err := parsers.NewESDTTransferParser(arg.Core.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:        arg.Core.AddressPubKeyConverter(),
		ShardCoordinator:       arg.ShardCoordinator,
		BuiltInFunctions:       builtInFuncs,
		ArgumentParser:         parsers.NewCallArgsParser(),
		EpochNotifier:          epochNotifier,
		RelayedTxV2EnableEpoch: arg.EpochConfig.EnableEpochs.RelayedTransactionsV2EnableEpoch,
		ESDTTransferParser:     esdtTransferParser,
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(arg.Economics, txTypeHandler, epochNotifier, enableEpochs.SCDeployEnableEpoch)
	if err != nil {
		return nil, err
	}

	argsParser := smartContract.NewArgumentParser()
	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:         vmContainer,
		ArgsParser:          argsParser,
		Hasher:              arg.Core.Hasher(),
		Marshalizer:         arg.Core.InternalMarshalizer(),
		AccountsDB:          arg.Accounts,
		BlockChainHook:      virtualMachineFactory.BlockChainHookImpl(),
		BuiltInFunctions:    builtInFuncs,
		PubkeyConv:          arg.Core.AddressPubKeyConverter(),
		ShardCoordinator:    arg.ShardCoordinator,
		ScrForwarder:        scForwarder,
		TxFeeHandler:        genesisFeeHandler,
		EconomicsFee:        genesisFeeHandler,
		TxTypeHandler:       txTypeHandler,
		GasHandler:          gasHandler,
		GasSchedule:         arg.GasSchedule,
		TxLogsProcessor:     arg.TxLogsProcessor,
		BadTxForwarder:      badTxForwarder,
		EpochNotifier:       epochNotifier,
		EnableEpochs:        enableEpochs,
		IsGenesisProcessing: true,
		ArwenChangeLocker:   &sync.RWMutex{}, // local Locker as to not interfere with the rest of the components
		VMOutputCacher:      txcache.NewDisabledCache(),
	}
	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewSCProcessor)
	if err != nil {
		return nil, err
	}

	argsNewMetaTxProcessor := processTransaction.ArgsNewMetaTxProcessor{
		Hasher:                                arg.Core.Hasher(),
		Marshalizer:                           arg.Core.InternalMarshalizer(),
		Accounts:                              arg.Accounts,
		PubkeyConv:                            arg.Core.AddressPubKeyConverter(),
		ShardCoordinator:                      arg.ShardCoordinator,
		ScProcessor:                           scProcessor,
		TxTypeHandler:                         txTypeHandler,
		EconomicsFee:                          genesisFeeHandler,
		ESDTEnableEpoch:                       enableEpochs.ESDTEnableEpoch,
		BuiltInFunctionOnMetachainEnableEpoch: enableEpochs.BuiltInFunctionOnMetaEnableEpoch,
		EpochNotifier:                         epochNotifier,
	}
	txProcessor, err := processTransaction.NewMetaTxProcessor(argsNewMetaTxProcessor)
	if err != nil {
		return nil, process.ErrNilTxProcessor
	}

	disabledRequestHandler := &disabled.RequestHandler{}
	disabledBlockTracker := &disabled.BlockTracker{}
	disabledBlockSizeComputationHandler := &disabled.BlockSizeComputationHandler{}
	disabledBalanceComputationHandler := &disabled.BalanceComputationHandler{}
	disabledScheduledTxsExecutionHandler := &disabled.ScheduledTxsExecutionHandler{}

	preProcFactory, err := metachain.NewPreProcessorsContainerFactory(
		arg.ShardCoordinator,
		arg.Data.StorageService(),
		arg.Core.InternalMarshalizer(),
		arg.Core.Hasher(),
		arg.Data.Datapool(),
		arg.Accounts,
		disabledRequestHandler,
		txProcessor,
		scProcessor,
		arg.Economics,
		gasHandler,
		disabledBlockTracker,
		arg.Core.AddressPubKeyConverter(),
		disabledBlockSizeComputationHandler,
		disabledBalanceComputationHandler,
		epochNotifier,
		enableEpochs.OptimizeGasUsedInCrossMiniBlocksEnableEpoch,
		enableEpochs.FrontRunningProtectionEnableEpoch,
		enableEpochs.ScheduledMiniBlocksEnableEpoch,
		txTypeHandler,
		disabledScheduledTxsExecutionHandler,
	)
	if err != nil {
		return nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, err
	}

	argsTransactionCoordinator := coordinator.ArgTransactionCoordinator{
		Hasher:                            arg.Core.Hasher(),
		Marshalizer:                       arg.Core.InternalMarshalizer(),
		ShardCoordinator:                  arg.ShardCoordinator,
		Accounts:                          arg.Accounts,
		MiniBlockPool:                     arg.Data.Datapool().MiniBlocks(),
		RequestHandler:                    disabledRequestHandler,
		PreProcessors:                     preProcContainer,
		InterProcessors:                   interimProcContainer,
		GasHandler:                        gasHandler,
		FeeHandler:                        genesisFeeHandler,
		BlockSizeComputation:              disabledBlockSizeComputationHandler,
		BalanceComputation:                disabledBalanceComputationHandler,
		EconomicsFee:                      genesisFeeHandler,
		TxTypeHandler:                     txTypeHandler,
		BlockGasAndFeesReCheckEnableEpoch: enableEpochs.BlockGasAndFeesReCheckEnableEpoch,
		TransactionsLogProcessor:          arg.TxLogsProcessor,
		EpochNotifier:                     epochNotifier,
		ScheduledTxsExecutionHandler:      disabledScheduledTxsExecutionHandler,
		ScheduledMiniBlocksEnableEpoch:    enableEpochs.ScheduledMiniBlocksEnableEpoch,
	}
	txCoordinator, err := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	if err != nil {
		return nil, err
	}

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:       vmContainer,
		EconomicsFee:      arg.Economics,
		BlockChainHook:    virtualMachineFactory.BlockChainHookImpl(),
		BlockChain:        arg.Data.Blockchain(),
		ArwenChangeLocker: &sync.RWMutex{},
		Bootstrapper:      syncDisabled.NewDisabledBootstrapper(),
	}
	queryService, err := smartContract.NewSCQueryService(argsNewSCQueryService)
	if err != nil {
		return nil, err
	}

	return &genesisProcessors{
		txCoordinator:  txCoordinator,
		systemSCs:      virtualMachineFactory.SystemSmartContractContainer(),
		blockchainHook: virtualMachineFactory.BlockChainHookImpl(),
		txProcessor:    txProcessor,
		scProcessor:    scProcessor,
		scrProcessor:   scProcessor,
		rwdProcessor:   nil,
		queryService:   queryService,
		vmContainer:    vmContainer,
	}, nil
}

// deploySystemSmartContracts deploys all the system smart contracts to the account state
func deploySystemSmartContracts(
	arg ArgsGenesisBlockCreator,
	txProcessor process.TransactionProcessor,
	systemSCs vm.SystemSCContainer,
) ([]data.TransactionHandler, error) {
	code := hex.EncodeToString([]byte("deploy"))
	vmType := hex.EncodeToString(factory.SystemVirtualMachine)
	codeMetadata := hex.EncodeToString((&vmcommon.CodeMetadata{}).ToBytes())
	deployTxData := strings.Join([]string{code, vmType, codeMetadata}, "@")
	rcvAddress := make([]byte, arg.Core.AddressPubKeyConverter().Len())

	systemSCAddresses := make([][]byte, 0)
	systemSCAddresses = append(systemSCAddresses, systemSCs.Keys()...)

	sort.Slice(systemSCAddresses, func(i, j int) bool {
		return bytes.Compare(systemSCAddresses[i], systemSCAddresses[j]) < 0
	})

	deploySystemSCTxs := make([]data.TransactionHandler, 0)

	for _, address := range systemSCAddresses {
		tx := &transaction.Transaction{
			Nonce:     0,
			Value:     big.NewInt(0),
			RcvAddr:   rcvAddress,
			SndAddr:   address,
			GasPrice:  0,
			GasLimit:  math.MaxUint64,
			Data:      []byte(deployTxData),
			Signature: nil,
		}

		deploySystemSCTxs = append(deploySystemSCTxs, tx)

		_, err := txProcessor.ProcessTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return deploySystemSCTxs, nil
}

// setStakedData sets the initial staked values to the staking smart contract
// it will register both categories of nodes: direct staked and delegated stake. This is done because it is the only
// way possible due to the fact that the delegation contract can not call a sandbox-ed processor suite and accounts state
// at genesis time
func setStakedData(
	arg ArgsGenesisBlockCreator,
	processors *genesisProcessors,
	nodesListSplitter genesis.NodesListSplitter,
) ([]data.TransactionHandler, error) {

	scQueryBlsKeys := &process.SCQuery{
		ScAddress: vm.StakingSCAddress,
		FuncName:  "isStaked",
	}

	stakingTxs := make([]data.TransactionHandler, 0)

	// create staking smart contract state for genesis - update fixed stake value from all
	oneEncoded := hex.EncodeToString(big.NewInt(1).Bytes())
	stakeValue := arg.GenesisNodePrice

	stakedNodes := nodesListSplitter.GetAllNodes()
	for _, nodeInfo := range stakedNodes {
		tx := &transaction.Transaction{
			Nonce:     0,
			Value:     new(big.Int).Set(stakeValue),
			RcvAddr:   vm.ValidatorSCAddress,
			SndAddr:   nodeInfo.AddressBytes(),
			GasPrice:  0,
			GasLimit:  math.MaxUint64,
			Data:      []byte("stake@" + oneEncoded + "@" + hex.EncodeToString(nodeInfo.PubKeyBytes()) + "@" + hex.EncodeToString([]byte("genesis"))),
			Signature: nil,
		}

		stakingTxs = append(stakingTxs, tx)

		_, err := processors.txProcessor.ProcessTransaction(tx)
		if err != nil {
			return nil, err
		}

		scQueryBlsKeys.Arguments = [][]byte{nodeInfo.PubKeyBytes()}
		vmOutput, err := processors.queryService.ExecuteQuery(scQueryBlsKeys)
		if err != nil {
			return nil, err
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			return nil, genesis.ErrBLSKeyNotStaked
		}
	}

	log.Debug("meta block genesis",
		"num nodes staked", len(stakedNodes),
	)

	return stakingTxs, nil
}
