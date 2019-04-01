package sharding_test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding/mock"
	"github.com/stretchr/testify/assert"
)

func createGenesisOneShardOneNode() *sharding.Genesis {
	g := &sharding.Genesis{}
	g.ConsensusGroupSize = 1
	g.MinNodesPerShard = 1
	g.InitialNodes = make([]*sharding.InitialNode, 1)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[0].Balance = "11"

	err := g.ProcessConfig()
	if err != nil {
		return nil
	}
	g.ProcessShardAssignment()
	g.CreateInitialNodesPubKeys()

	return g
}

func createGenesisTwoShardTwoNodes() *sharding.Genesis {
	g := &sharding.Genesis{}
	g.ConsensusGroupSize = 1
	g.MinNodesPerShard = 2
	g.InitialNodes = make([]*sharding.InitialNode, 4)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}
	g.InitialNodes[2] = &sharding.InitialNode{}
	g.InitialNodes[3] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	g.InitialNodes[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	g.InitialNodes[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"

	g.InitialNodes[0].Balance = "999"
	g.InitialNodes[1].Balance = "999"
	g.InitialNodes[2].Balance = "999"
	g.InitialNodes[3].Balance = "999"

	err := g.ProcessConfig()
	if err != nil {
		return nil
	}
	g.ProcessShardAssignment()
	g.CreateInitialNodesPubKeys()

	return g
}

func createGenesisTwoShard5Nodes() *sharding.Genesis {
	g := &sharding.Genesis{}
	g.ConsensusGroupSize = 1
	g.MinNodesPerShard = 2
	g.InitialNodes = make([]*sharding.InitialNode, 5)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}
	g.InitialNodes[2] = &sharding.InitialNode{}
	g.InitialNodes[3] = &sharding.InitialNode{}
	g.InitialNodes[4] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	g.InitialNodes[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	g.InitialNodes[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"
	g.InitialNodes[4].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7411"

	g.InitialNodes[0].Balance = "999"
	g.InitialNodes[1].Balance = "999"
	g.InitialNodes[2].Balance = "999"
	g.InitialNodes[3].Balance = "999"
	g.InitialNodes[4].Balance = "999"

	err := g.ProcessConfig()
	if err != nil {
		return nil
	}
	g.ProcessShardAssignment()
	g.CreateInitialNodesPubKeys()

	return g
}

func TestGenesis_NewGenesisConfigWrongFile(t *testing.T) {
	genesis, err := sharding.NewGenesisConfig("")

	assert.Nil(t, genesis)
	assert.NotNil(t, err)
}

func TestGenesis_InitialNodesPubKeysFromNil(t *testing.T) {
	genesis := sharding.Genesis{}
	inPubKeys := genesis.InitialNodesPubKeys()

	assert.NotNil(t, genesis)
	assert.Nil(t, inPubKeys)
}

func TestGenesis_ProcessConfigGenesisWithIncompleteDataShouldErr(t *testing.T) {
	g := sharding.Genesis{}

	g.InitialNodes = make([]*sharding.InitialNode, 2)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := g.ProcessConfig()

	assert.NotNil(t, g)
	assert.Equal(t, sharding.ErrCouldNotParsePubKey, err)
}

func TestGenesis_ProcessConfigInvalidConsensusGroupSizeShouldErr(t *testing.T) {
	g := sharding.Genesis{
		ConsensusGroupSize: 0,
		MinNodesPerShard:   0,
	}

	g.InitialNodes = make([]*sharding.InitialNode, 2)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := g.ProcessConfig()

	assert.NotNil(t, g)
	assert.Equal(t, sharding.ErrNegativeOrZeroConsensusGroupSize, err)
}

func TestGenesis_ProcessConfigInvalidConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	g := sharding.Genesis{
		ConsensusGroupSize: 3,
		MinNodesPerShard:   0,
	}

	g.InitialNodes = make([]*sharding.InitialNode, 2)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := g.ProcessConfig()

	assert.NotNil(t, g)
	assert.Equal(t, sharding.ErrNotEnoughValidators, err)
}

func TestGenesis_ProcessConfigInvalidMinNodesPerShardShouldErr(t *testing.T) {
	g := sharding.Genesis{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   0,
	}

	g.InitialNodes = make([]*sharding.InitialNode, 2)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := g.ProcessConfig()

	assert.NotNil(t, g)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestGenesis_ProcessConfigInvalidNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	g := sharding.Genesis{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   3,
	}

	g.InitialNodes = make([]*sharding.InitialNode, 2)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := g.ProcessConfig()

	assert.NotNil(t, g)
	assert.Equal(t, sharding.ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestGenesis_GenesisWithIncompleteBalance(t *testing.T) {
	g := sharding.Genesis{
		ConsensusGroupSize: 1,
		MinNodesPerShard:   1,
	}

	g.InitialNodes = make([]*sharding.InitialNode, 1)
	g.InitialNodes[0] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	_ = g.ProcessConfig()
	g.ProcessShardAssignment()
	g.CreateInitialNodesPubKeys()

	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(1, 0)
	adrConv := mock.NewAddressConverterFake(32, "")

	inBal, err := g.InitialNodesBalances(shardCoordinator, adrConv)

	assert.NotNil(t, g)
	assert.Nil(t, err)
	for _, val := range inBal {
		assert.Equal(t, big.NewInt(0), val)
	}
}

func TestGenesis_InitialNodesBalancesNil(t *testing.T) {
	genesis := sharding.Genesis{}
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(1, 0)
	adrConv := mock.NewAddressConverterFake(32, "")
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator, adrConv)

	assert.NotNil(t, genesis)
	assert.Equal(t, 0, len(inBalance))
	assert.Nil(t, err)
}

func TestGenesis_InitialNodesBalancesNilShardCoordinatorShouldErr(t *testing.T) {
	genesis := createGenesisOneShardOneNode()
	adrConv := mock.NewAddressConverterFake(32, "")
	inBalance, err := genesis.InitialNodesBalances(nil, adrConv)

	assert.NotNil(t, genesis)
	assert.Nil(t, inBalance)
	assert.Equal(t, sharding.ErrNilShardCoordinator, err)
}

func TestGenesis_InitialNodesBalancesNilAddrConverterShouldErr(t *testing.T) {
	genesis := createGenesisOneShardOneNode()
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(1, 0)
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator, nil)

	assert.NotNil(t, genesis)
	assert.Nil(t, inBalance)
	assert.Equal(t, sharding.ErrNilAddressConverter, err)
}

func TestGenesis_InitialNodesBalancesGood(t *testing.T) {
	genesis := createGenesisTwoShardTwoNodes()
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(2, 1)
	adrConv := mock.NewAddressConverterFake(32, "")
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator, adrConv)

	assert.NotNil(t, genesis)
	assert.Equal(t, 2, len(inBalance))
	assert.Nil(t, err)
}

