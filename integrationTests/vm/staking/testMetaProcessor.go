package staking

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/nodetype"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	mock3 "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	factory2 "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	mock2 "github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	economicsHandler "github.com/ElrondNetwork/elrond-go/process/economics"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	metaProcess "github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	"github.com/stretchr/testify/require"
)

const stakingV4InitEpoch = 1
const stakingV4EnableEpoch = 2

type HeaderInfo struct {
	Hash   []byte
	Header data.HeaderHandler
}

// TestMetaProcessor -
type TestMetaProcessor struct {
	MetaBlockProcessor  process.BlockProcessor
	NodesCoordinator    nodesCoordinator.NodesCoordinator
	BlockChain          data.ChainHandler
	ValidatorStatistics process.ValidatorStatisticsProcessor
	EpochStartTrigger   integrationTests.TestEpochStartTrigger
	BlockChainHandler   data.ChainHandler
}

// NewTestMetaProcessor -
func NewTestMetaProcessor(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfNodesPerShard uint32,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	numOfNodesInStakingQueue uint32,
	t *testing.T,
) *TestMetaProcessor {
	coreComponents, dataComponents, bootstrapComponents, statusComponents, stateComponents := createComponentHolders(numOfShards)
	epochStartTrigger := createEpochStartTrigger(coreComponents, dataComponents.StorageService())

	createStakingQueue(numOfNodesInStakingQueue, coreComponents, stateComponents)

	nc := createNodesCoordinator(numOfMetaNodes, numOfShards, numOfNodesPerShard, shardConsensusGroupSize, metaConsensusGroupSize, coreComponents, dataComponents, stateComponents, bootstrapComponents.NodesCoordinatorRegistryFactory())
	scp, blockChainHook, validatorsInfoCreator, metaVMFactory := createSystemSCProcessor(nc, coreComponents, stateComponents, bootstrapComponents, dataComponents)

	return &TestMetaProcessor{
		MetaBlockProcessor:  createMetaBlockProcessor(nc, scp, coreComponents, dataComponents, bootstrapComponents, statusComponents, stateComponents, validatorsInfoCreator, blockChainHook, metaVMFactory, epochStartTrigger),
		NodesCoordinator:    nc,
		BlockChain:          dataComponents.Blockchain(),
		ValidatorStatistics: validatorsInfoCreator,
		EpochStartTrigger:   epochStartTrigger,
		BlockChainHandler:   dataComponents.Blockchain(),
	}
}

func createStakingQueue(
	numOfNodesInStakingQueue uint32,
	coreComponents factory2.CoreComponentsHolder,
	stateComponents factory2.StateComponentsHolder,
) {
	owner := generateUniqueKey(50)
	var ownerWaitingNodes [][]byte
	for i := uint32(51); i < 51+numOfNodesInStakingQueue; i++ {
		ownerWaitingNodes = append(ownerWaitingNodes, generateUniqueKey(i))
	}

	saveOneKeyToWaitingList(stateComponents.AccountsAdapter(),
		ownerWaitingNodes[0],
		coreComponents.InternalMarshalizer(),
		owner,
		owner)

	_, _ = stateComponents.PeerAccounts().Commit()

	addKeysToWaitingList(stateComponents.AccountsAdapter(),
		ownerWaitingNodes[1:],
		coreComponents.InternalMarshalizer(),
		owner, owner)
	addValidatorData(stateComponents.AccountsAdapter(), owner, ownerWaitingNodes, big.NewInt(500000), coreComponents.InternalMarshalizer())

	_, _ = stateComponents.AccountsAdapter().Commit()
}

func createMetaBlockHeader2(epoch uint32, round uint64, prevHash []byte) *block.MetaBlock {
	hdr := block.MetaBlock{
		Epoch:                  epoch,
		Nonce:                  round,
		Round:                  round,
		PrevHash:               prevHash,
		Signature:              []byte("signature"),
		PubKeysBitmap:          []byte("pubKeysBitmap"),
		RootHash:               []byte("roothash"),
		ShardInfo:              make([]block.ShardData, 0),
		TxCount:                1,
		PrevRandSeed:           []byte("roothash"),
		RandSeed:               []byte("roothash" + strconv.Itoa(int(round))),
		AccumulatedFeesInEpoch: big.NewInt(0),
		AccumulatedFees:        big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
	}

	shardMiniBlockHeaders := make([]block.MiniBlockHeader, 0)
	shardMiniBlockHeader := block.MiniBlockHeader{
		Hash:            []byte("mb_hash" + strconv.Itoa(int(round))),
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxCount:         1,
	}
	shardMiniBlockHeaders = append(shardMiniBlockHeaders, shardMiniBlockHeader)
	shardData := block.ShardData{
		Nonce:                 round,
		ShardID:               0,
		HeaderHash:            []byte("hdr_hash" + strconv.Itoa(int(round))),
		TxCount:               1,
		ShardMiniBlockHeaders: shardMiniBlockHeaders,
		DeveloperFees:         big.NewInt(0),
		AccumulatedFees:       big.NewInt(0),
	}
	hdr.ShardInfo = append(hdr.ShardInfo, shardData)

	return &hdr
}

