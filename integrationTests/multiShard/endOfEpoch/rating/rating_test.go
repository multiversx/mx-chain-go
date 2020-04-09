package rating

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/redirects"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

func TestComputeRating_Metachain(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetDisplayByteSlice(logger.ToHexShort)

	//var fileForLogs *os.File
	//fileForLogs, err := prepareLogFile("")
	//
	//if err != nil {
	//	fmt.Errorf("%w creating a log file", err)
	//}
	//
	//defer func() {
	//	_ = fileForLogs.Close()
	//}()
	//
	//err = logger.RemoveLogObserver(os.Stdout)
	//
	//logger.SetLogLevel("*:DEBUG;peer:TRACE")

	epochsStandard := uint64(10)
	datas := []struct {
		consensusSize int
		shardSize     int
		waitingSize   int
		offlineNodes  int
		epochs        uint64
		roundDuration uint64
		rounds        uint64
	}{
		{
			consensusSize: 63,
			shardSize:     400,
			offlineNodes:  0,
			epochs:        epochsStandard,
			roundDuration: 6,
			rounds:        500,
			waitingSize:   100,
		},
		{
			consensusSize: 400,
			shardSize:     400,
			offlineNodes:  0,
			epochs:        epochsStandard,
			roundDuration: 6,
			rounds:        500,
			waitingSize:   100,
		},
		{
			consensusSize: 63,
			shardSize:     400,
			offlineNodes:  1,
			epochs:        epochsStandard,
			roundDuration: 6,
			rounds:        500,
			waitingSize:   100,
		},
		{
			consensusSize: 400,
			shardSize:     400,
			offlineNodes:  1,
			epochs:        epochsStandard,
			roundDuration: 6,
			rounds:        500,
			waitingSize:   100,
		},
	}

	for _, data := range datas {
		nodesPerShard := data.shardSize
		nbMetaNodes := data.shardSize
		nbShards := 1
		consensusGroupSize := data.consensusSize
		metaConsensusGroupSize := data.consensusSize
		waitingListsSize := data.waitingSize
		roundsPerEpoch := data.rounds
		offlineNodesCurrent := data.offlineNodes
		epochs := data.epochs
		roundDurationSeconds := data.roundDuration

		maxGasLimitPerBlock := uint64(100000)

		advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
		_ = advertiser.Bootstrap()

		seedAddress := integrationTests.GetConnectableAddress(advertiser)

		ratingsConfig := &config.RatingsConfig{}
		_ = core.LoadTomlFile(ratingsConfig, fmt.Sprintf("./%vratings.toml", consensusGroupSize))
		ratingsData, _ := rating.NewRatingsData(*ratingsConfig)
		rater, _ := rating.NewBlockSigningRater(ratingsData)

		coordinatorFactory := &integrationTests.IndexHashedNodesCoordinatorWithRaterFactory{
			PeerAccountListAndRatingHandler: rater,
		}

		nodesMap := createOneNodePerShard(
			nodesPerShard,
			nbMetaNodes,
			nbShards,
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
				node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
			}
		}

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

		round := uint64(1)
		//nonce := uint64(1)

		nbBlocksToProduce := epochs * roundsPerEpoch

		validatorRatings := make(map[uint32]map[string][]uint32)

		for i := uint32(0); i < uint32(nbShards); i++ {
			validatorRatings[i] = make(map[string][]uint32)
		}

		validatorRatings[core.MetachainShardId] = make(map[string][]uint32)

		folderName := fmt.Sprintf("%v_epochs%v_rounds%v", time.Now().Unix(), epochs, roundsPerEpoch)
		os.Mkdir(folderName, os.ModePerm)

		for i := uint64(0); i < nbBlocksToProduce; i++ {
			bodies, headers, _ := SimulateAllShardsProposeBlock(round, nodesMap, offlineNodesPks)
			SimulateSyncAllShardsWithRoundBlock(nodesMap, headers, bodies)

			if headers[core.MetachainShardId] != nil && headers[core.MetachainShardId].IsStartOfEpochBlock() {
				metachainNode := nodesMap[core.MetachainShardId][0]
				_, _ = metachainNode.InterceptorsContainer.Get("")
				metaValidatoStatisticProcessor := metachainNode.ValidatorStatisticsProcessor
				currentRootHash, _ := metaValidatoStatisticProcessor.RootHash()
				validatorInfos, _ := metaValidatoStatisticProcessor.GetValidatorInfoForRootHash(currentRootHash)

				for shardId, v := range validatorInfos {
					currentMap := validatorRatings[shardId]
					for _, validator := range v {
						key := fmt.Sprintf("%s", core.GetTrimmedPk(core.ToHex(validator.GetPublicKey())))
						currentMap[key] = append(currentMap[key], validator.TempRating)
					}
				}
				currEpoch := uint64(headers[core.MetachainShardId].GetEpoch())
				writeRatingsForCurrentEpoch(folderName, currEpoch, roundsPerEpoch, roundDurationSeconds, validatorRatings)
				writeValidatorInfos(folderName, currEpoch, validatorInfos)

				time.Sleep(1 * time.Second)
			}
			round++
		}

		writeRatingsForCurrentEpoch(folderName, epochs, roundsPerEpoch, roundDurationSeconds, validatorRatings)

		metachainNode := nodesMap[core.MetachainShardId][0]
		_, _ = metachainNode.InterceptorsContainer.Get("")
		metaValidatoStatisticProcessor := metachainNode.ValidatorStatisticsProcessor
		currentRootHash, _ := metaValidatoStatisticProcessor.RootHash()
		validatorInfos, _ := metaValidatoStatisticProcessor.GetValidatorInfoForRootHash(currentRootHash)

		writeValidatorInfos(folderName, epochs, validatorInfos)

		fdata, _ := os.Create(fmt.Sprintf("%s/data.out", folderName))
		defer fdata.Close()
		fdata.WriteString(fmt.Sprintf("%#v\n", data))
		fdata.WriteString(fmt.Sprintf("%#v]n", ratingsData))
	}
}

