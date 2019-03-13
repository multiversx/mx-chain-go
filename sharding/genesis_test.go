package sharding_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func createGenesisOneShardOneNode() sharding.Genesis {
	g := sharding.Genesis{}
	g.ConsensusGroupSize = 1
	g.MinNodesPerShard = 1
	g.InitialNodes = make([]*sharding.InitialNode, 1)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[0].Address = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[0].Balance = "11"

	g.ProcessConfig()
	g.ProcessShardAssigment()
	g.CreateInitialNodesAddress()
	g.CreateInitialNodesPubKeys()

	return g
}

func createGenesisTwoShardTwoNodes() sharding.Genesis {
	g := sharding.Genesis{}
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

	g.InitialNodes[0].Address = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[1].Address = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	g.InitialNodes[2].Address = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	g.InitialNodes[3].Address = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"

	g.InitialNodes[0].Balance = "999"
	g.InitialNodes[1].Balance = "999"
	g.InitialNodes[2].Balance = "999"
	g.InitialNodes[3].Balance = "999"

	g.ProcessConfig()
	g.ProcessShardAssigment()
	g.CreateInitialNodesAddress()
	g.CreateInitialNodesPubKeys()

	return g
}

func TestGenesis_NewGenesisConfigWrongFile(t *testing.T) {
	genesis, err := sharding.NewGenesisConfig("")

	assert.Nil(t, genesis)
	assert.NotNil(t, err)
}

func TestGenesis_InitialNodesAddressFromNil(t *testing.T) {
	genesis := sharding.Genesis{}
	inAddresses := genesis.InitialNodesAddress()

	assert.NotNil(t, genesis)
	assert.Nil(t, inAddresses)
}

func TestGenesis_InitialNodesAddress(t *testing.T) {
	genesis := createGenesisOneShardOneNode()
	inAddresses := genesis.InitialNodesAddress()

	assert.NotNil(t, genesis)
	assert.NotNil(t, inAddresses)
}

func TestGenesis_InitialNodesPubKeysFromNil(t *testing.T) {
	genesis := sharding.Genesis{}
	inPubKeys := genesis.InitialNodesPubKeys()

	assert.NotNil(t, genesis)
	assert.Nil(t, inPubKeys)
}

func TestGenesis_InitialNodesPubKeys(t *testing.T) {
	genesis := createGenesisOneShardOneNode()
	inPubKeys := genesis.InitialNodesAddress()

	assert.NotNil(t, genesis)
	assert.Equal(t, len(inPubKeys), 1)
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

func TestGenesis_InitialNodesAddressForShardNil(t *testing.T) {
	g := sharding.Genesis{}
	inAD, err := g.InitialNodesAddressForShard(0)

	assert.NotNil(t, g)
	assert.Nil(t, inAD)
	assert.NotNil(t, err)
}

func TestGenesis_InitialNodesAddressForShardWrongShard(t *testing.T) {
	g := createGenesisOneShardOneNode()
	inAD, err := g.InitialNodesAddressForShard(1)

	assert.NotNil(t, g)
	assert.Nil(t, inAD)
	assert.NotNil(t, err)
}

func TestGenesis_InitialNodesAddressForShardGood(t *testing.T) {
	g := createGenesisTwoShardTwoNodes()
	inAD, err := g.InitialNodesAddressForShard(1)

	assert.NotNil(t, g)
	assert.Equal(t, 2, len(inAD))
	assert.Nil(t, err)
}