func (tmp *TestMetaProcessor) Process(t *testing.T, fromRound, numOfRounds uint32) {
	for r := fromRound; r < fromRound+numOfRounds; r++ {
		currentHeader := tmp.BlockChainHandler.GetCurrentBlockHeader()
		currentHash := tmp.BlockChainHandler.GetCurrentBlockHeaderHash()
		if currentHeader == nil {
			currentHeader = tmp.BlockChain.GetGenesisHeader()
			currentHash = tmp.BlockChain.GetGenesisHeaderHash()
		}

		prevRandomness := currentHeader.GetRandSeed()
		fmt.Println(fmt.Sprintf("########################################### CREATEING HEADER FOR EPOCH %v in round %v",
			tmp.EpochStartTrigger.Epoch(),
			r,
		))

		newHdr := createMetaBlockHeader2(tmp.EpochStartTrigger.Epoch(), uint64(r), currentHash)
		newHdr.PrevRandSeed = prevRandomness
		createdHdr, _ := tmp.MetaBlockProcessor.CreateNewHeader(uint64(r), uint64(r))
		newHdr.SetEpoch(createdHdr.GetEpoch())

		newHdr2, newBodyHandler2, err := tmp.MetaBlockProcessor.CreateBlock(newHdr, func() bool { return true })
		require.Nil(t, err)
		err = tmp.MetaBlockProcessor.CommitBlock(newHdr2, newBodyHandler2)
		require.Nil(t, err)

		time.Sleep(time.Millisecond * 100)

		tmp.DisplayNodesConfig(tmp.EpochStartTrigger.Epoch(), 4)

		fmt.Println("#######################DISPLAYING VALIDAOTRS AFTEEEEEEEEEEEEEEEEER ")
		rootHash, _ := tmp.ValidatorStatistics.RootHash()
		allValidatorsInfo, err := tmp.ValidatorStatistics.GetValidatorInfoForRootHash(rootHash)
		require.Nil(t, err)
		displayValidatorsInfo(allValidatorsInfo, rootHash)
	}

}

func displayValidatorsInfo(validatorsInfoMap state.ShardValidatorsInfoMapHandler, rootHash []byte) {
	fmt.Println("#######################DISPLAYING VALIDAOTRS INFO for root hash ")
	for _, validators := range validatorsInfoMap.GetAllValidatorsInfo() {
		fmt.Println("PUBKEY: ", string(validators.GetPublicKey()), " SHARDID: ", validators.GetShardId(), " LIST: ", validators.GetList())
	}
}

func createEpochStartTrigger(coreComponents factory2.CoreComponentsHolder, storageService dataRetriever.StorageService) integrationTests.TestEpochStartTrigger {
	argsEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
		GenesisTime: time.Now(),
		Settings: &config.EpochStartConfig{
			MinRoundsBetweenEpochs: 10,
			RoundsPerEpoch:         10,
		},
		Epoch:              0,
		EpochStartNotifier: coreComponents.EpochStartNotifierWithConfirm(),
		Storage:            storageService,
		Marshalizer:        coreComponents.InternalMarshalizer(),
		Hasher:             coreComponents.Hasher(),
		AppStatusHandler:   coreComponents.StatusHandler(),
	}
	epochStartTrigger, _ := metachain.NewEpochStartTrigger(argsEpochStart)
	testTrigger := &metachain.TestTrigger{}
	testTrigger.SetTrigger(epochStartTrigger)
	return testTrigger
}

