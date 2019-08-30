package sharding_test

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func createNodesSetupOneShardOneNodeWithOneMeta() *sharding.NodesSetup {
	ns := &sharding.NodesSetup{}
	ns.ConsensusGroupSize = 1
	ns.MinNodesPerShard = 1
	ns.MetaChainConsensusGroupSize = 1
	ns.MetaChainMinNodes = 1
	ns.InitialNodes = make([]*sharding.InitialNode, 2)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	ns.InitialNodes[1] = &sharding.InitialNode{}
	ns.InitialNodes[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := ns.ProcessConfig()
	if err != nil {
		return nil
	}

	ns.ProcessMetaChainAssigment()
	ns.ProcessShardAssignment()
	ns.CreateInitialNodesPubKeys()

	return ns
}

func createNodesSetupTwoShardTwoNodesWithOneMeta() *sharding.NodesSetup {
	ns := &sharding.NodesSetup{}
	ns.ConsensusGroupSize = 1
	ns.MinNodesPerShard = 2
	ns.MetaChainConsensusGroupSize = 1
	ns.MetaChainMinNodes = 1
	ns.InitialNodes = make([]*sharding.InitialNode, 5)
	ns.InitialNodes[0] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417",
	}
	ns.InitialNodes[1] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419",
	}
	ns.InitialNodes[2] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418",
	}
	ns.InitialNodes[3] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417",
	}
	ns.InitialNodes[4] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416",
	}

	err := ns.ProcessConfig()
	if err != nil {
		return nil
	}

	ns.ProcessMetaChainAssigment()
	ns.ProcessShardAssignment()
	ns.CreateInitialNodesPubKeys()

	return ns
}

func createNodesSetupTwoShard5NodesWithMeta() *sharding.NodesSetup {
	ns := &sharding.NodesSetup{}
	ns.ConsensusGroupSize = 1
	ns.MinNodesPerShard = 2
	ns.MetaChainMinNodes = 1
	ns.MetaChainConsensusGroupSize = 1
	ns.InitialNodes = make([]*sharding.InitialNode, 6)
	ns.InitialNodes[0] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7410",
	}
	ns.InitialNodes[1] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419",
	}
	ns.InitialNodes[2] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418",
	}
	ns.InitialNodes[3] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417",
	}
	ns.InitialNodes[4] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416",
	}
	ns.InitialNodes[5] = &sharding.InitialNode{
		PubKey: "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7411",
	}

	err := ns.ProcessConfig()
	if err != nil {
		return nil
	}

	ns.ProcessMetaChainAssigment()
	ns.ProcessShardAssignment()
	ns.CreateInitialNodesPubKeys()

	return ns
}

func createNodesSetupTwoShard6NodesMeta() *sharding.NodesSetup {
	ns := &sharding.NodesSetup{}
	ns.ConsensusGroupSize = 1
	ns.MinNodesPerShard = 2
	ns.MetaChainMinNodes = 2
	ns.MetaChainConsensusGroupSize = 2
	ns.InitialNodes = make([]*sharding.InitialNode, 6)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}
	ns.InitialNodes[2] = &sharding.InitialNode{}
	ns.InitialNodes[3] = &sharding.InitialNode{}
	ns.InitialNodes[4] = &sharding.InitialNode{}
	ns.InitialNodes[5] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	ns.InitialNodes[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	ns.InitialNodes[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	ns.InitialNodes[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"
	ns.InitialNodes[4].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7411"
	ns.InitialNodes[5].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7410"

	err := ns.ProcessConfig()
	if err != nil {
		return nil
	}

	ns.ProcessMetaChainAssigment()
	ns.ProcessShardAssignment()
	ns.CreateInitialNodesPubKeys()

	return ns
}

func TestNodesSetup_NewNodesSetupWrongFile(t *testing.T) {
	ns, err := sharding.NewNodesSetup("", 0xFFFFFFFFFFFFFFFF)

	assert.Nil(t, ns)
	assert.NotNil(t, err)
}

func TestNodesSetup_NewNodesSetupWrongDataInFile(t *testing.T) {
	ns, err := sharding.NewNodesSetup("mock/invalidNodesSetupMock.json", 0xFFFFFFFFFFFFFFFF)

	assert.Nil(t, ns)
	assert.Equal(t, sharding.ErrNegativeOrZeroConsensusGroupSize, err)
}

func TestNodesSetup_NewNodesShouldWork(t *testing.T) {
	ns, err := sharding.NewNodesSetup("mock/nodesSetupMock.json", 0xFFFFFFFFFFFFFFFF)

	assert.NotNil(t, ns)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(ns.InitialNodes))
}

func TestNodesSetup_NewNodesShouldTrimInitialNodesList(t *testing.T) {
	ns, err := sharding.NewNodesSetup("mock/nodesSetupMock.json", 2)

	assert.NotNil(t, ns)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ns.InitialNodes))
}

