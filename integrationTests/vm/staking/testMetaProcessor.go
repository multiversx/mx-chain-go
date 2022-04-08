package staking

// nomindated proof of stake - polkadot
import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	factory2 "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
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
	ValidatorStatistics process.ValidatorStatisticsProcessor
	EpochStartTrigger   integrationTests.TestEpochStartTrigger
	BlockChainHandler   data.ChainHandler
}

// NewTestMetaProcessor -
func NewTestMetaProcessor(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfEligibleNodesPerShard uint32,
	numOfWaitingNodesPerShard uint32,
	numOfNodesToShufflePerShard uint32,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	numOfNodesInStakingQueue uint32,
) *TestMetaProcessor {
	coreComponents, dataComponents, bootstrapComponents, statusComponents, stateComponents := createComponentHolders(numOfShards)

	maxNodesConfig := createMaxNodesConfig(
		numOfMetaNodes,
		numOfShards,
		numOfEligibleNodesPerShard,
		numOfWaitingNodesPerShard,
		numOfNodesToShufflePerShard,
	)

	createStakingQueue(numOfNodesInStakingQueue, coreComponents.InternalMarshalizer(), stateComponents.AccountsAdapter())

	nc := createNodesCoordinator(
		numOfMetaNodes,
		numOfShards,
		numOfEligibleNodesPerShard,
		numOfWaitingNodesPerShard,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		coreComponents,
		dataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
		stateComponents,
		bootstrapComponents.NodesCoordinatorRegistryFactory(),
		maxNodesConfig,
	)

	gasScheduleNotifier := createGasScheduleNotifier()
	blockChainHook := createBlockChainHook(
		dataComponents, coreComponents,
		stateComponents.AccountsAdapter(),
		bootstrapComponents.ShardCoordinator(),
		gasScheduleNotifier,
	)

	metaVmFactory := createVMContainerFactory(
		coreComponents,
		gasScheduleNotifier,
		blockChainHook,
		stateComponents.PeerAccounts(),
		bootstrapComponents.ShardCoordinator(),
		nc,
	)
	vmContainer, _ := metaVmFactory.Create()

	validatorStatisticsProcessor := createValidatorStatisticsProcessor(
		dataComponents,
		coreComponents,
		nc,
		bootstrapComponents.ShardCoordinator(),
		stateComponents.PeerAccounts(),
	)
	scp := createSystemSCProcessor(
		nc,
		coreComponents,
		stateComponents,
		bootstrapComponents.ShardCoordinator(),
		maxNodesConfig,
		validatorStatisticsProcessor,
		vmContainer,
	)

	epochStartTrigger := createEpochStartTrigger(coreComponents, dataComponents.StorageService())

	return &TestMetaProcessor{
		MetaBlockProcessor: createMetaBlockProcessor(
			nc,
			scp,
			coreComponents,
			dataComponents,
			bootstrapComponents,
			statusComponents,
			stateComponents,
			validatorStatisticsProcessor,
			blockChainHook,
			metaVmFactory,
			epochStartTrigger,
		),
		NodesCoordinator:    nc,
		ValidatorStatistics: validatorStatisticsProcessor,
		EpochStartTrigger:   epochStartTrigger,
		BlockChainHandler:   dataComponents.Blockchain(),
	}
}

func createMaxNodesConfig(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfEligibleNodesPerShard uint32,
	numOfWaitingNodesPerShard uint32,
	numOfNodesToShufflePerShard uint32,
) []config.MaxNodesChangeConfig {
	totalEligible := numOfMetaNodes + numOfShards*numOfEligibleNodesPerShard
	totalWaiting := (numOfShards + 1) * numOfWaitingNodesPerShard

	maxNodesConfig := make([]config.MaxNodesChangeConfig, 0)
	maxNodesConfig = append(maxNodesConfig, config.MaxNodesChangeConfig{
		MaxNumNodes:            totalEligible + totalWaiting,
		NodesToShufflePerShard: numOfNodesToShufflePerShard,
	},
	)

	return maxNodesConfig
}

func createGasScheduleNotifier() core.GasScheduleNotifier {
	gasSchedule := arwenConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasSchedule, 1)
	return mock.NewGasScheduleNotifierMock(gasSchedule)
}

func createStakingQueue(
	numOfNodesInStakingQueue uint32,
	marshaller marshal.Marshalizer,
	accountsAdapter state.AccountsAdapter,
) {
	owner := generateUniqueKey(50)
	ownerWaitingNodes := make([][]byte, 0)
	for i := uint32(51); i < 51+numOfNodesInStakingQueue; i++ {
		ownerWaitingNodes = append(ownerWaitingNodes, generateUniqueKey(i))
	}

	testscommon.SaveOneKeyToWaitingList(
		accountsAdapter,
		ownerWaitingNodes[0],
		marshaller,
		owner,
		owner,
	)
	testscommon.AddKeysToWaitingList(
		accountsAdapter,
		ownerWaitingNodes[1:],
		marshaller,
		owner,
		owner,
	)
	testscommon.AddValidatorData(
		accountsAdapter,
		owner,
		ownerWaitingNodes,
		big.NewInt(50000),
		marshaller,
	)
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
			currentHeader = tmp.BlockChainHandler.GetGenesisHeader()
			currentHash = tmp.BlockChainHandler.GetGenesisHeaderHash()
		}

		prevRandomness := currentHeader.GetRandSeed()
		fmt.Println(fmt.Sprintf("########################################### CREATEING HEADER FOR EPOCH %v in round %v",
			tmp.EpochStartTrigger.Epoch(),
			r,
		))

		newHdr := createMetaBlockHeader2(tmp.EpochStartTrigger.Epoch(), uint64(r), currentHash)
		newHdr.PrevRandSeed = prevRandomness
		createdHdr, _ := tmp.MetaBlockProcessor.CreateNewHeader(uint64(r), uint64(r))
		_ = newHdr.SetEpoch(createdHdr.GetEpoch())

		newHdr2, newBodyHandler2, err := tmp.MetaBlockProcessor.CreateBlock(newHdr, func() bool { return true })
		require.Nil(t, err)
		err = tmp.MetaBlockProcessor.CommitBlock(newHdr2, newBodyHandler2)
		require.Nil(t, err)

		time.Sleep(time.Millisecond * 100)

		tmp.DisplayNodesConfig(tmp.EpochStartTrigger.Epoch())

		rootHash, _ := tmp.ValidatorStatistics.RootHash()
		allValidatorsInfo, err := tmp.ValidatorStatistics.GetValidatorInfoForRootHash(rootHash)
		require.Nil(t, err)
		displayValidatorsInfo(allValidatorsInfo)
	}

}

func displayValidatorsInfo(validatorsInfoMap state.ShardValidatorsInfoMapHandler) {
	fmt.Println("#######################DISPLAYING VALIDATORS INFO")
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

func (tmp *TestMetaProcessor) DisplayNodesConfig(epoch uint32) {
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

func generateUniqueKey(identifier uint32) []byte {
	neededLength := 15 //192
	uniqueIdentifier := fmt.Sprintf("address-%d", identifier)
	return []byte(strings.Repeat("0", neededLength-len(uniqueIdentifier)) + uniqueIdentifier)
}