func (tmp *TestMetaProcessor) DisplayNodesConfig(epoch uint32, numOfShards int) {
	eligible, _ := tmp.NodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	waiting, _ := tmp.NodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	leaving, _ := tmp.NodesCoordinator.GetAllLeavingValidatorsPublicKeys(epoch)
	shuffledOut, _ := tmp.NodesCoordinator.GetAllShuffledOutValidatorsPublicKeys(epoch)

	fmt.Println("############### Displaying nodes config in epoch " + strconv.Itoa(int(epoch)))

	for shard := range eligible {
		for _, pk := range eligible[shard] {
			fmt.Println("eligible", "pk", string(pk), "shardID", shard)
		}
		for _, pk := range waiting[shard] {
			fmt.Println("waiting", "pk", string(pk), "shardID", shard)
		}
		for _, pk := range leaving[shard] {
			fmt.Println("leaving", "pk", string(pk), "shardID", shard)
		}
		for _, pk := range shuffledOut[shard] {
			fmt.Println("shuffled out", "pk", string(pk), "shardID", shard)
		}
	}
}

// shuffler constants
const (
	shuffleBetweenShards = false
	adaptivity           = false
	hysteresis           = float32(0.2)
	initialRating        = 5
)

// TODO: Pass epoch config

func createSystemSCProcessor(
	nc nodesCoordinator.NodesCoordinator,
	coreComponents factory2.CoreComponentsHolder,
	stateComponents factory2.StateComponentsHandler,
	bootstrapComponents factory2.BootstrapComponentsHolder,
	dataComponents factory2.DataComponentsHolder,
) (process.EpochStartSystemSCProcessor, process.BlockChainHookHandler, process.ValidatorStatisticsProcessor, process.VirtualMachinesContainerFactory) {
	args, blockChainHook, validatorsInfOCreator, metaVMFactory := createFullArgumentsForSystemSCProcessing(nc,
		coreComponents,
		stateComponents,
		bootstrapComponents,
		dataComponents,
	)
	s, _ := metachain.NewSystemSCProcessor(args)
	return s, blockChainHook, validatorsInfOCreator, metaVMFactory
}

func generateUniqueKey(identifier uint32) []byte {
	neededLength := 15 //192
	uniqueIdentifier := fmt.Sprintf("address-%d", identifier)
	return []byte(strings.Repeat("0", neededLength-len(uniqueIdentifier)) + uniqueIdentifier)
}

// TODO: MAYBE USE factory from mainFactory.CreateNodesCoordinator
func createNodesCoordinator(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfNodesPerShard uint32,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	coreComponents factory2.CoreComponentsHolder,
	dataComponents factory2.DataComponentsHolder,
	stateComponents factory2.StateComponentsHandler,
	nodesCoordinatorRegistryFactory nodesCoordinator.NodesCoordinatorRegistryFactory,
) nodesCoordinator.NodesCoordinator {
	validatorsMap := generateGenesisNodeInfoMap(numOfMetaNodes, numOfShards, numOfNodesPerShard, 0)
	validatorsMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)

	waitingMap := generateGenesisNodeInfoMap(numOfMetaNodes, numOfShards, numOfNodesPerShard, numOfMetaNodes+numOfShards*numOfNodesPerShard)
	waitingMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(waitingMap)

	// TODO: HERE SAVE ALL ACCOUNTS
	var allPubKeys [][]byte

	for shardID, vals := range validatorsMapForNodesCoordinator {
		for _, val := range vals {
			peerAccount, _ := state.NewPeerAccount(val.PubKey())
			peerAccount.SetTempRating(initialRating)
			peerAccount.ShardId = shardID
			peerAccount.BLSPublicKey = val.PubKey()
			peerAccount.List = string(common.EligibleList)
			stateComponents.PeerAccounts().SaveAccount(peerAccount)
			allPubKeys = append(allPubKeys, val.PubKey())
		}
	}

	for shardID, vals := range waitingMapForNodesCoordinator {
		for _, val := range vals {
			peerAccount, _ := state.NewPeerAccount(val.PubKey())
			peerAccount.SetTempRating(initialRating)
			peerAccount.ShardId = shardID
			peerAccount.BLSPublicKey = val.PubKey()
			peerAccount.List = string(common.WaitingList)
			stateComponents.PeerAccounts().SaveAccount(peerAccount)
			allPubKeys = append(allPubKeys, val.PubKey())
		}
	}

	for idx, pubKey := range allPubKeys {
		registerValidatorKeys(stateComponents.AccountsAdapter(), []byte(string(pubKey)+strconv.Itoa(idx)), []byte(string(pubKey)+strconv.Itoa(idx)), [][]byte{pubKey}, big.NewInt(2000), coreComponents.InternalMarshalizer())
	}

	maxNodesConfig := make([]config.MaxNodesChangeConfig, 0)
	maxNodesConfig = append(maxNodesConfig, config.MaxNodesChangeConfig{MaxNumNodes: 24, NodesToShufflePerShard: 2})

	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:                     numOfNodesPerShard,
		NodesMeta:                      numOfMetaNodes,
		Hysteresis:                     hysteresis,
		Adaptivity:                     adaptivity,
		ShuffleBetweenShards:           shuffleBetweenShards,
		MaxNodesEnableConfig:           maxNodesConfig,
		WaitingListFixEnableEpoch:      0,
		BalanceWaitingListsEnableEpoch: 0,
		StakingV4EnableEpoch:           stakingV4EnableEpoch,
	}
	nodeShuffler, _ := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)

	cache, _ := lrucache.NewCache(10000)
	argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ShardConsensusGroupSize:         shardConsensusGroupSize,
		MetaConsensusGroupSize:          metaConsensusGroupSize,
		Marshalizer:                     coreComponents.InternalMarshalizer(),
		Hasher:                          coreComponents.Hasher(),
		ShardIDAsObserver:               core.MetachainShardId,
		NbShards:                        numOfShards,
		EligibleNodes:                   validatorsMapForNodesCoordinator,
		WaitingNodes:                    waitingMapForNodesCoordinator,
		SelfPublicKey:                   validatorsMap[core.MetachainShardId][0].PubKeyBytes(),
		ConsensusGroupCache:             cache,
		ShuffledOutHandler:              &mock2.ShuffledOutHandlerStub{},
		ChanStopNode:                    coreComponents.ChanStopNodeProcess(),
		IsFullArchive:                   false,
		Shuffler:                        nodeShuffler,
		BootStorer:                      dataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
		EpochStartNotifier:              coreComponents.EpochStartNotifierWithConfirm(),
		StakingV4EnableEpoch:            stakingV4EnableEpoch,
		NodesCoordinatorRegistryFactory: nodesCoordinatorRegistryFactory,
		NodeTypeProvider:                nodetype.NewNodeTypeProvider(core.NodeTypeValidator),
	}

	baseNodesCoordinator, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
	if err != nil {
		fmt.Println("error creating node coordinator")
	}

	nodesCoord, err := nodesCoordinator.NewIndexHashedNodesCoordinatorWithRater(baseNodesCoordinator, coreComponents.Rater())
	if err != nil {
		fmt.Println("error creating node coordinator")
	}

	return nodesCoord
}

