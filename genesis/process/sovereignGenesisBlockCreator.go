package process

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/genesis/process/intermediate"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks/counters"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
)

type sovereignGenesisBlockCreator struct {
	*genesisBlockCreator
}

// NewSovereignGenesisBlockCreator creates a new sovereign genesis block creator instance
func NewSovereignGenesisBlockCreator(gbc *genesisBlockCreator) (*sovereignGenesisBlockCreator, error) {
	if gbc == nil {
		return nil, errNilGenesisBlockCreator
	}

	return &sovereignGenesisBlockCreator{
		genesisBlockCreator: gbc,
	}, nil
}

// CreateGenesisBlocks will create sovereign genesis blocks
func (gbc *sovereignGenesisBlockCreator) CreateGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	if !mustDoGenesisProcess(gbc.arg) {
		return gbc.createSovereignEmptyGenesisBlocks()
	}

	if mustDoHardForkImportProcess(gbc.arg) {
		err := gbc.arg.importHandler.ImportAll()
		if err != nil {
			return nil, err
		}

		err = gbc.computeSovereignDNSAddresses(gbc.arg.EpochConfig.EnableEpochs)
		if err != nil {
			return nil, err
		}
	}

	shardIDs := make([]uint32, 1)
	shardIDs[0] = core.SovereignChainShardId
	argsCreateBlock, err := gbc.baseCreateGenesisBlocks(shardIDs)
	if err != nil {
		return nil, err
	}

	return gbc.createSovereignHeaders(
		argsCreateBlock.mapArgsGenesisBlockCreator,
		argsCreateBlock.mapHardForkBlockProcessor,
		argsCreateBlock.mapBodies,
		argsCreateBlock.shardIDs,
		argsCreateBlock.genesisBlocks,
	)
}

func (gbc *sovereignGenesisBlockCreator) createSovereignEmptyGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	err := gbc.computeSovereignDNSAddresses(createGenesisConfig())
	if err != nil {
		return nil, err
	}

	round, nonce, epoch := getGenesisBlocksRoundNonceEpoch(gbc.arg)

	mapEmptyGenesisBlocks := make(map[uint32]data.HeaderHandler, 1)
	mapEmptyGenesisBlocks[core.SovereignChainShardId] = &block.SovereignChainHeader{
		Header: &block.Header{
			Round:     round,
			Nonce:     nonce,
			Epoch:     epoch,
			TimeStamp: gbc.arg.GenesisTime,
			ShardID:   core.SovereignChainShardId,
		},
	}

	return mapEmptyGenesisBlocks, nil
}

// in case of hardfork initial smart contracts deployment is not called as they are all imported from previous state
func (gbc *sovereignGenesisBlockCreator) computeSovereignDNSAddresses(enableEpochsConfig config.EnableEpochs) error {
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
	}
	blockChainHook, err := hooks.CreateBlockChainHook(gbc.arg.ChainRunType, argsHook)
	if err != nil {
		return err
	}

	initialAddresses, err := factory.DecodeAddresses(gbc.arg.Core.AddressPubKeyConverter(), gbc.arg.MapDNSV2Addresses)
	if err != nil {
		return err
	}
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

func (gbc *genesisBlockCreator) createSovereignHeaders(
	mapArgsGenesisBlockCreator map[uint32]ArgsGenesisBlockCreator,
	mapHardForkBlockProcessor map[uint32]update.HardForkBlockProcessor,
	mapBodies map[uint32]*block.Body,
	shardIDs []uint32,
	genesisBlocks map[uint32]data.HeaderHandler,
) (map[uint32]data.HeaderHandler, error) {
	var nodesListSplitter genesis.NodesListSplitter
	var err error

	nodesListSplitter, err = intermediate.NewNodesListSplitter(gbc.arg.InitialNodesSetup, gbc.arg.AccountsParser)
	if err != nil {
		return nil, err
	}

	allScAddresses := make([][]byte, 0)

	shardID := core.SovereignChainShardId
	log.Debug("genesisBlockCreator.createHeaders", "shard", shardID)
	var genesisBlock data.HeaderHandler
	var scResults [][]byte

	genesisBlock, scResults, gbc.initialIndexingData[shardID], err = createSovereignShardGenesisBlock(
		mapArgsGenesisBlockCreator[shardID],
		mapBodies[shardID],
		nodesListSplitter,
		mapHardForkBlockProcessor[shardID],
	)

	if err != nil {
		return nil, fmt.Errorf("'%w' while generating genesis block for shard %d", err, shardID)
	}

	allScAddresses = append(allScAddresses, scResults...)
	genesisBlocks[shardID] = genesisBlock
	err = gbc.saveGenesisBlock(genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("'%w' while saving genesis block for shard %d", err, shardID)
	}

	err = gbc.checkDelegationsAgainstDeployedSC(allScAddresses, gbc.arg)
	if err != nil {
		return nil, err
	}

	gb := genesisBlocks[shardID]

	log.Info("genesisBlockCreator.createHeaders",
		"shard", gb.GetShardID(),
		"nonce", gb.GetNonce(),
		"round", gb.GetRound(),
		"root hash", gb.GetRootHash(),
	)

	return genesisBlocks, nil
}

