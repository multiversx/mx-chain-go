package sharding_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
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

func TestGenesis_GenesisWithUncomlpeteData(t *testing.T) {
	g := sharding.Genesis{}

	g.InitialNodes = make([]*sharding.InitialNode, 2)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := g.ProcessConfig()

	assert.NotNil(t, g)
	assert.NotNil(t, err)
}

func TestGenesis_GenesisWithUncomlpeteBalance(t *testing.T) {
	g := sharding.Genesis{}

	g.InitialNodes = make([]*sharding.InitialNode, 1)
	g.InitialNodes[0] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := g.ProcessConfig()
	g.ProcessShardAssignment()
	g.CreateInitialNodesPubKeys()

	inBal, err := g.InitialNodesBalances(0)

	assert.NotNil(t, g)
	assert.Nil(t, err)
	for _, val := range inBal {
		assert.Equal(t, big.NewInt(0), val)
	}
}

func TestGenesis_InitialNodesBalancesNil(t *testing.T) {
	genesis := sharding.Genesis{}
	inBalance, err := genesis.InitialNodesBalances(0)

	assert.NotNil(t, genesis)
	assert.Nil(t, inBalance)
	assert.NotNil(t, err)
}

func TestGenesis_InitialNodesBalancesWrongShard(t *testing.T) {
	genesis := createGenesisOneShardOneNode()
	inBalance, err := genesis.InitialNodesBalances(1)

	assert.NotNil(t, genesis)
	assert.Nil(t, inBalance)
	assert.NotNil(t, err)
}

func TestGenesis_InitialNodesBalancesGood(t *testing.T) {
	genesis := createGenesisTwoShardTwoNodes()
	inBalance, err := genesis.InitialNodesBalances(1)

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
	inBalance, err := genesis.InitialNodesBalances(1)

	assert.NotNil(t, genesis)
	assert.Equal(t, 2, len(inBalance))
	assert.Nil(t, err)
}