func generateGenesisNodeInfoMap(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfNodesPerShard uint32,
	startIdx uint32,
) map[uint32][]nodesCoordinator.GenesisNodeInfoHandler {
	validatorsMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
	id := startIdx
	for shardId := uint32(0); shardId < numOfShards; shardId++ {
		for n := uint32(0); n < numOfNodesPerShard; n++ {
			addr := generateUniqueKey(id)
			validator := mock2.NewNodeInfo(addr, addr, shardId, initialRating)
			validatorsMap[shardId] = append(validatorsMap[shardId], validator)
			id++
		}
	}

	for n := uint32(0); n < numOfMetaNodes; n++ {
		addr := generateUniqueKey(id)
		validator := mock2.NewNodeInfo(addr, addr, core.MetachainShardId, initialRating)
		validatorsMap[core.MetachainShardId] = append(validatorsMap[core.MetachainShardId], validator)
		id++
	}

	return validatorsMap
}

func createGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for ShardID := uint32(0); ShardID < shardCoordinator.NumberOfShards(); ShardID++ {
		genesisBlocks[ShardID] = createGenesisBlock(ShardID)
	}

	genesisBlocks[core.MetachainShardId] = createGenesisMetaBlock()

	return genesisBlocks
}

func createGenesisBlock(ShardID uint32) *block.Header {
	rootHash := []byte("roothash")
	return &block.Header{
		Nonce:           0,
		Round:           0,
		Signature:       rootHash,
		RandSeed:        rootHash,
		PrevRandSeed:    rootHash,
		ShardID:         ShardID,
		PubKeysBitmap:   rootHash,
		RootHash:        rootHash,
		PrevHash:        rootHash,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
}

func createGenesisMetaBlock() *block.MetaBlock {
	rootHash := []byte("roothash")
	return &block.MetaBlock{
		Nonce:                  0,
		Round:                  0,
		Signature:              rootHash,
		RandSeed:               rootHash,
		PrevRandSeed:           rootHash,
		PubKeysBitmap:          rootHash,
		RootHash:               rootHash,
		PrevHash:               rootHash,
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
	}
}

func createFullArgumentsForSystemSCProcessing(
	nc nodesCoordinator.NodesCoordinator,
	coreComponents factory2.CoreComponentsHolder,
	stateComponents factory2.StateComponentsHandler,
	bootstrapComponents factory2.BootstrapComponentsHolder,
	dataComponents factory2.DataComponentsHolder,
) (metachain.ArgsNewEpochStartSystemSCProcessing, process.BlockChainHookHandler, process.ValidatorStatisticsProcessor, process.VirtualMachinesContainerFactory) {
	nodesSetup := &mock.NodesSetupStub{}
	argsValidatorsProcessor := peer.ArgValidatorStatisticsProcessor{
		Marshalizer:                          coreComponents.InternalMarshalizer(),
		NodesCoordinator:                     nc,
		ShardCoordinator:                     bootstrapComponents.ShardCoordinator(),
		DataPool:                             dataComponents.Datapool(),
		StorageService:                       dataComponents.StorageService(),
		PubkeyConv:                           coreComponents.AddressPubKeyConverter(),
		PeerAdapter:                          stateComponents.PeerAccounts(),
		Rater:                                coreComponents.Rater(),
		RewardsHandler:                       &mock3.RewardsHandlerStub{},
		NodesSetup:                           nodesSetup,
		MaxComputableRounds:                  1,
		MaxConsecutiveRoundsOfRatingDecrease: 2000,
		EpochNotifier:                        coreComponents.EpochNotifier(),
		StakingV2EnableEpoch:                 0,
		StakingV4EnableEpoch:                 stakingV4EnableEpoch,
	}
	validatorStatisticsProcessor, _ := peer.NewValidatorStatisticsProcessor(argsValidatorsProcessor)

	gasSchedule := arwenConfig.MakeGasMapForTests()
	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasSchedule)
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:      gasScheduleNotifier,
		MapDNSAddresses:  make(map[string]struct{}),
		Marshalizer:      coreComponents.InternalMarshalizer(),
		Accounts:         stateComponents.AccountsAdapter(),
		ShardCoordinator: bootstrapComponents.ShardCoordinator(),
		EpochNotifier:    coreComponents.EpochNotifier(),
	}
	builtInFuncs, _, _ := builtInFunctions.CreateBuiltInFuncContainerAndNFTStorageHandler(argsBuiltIn)

	argsHook := hooks.ArgBlockChainHook{
		Accounts:           stateComponents.AccountsAdapter(),
		PubkeyConv:         coreComponents.AddressPubKeyConverter(),
		StorageService:     dataComponents.StorageService(),
		BlockChain:         dataComponents.Blockchain(),
		ShardCoordinator:   bootstrapComponents.ShardCoordinator(),
		Marshalizer:        coreComponents.InternalMarshalizer(),
		Uint64Converter:    coreComponents.Uint64ByteSliceConverter(),
		NFTStorageHandler:  &testscommon.SimpleNFTStorageHandlerStub{},
		BuiltInFunctions:   builtInFuncs,
		DataPool:           dataComponents.Datapool(),
		CompiledSCPool:     dataComponents.Datapool().SmartContracts(),
		EpochNotifier:      coreComponents.EpochNotifier(),
		NilCompiledSCStore: true,
	}

	defaults.FillGasMapInternal(gasSchedule, 1)
	signVerifer, _ := disabled.NewMessageSignVerifier(&cryptoMocks.KeyGenStub{})
	blockChainHookImpl, _ := hooks.NewBlockChainHookImpl(argsHook)
	argsNewVMContainerFactory := metaProcess.ArgsNewVMContainerFactory{
		BlockChainHook:      blockChainHookImpl,
		PubkeyConv:          argsHook.PubkeyConv,
		Economics:           coreComponents.EconomicsData(),
		MessageSignVerifier: signVerifer,
		GasSchedule:         gasScheduleNotifier,
		NodesConfigProvider: nodesSetup,
		Hasher:              coreComponents.Hasher(),
		Marshalizer:         coreComponents.InternalMarshalizer(),
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost:  "1000",
				OwnerAddress:     "aaaaaa",
				DelegationTicker: "DEL",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        "50",
					MinPassThreshold: "50",
					MinVetoThreshold: "50",
				},
				FirstWhitelistedAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             24, // TODO HERE ADD MAX NUM NODES
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
				StakeLimitPercentage:                 100.0,
				NodeLimitPercentage:                  100.0,
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "100",
				MinStakeAmount:      "100",
				ConfigChangeAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 100,
			},
		},
		ValidatorAccountsDB: stateComponents.PeerAccounts(),
		ChanceComputer:      &mock3.ChanceComputerStub{},
		EpochNotifier:       coreComponents.EpochNotifier(),
		EpochConfig: &config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				StakingV2EnableEpoch:               0,
				StakeEnableEpoch:                   0,
				DelegationManagerEnableEpoch:       0,
				DelegationSmartContractEnableEpoch: 0,
				StakeLimitsEnableEpoch:             10,
				StakingV4InitEnableEpoch:           stakingV4InitEpoch,
				StakingV4EnableEpoch:               stakingV4EnableEpoch,
			},
		},
		ShardCoordinator: bootstrapComponents.ShardCoordinator(),
		NodesCoordinator: nc,
	}

	metaVmFactory, _ := metaProcess.NewVMContainerFactory(argsNewVMContainerFactory)
	vmContainer, _ := metaVmFactory.Create()
	systemVM, _ := vmContainer.Get(vmFactory.SystemVirtualMachine)
	stakingSCprovider, _ := metachain.NewStakingDataProvider(systemVM, "1000")

	maxNodesConfig := make([]config.MaxNodesChangeConfig, 0)
	for i := 0; i < 444; i++ {
		maxNodesConfig = append(maxNodesConfig, config.MaxNodesChangeConfig{MaxNumNodes: 24, NodesToShufflePerShard: 2})
	}

	args := metachain.ArgsNewEpochStartSystemSCProcessing{
		SystemVM:                systemVM,
		UserAccountsDB:          stateComponents.AccountsAdapter(),
		PeerAccountsDB:          stateComponents.PeerAccounts(),
		Marshalizer:             coreComponents.InternalMarshalizer(),
		StartRating:             initialRating,
		ValidatorInfoCreator:    validatorStatisticsProcessor,
		EndOfEpochCallerAddress: vm.EndOfEpochAddress,
		StakingSCAddress:        vm.StakingSCAddress,
		ChanceComputer:          &mock3.ChanceComputerStub{},
		EpochNotifier:           coreComponents.EpochNotifier(),
		GenesisNodesConfig:      nodesSetup,
		StakingDataProvider:     stakingSCprovider,
		NodesConfigProvider:     nc,
		ShardCoordinator:        bootstrapComponents.ShardCoordinator(),
		ESDTOwnerAddressBytes:   bytes.Repeat([]byte{1}, 32),
		EpochConfig: config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				StakingV2EnableEpoch:      0,
				ESDTEnableEpoch:           1000000,
				StakingV4InitEnableEpoch:  stakingV4InitEpoch,
				StakingV4EnableEpoch:      stakingV4EnableEpoch,
				MaxNodesChangeEnableEpoch: maxNodesConfig,
			},
		},
		MaxNodesEnableConfig: maxNodesConfig,
	}

	return args, blockChainHookImpl, validatorStatisticsProcessor, metaVmFactory
}

