package rating

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

type testData struct {
	consensusSize     int
	shardSize         int
	metaConsensusSize int
	metaShardSize     int
	waitingSize       int
	offlineNodes      int
	epochs            int
	roundDuration     int
	roundsPerEpoch    int
	numShards         int
}

func TestComputeRating_Metachain(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	datas := createTestData()

	for _, testData := range datas {
		computeRatings(testData)
	}
}

func computeRatings(testData testData) {
	nodesPerShard := testData.shardSize
	numMetaNodes := testData.metaShardSize
	numShards := testData.numShards
	consensusGroupSize := testData.consensusSize
	metaConsensusGroupSize := testData.metaConsensusSize
	waitingListsSize := testData.waitingSize
	roundsPerEpoch := testData.roundsPerEpoch
	offlineNodesCurrent := testData.offlineNodes
	epochs := testData.epochs
	roundDurationSeconds := testData.roundDuration

	maxGasLimitPerBlock := uint64(100000)

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	ratingsConfig := &config.RatingsConfig{}
	_ = core.LoadTomlFile(ratingsConfig, "./ratings.toml")

	args := rating.RatingsDataArg{
		Config:                   *ratingsConfig,
		ShardConsensusSize:       uint32(testData.consensusSize),
		MetaConsensusSize:        uint32(testData.consensusSize),
		ShardMinNodes:            uint32(testData.shardSize),
		MetaMinNodes:             uint32(testData.shardSize),
		RoundDurationMiliseconds: uint64(testData.roundDuration * 1000),
	}

	ratingsData, _ := rating.NewRatingsData(args)
	rater, _ := rating.NewBlockSigningRater(ratingsData)

	coordinatorFactory := &integrationTests.IndexHashedNodesCoordinatorWithRaterFactory{
		PeerAccountListAndRatingHandler: rater,
	}

	nodesMap := createOneNodePerShard(
		nodesPerShard,
		numMetaNodes,
		numShards,
		consensusGroupSize,
		metaConsensusGroupSize,
		waitingListsSize,
		seedAddress,
		coordinatorFactory,
		ratingsData,
	)

	delete(nodesMap, 0)

	gasPrice := uint64(10)
	gasLimit := uint64(100)

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Node.Stop()
			}
		}
	}()

	for _, nodes := range nodesMap {
		integrationTests.SetEconomicsParameters(nodes, maxGasLimitPerBlock, gasPrice, gasLimit)
		nodes[0].Node.Start()
		//integrationTests.DisplayAndStartNodes(nodes)
		for _, node := range nodes {
			node.EpochStartTrigger.SetRoundsPerEpoch(uint64(roundsPerEpoch))
		}
	}

	offlineNodesPks := setOfflineNodes(nodesMap, offlineNodesCurrent)

	round := uint64(1)
	//nonce := uint64(1)

	nbBlocksToProduce := (epochs + 1) * roundsPerEpoch

	validatorRatings := make(map[uint32]map[string][]uint32)

	for i := uint32(0); i < uint32(numShards); i++ {
		validatorRatings[i] = make(map[string][]uint32)
	}

	validatorRatings[core.MetachainShardId] = make(map[string][]uint32)

	folderName := fmt.Sprintf("%v_epochs%v_rounds%v", time.Now().Unix(), epochs, roundsPerEpoch)
	_ = os.Mkdir(folderName, os.ModePerm)

	writeOutputComputed(testData, folderName, ratingsData)

	startTime := time.Now().Unix()
	for i := 0; i < nbBlocksToProduce; i++ {

		bodies, headers, _ := SimulateAllShardsProposeBlock(round, nodesMap, offlineNodesPks)
		SimulateSyncAllShardsWithRoundBlock(nodesMap, headers, bodies)

		if headers[core.MetachainShardId] != nil && headers[core.MetachainShardId].IsStartOfEpochBlock() {
			validatorInfos := getValidatorInfos(nodesMap)

			for shardId, v := range validatorInfos {
				currentMap := validatorRatings[shardId]
				for _, validator := range v {
					key := fmt.Sprintf("%s", core.GetTrimmedPk(core.ToHex(validator.GetPublicKey())))
					currentMap[key] = append(currentMap[key], validator.TempRating)
				}
			}

			endtime := time.Now().Unix()
			neededTime := endtime - startTime
			fmt.Println(fmt.Sprintf("Time for epoch %v seconds", neededTime))
			startTime = endtime

			currEpoch := uint64(headers[core.MetachainShardId].GetEpoch())
			writeRatingsForCurrentEpoch(folderName, int(currEpoch), roundsPerEpoch, roundDurationSeconds, validatorRatings, neededTime)
			writeValidatorInfos(folderName, int(currEpoch), validatorInfos)

			time.Sleep(1 * time.Second)
		}
		round++
	}

	validatorInfos := getValidatorInfos(nodesMap)
	writeValidatorInfos(folderName, epochs, validatorInfos)
	writeRatingsForCurrentEpoch(folderName, epochs, roundsPerEpoch, roundDurationSeconds, validatorRatings, 0)
}

