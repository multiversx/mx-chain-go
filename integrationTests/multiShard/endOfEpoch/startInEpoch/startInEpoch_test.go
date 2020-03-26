package startInEpoch

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestStartInEpochForAShardNodeInMultiShardedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	totalNodesPerShard := 4
	numNodesPerShardOnline := totalNodesPerShard - 1
	shardCnsSize := 2
	metaCnsSize := 3
	numMetachainNodes := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		numNodesPerShardOnline,
		numMetachainNodes,
		numOfShards,
		shardCnsSize,
		metaCnsSize,
		integrationTests.GetConnectableAddress(advertiser),
	)

	nodes := convertToSlice(nodesMap)

	nodeToJoinLate := nodes[numNodesPerShardOnline] // will return the last node in shard 0 which was not used in consensus
	_ = nodeToJoinLate.Messenger.Close()            // set not offline

	nodes = append(nodes[:numNodesPerShardOnline], nodes[numNodesPerShardOnline+1:]...)
	nodes = append(nodes[:2*numNodesPerShardOnline], nodes[2*numNodesPerShardOnline+1:]...)

	roundsPerEpoch := uint64(10)
	for _, node := range nodes {
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * numNodesPerShardOnline
	}
	idxProposers[numOfShards] = numOfShards * numNodesPerShardOnline

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	initialVal := big.NewInt(10000000)
	sendValue := big.NewInt(5)
	integrationTests.MintAllNodes(nodes, initialVal)
	receiverAddress := []byte("12345678901234567890123456789012")

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)

	/////////----- wait for epoch end period
	epoch := uint32(1)
	nrRoundsToPropagateMultiShard := uint64(5)
	for i := uint64(0); i <= (uint64(epoch)*roundsPerEpoch)+nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++

		for _, node := range nodes {
			integrationTests.CreateAndSendTransaction(node, sendValue, receiverAddress, "")
		}

		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)

	endOfEpoch.VerifyThatNodesHaveCorrectEpoch(t, epoch, nodes)
	endOfEpoch.VerifyIfAddedShardHeadersAreWithNewEpoch(t, nodes)

	epochHandler := &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return epoch
		},
	}
	for _, node := range nodes {
		_ = dataRetriever.SetEpochHandlerToHdrResolver(node.ResolversContainer, epochHandler)
	}

	nodesConfig := sharding.NodesSetup{
		StartTime:     time.Now().Unix(),
		RoundDuration: 4000,
		InitialNodes:  getInitialNodes(nodesMap),
	}
	nodesConfig.SetNumberOfShards(uint32(numOfShards))

	messenger := integrationTests.CreateMessengerWithKadDht(context.Background(), integrationTests.GetConnectableAddress(advertiser))
	_ = messenger.Bootstrap()
	time.Sleep(integrationTests.P2pBootstrapDelay)
	argsBootstrapHandler := bootstrap.ArgsEpochStartBootstrap{
		PublicKey:          nodeToJoinLate.NodeKeys.Pk,
		Marshalizer:        integrationTests.TestMarshalizer,
		Hasher:             integrationTests.TestHasher,
		Messenger:          messenger,
		GeneralConfig:      getGeneralConfig(),
		EconomicsData:      integrationTests.CreateEconomicsData(),
		SingleSigner:       &mock.SignerMock{},
		BlockSingleSigner:  &mock.SignerMock{},
		KeyGen:             &mock.KeyGenMock{},
		BlockKeyGen:        &mock.KeyGenMock{},
		GenesisNodesConfig: &nodesConfig,
		PathManager:        &mock.PathManagerStub{},
		WorkingDir:         "test_directory",
		DefaultDBPath:      "test_db",
		DefaultEpochString: "test_epoch",
		DefaultShardString: "test_shard",
		Rater:              &mock.RaterMock{},
	}
	epochStartBootstrap, err := bootstrap.NewEpochStartBootstrap(argsBootstrapHandler)
	assert.Nil(t, err)

	params, err := epochStartBootstrap.Bootstrap()
	assert.NoError(t, err)
	assert.Equal(t, epoch, params.Epoch)
	assert.Equal(t, uint32(0), params.SelfShardId)
	assert.Equal(t, uint32(2), params.NumOfShards)
}

func convertToSlice(originalMap map[uint32][]*integrationTests.TestProcessorNode) []*integrationTests.TestProcessorNode {
	sliceToRet := make([]*integrationTests.TestProcessorNode, 0)
	for _, nodesPerShard := range originalMap {
		for _, node := range nodesPerShard {
			sliceToRet = append(sliceToRet, node)
		}
	}

	return sliceToRet
}

func getInitialNodes(nodesMap map[uint32][]*integrationTests.TestProcessorNode) []*sharding.InitialNode {
	sliceToRet := make([]*sharding.InitialNode, 0)
	for _, nodesPerShard := range nodesMap {
		for _, node := range nodesPerShard {
			pubKeyBytes, _ := node.NodeKeys.Pk.ToByteArray()
			addressBytes := node.OwnAccount.Address.Bytes()
			entry := &sharding.InitialNode{
				PubKey:   hex.EncodeToString(pubKeyBytes),
				Address:  hex.EncodeToString(addressBytes),
				NodeInfo: sharding.NodeInfo{},
			}
			sliceToRet = append(sliceToRet, entry)
		}
	}

	return sliceToRet
}

func getGeneralConfig() config.Config {
	return config.Config{
		EpochStartConfig: config.EpochStartConfig{
			MinRoundsBetweenEpochs: 5,
			RoundsPerEpoch:         10,
		},
		WhiteListPool: config.CacheConfig{
			Size:   10000,
			Type:   "LRU",
			Shards: 1,
		},
		StoragePruning: config.StoragePruningConfig{
			Enabled:             false,
			FullArchive:         true,
			NumEpochsToKeep:     3,
			NumActivePersisters: 3,
		},
		AccountsTrieStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Size: 10000, Type: "LRU", Shards: 1,
			},
			DB: config.DBConfig{
				FilePath:          "AccountsDB",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerAccountsTrieStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Size: 10000, Type: "LRU", Shards: 1,
			},
			DB: config.DBConfig{
				FilePath:          "AccountsDB",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		TxDataPool: config.CacheConfig{
			Size: 10000, Type: "LRU", Shards: 1,
		},
		UnsignedTransactionDataPool: config.CacheConfig{
			Size: 10000, Type: "LRU", Shards: 1,
		},
		RewardTransactionDataPool: config.CacheConfig{
			Size: 10000, Type: "LRU", Shards: 1,
		},
		HeadersPoolConfig: config.HeadersPoolConfig{
			MaxHeadersPerShard:            10,
			NumElementsToRemoveOnEviction: 1,
		},
		TxBlockBodyDataPool: config.CacheConfig{
			Size: 10000, Type: "LRU", Shards: 1,
		},
		PeerBlockBodyDataPool: config.CacheConfig{
			Size: 10000, Type: "LRU", Shards: 1,
		},
		TrieNodesDataPool: config.CacheConfig{
			Size: 10000, Type: "LRU", Shards: 1,
		},
	}
}
