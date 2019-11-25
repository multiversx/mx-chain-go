package integrationTests

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/core"
	kmultisig "github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	scToProtocol2 "github.com/ElrondNetwork/elrond-go/process/scToProtocol"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"strconv"
)

const (
	validatorIncreaseRatingStep   = "validatorIncreaseRatingStep"
	validatorDecreaseRatingStep   = "validatorDecreaseRatingStep"
	proposerIncreaseRatingStepKey = "proposerIncreaseRatingStep"
	proposerDecreaseRatingStepKey = "proposerDecreaseRatingStep"
	minRating                     = uint32(1)
	maxRating                     = uint32(10)
)

type TestProcessorNodeWithRater struct {
	TestProcessorNode
}

// CreateNodesWithNodesCoordinator returns a map with nodes per shard each using a real nodes coordinator
func CreateNodesWithNodesCoordinatorWithRater(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
) map[uint32][]*TestProcessorNode {
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))
	nodesMap := make(map[uint32][]*TestProcessorNode)

	for shardId, validatorList := range validatorsMap {
		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Hasher:                  TestHasher,
			ShardId:                 shardId,
			NbShards:                uint32(nbShards),
			Nodes:                   validatorsMap,
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
		}

		ratingValues := make(map[string]int32, 0)
		ratingValues[validatorIncreaseRatingStep] = 1
		ratingValues[validatorDecreaseRatingStep] = -2
		ratingValues[proposerIncreaseRatingStepKey] = 3
		ratingValues[proposerDecreaseRatingStepKey] = -4

		ratingsData, _ := economics.NewRatingsData(uint32(5), uint32(1), uint32(10), "mockRater", ratingValues)

		rater, _ := rating.NewBlockSigningRater(ratingsData)

		nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinatorWithRater(argumentsNodesCoordinator, rater)

		if err != nil {
			fmt.Println("Error creating node coordinator")
		}

		nodesList := make([]*TestProcessorNode, len(validatorList))
		for i := range validatorList {
			nodesList[i] = NewTestProcessorNodeWithCustomNodesCoordinatorAndRater(
				uint32(nbShards),
				shardId,
				seedAddress,
				nodesCoordinator,
				cp,
				i,
				rater,
			)
		}
		nodesMap[shardId] = nodesList
	}

	return nodesMap
}