func writeOutputComputed(testData testData, folderName string, ratingsData *rating.RatingsData) {
	computedRatingsData := filepath.Join(folderName, "computed_ratings.info")
	computedRatingsDataStr := createStringFromRatingsData(ratingsData)
	_ = ioutil.WriteFile(computedRatingsData, []byte(computedRatingsDataStr), os.ModePerm)

	outData, _ := os.Create(fmt.Sprintf("%s/data_%v_%v_%v_%v_%v.out",
		folderName,
		testData.epochs,
		testData.roundsPerEpoch,
		testData.shardSize,
		testData.consensusSize,
		testData.offlineNodes))
	defer func() {
		_ = outData.Close()
	}()
	_, _ = outData.WriteString(fmt.Sprintf("%#v\n", testData))
}

func createTestData() []testData {
	epochsStandard := 50
	rounds := 600
	shardSize := 400
	waitingSize := 100
	roundDuration := 6
	numShards := 1
	datas := []testData{
		{
			consensusSize:     63,
			shardSize:         shardSize,
			metaConsensusSize: 63,
			metaShardSize:     shardSize,
			waitingSize:       waitingSize,
			offlineNodes:      0,
			epochs:            epochsStandard,
			roundDuration:     roundDuration,
			roundsPerEpoch:    rounds,
			numShards:         numShards,
		},
		{
			consensusSize:     shardSize,
			shardSize:         shardSize,
			metaConsensusSize: shardSize,
			metaShardSize:     shardSize,
			waitingSize:       waitingSize,
			offlineNodes:      0,
			epochs:            epochsStandard,
			roundDuration:     roundDuration,
			roundsPerEpoch:    rounds,
			numShards:         numShards,
		},
		{
			numShards:         numShards,
			consensusSize:     63,
			shardSize:         shardSize,
			metaConsensusSize: 63,
			metaShardSize:     shardSize,
			offlineNodes:      1,
			epochs:            epochsStandard,
			roundDuration:     roundDuration,
			roundsPerEpoch:    rounds,
			waitingSize:       waitingSize,
		},
		{
			numShards:         numShards,
			consensusSize:     shardSize,
			shardSize:         shardSize,
			metaConsensusSize: shardSize,
			metaShardSize:     shardSize,
			offlineNodes:      1,
			epochs:            epochsStandard,
			roundDuration:     roundDuration,
			roundsPerEpoch:    rounds,
			waitingSize:       waitingSize,
		},
	}
	return datas
}

func getValidatorInfos(nodesMap map[uint32][]*integrationTests.TestProcessorNode) map[uint32][]*state.ValidatorInfo {
	metachainNode := nodesMap[core.MetachainShardId][0]
	metaValidatoStatisticProcessor := metachainNode.ValidatorStatisticsProcessor
	currentRootHash, _ := metaValidatoStatisticProcessor.RootHash()
	validatorInfos, _ := metaValidatoStatisticProcessor.GetValidatorInfoForRootHash(currentRootHash)
	return validatorInfos
}

func setOfflineNodes(nodesMap map[uint32][]*integrationTests.TestProcessorNode, offlineNodesCurrent int) map[uint32][][]byte {
	offlineNodesPks := make(map[uint32][][]byte)
	shardValidators, _ := nodesMap[core.MetachainShardId][0].NodesCoordinator.GetAllEligibleValidatorsPublicKeys(0)
	shardWaiting, _ := nodesMap[core.MetachainShardId][0].NodesCoordinator.GetAllWaitingValidatorsPublicKeys(0)

	for i := range shardValidators {
		shardValidators[i] = append(shardValidators[i], shardWaiting[i]...)
	}

	for shardId, shardPks := range shardValidators {
		sort.Slice(shardPks, func(i, j int) bool {
			return bytes.Compare(shardPks[i], shardPks[j]) < 0
		})

		for i := 0; i < offlineNodesCurrent; i++ {
			offlineNodesPks[shardId] = append(offlineNodesPks[shardId], shardPks[i])
		}
	}
	return offlineNodesPks
}

