package staking

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
	"github.com/ElrondNetwork/elrond-go/common"
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
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	"github.com/stretchr/testify/require"
)

const (
	stakingV4InitEpoch   = 1
	stakingV4EnableEpoch = 2
	addressLength        = 15
	nodePrice            = 1000
)

type NodesConfig struct {
	eligible    map[uint32][][]byte
	waiting     map[uint32][][]byte
	leaving     map[uint32][][]byte
	shuffledOut map[uint32][][]byte
	queue       [][]byte
	auction     [][]byte
}

// TestMetaProcessor -
type TestMetaProcessor struct {
	MetaBlockProcessor  process.BlockProcessor
	NodesCoordinator    nodesCoordinator.NodesCoordinator
	ValidatorStatistics process.ValidatorStatisticsProcessor
	EpochStartTrigger   integrationTests.TestEpochStartTrigger
	BlockChainHandler   data.ChainHandler
	NodesConfig         NodesConfig
	CurrentRound        uint64
	AccountsAdapter     state.AccountsAdapter
	Marshaller          marshal.Marshalizer

	metaConsensusGroupSize uint32
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

	queue := createStakingQueue(
		numOfNodesInStakingQueue,
		maxNodesConfig[0].MaxNumNodes,
		coreComponents.InternalMarshalizer(),
		stateComponents.AccountsAdapter(),
	)

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

	eligible, _ := nc.GetAllEligibleValidatorsPublicKeys(0)
	waiting, _ := nc.GetAllWaitingValidatorsPublicKeys(0)
	shuffledOut, _ := nc.GetAllShuffledOutValidatorsPublicKeys(0)

	return &TestMetaProcessor{
		AccountsAdapter: stateComponents.AccountsAdapter(),
		Marshaller:      coreComponents.InternalMarshalizer(),
		NodesConfig: NodesConfig{
			eligible:    eligible,
			waiting:     waiting,
			shuffledOut: shuffledOut,
			queue:       queue,
			auction:     make([][]byte, 0),
		},
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
			vmContainer,
		),
		CurrentRound:           1,
		NodesCoordinator:       nc,
		metaConsensusGroupSize: uint32(metaConsensusGroupSize),
		ValidatorStatistics:    validatorStatisticsProcessor,
		EpochStartTrigger:      epochStartTrigger,
		BlockChainHandler:      dataComponents.Blockchain(),
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
	totalNumOfNodes uint32,
	marshaller marshal.Marshalizer,
	accountsAdapter state.AccountsAdapter,
) [][]byte {
	owner := generateAddress(totalNumOfNodes)
	totalNumOfNodes += 1
	ownerWaitingNodes := make([][]byte, 0)
	for i := totalNumOfNodes; i < totalNumOfNodes+numOfNodesInStakingQueue; i++ {
		ownerWaitingNodes = append(ownerWaitingNodes, generateAddress(i))
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
		big.NewInt(int64(2*nodePrice*numOfNodesInStakingQueue)),
		marshaller,
	)

	return ownerWaitingNodes
}