func (tpn *TestProcessorNodeWithRater) initBlockProcessor() {
	var err error

	tpn.ForkDetector = &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
	}
	var peerDataPool peer.DataPool = tpn.MetaDataPool
	if tpn.ShardCoordinator.SelfId() < tpn.ShardCoordinator.NumberOfShards() {
		peerDataPool = tpn.ShardDataPool
	}

	initialNodes := make([]*sharding.InitialNode, 0)

	for _, pks := range tpn.NodesCoordinator.GetAllValidatorsPublicKeys() {
		for _, pk := range pks {
			validator, _, _ := tpn.NodesCoordinator.GetValidatorWithPublicKey(pk)
			n := &sharding.InitialNode{
				PubKey:   core.ToHex(validator.PubKey()),
				Address:  core.ToHex(validator.Address()),
				NodeInfo: sharding.NodeInfo{},
			}
			initialNodes = append(initialNodes, n)
		}
	}

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes:     initialNodes,
		PeerAdapter:      tpn.PeerState,
		AdrConv:          TestAddressConverter,
		NodesCoordinator: tpn.NodesCoordinator,
		ShardCoordinator: tpn.ShardCoordinator,
		DataPool:         peerDataPool,
		StorageService:   tpn.Storage,
		Marshalizer:      TestMarshalizer,
		StakeValue:       tpn.EconomicsData.StakeValue(),
		Rater:            tpn.Rater,
	}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	tpn.ValidatorStatisticsProcessor = validatorStatistics

	argumentsBase := block.ArgBaseProcessor{
		Accounts:                     tpn.AccntState,
		ForkDetector:                 tpn.ForkDetector,
		Hasher:                       TestHasher,
		Marshalizer:                  TestMarshalizer,
		Store:                        tpn.Storage,
		ShardCoordinator:             tpn.ShardCoordinator,
		NodesCoordinator:             tpn.NodesCoordinator,
		SpecialAddressHandler:        tpn.SpecialAddressHandler,
		Uint64Converter:              TestUint64Converter,
		StartHeaders:                 tpn.GenesisBlocks,
		RequestHandler:               tpn.RequestHandler,
		Core:                         nil,
		BlockChainHook:               tpn.BlockChainHookImpl,
		ValidatorStatisticsProcessor: tpn.ValidatorStatisticsProcessor,
		Rounder:                      &mock.RounderMock{},
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		argumentsBase.TxCoordinator = tpn.TxCoordinator

		argsStakingToPeer := scToProtocol2.ArgStakingToPeer{
			AdrConv:      TestAddressConverter,
			Hasher:       TestHasher,
			Marshalizer:  TestMarshalizer,
			PeerState:    tpn.PeerState,
			BaseState:    tpn.AccntState,
			ArgParser:    tpn.ArgsParser,
			CurrTxs:      tpn.MetaDataPool.CurrentBlockTxs(),
			ScDataGetter: tpn.ScDataGetter,
		}
		scToProtocol, _ := scToProtocol2.NewStakingToPeer(argsStakingToPeer)
		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:   argumentsBase,
			DataPool:           tpn.MetaDataPool,
			SCDataGetter:       tpn.ScDataGetter,
			SCToProtocol:       scToProtocol,
			PeerChangesHandler: scToProtocol,
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)
	} else {
		argumentsBase.BlockChainHook = tpn.BlockChainHookImpl
		argumentsBase.TxCoordinator = tpn.TxCoordinator
		arguments := block.ArgShardProcessor{
			ArgBaseProcessor: argumentsBase,
			DataPool:         tpn.ShardDataPool,
			TxsPoolsCleaner:  &mock.TxPoolsCleanerMock{},
		}

		tpn.BlockProcessor, err = block.NewShardProcessor(arguments)
	}

	if err != nil {
		fmt.Printf("Error creating blockprocessor: %s\n", err.Error())
	}
}

// NewTestProcessorNodeWithCustomNodesCoordinator returns a new TestProcessorNode instance with custom NodesCoordinator
func NewTestProcessorNodeWithCustomNodesCoordinatorAndRater(
	maxShards uint32,
	nodeShardId uint32,
	initialNodeAddr string,
	nodesCoordinator sharding.NodesCoordinator,
	cp *CryptoParams,
	keyIndex int,
	rater sharding.Rater,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	tpn := &TestProcessorNodeWithRater{
		TestProcessorNode{
			ShardCoordinator: shardCoordinator,
			Messenger:        messenger,
			NodesCoordinator: nodesCoordinator,
		},
	}

	tpn.BlockProcessorInitializer = &blockProcessorInitializer{InitBlockProcessorCalled: tpn.initBlockProcessor}

	tpn.NodeKeys = cp.Keys[nodeShardId][keyIndex]
	tpn.Rater = rater
	llsig := &kmultisig.KyberMultiSignerBLS{}
	blsHasher := blake2b.Blake2b{HashSize: factory.BlsHashSize}

	pubKeysMap := PubKeysMapFromKeysMap(cp.Keys)

	tpn.MultiSigner, _ = multisig.NewBLSMultisig(
		llsig,
		blsHasher,
		pubKeysMap[nodeShardId],
		tpn.NodeKeys.Sk,
		cp.KeyGen,
		0,
	)
	if tpn.MultiSigner == nil {
		fmt.Println("Error generating multisigner")
	}
	accountShardId := nodeShardId
	if nodeShardId == sharding.MetachainShardId {
		accountShardId = 0
	}

	tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, accountShardId)
	tpn.initDataPools()
	tpn.initTestNode()

	return &tpn.TestProcessorNode
}