func TestGenesis_InitialNodesPubKeysForShardNil(t *testing.T) {
	g := sharding.Genesis{}
	inPK, err := g.InitialNodesPubKeysForShard(0)

	assert.NotNil(t, g)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestGenesis_InitialNodesPubKeysForShardWrongShard(t *testing.T) {
	g := createGenesisOneShardOneNode()
	inPK, err := g.InitialNodesPubKeysForShard(1)

	assert.NotNil(t, g)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestGenesis_InitialNodesPubKeysForShardGood(t *testing.T) {
	g := createGenesisTwoShardTwoNodes()
	inPK, err := g.InitialNodesPubKeysForShard(1)

	assert.NotNil(t, g)
	assert.Equal(t, len(inPK), 2)
	assert.Nil(t, err)
}

func TestGenesis_Initial5NodesBalancesGood(t *testing.T) {
	genesis := createGenesisTwoShard5Nodes()
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(2, 1)
	adrConv := mock.NewAddressConverterFake(32, "")
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator, adrConv)

	assert.NotNil(t, genesis)
	assert.Equal(t, 3, len(inBalance))
	assert.Nil(t, err)
}

func TestGenesis_PublicKeyNotGood(t *testing.T) {
	genesis := createGenesisTwoShard5Nodes()

	_, err := genesis.GetShardIDFromPubKey([]byte("5126b6505a73e59a994caa8f956f8c335d4399229de42102bb4814ca261c7419"))

	assert.NotNil(t, genesis)
	assert.NotNil(t, err)
}

func TestGenesis_PublicKeyGood(t *testing.T) {
	genesis := createGenesisTwoShard5Nodes()
	publicKey, err := hex.DecodeString("5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417")

	selfId, err := genesis.GetShardIDFromPubKey(publicKey)

	assert.NotNil(t, genesis)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), selfId)
}