func writeValidatorInfos(folderName string, epoch int, validatorInfos map[uint32][]*state.ValidatorInfo) {
	f, _ := os.Create(fmt.Sprintf("%s/validatorInfos_%d.csv", folderName, epoch))
	defer func() {
		_ = f.Close()
	}()

	headerCSV :=
		"PublicKey," +
			"ShardId," +
			"List," +
			"Index," +
			"TempRating," +
			"Rating," +
			"RewardAddress," +
			"LeaderSuccess," +
			"LeaderFailure," +
			"ValidatorSuccess," +
			"ValidatorFailure," +
			"NumSelectedInSuccessBlocks," +
			"AccumulatedFees," +
			"TotalLeaderSuccess," +
			"TotalLeaderFailure," +
			"TotalValidatorSuccess," +
			"TotalValidatorFailure,"

	_, _ = f.WriteString(headerCSV + "\n")

	for shardId, shardVis := range validatorInfos {
		if shardId != core.MetachainShardId {
			continue
		}
		for _, vi := range shardVis {
			_, _ = f.WriteString(fmt.Sprintf("%s,%v,%s,%v,%v,%v,%s,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
				core.GetTrimmedPk(core.ToHex(vi.PublicKey)),
				vi.ShardId,
				vi.List,
				vi.Index,
				vi.TempRating,
				vi.Rating,
				core.GetTrimmedPk(core.ToHex(vi.RewardAddress)),
				vi.LeaderSuccess,
				vi.LeaderFailure,
				vi.ValidatorSuccess,
				vi.ValidatorFailure,
				vi.NumSelectedInSuccessBlocks,
				vi.AccumulatedFees,
				vi.TotalLeaderSuccess,
				vi.TotalLeaderFailure,
				vi.TotalValidatorSuccess,
				vi.TotalValidatorFailure,
			))
		}
	}
}

func writeRatingsForCurrentEpoch(
	folderName string,
	currentEpoch int,
	roundsPerEpoch int,
	roundDurationSeconds int,
	validatorRatingsMap map[uint32]map[string][]uint32,
	neededTime int64) {
	f, _ := os.Create(fmt.Sprintf("%s/ratings_%v_neededSeconds_%v.csv", folderName, currentEpoch, neededTime))
	defer func() {
		_ = f.Close()
	}()

	headerCSV := "pk"
	for i := 0; i < currentEpoch; i++ {
		headerCSV = fmt.Sprintf("%s,%v(%sh)", headerCSV, i, fmt.Sprintf("%.2f", float32((i+1)*roundsPerEpoch*roundDurationSeconds)/3600))
	}
	_, _ = f.WriteString(headerCSV + "\n")

	for shardId, validatorRatings := range validatorRatingsMap {
		if shardId != core.MetachainShardId {
			continue
		}
		var keys []string
		for k := range validatorRatings {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		_, _ = f.WriteString(fmt.Sprintf("%d\n", shardId))

		for _, key := range keys {
			ratingPerEpoch := validatorRatings[key]
			csvRatings := core.GetTrimmedPk(key)
			for _, rt := range ratingPerEpoch {
				cellValue := fmt.Sprintf("%d", rt)
				csvRatings = fmt.Sprintf("%s,%s", csvRatings, cellValue)
			}

			_, _ = f.WriteString(csvRatings + "\n")
		}
	}
}

func createOneNodePerShard(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	consensusGroupSize int,
	metaConsensusGroupSize int,
	nodesInWaitingListPerShard int,
	seedAddress string,
	coordinatorFactory *integrationTests.IndexHashedNodesCoordinatorWithRaterFactory,
	ratingsData *rating.RatingsData,
) map[uint32][]*integrationTests.TestProcessorNode {
	cp := integrationTests.CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	pubKeys := integrationTests.PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := integrationTests.GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))
	validatorsMapForNodesCoordinator, _ := sharding.NodesInfoToValidators(validatorsMap)

	cpWaiting := integrationTests.CreateCryptoParams(nodesInWaitingListPerShard, nodesInWaitingListPerShard, uint32(nbShards))
	pubKeysWaiting := integrationTests.PubKeysMapFromKeysMap(cpWaiting.Keys)
	waitingMap := integrationTests.GenValidatorsFromPubKeys(pubKeysWaiting, uint32(nbShards))
	waitingMapForNodesCoordinator, _ := sharding.NodesInfoToValidators(waitingMap)

	nodesMap := make(map[uint32][]*integrationTests.TestProcessorNode)

	nodesSetup := &mock.NodesSetupStub{InitialNodesInfoCalled: func() (m map[uint32][]sharding.GenesisNodeInfoHandler, m2 map[uint32][]sharding.GenesisNodeInfoHandler) {
		return validatorsMap, waitingMap
	}}

	for shardId := range validatorsMap {
		nodesList := make([]*integrationTests.TestProcessorNode, 1)
		nodesListWaiting := make([]*integrationTests.TestProcessorNode, 0)

		dataCache, _ := lrucache.NewCache(10000)
		nodesList[0] = integrationTests.CreateNode(
			nodesPerShard,
			nbMetaNodes,
			consensusGroupSize,
			metaConsensusGroupSize,
			shardId,
			nbShards,
			validatorsMapForNodesCoordinator,
			waitingMapForNodesCoordinator,
			0,
			seedAddress,
			cp,
			dataCache,
			coordinatorFactory,
			nodesSetup,
			ratingsData,
		)

		nodesMap[shardId] = append(nodesList, nodesListWaiting...)
	}

	return nodesMap
}

