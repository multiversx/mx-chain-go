package rating

import (
	"bytes"
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
	"github.com/ElrondNetwork/elrond-go/core/atomic"
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

	//_ = logger.SetDisplayByteSlice(logger.ToHexShort)
	//
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
	//logger.SetLogLevel("*:DEBUG")

	datas := []struct {
		consensusSize int
		shardSize     int
		waitingSize   int
		offlineNodes  int
		epochs        uint64
		roundDuration uint64
		rounds        uint64
	}{{
		consensusSize: 63,
		shardSize:     400,
		offlineNodes:  0,
		epochs:        101,
		roundDuration: 6,
		rounds:        500,
		waitingSize:   100,
	},
		{
			consensusSize: 400,
			shardSize:     400,
			offlineNodes:  0,
			epochs:        101,
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

		randomness := atomic.Uint64{}

		integrationTests.TestMultiSig = &mock.BelNevMock{
			AggregateSigsMock: func(bitmap []byte) (bytes []byte, err error) {
				randomness.Set(randomness.Get() + 1)
				return integrationTests.TestHasher.Compute(fmt.Sprintf("%d_%d", time.Now().UnixNano(), randomness)), nil
			},
		}

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
			bodies, headers, _ := integrationTests.SimulateAllShardsProposeBlock(round, nodesMap, offlineNodesPks)
			integrationTests.SimulateSyncAllShardsWithRoundBlock(nodesMap, headers, bodies)

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

				writeRatingsForCurrentEpoch(folderName, uint64(headers[core.MetachainShardId].GetEpoch()), roundsPerEpoch, roundDurationSeconds, validatorRatings)

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

		f, _ := os.Create(fmt.Sprintf("%s/validatorInfos.csv", folderName))
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

		fdata, _ := os.Create(fmt.Sprintf("%s/data.out", folderName))
		defer fdata.Close()
		fdata.WriteString(fmt.Sprintf("%#v\n", data))
		fdata.WriteString(fmt.Sprintf("%#v]n", ratingsData))
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
		_, headers, _ := integrationTests.SimulateAllShardsProposeBlock(round, nodesMap, nil)

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