func createAccountsDB(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountFactory state.AccountFactory,
	trieStorageManager common.StorageManager,
) *state.AccountsDB {
	tr, _ := trie.NewTrie(trieStorageManager, marshalizer, hasher, 5)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(10, testscommon.NewMemDbMock(), marshalizer)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)
	adb, _ := state.NewAccountsDB(tr, hasher, marshalizer, accountFactory, spm, common.Normal)
	return adb
}

func createEconomicsData() process.EconomicsDataHandler {
	maxGasLimitPerBlock := strconv.FormatUint(1500000000, 10)
	minGasPrice := strconv.FormatUint(10, 10)
	minGasLimit := strconv.FormatUint(10, 10)

	argsNewEconomicsData := economicsHandler.ArgsNewEconomicsData{
		Economics: &config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				GenesisTotalSupply: "2000000000000000000000",
				MinimumInflation:   0,
				YearSettings: []*config.YearSetting{
					{
						Year:             0,
						MaximumInflation: 0.01,
					},
				},
			},
			RewardsSettings: config.RewardsSettings{
				RewardsConfigByEpoch: []config.EpochRewardSettings{
					{
						LeaderPercentage:                 0.1,
						DeveloperPercentage:              0.1,
						ProtocolSustainabilityPercentage: 0.1,
						ProtocolSustainabilityAddress:    "protocol",
						TopUpGradientPoint:               "300000000000000000000",
						TopUpFactor:                      0.25,
					},
				},
			},
			FeeSettings: config.FeeSettings{
				GasLimitSettings: []config.GasLimitSetting{
					{
						MaxGasLimitPerBlock:         maxGasLimitPerBlock,
						MaxGasLimitPerMiniBlock:     maxGasLimitPerBlock,
						MaxGasLimitPerMetaBlock:     maxGasLimitPerBlock,
						MaxGasLimitPerMetaMiniBlock: maxGasLimitPerBlock,
						MaxGasLimitPerTx:            maxGasLimitPerBlock,
						MinGasLimit:                 minGasLimit,
					},
				},
				MinGasPrice:      minGasPrice,
				GasPerDataByte:   "1",
				GasPriceModifier: 1.0,
			},
		},
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  &epochNotifier.EpochNotifierStub{},
		BuiltInFunctionsCostHandler:    &mock.BuiltInCostHandlerStub{},
	}
	economicsData, _ := economicsHandler.NewEconomicsData(argsNewEconomicsData)
	return economicsData
}