func createBlock(currentNode *integrationTests.TestProcessorNode,
	shardId uint32,
	round uint64,
	bodyMap map[uint32]data.BodyHandler,
	headerMap map[uint32]data.HeaderHandler,
	newRandomness map[uint32][]byte,
	mutex *sync.Mutex,
	wg *sync.WaitGroup,
	offlinePks [][]byte) {

	defer wg.Done()

	currentBlockHeader := currentNode.BlockChain.GetCurrentBlockHeader()
	if check.IfNil(currentBlockHeader) {
		currentBlockHeader = currentNode.BlockChain.GetGenesisHeader()
	}

	epoch := currentBlockHeader.GetEpoch()
	prevRandomness := currentBlockHeader.GetRandSeed()

	nodesCoordinator := currentNode.NodesCoordinator

	marshalled, _ := integrationTests.TestMarshalizer.Marshal(currentBlockHeader)
	hash := integrationTests.TestHasher.Compute(string(marshalled))

	fmt.Println(fmt.Sprintf("Previous header hash %s and round %v and prevRandSeed %s",
		core.GetTrimmedPk(core.ToHex(hash)),
		currentBlockHeader.GetNonce(),
		core.GetTrimmedPk(core.ToHex(currentBlockHeader.GetRandSeed()))))

	pubKeys, err := nodesCoordinator.GetConsensusValidatorsPublicKeys(prevRandomness, round, shardId, epoch)

	if err != nil {
		fmt.Println("Error getting the validators public keys: ", err)
	}

	proposerKey := []byte(pubKeys[0])
	for _, pk := range offlinePks {
		if bytes.Equal(pk, proposerKey) {
			fmt.Println(fmt.Sprintf("Not proposed by %s", core.GetTrimmedPk(core.ToHex(proposerKey))))
			return
		}
	}

	// first node is block proposer
	var body data.BodyHandler
	var header data.HeaderHandler
	for i := 0; i < 10; i++ {
		body, header, _ = currentNode.ProposeBlock(round, currentBlockHeader.GetNonce()+1)
		if body != nil && header != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	pubKeysLen := len(pubKeys)
	signersLen := pubKeysLen*2/3 + 1

	missingSigs := 0
	remainingPks := make(map[string]bool)

	for i := 0; i < len(pubKeys); i++ {
		found := false
		for _, pk := range offlinePks {
			if bytes.Equal(pk, []byte(pubKeys[i])) {
				missingSigs++
				found = true
				break
			}
		}
		if !found {
			remainingPks[pubKeys[i]] = true
		}

	}

	if pubKeysLen-missingSigs < signersLen {
		return
	}

	if header == nil {
		return
	}

	header.SetPrevRandSeed(prevRandomness)
	header = SimulateDoConsensusSigningOnBlock(header, pubKeys, remainingPks)

	if header == nil {
		return
	}

	mutex.Lock()
	bodyMap[shardId] = body
	headerMap[shardId] = header
	//consensusNodesMap[shardId] = consensusNodes
	newRandomness[shardId] = headerMap[shardId].GetRandSeed()
	mutex.Unlock()

	currentNode.CommitBlock(body, header)

}

// SimulateAllShardsProposeBlock simulates each shard selecting a consensus group and proposing/broadcasting/committing a block
func SimulateAllShardsProposeBlock(
	round uint64,
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
	offlineNodesMap map[uint32][][]byte,
) (
	map[uint32]data.BodyHandler,
	map[uint32]data.HeaderHandler,
	map[uint32][]*integrationTests.TestProcessorNode,
) {

	bodyMap := make(map[uint32]data.BodyHandler)
	headerMap := make(map[uint32]data.HeaderHandler)
	consensusNodesMap := make(map[uint32][]*integrationTests.TestProcessorNode)
	newRandomness := make(map[uint32][]byte)

	wg := &sync.WaitGroup{}

	wg.Add(len(nodesMap))
	mutMaps := &sync.Mutex{}
	// propose blocks
	for shardId := range nodesMap {
		currentNode := nodesMap[shardId][0]
		offlineKeys := offlineNodesMap[shardId]
		go createBlock(currentNode, shardId, round, bodyMap, headerMap, newRandomness, mutMaps, wg, offlineKeys)
	}
	wg.Wait()

	time.Sleep(1 * time.Millisecond)

	return bodyMap, headerMap, consensusNodesMap
}

// SimulateSyncAllShardsWithRoundBlock enforces all nodes in each shard synchronizing the block for the given round
func SimulateSyncAllShardsWithRoundBlock(
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
	headerMap map[uint32]data.HeaderHandler,
	bodyMap map[uint32]data.BodyHandler,
) {
	for shard, nodeList := range nodesMap {
		for _, node := range nodeList {
			for _, header := range headerMap {
				if header.GetShardID() == shard {
					continue
				}
				marshalizedHeader, _ := integrationTests.TestMarshalizer.Marshal(header)
				headerHash := integrationTests.TestHasher.Compute(string(marshalizedHeader))

				if shard == core.MetachainShardId {
					node.DataPool.Headers().AddHeader(headerHash, header)
				} else {
					if header.GetShardID() == core.MetachainShardId {
						node.DataPool.Headers().AddHeader(headerHash, header)
					}
				}
			}

			for _, body := range bodyMap {
				actualBody := body.(*block.Body)
				for _, miniBlock := range actualBody.MiniBlocks {
					marshalizedHeader, _ := integrationTests.TestMarshalizer.Marshal(miniBlock)
					headerHash := integrationTests.TestHasher.Compute(string(marshalizedHeader))
					node.DataPool.MiniBlocks().Put(headerHash, miniBlock)
				}
			}
		}
	}
}

// SimulateDoConsensusSigningOnBlock simulates a consensus aggregated signature on the provided block
func SimulateDoConsensusSigningOnBlock(
	blockHeader data.HeaderHandler,
	pubKeys []string,
	remainingKeys map[string]bool,
) data.HeaderHandler {
	pubKeysLen := len(pubKeys)
	signersLen := pubKeysLen*2/3 + 1

	// set bitmap for all consensus nodes signing
	bitmap := make([]byte, pubKeysLen/8+1)
	bitmap[signersLen/8] >>= uint8(8 - (signersLen % 8))

	for i, pk := range pubKeys {
		if remainingKeys[pk] {
			bitmap[i/8] |= 1 << (uint16(i) % 8)
		}
	}

	blockHeaderHash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, blockHeader)
	randSeed := integrationTests.TestHasher.Compute(core.ToHex(blockHeaderHash))

	blockHeader.SetSignature(blockHeaderHash)
	blockHeader.SetRandSeed(randSeed)
	blockHeader.SetPubKeysBitmap(bitmap)
	blockHeader.SetLeaderSignature([]byte("leader sign"))

	return blockHeader
}

