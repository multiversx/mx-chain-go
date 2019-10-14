package sharding_test

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

var (
	PubKeys = []string{
		"41378f754e2c7b2745208c3ed21b151d297acdc84c3aca00b9e292cf28ec2d444771070157ea7760ed83c26f4fed387d0077e00b563a95825dac2cbc349fc0025ccf774e37b0a98ad9724d30e90f8c29b4091ccb738ed9ffc0573df776ee9ea30b3c038b55e532760ea4a8f152f2a52848020e5cee1cc537f2c2323399723081",
		"52f3bf5c01771f601ec2137e267319ab6716ef6ff5dfddaea48b42d955f631167f2ce19296a202bb8fd174f4e94f8c85f619df85a7f9f8de0f3768e5e6d8c48187b767deccf9829be246aa331aa86d182eb8fa28ea8a3e45d357ed1647a9be020a5569d686253a6f89e9123c7f21f302e82f67d3e3cd69cf267b9910a663ef32",
		"5e91c426c5c8f5f805f86de1e0653e2ec33853772e583b88e9f0f201089d03d8570759c3c3ab610ce573493c33ba0adf954c8939dba5d5ef7f2be4e87145d8153fc5b4fb91cecb8d9b1f62e080743fbf69c8c3096bf07980bb82cb450ba9b902673373d5b671ea73620cc5bc4d36f7a0f5ca3684d4c8aa5c1b425ab2a8673140",
		"73972bf46dca59fba211c58f11b530f8e9d6392c499655ce760abc6458fd9c6b54b9676ee4b95aa32f6c254c9aad2f63a6195cd65d837a4320d7b8e915ba3a7123c8f4983b201035573c0752bb54e9021eb383b40d302447b62ea7a3790c89c47f5ab81d183f414e87611a31ff635ad22e969495356d5bc44eec7917aaad4c5e",
		"7391ccce066ab5674304b10220643bc64829afa626a165f1e7a6618e260fa68f8e79018ac5964f7a1b8dd419645049042e34ebe7f2772def71e6176ce9daf50a57c17ee2a7445b908fe47e8f978380fcc2654a19925bf73db2402b09dde515148081f8ca7c331fbedec689de1b7bfce6bf106e4433557c29752c12d0a009f47a",
		"24dea9b5c79174c558c38316b2df25b956c53f0d0128b7427d219834867cc1b0868b7faff0205fe23e5ffdf276acfad6423890c782c7be7b98a31d501e4276a015a54d9849109322130fc9a9cb61d183318d50fcde44fabcbf600051c7cb950304b05e82f90f2ac4647016f39439608cd64ccc82fe6e996289bb2150e4e3ab08",
	}

	Address = []string{
		"9e95a4e46da335a96845b4316251fc1bb197e1b8136d96ecc62bf6604eca9e49",
		"7a330039e77ca06bc127319fd707cc4911a80db489a39fcfb746283a05f61836",
		"131e2e717f2d33bdf7850c12b03dfe41ea8a5e76fdd6d4f23aebe558603e746f",
		"4c9e66b605882c1099088f26659692f084e41dc0dedfaedf6a6409af21c02aac",
		"90a66900634b206d20627fbaec432ebfbabeaf30b9e338af63191435e2e37022",
		"63f702e061385324a25dc4f1bcfc7e4f4692bcd80de71bd4dd7d6e2f67f92481",
	}
)

func createNodesSetupOneShardOneNodeWithOneMeta() *sharding.NodesSetup {
	noOfInitialNodes := 2
	ns := &sharding.NodesSetup{}
	ns.ConsensusGroupSize = 1
	ns.MinNodesPerShard = 1
	ns.MetaChainConsensusGroupSize = 1
	ns.MetaChainMinNodes = 1
	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)
	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[0].PubKey = PubKeys[0]
	ns.InitialNodes[0].Address = Address[0]
	ns.InitialNodes[1] = &sharding.InitialNode{}
	ns.InitialNodes[1].PubKey = PubKeys[1]
	ns.InitialNodes[1].Address = Address[1]
	err := ns.ProcessConfig()
	if err != nil {
		return nil
	}

	ns.ProcessMetaChainAssigment()
	ns.ProcessShardAssignment()
	ns.CreateInitialNodesInfo()

	return ns
}

func createNodesSetupTwoShardTwoNodesWithOneMeta() *sharding.NodesSetup {
	noOfInitialNodes := 6
	ns := &sharding.NodesSetup{}
	ns.ConsensusGroupSize = 1
	ns.MinNodesPerShard = 2
	ns.MetaChainConsensusGroupSize = 1
	ns.MetaChainMinNodes = 2
	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()
	if err != nil {
		return nil
	}

	ns.ProcessMetaChainAssigment()
	ns.ProcessShardAssignment()
	ns.CreateInitialNodesInfo()

	return ns
}

func createNodesSetupTwoShard5NodesWithMeta() *sharding.NodesSetup {
	noOfInitialNodes := 5
	ns := &sharding.NodesSetup{}
	ns.ConsensusGroupSize = 1
	ns.MinNodesPerShard = 2
	ns.MetaChainConsensusGroupSize = 1
	ns.MetaChainMinNodes = 1
	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()
	if err != nil {
		return nil
	}

	ns.ProcessMetaChainAssigment()
	ns.ProcessShardAssignment()
	ns.CreateInitialNodesInfo()

	return ns
}