func createSovereignShardGenesisBlock(
	arg ArgsGenesisBlockCreator,
	body *block.Body,
	nodesListSplitter genesis.NodesListSplitter,
	hardForkBlockProcessor update.HardForkBlockProcessor,
) (data.HeaderHandler, [][]byte, *genesis.IndexingData, error) {
	if mustDoHardForkImportProcess(arg) {
		return createShardGenesisBlockAfterHardFork(arg, body, hardForkBlockProcessor)
	}

	indexingData := &genesis.IndexingData{
		DelegationTxs:      make([]data.TransactionHandler, 0),
		ScrsTxs:            make(map[string]data.TransactionHandler),
		StakingTxs:         make([]data.TransactionHandler, 0),
		DeploySystemScTxs:  make([]data.TransactionHandler, 0),
		DeployInitialScTxs: make([]data.TransactionHandler, 0),
	}

	processors, err := createProcessorsForShardGenesisBlock(arg, createGenesisConfig(), createGenesisRoundConfig())
	if err != nil {
		return nil, nil, nil, err
	}

	deployMetrics := &deployedScMetrics{}

	scAddresses, scTxs, err := deployInitialSmartContracts(processors, arg, deployMetrics)
	if err != nil {
		return nil, nil, nil, err
	}
	indexingData.DeployInitialScTxs = scTxs

	numSetBalances, err := setBalancesToTrie(arg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while setting the balances to trie",
			err, arg.ShardCoordinator.SelfId())
	}

	numStaked, err := increaseStakersNonces(processors, arg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while incrementing nonces",
			err, arg.ShardCoordinator.SelfId())
	}

	delegationResult, delegationTxs, err := executeDelegation(processors, arg, nodesListSplitter)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while execution delegation",
			err, arg.ShardCoordinator.SelfId())
	}
	indexingData.DelegationTxs = delegationTxs

	numCrossShardDelegations, err := incrementNoncesForCrossShardDelegations(processors, arg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while incrementing crossshard nonce",
			err, arg.ShardCoordinator.SelfId())
	}

	processors2, err := createProcessorsForMetaGenesisBlock(arg, createGenesisConfig(), createGenesisRoundConfig())
	if err != nil {
		return nil, nil, nil, err
	}

	deploySystemSCTxs, err := deploySystemSmartContracts(arg, processors2.txProcessor, processors2.systemSCs)
	if err != nil {
		return nil, nil, nil, err
	}
	indexingData.DeploySystemScTxs = deploySystemSCTxs

	stakingTxs, err := setSovereignStakedData(arg, processors2, nodesListSplitter)
	if err != nil {
		return nil, nil, nil, err
	}
	indexingData.StakingTxs = stakingTxs

	rootHash, err := arg.Accounts.Commit()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while commiting",
			err, arg.ShardCoordinator.SelfId())
	}

	scrsTxs := processors.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
	scrsTxs2 := processors2.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)

	allScrsTxs := make(map[string]data.TransactionHandler)
	for scrTxIn1, scrTx := range scrsTxs {
		allScrsTxs[scrTxIn1] = scrTx
	}

	for scrTxIn1, scrTx := range scrsTxs2 {
		allScrsTxs[scrTxIn1] = scrTx
	}

	indexingData.ScrsTxs = allScrsTxs

	log.Info("STARTING TO CREATE STUFF FOR META")

	log.Info("ENDED TO CREATE STUFF FOR META")

	log.Debug("shard block genesis",
		"shard ID", arg.ShardCoordinator.SelfId(),
		"num delegation SC deployed", deployMetrics.numDelegation,
		"num other SC deployed", deployMetrics.numOtherTypes,
		"num set balances", numSetBalances,
		"num staked directly", numStaked,
		"total staked on a delegation SC", delegationResult.NumTotalStaked,
		"total delegation nodes", delegationResult.NumTotalDelegated,
		"cross shard delegation calls", numCrossShardDelegations,
		"resulted roothash", rootHash,
	)

	round, nonce, epoch := getGenesisBlocksRoundNonceEpoch(arg)
	header := &block.Header{
		Epoch:           epoch,
		Round:           round,
		Nonce:           nonce,
		ShardID:         arg.ShardCoordinator.SelfId(),
		BlockBodyType:   block.StateBlock,
		PubKeysBitmap:   []byte{1},
		Signature:       rootHash,
		RootHash:        rootHash,
		PrevRandSeed:    rootHash,
		RandSeed:        rootHash,
		TimeStamp:       arg.GenesisTime,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		ChainID:         []byte(arg.Core.ChainID()),
		SoftwareVersion: []byte(""),
	}

	err = processors.vmContainer.Close()
	if err != nil {
		return nil, nil, nil, err
	}

	err = processors.vmContainersFactory.Close()
	if err != nil {
		return nil, nil, nil, err
	}

	return header, scAddresses, indexingData, nil
}

func setSovereignStakedData(
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
		senderAcc, err := arg.Accounts.LoadAccount(nodeInfo.AddressBytes())
		if err != nil {
			return nil, err
		}

		tx := &transaction.Transaction{
			Nonce:     senderAcc.GetNonce(),
			Value:     new(big.Int).Set(stakeValue),
			RcvAddr:   vm.ValidatorSCAddress,
			SndAddr:   nodeInfo.AddressBytes(),
			GasPrice:  0,
			GasLimit:  math.MaxUint64,
			Data:      []byte("stake@" + oneEncoded + "@" + hex.EncodeToString(nodeInfo.PubKeyBytes()) + "@" + hex.EncodeToString([]byte("genesis"))),
			Signature: nil,
		}

		stakingTxs = append(stakingTxs, tx)

		_, err = processors.txProcessor.ProcessTransaction(tx)
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
