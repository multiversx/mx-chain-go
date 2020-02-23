package integrationTests

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NewTestProcessorNodeWithStateCheckpointModulus creates a new testNodeProcessor with custom state checkpoint modulus
func NewTestProcessorNodeWithStateCheckpointModulus(
	maxShards uint32,
	nodeShardId uint32,
	txSignPrivKeyShardId uint32,
	initialNodeAddr string,
	stateCheckpointModulus uint,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	kg := &mock.KeyGenMock{}
	sk, pk := kg.GeneratePair()

	pkBytes := make([]byte, 128)
	address := make([]byte, 32)

	pkBytes[0] = 1
	address[0] = 1

	nodesCoordinator := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []sharding.Validator, err error) {
			v, _ := sharding.NewValidator(pkBytes, address)
			return []sharding.Validator{v}, nil
		},
		GetAllEligibleValidatorsPublicKeysCalled: func(shardId uint32) (m map[uint32][][]byte, err error) {
			ma := map[uint32][][]byte{
				0: [][]byte{pkBytes},
			}
			return ma, nil
		},
	}

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	tpn := &TestProcessorNode{
		ShardCoordinator:  shardCoordinator,
		Messenger:         messenger,
		NodesCoordinator:  nodesCoordinator,
		HeaderSigVerifier: &mock.HeaderSigVerifierStub{},
		ChainID:           ChainID,
	}

	tpn.NodeKeys = &TestKeyPair{
		Sk: sk,
		Pk: pk,
	}
	tpn.MultiSigner = TestMultiSig
	tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, txSignPrivKeyShardId)
	tpn.initDataPools()
	tpn.SpecialAddressHandler = mock.NewSpecialAddressHandlerMock(
		TestAddressConverter,
		tpn.ShardCoordinator,
		tpn.NodesCoordinator,
	)
	tpn.initHeaderValidator()
	tpn.initRounder()
	tpn.initStorage()
	tpn.initAccountDBs()
	tpn.initChainHandler()
	tpn.initEconomicsData()
	tpn.initRequestedItemsHandler()
	tpn.initResolvers()
	tpn.initValidatorStatistics()
	rootHash, _ := tpn.ValidatorStatisticsProcessor.RootHash()
	tpn.GenesisBlocks = CreateGenesisBlocks(
		tpn.AccntState,
		TestAddressConverter,
		&sharding.NodesSetup{},
		tpn.ShardCoordinator,
		tpn.Storage,
		tpn.BlockChain,
		TestMarshalizer,
		TestHasher,
		TestUint64Converter,
		tpn.DataPool,
		tpn.EconomicsData.EconomicsData,
		rootHash,
	)
	tpn.initBlockTracker()
	tpn.initInterceptors()
	tpn.initInnerProcessors()
	tpn.SCQueryService, _ = smartContract.NewSCQueryService(tpn.VMContainer, tpn.EconomicsData.MaxGasLimitPerBlock())
	tpn.initBlockProcessor(stateCheckpointModulus)
	tpn.BroadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
		TestMarshalizer,
		tpn.Messenger,
		tpn.ShardCoordinator,
		tpn.OwnAccount.SkTxSign,
		tpn.OwnAccount.SingleSigner,
	)
	tpn.setGenesisBlock()
	tpn.initNode()
	tpn.SCQueryService, _ = smartContract.NewSCQueryService(tpn.VMContainer, tpn.EconomicsData.MaxGasLimitPerBlock())
	tpn.addHandlersForCounters()
	tpn.addGenesisBlocksIntoStorage()

	return tpn
}