// ######

func registerValidatorKeys(
	accountsDB state.AccountsAdapter,
	ownerAddress []byte,
	rewardAddress []byte,
	stakedKeys [][]byte,
	totalStake *big.Int,
	marshaller marshal.Marshalizer,
) {
	addValidatorData(accountsDB, ownerAddress, stakedKeys, totalStake, marshaller)
	addStakingData(accountsDB, ownerAddress, rewardAddress, stakedKeys, marshaller)
	_, err := accountsDB.Commit()
	if err != nil {
		fmt.Println("ERROR REGISTERING VALIDATORS ", err)
	}
	//log.LogIfError(err)
}

func addValidatorData(
	accountsDB state.AccountsAdapter,
	ownerKey []byte,
	registeredKeys [][]byte,
	totalStake *big.Int,
	marshaller marshal.Marshalizer,
) {
	validatorSC := loadSCAccount(accountsDB, vm.ValidatorSCAddress)
	validatorData := &systemSmartContracts.ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   ownerKey,
		TotalStakeValue: totalStake,
		LockedStake:     big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		BlsPubKeys:      registeredKeys,
		NumRegistered:   uint32(len(registeredKeys)),
	}

	marshaledData, _ := marshaller.Marshal(validatorData)
	_ = validatorSC.DataTrieTracker().SaveKeyValue(ownerKey, marshaledData)

	_ = accountsDB.SaveAccount(validatorSC)
}

