package sharding_test

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func createNodesOneShardOneNode() *sharding.Nodes {
	g := &sharding.Nodes{}
	g.ConsensusGroupSize = 1
	g.MinNodesPerShard = 1
	g.InitialNodes = make([]*sharding.InitialNode, 1)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := g.ProcessConfig()
	if err != nil {
		return nil
	}

	if g.MetaChainActive {
		g.ProcessMetaChainAssigment()
	}

	g.ProcessShardAssignment()
	g.CreateInitialNodesPubKeys()

	return g
}

func createNodesTwoShardTwoNodes() *sharding.Nodes {
	g := &sharding.Nodes{}
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

	err := g.ProcessConfig()
	if err != nil {
		return nil
	}

	if g.MetaChainActive {
		g.ProcessMetaChainAssigment()
	}

	g.ProcessShardAssignment()
	g.CreateInitialNodesPubKeys()

	return g
}

func createNodesTwoShard5Nodes() *sharding.Nodes {
	g := &sharding.Nodes{}
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

	err := g.ProcessConfig()
	if err != nil {
		return nil
	}

	if g.MetaChainActive {
		g.ProcessMetaChainAssigment()
	}

	g.ProcessShardAssignment()
	g.CreateInitialNodesPubKeys()

	return g
}

func createNodesTwoShard6NodesMeta() *sharding.Nodes {
	g := &sharding.Nodes{}
	g.ConsensusGroupSize = 1
	g.MinNodesPerShard = 2
	g.MetaChainActive = true
	g.MetaChainMinNodes = 2
	g.MetaChainConsensusGroupSize = 2
	g.InitialNodes = make([]*sharding.InitialNode, 6)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}
	g.InitialNodes[2] = &sharding.InitialNode{}
	g.InitialNodes[3] = &sharding.InitialNode{}
	g.InitialNodes[4] = &sharding.InitialNode{}
	g.InitialNodes[5] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	g.InitialNodes[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	g.InitialNodes[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"
	g.InitialNodes[4].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7411"
	g.InitialNodes[5].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7410"

	err := g.ProcessConfig()
	if err != nil {
		return nil
	}

	if g.MetaChainActive {
		g.ProcessMetaChainAssigment()
	}

	g.ProcessShardAssignment()
	g.CreateInitialNodesPubKeys()

	return g
}

func TestNodes_NewNodesConfigWrongFile(t *testing.T) {
	nodes, err := sharding.NewNodesConfig("")

	assert.Nil(t, nodes)
	assert.NotNil(t, err)
}

func TestNodes_InitialNodesPubKeysFromNil(t *testing.T) {
	nodes := sharding.Nodes{}
	inPubKeys := nodes.InitialNodesPubKeys()

	assert.NotNil(t, nodes)
	assert.Nil(t, inPubKeys)
}

func TestNodes_ProcessConfigNodesWithIncompleteDataShouldErr(t *testing.T) {
	g := sharding.Nodes{}

	g.InitialNodes = make([]*sharding.InitialNode, 2)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := g.ProcessConfig()

	assert.NotNil(t, g)
	assert.Equal(t, sharding.ErrCouldNotParsePubKey, err)
}

func TestNodes_ProcessConfigInvalidConsensusGroupSizeShouldErr(t *testing.T) {
	g := sharding.Nodes{
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

func TestNodes_ProcessConfigInvalidMetaConsensusGroupSizeShouldErr(t *testing.T) {
	g := sharding.Nodes{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 0,
		MetaChainMinNodes:           0,
		MetaChainActive:             true,
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

func TestNodes_ProcessConfigInvalidConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	g := sharding.Nodes{
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

func TestNodes_ProcessConfigInvalidMetaConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	g := sharding.Nodes{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 1,
		MetaChainMinNodes:           0,
		MetaChainActive:             true,
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

func TestNodes_ProcessConfigInvalidMinNodesPerShardShouldErr(t *testing.T) {
	g := sharding.Nodes{
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

func TestNodes_ProcessConfigInvalidMetaMinNodesPerShardShouldErr(t *testing.T) {
	g := sharding.Nodes{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 1,
		MetaChainMinNodes:           0,
		MetaChainActive:             true,
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

func TestNodes_ProcessConfigInvalidNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	g := sharding.Nodes{
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

func TestNodes_ProcessConfigInvalidMetaNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	g := sharding.Nodes{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainActive:             true,
		MetaChainConsensusGroupSize: 2,
		MetaChainMinNodes:           3,
	}

	g.InitialNodes = make([]*sharding.InitialNode, 3)
	g.InitialNodes[0] = &sharding.InitialNode{}
	g.InitialNodes[1] = &sharding.InitialNode{}
	g.InitialNodes[2] = &sharding.InitialNode{}

	g.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	g.InitialNodes[2].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"

	err := g.ProcessConfig()

	assert.NotNil(t, g)
	assert.Equal(t, sharding.ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestNodes_InitialNodesPubKeysForShardNil(t *testing.T) {
	g := sharding.Nodes{}
	inPK, err := g.InitialNodesPubKeysForShard(0)

	assert.NotNil(t, g)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestNodes_InitialNodesPubKeysForShardWrongShard(t *testing.T) {
	g := createNodesOneShardOneNode()
	inPK, err := g.InitialNodesPubKeysForShard(1)

	assert.NotNil(t, g)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestNodes_InitialNodesPubKeysForShardGood(t *testing.T) {
	g := createNodesTwoShardTwoNodes()
	inPK, err := g.InitialNodesPubKeysForShard(1)

	assert.NotNil(t, g)
	assert.Equal(t, len(inPK), 2)
	assert.Nil(t, err)
}

func TestNodes_InitialNodesPubKeysForShardWrongMeta(t *testing.T) {
	g := createNodesTwoShardTwoNodes()
	metaId := sharding.MetachainShardId
	inPK, err := g.InitialNodesPubKeysForShard(metaId)

	assert.NotNil(t, g)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestNodes_InitialNodesPubKeysForShardGoodMeta(t *testing.T) {
	g := createNodesTwoShard6NodesMeta()
	metaId := sharding.MetachainShardId
	inPK, err := g.InitialNodesPubKeysForShard(metaId)

	assert.NotNil(t, g)
	assert.Equal(t, len(inPK), 2)
	assert.Nil(t, err)
}

func TestNodes_PublicKeyNotGood(t *testing.T) {
	nodes := createNodesTwoShard6NodesMeta()

	_, err := nodes.GetShardIDForPubKey([]byte("5126b6505a73e59a994caa8f956f8c335d4399229de42102bb4814ca261c7419"))

	assert.NotNil(t, nodes)
	assert.NotNil(t, err)
}

func TestNodes_PublicKeyGood(t *testing.T) {
	nodes := createNodesTwoShard5Nodes()
	publicKey, err := hex.DecodeString("5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417")

	selfId, err := nodes.GetShardIDForPubKey(publicKey)

	assert.NotNil(t, nodes)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), selfId)
}

func TestNodes_ShardPublicKeyGoodMeta(t *testing.T) {
	nodes := createNodesTwoShard6NodesMeta()
	publicKey, err := hex.DecodeString("5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417")

	selfId, err := nodes.GetShardIDForPubKey(publicKey)

	assert.NotNil(t, nodes)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), selfId)
}

func TestNodes_MetaPublicKeyGoodMeta(t *testing.T) {
	nodes := createNodesTwoShard6NodesMeta()
	metaId := sharding.MetachainShardId
	publicKey, err := hex.DecodeString("5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418")

	selfId, err := nodes.GetShardIDForPubKey(publicKey)

	assert.NotNil(t, nodes)
	assert.Nil(t, err)
	assert.Equal(t, metaId, selfId)
}