func createStringFromRatingsData(ratingsData *rating.RatingsData) string {
	metaChainStepHandler := ratingsData.MetaChainRatingsStepHandler()
	shardChainHandler := ratingsData.ShardChainRatingsStepHandler()
	computedRatingsDataStr := fmt.Sprintf(
		"meta:\n"+
			"ProposerIncrease=%v\n"+
			"ProposerDecrease=%v\n"+
			"ValidatorIncrease=%v\n"+
			"ValidatorDecrease=%v\n\n"+
			"shard:\n"+
			"ProposerIncrease=%v\n"+
			"ProposerDecrease=%v\n"+
			"ValidatorIncrease=%v\n"+
			"ValidatorDecrease=%v",
		metaChainStepHandler.ProposerIncreaseRatingStep(),
		metaChainStepHandler.ProposerDecreaseRatingStep(),
		metaChainStepHandler.ValidatorIncreaseRatingStep(),
		metaChainStepHandler.ValidatorDecreaseRatingStep(),
		shardChainHandler.ProposerIncreaseRatingStep(),
		shardChainHandler.ProposerDecreaseRatingStep(),
		shardChainHandler.ValidatorIncreaseRatingStep(),
		shardChainHandler.ValidatorDecreaseRatingStep(),
	)
	return computedRatingsDataStr
}