func addStakingData(
	accountsDB state.AccountsAdapter,
	ownerAddress []byte,
	rewardAddress []byte,
	stakedKeys [][]byte,
	marshaller marshal.Marshalizer,
) {
	stakedData := &systemSmartContracts.StakedDataV2_0{
		Staked:        true,
		RewardAddress: rewardAddress,
		OwnerAddress:  ownerAddress,
		StakeValue:    big.NewInt(100),
	}
	marshaledData, _ := marshaller.Marshal(stakedData)

	stakingSCAcc := loadSCAccount(accountsDB, vm.StakingSCAddress)
	for _, key := range stakedKeys {
		_ = stakingSCAcc.DataTrieTracker().SaveKeyValue(key, marshaledData)
	}

	_ = accountsDB.SaveAccount(stakingSCAcc)
}

func loadSCAccount(accountsDB state.AccountsAdapter, address []byte) state.UserAccountHandler {
	acc, _ := accountsDB.LoadAccount(address)
	stakingSCAcc := acc.(state.UserAccountHandler)

	return stakingSCAcc
}

func prepareStakingContractWithData(
	accountsDB state.AccountsAdapter,
	stakedKey []byte,
	waitingKey []byte,
	marshalizer marshal.Marshalizer,
	rewardAddress []byte,
	ownerAddress []byte,
) {
	addStakingData(accountsDB, ownerAddress, rewardAddress, [][]byte{stakedKey}, marshalizer)
	saveOneKeyToWaitingList(accountsDB, waitingKey, marshalizer, rewardAddress, ownerAddress)
	addValidatorData(accountsDB, rewardAddress, [][]byte{stakedKey, waitingKey}, big.NewInt(10000000000), marshalizer)

	_, _ = accountsDB.Commit()

}

