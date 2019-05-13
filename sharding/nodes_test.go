package sharding_test

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func createNodesOneShardOneNode() *sharding.Nodes {
	nodes := &sharding.Nodes{}
	nodes.ConsensusGroupSize = 1
	nodes.MinNodesPerShard = 1
	nodes.InitialNodes = make([]*sharding.InitialNode, 1)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := nodes.ProcessConfig()
	if err != nil {
		return nil
	}

	if nodes.MetaChainActive {
		nodes.ProcessMetaChainAssigment()
	}

	nodes.ProcessShardAssignment()
	nodes.CreateInitialNodesPubKeys()

	return nodes
}

func createNodesTwoShardTwoNodes() *sharding.Nodes {
	nodes := &sharding.Nodes{}
	nodes.ConsensusGroupSize = 1
	nodes.MinNodesPerShard = 2
	nodes.InitialNodes = make([]*sharding.InitialNode, 4)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}
	nodes.InitialNodes[2] = &sharding.InitialNode{}
	nodes.InitialNodes[3] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	nodes.InitialNodes[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	nodes.InitialNodes[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"

	err := nodes.ProcessConfig()
	if err != nil {
		return nil
	}

	if nodes.MetaChainActive {
		nodes.ProcessMetaChainAssigment()
	}

	nodes.ProcessShardAssignment()
	nodes.CreateInitialNodesPubKeys()

	return nodes
}

func createNodesTwoShard5Nodes() *sharding.Nodes {
	nodes := &sharding.Nodes{}
	nodes.ConsensusGroupSize = 1
	nodes.MinNodesPerShard = 2
	nodes.InitialNodes = make([]*sharding.InitialNode, 5)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}
	nodes.InitialNodes[2] = &sharding.InitialNode{}
	nodes.InitialNodes[3] = &sharding.InitialNode{}
	nodes.InitialNodes[4] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	nodes.InitialNodes[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	nodes.InitialNodes[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"
	nodes.InitialNodes[4].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7411"

	err := nodes.ProcessConfig()
	if err != nil {
		return nil
	}

	if nodes.MetaChainActive {
		nodes.ProcessMetaChainAssigment()
	}

	nodes.ProcessShardAssignment()
	nodes.CreateInitialNodesPubKeys()

	return nodes
}

func createNodesTwoShard6NodesMeta() *sharding.Nodes {
	nodes := &sharding.Nodes{}
	nodes.ConsensusGroupSize = 1
	nodes.MinNodesPerShard = 2
	nodes.MetaChainActive = true
	nodes.MetaChainMinNodes = 2
	nodes.MetaChainConsensusGroupSize = 2
	nodes.InitialNodes = make([]*sharding.InitialNode, 6)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}
	nodes.InitialNodes[2] = &sharding.InitialNode{}
	nodes.InitialNodes[3] = &sharding.InitialNode{}
	nodes.InitialNodes[4] = &sharding.InitialNode{}
	nodes.InitialNodes[5] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	nodes.InitialNodes[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	nodes.InitialNodes[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"
	nodes.InitialNodes[4].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7411"
	nodes.InitialNodes[5].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7410"

	err := nodes.ProcessConfig()
	if err != nil {
		return nil
	}

	if nodes.MetaChainActive {
		nodes.ProcessMetaChainAssigment()
	}

	nodes.ProcessShardAssignment()
	nodes.CreateInitialNodesPubKeys()

	return nodes
}

func TestNodes_NewNodesConfigWrongFile(t *testing.T) {
	nodes, err := sharding.NewNodesConfig("")

	assert.Nil(t, nodes)
	assert.NotNil(t, err)
}

func TestNodes_NewNodesConfigWrongDataInFile(t *testing.T) {
	nodes, err := sharding.NewNodesConfig("mock/invalidNodesMock.json")

	assert.Nil(t, nodes)
	assert.Equal(t, sharding.ErrNegativeOrZeroConsensusGroupSize, err)
}

func TestNodes_NewNodesShouldWork(t *testing.T) {
	nodes, err := sharding.NewNodesConfig("mock/nodesMock.json")

	assert.NotNil(t, nodes)
	assert.Nil(t, err)
}

func TestNodes_InitialNodesPubKeysFromNil(t *testing.T) {
	nodes := sharding.Nodes{}
	inPubKeys := nodes.InitialNodesPubKeys()

	assert.NotNil(t, nodes)
	assert.Nil(t, inPubKeys)
}

func TestNodes_ProcessConfigNodesWithIncompleteDataShouldErr(t *testing.T) {
	nodes := sharding.Nodes{}

	nodes.InitialNodes = make([]*sharding.InitialNode, 2)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := nodes.ProcessConfig()

	assert.NotNil(t, nodes)
	assert.Equal(t, sharding.ErrCouldNotParsePubKey, err)
}

func TestNodes_ProcessConfigInvalidConsensusGroupSizeShouldErr(t *testing.T) {
	nodes := sharding.Nodes{
		ConsensusGroupSize: 0,
		MinNodesPerShard:   0,
	}

	nodes.InitialNodes = make([]*sharding.InitialNode, 2)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := nodes.ProcessConfig()

	assert.NotNil(t, nodes)
	assert.Equal(t, sharding.ErrNegativeOrZeroConsensusGroupSize, err)
}

func TestNodes_ProcessConfigInvalidMetaConsensusGroupSizeShouldErr(t *testing.T) {
	nodes := sharding.Nodes{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 0,
		MetaChainMinNodes:           0,
		MetaChainActive:             true,
	}

	nodes.InitialNodes = make([]*sharding.InitialNode, 2)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := nodes.ProcessConfig()

	assert.NotNil(t, nodes)
	assert.Equal(t, sharding.ErrNegativeOrZeroConsensusGroupSize, err)
}

func TestNodes_ProcessConfigInvalidConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	nodes := sharding.Nodes{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   0,
	}

	nodes.InitialNodes = make([]*sharding.InitialNode, 2)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := nodes.ProcessConfig()

	assert.NotNil(t, nodes)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodes_ProcessConfigInvalidMetaConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	nodes := sharding.Nodes{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 1,
		MetaChainMinNodes:           0,
		MetaChainActive:             true,
	}

	nodes.InitialNodes = make([]*sharding.InitialNode, 2)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := nodes.ProcessConfig()

	assert.NotNil(t, nodes)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodes_ProcessConfigInvalidMinNodesPerShardShouldErr(t *testing.T) {
	nodes := sharding.Nodes{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   0,
	}

	nodes.InitialNodes = make([]*sharding.InitialNode, 2)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := nodes.ProcessConfig()

	assert.NotNil(t, nodes)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodes_ProcessConfigInvalidMetaMinNodesPerShardShouldErr(t *testing.T) {
	nodes := sharding.Nodes{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 1,
		MetaChainMinNodes:           0,
		MetaChainActive:             true,
	}

	nodes.InitialNodes = make([]*sharding.InitialNode, 2)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := nodes.ProcessConfig()

	assert.NotNil(t, nodes)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodes_ProcessConfigInvalidNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	nodes := sharding.Nodes{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   3,
	}

	nodes.InitialNodes = make([]*sharding.InitialNode, 2)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := nodes.ProcessConfig()

	assert.NotNil(t, nodes)
	assert.Equal(t, sharding.ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestNodes_ProcessConfigInvalidMetaNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	nodes := sharding.Nodes{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainActive:             true,
		MetaChainConsensusGroupSize: 2,
		MetaChainMinNodes:           3,
	}

	nodes.InitialNodes = make([]*sharding.InitialNode, 3)
	nodes.InitialNodes[0] = &sharding.InitialNode{}
	nodes.InitialNodes[1] = &sharding.InitialNode{}
	nodes.InitialNodes[2] = &sharding.InitialNode{}

	nodes.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	nodes.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	nodes.InitialNodes[2].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"

	err := nodes.ProcessConfig()

	assert.NotNil(t, nodes)
	assert.Equal(t, sharding.ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestNodes_InitialNodesPubKeysForShardNil(t *testing.T) {
	nodes := sharding.Nodes{}
	inPK, err := nodes.InitialNodesPubKeysForShard(0)

	assert.NotNil(t, nodes)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestNodes_InitialNodesPubKeysForShardWrongShard(t *testing.T) {
	nodes := createNodesOneShardOneNode()
	inPK, err := nodes.InitialNodesPubKeysForShard(1)

	assert.NotNil(t, nodes)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestNodes_InitialNodesPubKeysForShardGood(t *testing.T) {
	nodes := createNodesTwoShardTwoNodes()
	inPK, err := nodes.InitialNodesPubKeysForShard(1)

	assert.NotNil(t, nodes)
	assert.Equal(t, len(inPK), 2)
	assert.Nil(t, err)
}

func TestNodes_InitialNodesPubKeysForShardWrongMeta(t *testing.T) {
	nodes := createNodesTwoShardTwoNodes()
	metaId := sharding.MetachainShardId
	inPK, err := nodes.InitialNodesPubKeysForShard(metaId)

	assert.NotNil(t, nodes)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestNodes_InitialNodesPubKeysForShardGoodMeta(t *testing.T) {
	nodes := createNodesTwoShard6NodesMeta()
	metaId := sharding.MetachainShardId
	inPK, err := nodes.InitialNodesPubKeysForShard(metaId)

	assert.NotNil(t, nodes)
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
