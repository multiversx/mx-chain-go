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
	MinNumberOfShardNodesCalled          func() uint32
	MinNumberOfMetaNodesCalled           func() uint32
	GetHysteresisCalled                  func() float32
	GetAdaptivityCalled                  func() bool
	MinNumberOfNodesWithHysteresisCalled func() uint32
}

// MinNumberOfShardNodes -
func (n *NodesSetupStub) MinNumberOfShardNodes() uint32 {
	if n.MinNumberOfShardNodesCalled != nil {
		return n.MinNumberOfShardNodesCalled()
	}

	return 1
}

// MinNumberOfMetaNodes -
func (n *NodesSetupStub) MinNumberOfMetaNodes() uint32 {
	if n.MinNumberOfMetaNodesCalled != nil {
		return n.MinNumberOfMetaNodesCalled()
	}

	return 1
}

// GetHysteresis -
func (n *NodesSetupStub) GetHysteresis() float32 {
	if n.GetHysteresisCalled != nil {
		return n.GetHysteresisCalled()
	}

	return 0
}

// GetAdaptivity -
func (n *NodesSetupStub) GetAdaptivity() bool {
	if n.GetAdaptivityCalled != nil {
		return n.GetAdaptivityCalled()
	}

	return false
}

// MinNumberOfNodes -
func (n *NodesSetupStub) MinNumberOfNodes() uint32 {
	if n.MinNumberOfNodesCalled != nil {
		return n.MinNumberOfNodesCalled()
	}
	return 2
}

// GetStartTime -
func (n *NodesSetupStub) GetStartTime() int64 {
	if n.GetStartTimeCalled != nil {
		return n.GetStartTimeCalled()
	}
	return 0
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

// GetMinTransactionVersion -
func (n *NodesSetupStub) GetMinTransactionVersion() uint32 {
	if n.GetMinTransactionVersionCalled != nil {
		return n.GetMinTransactionVersionCalled()
	}
	return 1
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