func TestNodesSetup_InitialNodesPubKeysFromNil(t *testing.T) {
	ns := sharding.NodesSetup{}
	inPubKeys := ns.InitialNodesPubKeys()

	assert.NotNil(t, ns)
	assert.Nil(t, inPubKeys)
}

func TestNodesSetup_ProcessConfigNodesWithIncompleteDataShouldErr(t *testing.T) {
	ns := sharding.NodesSetup{}

	ns.InitialNodes = make([]*sharding.InitialNode, 2)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrCouldNotParsePubKey, err)
}

func TestNodesSetup_ProcessConfigInvalidConsensusGroupSizeShouldErr(t *testing.T) {
	ns := sharding.NodesSetup{
		ConsensusGroupSize: 0,
		MinNodesPerShard:   0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, 2)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	ns.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrNegativeOrZeroConsensusGroupSize, err)
}

func TestNodesSetup_ProcessConfigInvalidMetaConsensusGroupSizeShouldErr(t *testing.T) {
	ns := sharding.NodesSetup{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 0,
		MetaChainMinNodes:           0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, 2)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	ns.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrNegativeOrZeroConsensusGroupSize, err)
}

func TestNodesSetup_ProcessConfigInvalidConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	ns := sharding.NodesSetup{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, 2)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	ns.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodesSetup_ProcessConfigInvalidMetaConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	ns := sharding.NodesSetup{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 1,
		MetaChainMinNodes:           0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, 2)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	ns.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodesSetup_ProcessConfigInvalidMinNodesPerShardShouldErr(t *testing.T) {
	ns := sharding.NodesSetup{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, 2)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	ns.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodesSetup_ProcessConfigInvalidMetaMinNodesPerShardShouldErr(t *testing.T) {
	ns := sharding.NodesSetup{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 1,
		MetaChainMinNodes:           0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, 2)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	ns.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodesSetup_ProcessConfigInvalidNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	ns := sharding.NodesSetup{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   3,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, 2)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	ns.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestNodesSetup_ProcessConfigInvalidMetaNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	ns := sharding.NodesSetup{
		ConsensusGroupSize: 1,
		MinNodesPerShard:   1,

		MetaChainConsensusGroupSize: 2,
		MetaChainMinNodes:           3,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, 3)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}
	ns.InitialNodes[2] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	ns.InitialNodes[1].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	ns.InitialNodes[2].PubKey = "3336b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardNil(t *testing.T) {
	ns := sharding.NodesSetup{}
	inPK, err := ns.InitialNodesPubKeysForShard(0)

	assert.NotNil(t, ns)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardWrongShard(t *testing.T) {
	ns := createNodesSetupOneShardOneNodeWithOneMeta()
	inPK, err := ns.InitialNodesPubKeysForShard(1)

	assert.NotNil(t, ns)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardGood(t *testing.T) {
	ns := createNodesSetupTwoShardTwoNodesWithOneMeta()
	inPK, err := ns.InitialNodesPubKeysForShard(1)

	assert.NotNil(t, ns)
	assert.Equal(t, len(inPK), 2)
	assert.Nil(t, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardGoodMeta(t *testing.T) {
	ns := createNodesSetupTwoShard6NodesMeta()
	metaId := sharding.MetachainShardId
	inPK, err := ns.InitialNodesPubKeysForShard(metaId)

	assert.NotNil(t, ns)
	assert.Equal(t, len(inPK), 2)
	assert.Nil(t, err)
}

func TestNodesSetup_PublicKeyNotGood(t *testing.T) {
	ns := createNodesSetupTwoShard6NodesMeta()

	_, err := ns.GetShardIDForPubKey([]byte("5126b6505a73e59a994caa8f956f8c335d4399229de42102bb4814ca261c7419"))

	assert.NotNil(t, ns)
	assert.NotNil(t, err)
}

func TestNodesSetup_PublicKeyGood(t *testing.T) {
	ns := createNodesSetupTwoShard5NodesWithMeta()
	publicKey, err := hex.DecodeString("5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417")

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	assert.NotNil(t, ns)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), selfId)
}

func TestNodesSetup_ShardPublicKeyGoodMeta(t *testing.T) {
	ns := createNodesSetupTwoShard6NodesMeta()
	publicKey, err := hex.DecodeString("5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417")

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	assert.NotNil(t, ns)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), selfId)
}

func TestNodesSetup_MetaPublicKeyGoodMeta(t *testing.T) {
	ns := createNodesSetupTwoShard6NodesMeta()
	metaId := sharding.MetachainShardId
	publicKey, err := hex.DecodeString("5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418")

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	assert.NotNil(t, ns)
	assert.Nil(t, err)
	assert.Equal(t, metaId, selfId)
}
