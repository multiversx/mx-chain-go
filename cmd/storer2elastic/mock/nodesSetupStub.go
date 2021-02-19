package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// NodesSetupStub -
type NodesSetupStub struct {
	InitialNodesInfoForShardCalled       func(shardId uint32) ([]sharding.GenesisNodeInfoHandler, []sharding.GenesisNodeInfoHandler, error)
	InitialNodesInfoCalled               func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler)
	GetStartTimeCalled                   func() int64
	GetRoundDurationCalled               func() uint64
	GetChainIdCalled                     func() string
	GetMinTransactionVersionCalled       func() uint32
	GetShardConsensusGroupSizeCalled     func() uint32
	GetMetaConsensusGroupSizeCalled      func() uint32
	NumberOfShardsCalled                 func() uint32
	MinNumberOfNodesCalled               func() uint32
	MinNumberOfNodesWithHysteresisCalled func() uint32
}

// MinNumberOfNodes -
func (n *NodesSetupStub) MinNumberOfNodes() uint32 {
	if n.MinNumberOfNodesCalled != nil {
		return n.MinNumberOfNodesCalled()
	}
	return 1
}

// GetStartTime -
func (n *NodesSetupStub) GetStartTime() int64 {
	if n.GetStartTimeCalled != nil {
		return n.GetStartTimeCalled()
	}
	return 0
}

// GetMinTransactionVersion -
func (n *NodesSetupStub) GetMinTransactionVersion() uint32 {
	if n.GetMinTransactionVersionCalled != nil {
		return n.GetMinTransactionVersionCalled()
	}
	return 1
}

// GetRoundDuration -
func (n *NodesSetupStub) GetRoundDuration() uint64 {
	if n.GetRoundDurationCalled != nil {
		return n.GetRoundDurationCalled()
	}
	return 0
}

// GetChainId -
func (n *NodesSetupStub) GetChainId() string {
	if n.GetChainIdCalled != nil {
		return n.GetChainIdCalled()
	}
	return "chainID"
}

// GetShardConsensusGroupSize -
func (n *NodesSetupStub) GetShardConsensusGroupSize() uint32 {
	if n.GetShardConsensusGroupSizeCalled != nil {
		return n.GetShardConsensusGroupSizeCalled()
	}
	return 0
}

// GetMetaConsensusGroupSize -
func (n *NodesSetupStub) GetMetaConsensusGroupSize() uint32 {
	if n.GetMetaConsensusGroupSizeCalled != nil {
		return n.GetMetaConsensusGroupSizeCalled()
	}
	return 0
}

// NumberOfShards -
func (n *NodesSetupStub) NumberOfShards() uint32 {
	if n.NumberOfShardsCalled != nil {
		return n.NumberOfShardsCalled()
	}
	return 0
}

// InitialNodesInfoForShard -
func (n *NodesSetupStub) InitialNodesInfoForShard(shardId uint32) ([]sharding.GenesisNodeInfoHandler, []sharding.GenesisNodeInfoHandler, error) {
	if n.InitialNodesInfoForShardCalled != nil {
		return n.InitialNodesInfoForShardCalled(shardId)
	}
	return nil, nil, nil
}

// InitialNodesInfo -
func (n *NodesSetupStub) InitialNodesInfo() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
	if n.InitialNodesInfoCalled != nil {
		return n.InitialNodesInfoCalled()
	}
	return nil, nil
}

// MinNumberOfNodesWithHysteresis -
func (n *NodesSetupStub) MinNumberOfNodesWithHysteresis() uint32 {
	if n.MinNumberOfNodesWithHysteresisCalled != nil {
		return n.MinNumberOfNodesWithHysteresisCalled()
	}
	return n.MinNumberOfNodes()
}

// IsInterfaceNil -
func (n *NodesSetupStub) IsInterfaceNil() bool {
	return n == nil
}
