package rating

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/redirects"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

func TestComputeRating_SingleMetachainShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetDisplayByteSlice(logger.ToHexShort)

	var fileForLogs *os.File
	fileForLogs, err := prepareLogFile("")

	if err != nil {
		fmt.Errorf("%w creating a log file", err)
	}

	defer func() {
		_ = fileForLogs.Close()
	}()

	err = logger.RemoveLogObserver(os.Stdout)

	logger.SetLogLevel("*:DEBUG")

	roundsPerEpoch := uint64(500)
	epochs := uint64(50)
	roundDurationSeconds := uint64(6)

	nodesPerShard := 63
	nbMetaNodes := 63
	nbShards := 5
	consensusGroupSize := 51
	waintingListsSize := 7
	maxGasLimitPerBlock := uint64(100000)

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	ratingsConfig := &config.RatingsConfig{}
	_ = core.LoadTomlFile(ratingsConfig, "./ratings.toml")
	ratingsData, _ := rating.NewRatingsData(*ratingsConfig)
	rater, _ := rating.NewBlockSigningRater(ratingsData)

	integrationTests.TestMultiSig = &mock.BelNevMock{
		AggregateSigsMock: func(bitmap []byte) (bytes []byte, err error) {
			return integrationTests.TestHasher.Compute(fmt.Sprintf("%d", time.Now().UnixNano())), nil
		},
	}

	coordinatorFactory := &integrationTests.IndexHashedNodesCoordinatorWithRaterFactory{
		RaterHandler: rater,
	}

	nodesMap := createOneNodePerShard(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		waintingListsSize,
		seedAddress,
		coordinatorFactory,
	)

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

	round := uint64(1)
	nonce := uint64(1)

	nbBlocksToProduce := epochs * roundsPerEpoch

	validatorRatings := make(map[uint32]map[string][]uint32)

	for i := uint32(0); i < uint32(nbShards); i++ {
		validatorRatings[i] = make(map[string][]uint32)
	}

	validatorRatings[core.MetachainShardId] = make(map[string][]uint32)

	folderName := fmt.Sprintf("%v_epochs%v_rounds%v", time.Now().Unix(), epochs, roundsPerEpoch)
	os.Mkdir(folderName, os.ModePerm)

	for i := uint64(0); i < nbBlocksToProduce; i++ {
		bodies, headers, _ := integrationTests.SimulateAllShardsProposeBlock(round, nonce, nodesMap)
		integrationTests.SimulateSyncAllShardsWithRoundBlock(nodesMap, headers, bodies)

		if headers[core.MetachainShardId].IsStartOfEpochBlock() {
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

			writeRatingsToCurrentEpoch(folderName, uint64(headers[core.MetachainShardId].GetEpoch()), roundsPerEpoch, roundDurationSeconds, validatorRatings)

			time.Sleep(1 * time.Second)
		}
		round++
		nonce++

	}

	writeRatingsToCurrentEpoch(folderName, epochs, roundsPerEpoch, roundDurationSeconds, validatorRatings)
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

func writeRatingsToCurrentEpoch(folderName string, currentEpoch uint64, roundsPerEpoch uint64, roundDurationSeconds uint64, validatorRatingsMap map[uint32]map[string][]uint32) {
	f, _ := os.Create(fmt.Sprintf("%s/%v.csv", folderName, currentEpoch))
	defer f.Close()

	headerCSV := "pk"
	for i := uint64(0); i < currentEpoch; i++ {
		headerCSV = fmt.Sprintf("%s,%v(%sh)", headerCSV, i, fmt.Sprintf("%.2f", float32((i+1)*roundsPerEpoch*roundDurationSeconds)/3600))
	}
	f.WriteString(headerCSV + "\n")

	for shardId, validatorRatings := range validatorRatingsMap {

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
				if prevValue != 1000000 && (prevValue == rt || rt == 500000) {
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
	waintingListsSize := 100
	maxGasLimitPerBlock := uint64(100000)

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	rater, _ := rating.NewBlockSigningRater(integrationTests.CreateRatingsData())

	coordinatorFactory := &integrationTests.IndexHashedNodesCoordinatorWithRaterFactory{
		RaterHandler: rater,
	}

	nodesMap := createOneNodePerShard(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		waintingListsSize,
		seedAddress,
		coordinatorFactory)

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
		_, headers, _ := integrationTests.SimulateAllShardsProposeBlock(round, nonce, nodesMap)

		//for shardId, validator := range nodesMap{
		//
		//	if shardId
		//
		//	err = tpn.BlockProcessor.ProcessBlock(
		//		header,
		//		body,
		//		func() time.Duration {
		//			return time.Second * 2
		//		},
		//	)
		//	if err != nil {
		//		return err
		//	}
		//
		//	err = tpn.BlockProcessor.CommitBlock(header, body)
		//}

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
	nodesInWaitingListPerShard int,
	seedAddress string,
	coordinatorFactory *integrationTests.IndexHashedNodesCoordinatorWithRaterFactory,
) map[uint32][]*integrationTests.TestProcessorNode {
	cp := integrationTests.CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	pubKeys := integrationTests.PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := integrationTests.GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))

	cpWaiting := integrationTests.CreateCryptoParams(nodesInWaitingListPerShard, nodesInWaitingListPerShard, uint32(nbShards))
	pubKeysWaiting := integrationTests.PubKeysMapFromKeysMap(cpWaiting.Keys)
	waitingMap := integrationTests.GenValidatorsFromPubKeys(pubKeysWaiting, uint32(nbShards))

	nodesMap := make(map[uint32][]*integrationTests.TestProcessorNode)

	for shardId := range validatorsMap {
		nodesList := make([]*integrationTests.TestProcessorNode, 1)
		nodesListWaiting := make([]*integrationTests.TestProcessorNode, 0)

		dataCache, _ := lrucache.NewCache(10000)
		nodesList[0] = integrationTests.CreateNode(
			nodesPerShard,
			nbMetaNodes,
			consensusGroupSize,
			nbMetaNodes,
			shardId,
			nbShards,
			validatorsMap,
			waitingMap,
			0,
			seedAddress,
			cp,
			dataCache,
			coordinatorFactory,
		)

		nodesMap[shardId] = append(nodesList, nodesListWaiting...)
	}

	return nodesMap
}