func writeValidatorInfos(folderName string, epoch uint64, validatorInfos map[uint32][]*state.ValidatorInfo) {
	f, _ := os.Create(fmt.Sprintf("%s/validatorInfos_%d.csv", folderName, epoch))
	defer f.Close()

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

	f.WriteString(headerCSV + "\n")

	for shardId, shardVis := range validatorInfos {
		if shardId < 100 {
			continue
		}
		for _, vi := range shardVis {
			f.WriteString(fmt.Sprintf("%s,%v,%s,%v,%v,%v,%s,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
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

func prepareLogFile(workingDir string) (*os.File, error) {
	fileForLog, err := core.CreateFile("elrond-go", workingDir, "log")
	if err != nil {
		return nil, err
	}

	//we need this function as to close file.Close() when the code panics and the defer func associated
	//with the file pointer in the main func will never be reached
	runtime.SetFinalizer(fileForLog, func(f *os.File) {
		_ = f.Close()
	})

	err = redirects.RedirectStderr(fileForLog)
	if err != nil {
		return nil, err
	}

	err = logger.AddLogObserver(fileForLog, &logger.PlainFormatter{})
	if err != nil {
		return nil, fmt.Errorf("%w adding file log observer", err)
	}

	return fileForLog, nil
}

func writeRatingsForCurrentEpoch(folderName string, currentEpoch uint64, roundsPerEpoch uint64, roundDurationSeconds uint64, validatorRatingsMap map[uint32]map[string][]uint32) {
	f, _ := os.Create(fmt.Sprintf("%s/%v.csv", folderName, currentEpoch))
	defer f.Close()

	headerCSV := "pk"
	for i := uint64(0); i < currentEpoch; i++ {
		headerCSV = fmt.Sprintf("%s,%v(%sh)", headerCSV, i, fmt.Sprintf("%.2f", float32((i+1)*roundsPerEpoch*roundDurationSeconds)/3600))
	}
	f.WriteString(headerCSV + "\n")

	for shardId, validatorRatings := range validatorRatingsMap {
		if shardId < 100 {
			continue
		}
		var keys []string
		for k := range validatorRatings {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		f.WriteString(fmt.Sprintf("%d\n", shardId))

		for _, key := range keys {
			ratingPerEpoch := validatorRatings[key]
			csvRatings := core.GetTrimmedPk(key)
			prevValue := uint32(0)
			for _, rt := range ratingPerEpoch {
				cellValue := fmt.Sprintf("%d", rt)
				if prevValue != 10000000 && (prevValue == rt || rt == 5000000) {
					cellValue = "waiting"
				}
				csvRatings = fmt.Sprintf("%s,%s", csvRatings, cellValue)
				prevValue = rt
			}

			f.WriteString(csvRatings + "\n")
		}
	}
}

func TestEpochChangeWithNodesShufflingAndRater(t *testing.T) {

	t.Skip("this is not a short test")

	_ = logger.SetDisplayByteSlice(logger.ToHexShort)

	nodesPerShard := 400
	nbMetaNodes := 400
	nbShards := 1
	consensusGroupSize := 63
	metaConsensusGroupSize := 63
	waintingListsSize := 100
	maxGasLimitPerBlock := uint64(100000)

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	ratingsData := integrationTests.CreateRatingsData()
	rater, _ := rating.NewBlockSigningRater(ratingsData)

	coordinatorFactory := &integrationTests.IndexHashedNodesCoordinatorWithRaterFactory{
		PeerAccountListAndRatingHandler: rater,
	}

	nodesMap := createOneNodePerShard(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		metaConsensusGroupSize,
		waintingListsSize,
		seedAddress,
		coordinatorFactory,
		ratingsData)

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

	roundsPerEpoch := uint64(100)
	for _, nodes := range nodesMap {
		integrationTests.SetEconomicsParameters(nodes, maxGasLimitPerBlock, gasPrice, gasLimit)
		nodes[0].Node.Start()
		//integrationTests.DisplayAndStartNodes(nodes)
		for _, node := range nodes {
			node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
		}
	}

	round := uint64(1)
	nonce := uint64(1)
	nbBlocksToProduce := uint64(10 * roundsPerEpoch)
	expectedLastEpoch := uint32(nbBlocksToProduce / roundsPerEpoch)

	for i := uint64(0); i < nbBlocksToProduce; i++ {
		_, headers, _ := SimulateAllShardsProposeBlock(round, nodesMap, nil)

		if headers[core.MetachainShardId].IsStartOfEpochBlock() {
			_, _ = nodesMap[core.MetachainShardId][0].InterceptorsContainer.Get("")
			metaValidatoStatisticProcessor := nodesMap[core.MetachainShardId][0].ValidatorStatisticsProcessor
			currentRootHash, _ := metaValidatoStatisticProcessor.RootHash()
			validatorInfos, _ := metaValidatoStatisticProcessor.GetValidatorInfoForRootHash(currentRootHash)
			for _, v := range validatorInfos {
				for _, validator := range v {
					fmt.Println(validator.TempRating)
				}
			}
		}
		round++
		nonce++
	}

	for _, nodes := range nodesMap {
		endOfEpoch.VerifyThatNodesHaveCorrectEpoch(t, expectedLastEpoch, nodes)
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
	i uint32,
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

	// TODO: remove if start of epoch block needs to be validated by the new epoch nodes
	epoch := currentBlockHeader.GetEpoch()
	prevRandomness := currentBlockHeader.GetRandSeed()

	nodesCoordinator := currentNode.NodesCoordinator

	pubKeys, err := nodesCoordinator.GetConsensusValidatorsPublicKeys(prevRandomness, round, i, epoch)

	if err != nil {
		fmt.Println("Error getting the validators public keys: ", err)
	}

	proposerKey := []byte(pubKeys[0])
	for _, pk := range offlinePks {
		if bytes.Equal(pk, proposerKey) {
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

	header.SetPrevRandSeed(prevRandomness)
	header = SimulateDoConsensusSigningOnBlock(header, pubKeys, remainingPks)

	if header == nil {
		return
	}

	mutex.Lock()
	bodyMap[i] = body
	headerMap[i] = header
	//consensusNodesMap[i] = consensusNodes
	newRandomness[i] = headerMap[i].GetRandSeed()
	mutex.Unlock()

	currentNode.CommitBlock(body, header)

}

// AllShardsProposeBlock simulates each shard selecting a consensus group and proposing/broadcasting/committing a block
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
	for i := range nodesMap {
		currentNode := nodesMap[i][0]
		offlineKeys := offlineNodesMap[i]
		go createBlock(currentNode, i, round, bodyMap, headerMap, newRandomness, mutMaps, wg, offlineKeys)
	}
	wg.Wait()

	time.Sleep(1 * time.Millisecond)

	return bodyMap, headerMap, consensusNodesMap
}

// SyncAllShardsWithRoundBlock enforces all nodes in each shard synchronizing the block for the given round
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

// DoConsensusSigningOnBlock simulates a consensus aggregated signature on the provided block
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

	blockHeader.SetSignature(blockHeaderHash)
	blockHeader.SetPubKeysBitmap(bitmap)
	blockHeader.SetLeaderSignature([]byte("leader sign"))

	return blockHeader
}