func createNodesSetupTwoShard6NodesMeta() *sharding.NodesSetup {
	noOfInitialNodes := 6
	ns := &sharding.NodesSetup{}
	ns.ConsensusGroupSize = 1
	ns.MinNodesPerShard = 2
	ns.MetaChainMinNodes = 2
	ns.MetaChainConsensusGroupSize = 2
	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()
	if err != nil {
		return nil
	}

	ns.ProcessMetaChainAssigment()
	ns.ProcessShardAssignment()
	ns.CreateInitialNodesInfo()

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
	inPubKeys := ns.InitialNodesInfo()

	assert.NotNil(t, ns)
	assert.Nil(t, inPubKeys)
}

func TestNodesSetup_ProcessConfigNodesWithIncompleteDataShouldErr(t *testing.T) {
	noOfInitialNodes := 2
	ns := sharding.NodesSetup{}

	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	ns.InitialNodes[0] = &sharding.InitialNode{}
	ns.InitialNodes[1] = &sharding.InitialNode{}

	ns.InitialNodes[0].PubKey = PubKeys[0]
	ns.InitialNodes[0].Address = Address[0]

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrCouldNotParsePubKey, err)
}

func TestNodesSetup_ProcessConfigInvalidConsensusGroupSizeShouldErr(t *testing.T) {
	noOfInitialNodes := 2
	ns := sharding.NodesSetup{
		ConsensusGroupSize: 0,
		MinNodesPerShard:   0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrNegativeOrZeroConsensusGroupSize, err)
}

func TestNodesSetup_ProcessConfigInvalidMetaConsensusGroupSizeShouldErr(t *testing.T) {
	noOfInitialNodes := 2
	ns := sharding.NodesSetup{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 0,
		MetaChainMinNodes:           0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrNegativeOrZeroConsensusGroupSize, err)
}

func TestNodesSetup_ProcessConfigInvalidConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	noOfInitialNodes := 2
	ns := sharding.NodesSetup{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodesSetup_ProcessConfigInvalidMetaConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	noOfInitialNodes := 2
	ns := sharding.NodesSetup{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 1,
		MetaChainMinNodes:           0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, 2)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodesSetup_ProcessConfigInvalidMinNodesPerShardShouldErr(t *testing.T) {
	noOfInitialNodes := 2
	ns := sharding.NodesSetup{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodesSetup_ProcessConfigInvalidMetaMinNodesPerShardShouldErr(t *testing.T) {
	noOfInitialNodes := 1
	ns := sharding.NodesSetup{
		ConsensusGroupSize:          1,
		MinNodesPerShard:            1,
		MetaChainConsensusGroupSize: 1,
		MetaChainMinNodes:           0,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrMinNodesPerShardSmallerThanConsensusSize, err)
}

func TestNodesSetup_ProcessConfigInvalidNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	noOfInitialNodes := 2
	ns := sharding.NodesSetup{
		ConsensusGroupSize: 2,
		MinNodesPerShard:   3,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestNodesSetup_ProcessConfigInvalidMetaNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	noOfInitialNodes := 3
	ns := sharding.NodesSetup{
		ConsensusGroupSize: 1,
		MinNodesPerShard:   1,

		MetaChainConsensusGroupSize: 2,
		MetaChainMinNodes:           3,
	}

	ns.InitialNodes = make([]*sharding.InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
		ns.InitialNodes[i] = &sharding.InitialNode{}
		ns.InitialNodes[i].PubKey = PubKeys[i]
		ns.InitialNodes[i].Address = Address[i]
	}

	err := ns.ProcessConfig()

	assert.NotNil(t, ns)
	assert.Equal(t, sharding.ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardNil(t *testing.T) {
	ns := sharding.NodesSetup{}
	inPK, err := ns.InitialNodesInfoForShard(0)

	assert.NotNil(t, ns)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardWrongShard(t *testing.T) {
	ns := createNodesSetupOneShardOneNodeWithOneMeta()
	inPK, err := ns.InitialNodesInfoForShard(1)

	assert.NotNil(t, ns)
	assert.Nil(t, inPK)
	assert.NotNil(t, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardGood(t *testing.T) {
	ns := createNodesSetupTwoShardTwoNodesWithOneMeta()
	inPK, err := ns.InitialNodesInfoForShard(1)

	assert.NotNil(t, ns)
	assert.Equal(t, 2, len(inPK))
	assert.Nil(t, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardGoodMeta(t *testing.T) {
	ns := createNodesSetupTwoShard6NodesMeta()
	metaId := sharding.MetachainShardId
	inPK, err := ns.InitialNodesInfoForShard(metaId)

	assert.NotNil(t, ns)
	assert.Equal(t, 2, len(inPK))
	assert.Nil(t, err)
}

func TestNodesSetup_PublicKeyNotGood(t *testing.T) {
	ns := createNodesSetupTwoShard6NodesMeta()

	_, err := ns.GetShardIDForPubKey([]byte(PubKeys[0]))

	assert.NotNil(t, ns)
	assert.NotNil(t, err)
}

func TestNodesSetup_PublicKeyGood(t *testing.T) {
	ns := createNodesSetupTwoShard5NodesWithMeta()
	publicKey, err := hex.DecodeString(PubKeys[2])

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	assert.NotNil(t, ns)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), selfId)
}

func TestNodesSetup_ShardPublicKeyGoodMeta(t *testing.T) {
	ns := createNodesSetupTwoShard6NodesMeta()
	publicKey, err := hex.DecodeString(PubKeys[2])

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	assert.NotNil(t, ns)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), selfId)
}

func TestNodesSetup_MetaPublicKeyGoodMeta(t *testing.T) {
	ns := createNodesSetupTwoShard6NodesMeta()
	metaId := sharding.MetachainShardId
	publicKey, err := hex.DecodeString(PubKeys[0])

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	assert.NotNil(t, ns)
	assert.Nil(t, err)
	assert.Equal(t, metaId, selfId)
}