func createEpochStartTrigger(
	coreComponents factory2.CoreComponentsHolder,
	storageService dataRetriever.StorageService,
) integrationTests.TestEpochStartTrigger {
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

func (tmp *TestMetaProcessor) Process(t *testing.T, numOfRounds uint64) {
	for r := tmp.CurrentRound; r < tmp.CurrentRound+numOfRounds; r++ {
		currentHeader, currentHash := tmp.getCurrentHeaderInfo()

		_, err := tmp.MetaBlockProcessor.CreateNewHeader(r, r)
		require.Nil(t, err)

		fmt.Println(fmt.Sprintf("############## CREATING HEADER FOR EPOCH %v in round %v ##############",
			tmp.EpochStartTrigger.Epoch(),
			r,
		))

		header := createMetaBlockToCommit(
			tmp.EpochStartTrigger.Epoch(),
			r,
			currentHash,
			currentHeader.GetRandSeed(),
			tmp.metaConsensusGroupSize/8+1,
		)
		newHeader, blockBody, err := tmp.MetaBlockProcessor.CreateBlock(header, func() bool { return true })
		require.Nil(t, err)

		err = tmp.MetaBlockProcessor.CommitBlock(newHeader, blockBody)
		require.Nil(t, err)

		time.Sleep(time.Millisecond * 40)

		tmp.updateNodesConfig(tmp.EpochStartTrigger.Epoch())
	}

	tmp.CurrentRound += numOfRounds
}

func displayValidators(list string, pubKeys [][]byte, shardID uint32) {
	pubKeysToDisplay := pubKeys
	if len(pubKeys) > 6 {
		pubKeysToDisplay = make([][]byte, 0)
		pubKeysToDisplay = append(pubKeysToDisplay, pubKeys[:3]...)
		pubKeysToDisplay = append(pubKeysToDisplay, [][]byte{[]byte("...")}...)
		pubKeysToDisplay = append(pubKeysToDisplay, pubKeys[len(pubKeys)-3:]...)
	}

	for _, pk := range pubKeysToDisplay {
		fmt.Println(list, "pk", string(pk), "shardID", shardID)
	}
}

func (tmp *TestMetaProcessor) updateNodesConfig(epoch uint32) {
	eligible, _ := tmp.NodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	waiting, _ := tmp.NodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	leaving, _ := tmp.NodesCoordinator.GetAllLeavingValidatorsPublicKeys(epoch)
	shuffledOut, _ := tmp.NodesCoordinator.GetAllShuffledOutValidatorsPublicKeys(epoch)

	for shard := range eligible {
		displayValidators("eligible", eligible[shard], shard)
		displayValidators("waiting", waiting[shard], shard)
		displayValidators("leaving", leaving[shard], shard)
		displayValidators("shuffled", shuffledOut[shard], shard)
	}

	rootHash, _ := tmp.ValidatorStatistics.RootHash()
	validatorsInfoMap, _ := tmp.ValidatorStatistics.GetValidatorInfoForRootHash(rootHash)

	auction := make([][]byte, 0)
	fmt.Println("####### Auction list")
	for _, validator := range validatorsInfoMap.GetAllValidatorsInfo() {
		if validator.GetList() == string(common.AuctionList) {
			auction = append(auction, validator.GetPublicKey())
		}
	}
	displayValidators("auction", auction, 0)
	queue := tmp.searchPreviousFromHead()
	fmt.Println("##### STAKING QUEUE")
	displayValidators("queue", queue, 0)

	tmp.NodesConfig.eligible = eligible
	tmp.NodesConfig.waiting = waiting
	tmp.NodesConfig.shuffledOut = shuffledOut
	tmp.NodesConfig.leaving = leaving
	tmp.NodesConfig.auction = auction
	tmp.NodesConfig.queue = queue
}

func loadSCAccount(accountsDB state.AccountsAdapter, address []byte) state.UserAccountHandler {
	acc, _ := accountsDB.LoadAccount(address)
	stakingSCAcc := acc.(state.UserAccountHandler)

	return stakingSCAcc
}

func (tmp *TestMetaProcessor) searchPreviousFromHead() [][]byte {
	stakingSCAcc := loadSCAccount(tmp.AccountsAdapter, vm.StakingSCAddress)

	waitingList := &systemSmartContracts.WaitingList{
		FirstKey:      make([]byte, 0),
		LastKey:       make([]byte, 0),
		Length:        0,
		LastJailedKey: make([]byte, 0),
	}
	marshaledData, _ := stakingSCAcc.DataTrieTracker().RetrieveValue([]byte("waitingList"))
	if len(marshaledData) == 0 {
		return nil
	}

	err := tmp.Marshaller.Unmarshal(waitingList, marshaledData)
	if err != nil {
		return nil
	}

	index := uint32(1)
	nextKey := make([]byte, len(waitingList.FirstKey))
	copy(nextKey, waitingList.FirstKey)

	allPubKeys := make([][]byte, 0)
	for len(nextKey) != 0 && index <= waitingList.Length {
		allPubKeys = append(allPubKeys, nextKey)

		element, errGet := tmp.getWaitingListElement(nextKey)
		if errGet != nil {
			return nil
		}

		nextKey = make([]byte, len(element.NextKey))
		if len(element.NextKey) == 0 {
			break
		}
		index++
		copy(nextKey, element.NextKey)
	}
	return allPubKeys
}

func (tmp *TestMetaProcessor) getWaitingListElement(key []byte) (*systemSmartContracts.ElementInList, error) {
	stakingSCAcc := loadSCAccount(tmp.AccountsAdapter, vm.StakingSCAddress)

	marshaledData, _ := stakingSCAcc.DataTrieTracker().RetrieveValue(key)
	if len(marshaledData) == 0 {
		return nil, vm.ErrElementNotFound
	}

	element := &systemSmartContracts.ElementInList{}
	err := tmp.Marshaller.Unmarshal(element, marshaledData)
	if err != nil {
		return nil, err
	}

	return element, nil
}

func (tmp *TestMetaProcessor) getCurrentHeaderInfo() (data.HeaderHandler, []byte) {
	currentHeader := tmp.BlockChainHandler.GetCurrentBlockHeader()
	currentHash := tmp.BlockChainHandler.GetCurrentBlockHeaderHash()
	if currentHeader == nil {
		currentHeader = tmp.BlockChainHandler.GetGenesisHeader()
		currentHash = tmp.BlockChainHandler.GetGenesisHeaderHash()
	}

	return currentHeader, currentHash
}

func createMetaBlockToCommit(
	epoch uint32,
	round uint64,
	prevHash []byte,
	prevRandSeed []byte,
	consensusSize uint32,
) *block.MetaBlock {
	roundStr := strconv.Itoa(int(round))
	hdr := block.MetaBlock{
		Epoch:                  epoch,
		Nonce:                  round,
		Round:                  round,
		PrevHash:               prevHash,
		Signature:              []byte("signature"),
		PubKeysBitmap:          []byte(strings.Repeat("f", int(consensusSize))),
		RootHash:               []byte("roothash"),
		ShardInfo:              make([]block.ShardData, 0),
		TxCount:                1,
		PrevRandSeed:           prevRandSeed,
		RandSeed:               []byte("roothash" + roundStr),
		AccumulatedFeesInEpoch: big.NewInt(0),
		AccumulatedFees:        big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
	}

	shardMiniBlockHeaders := make([]block.MiniBlockHeader, 0)
	shardMiniBlockHeader := block.MiniBlockHeader{
		Hash:            []byte("mb_hash" + roundStr),
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxCount:         1,
	}
	shardMiniBlockHeaders = append(shardMiniBlockHeaders, shardMiniBlockHeader)
	shardData := block.ShardData{
		Nonce:                 round,
		ShardID:               0,
		HeaderHash:            []byte("hdr_hash" + roundStr),
		TxCount:               1,
		ShardMiniBlockHeaders: shardMiniBlockHeaders,
		DeveloperFees:         big.NewInt(0),
		AccumulatedFees:       big.NewInt(0),
	}
	hdr.ShardInfo = append(hdr.ShardInfo, shardData)

	return &hdr
}

func generateAddress(identifier uint32) []byte {
	uniqueIdentifier := fmt.Sprintf("address-%d", identifier)
	return []byte(strings.Repeat("0", addressLength-len(uniqueIdentifier)) + uniqueIdentifier)
}