func saveOneKeyToWaitingList(
	accountsDB state.AccountsAdapter,
	waitingKey []byte,
	marshalizer marshal.Marshalizer,
	rewardAddress []byte,
	ownerAddress []byte,
) {
	stakingSCAcc := loadSCAccount(accountsDB, vm.StakingSCAddress)
	stakedData := &systemSmartContracts.StakedDataV2_0{
		Waiting:       true,
		RewardAddress: rewardAddress,
		OwnerAddress:  ownerAddress,
		StakeValue:    big.NewInt(100),
	}
	marshaledData, _ := marshalizer.Marshal(stakedData)
	_ = stakingSCAcc.DataTrieTracker().SaveKeyValue(waitingKey, marshaledData)

	waitingKeyInList := []byte("w_" + string(waitingKey))
	waitingListHead := &systemSmartContracts.WaitingList{
		FirstKey: waitingKeyInList,
		LastKey:  waitingKeyInList,
		Length:   1,
	}
	marshaledData, _ = marshalizer.Marshal(waitingListHead)
	_ = stakingSCAcc.DataTrieTracker().SaveKeyValue([]byte("waitingList"), marshaledData)

	waitingListElement := &systemSmartContracts.ElementInList{
		BLSPublicKey: waitingKey,
		PreviousKey:  waitingKeyInList,
		NextKey:      make([]byte, 0),
	}
	marshaledData, _ = marshalizer.Marshal(waitingListElement)
	_ = stakingSCAcc.DataTrieTracker().SaveKeyValue(waitingKeyInList, marshaledData)

	_ = accountsDB.SaveAccount(stakingSCAcc)
}

func addKeysToWaitingList(
	accountsDB state.AccountsAdapter,
	waitingKeys [][]byte,
	marshalizer marshal.Marshalizer,
	rewardAddress []byte,
	ownerAddress []byte,
) {
	stakingSCAcc := loadSCAccount(accountsDB, vm.StakingSCAddress)

	for _, waitingKey := range waitingKeys {
		stakedData := &systemSmartContracts.StakedDataV2_0{
			Waiting:       true,
			RewardAddress: rewardAddress,
			OwnerAddress:  ownerAddress,
			StakeValue:    big.NewInt(100),
		}
		marshaledData, _ := marshalizer.Marshal(stakedData)
		_ = stakingSCAcc.DataTrieTracker().SaveKeyValue(waitingKey, marshaledData)
	}

	marshaledData, _ := stakingSCAcc.DataTrieTracker().RetrieveValue([]byte("waitingList"))
	waitingListHead := &systemSmartContracts.WaitingList{}
	_ = marshalizer.Unmarshal(waitingListHead, marshaledData)

	waitingListAlreadyHasElements := waitingListHead.Length > 0
	waitingListLastKeyBeforeAddingNewKeys := waitingListHead.LastKey

	waitingListHead.Length += uint32(len(waitingKeys))
	lastKeyInList := []byte("w_" + string(waitingKeys[len(waitingKeys)-1]))
	waitingListHead.LastKey = lastKeyInList

	marshaledData, _ = marshalizer.Marshal(waitingListHead)
	_ = stakingSCAcc.DataTrieTracker().SaveKeyValue([]byte("waitingList"), marshaledData)

	numWaitingKeys := len(waitingKeys)
	previousKey := waitingListHead.LastKey
	for i, waitingKey := range waitingKeys {

		waitingKeyInList := []byte("w_" + string(waitingKey))
		waitingListElement := &systemSmartContracts.ElementInList{
			BLSPublicKey: waitingKey,
			PreviousKey:  previousKey,
			NextKey:      make([]byte, 0),
		}

		if i < numWaitingKeys-1 {
			nextKey := []byte("w_" + string(waitingKeys[i+1]))
			waitingListElement.NextKey = nextKey
		}

		marshaledData, _ = marshalizer.Marshal(waitingListElement)
		_ = stakingSCAcc.DataTrieTracker().SaveKeyValue(waitingKeyInList, marshaledData)

		previousKey = waitingKeyInList
	}

	if waitingListAlreadyHasElements {
		marshaledData, _ = stakingSCAcc.DataTrieTracker().RetrieveValue(waitingListLastKeyBeforeAddingNewKeys)
	} else {
		marshaledData, _ = stakingSCAcc.DataTrieTracker().RetrieveValue(waitingListHead.FirstKey)
	}

	waitingListElement := &systemSmartContracts.ElementInList{}
	_ = marshalizer.Unmarshal(waitingListElement, marshaledData)
	waitingListElement.NextKey = []byte("w_" + string(waitingKeys[0]))
	marshaledData, _ = marshalizer.Marshal(waitingListElement)

	if waitingListAlreadyHasElements {
		_ = stakingSCAcc.DataTrieTracker().SaveKeyValue(waitingListLastKeyBeforeAddingNewKeys, marshaledData)
	} else {
		_ = stakingSCAcc.DataTrieTracker().SaveKeyValue(waitingListHead.FirstKey, marshaledData)
	}

	_ = accountsDB.SaveAccount(stakingSCAcc)
}
