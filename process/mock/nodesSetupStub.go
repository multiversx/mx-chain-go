package mock

import "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"

// NodesSetupStub -
type NodesSetupStub struct {
	InitialNodesInfoForShardCalled            func(shardId uint32) ([]nodesCoordinator.GenesisNodeInfoHandler, []nodesCoordinator.GenesisNodeInfoHandler, error)
	InitialNodesInfoCalled                    func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
	GetStartTimeCalled                        func() int64
	GetRoundDurationCalled                    func() uint64
	GetShardConsensusGroupSizeCalled          func() uint32
	GetMetaConsensusGroupSizeCalled           func() uint32
	NumberOfShardsCalled                      func() uint32
	MinNumberOfNodesCalled                    func() uint32
	AllInitialNodesCalled                     func() []nodesCoordinator.GenesisNodeInfoHandler
	GetAdaptivityCalled                       func() bool
	GetHysteresisCalled                       func() float32
	GetShardIDForPubKeyCalled                 func(pubkey []byte) (uint32, error)
	InitialEligibleNodesPubKeysForShardCalled func(shardId uint32) ([]string, error)
	InitialNodesPubKeysCalled                 func() map[uint32][]string
	MinNumberOfMetaNodesCalled                func() uint32
	MinNumberOfShardNodesCalled               func() uint32
	MinNumberOfNodesWithHysteresisCalled      func() uint32
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

// GetRoundDuration -
func (n *NodesSetupStub) GetRoundDuration() uint64 {
	if n.GetRoundDurationCalled != nil {
		return n.GetRoundDurationCalled()
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

func (n *NodesSetupStub) GetShardConsensusGroupSize() uint32 {
	if n.GetMetaConsensusGroupSizeCalled != nil {
		return n.GetShardConsensusGroupSizeCalled()
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
func (n *NodesSetupStub) InitialNodesInfoForShard(shardId uint32) ([]nodesCoordinator.GenesisNodeInfoHandler, []nodesCoordinator.GenesisNodeInfoHandler, error) {
	if n.InitialNodesInfoForShardCalled != nil {
		return n.InitialNodesInfoForShardCalled(shardId)
	}
	return nil, nil, nil
}

// InitialNodesInfo -
func (n *NodesSetupStub) InitialNodesInfo() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
	if n.InitialNodesInfoCalled != nil {
		return n.InitialNodesInfoCalled()
	}
	return nil, nil
}

// AllInitialNodes -
func (n *NodesSetupStub) AllInitialNodes() []nodesCoordinator.GenesisNodeInfoHandler {
	if n.AllInitialNodesCalled != nil {
		return n.AllInitialNodesCalled()
	}
	return nil
}

// GetAdaptivity -
func (n *NodesSetupStub) GetAdaptivity() bool {
	if n.GetAdaptivityCalled != nil {
		return n.GetAdaptivityCalled()
	}

	return false
}

// GetHysteresis -
func (n *NodesSetupStub) GetHysteresis() float32 {
	if n.GetHysteresisCalled != nil {
		return n.GetHysteresisCalled()
	}

	return 0
}

// GetShardIDForPubKey -
func (n *NodesSetupStub) GetShardIDForPubKey(pubkey []byte) (uint32, error) {
	if n.GetShardIDForPubKeyCalled != nil {
		return n.GetShardIDForPubKeyCalled(pubkey)
	}
	return 0, nil
}

// InitialEligibleNodesPubKeysForShard -
func (n *NodesSetupStub) InitialEligibleNodesPubKeysForShard(shardId uint32) ([]string, error) {
	if n.InitialEligibleNodesPubKeysForShardCalled != nil {
		return n.InitialEligibleNodesPubKeysForShardCalled(shardId)
	}

	return []string{"val1", "val2"}, nil
}

// InitialNodesPubKeys -
func (n *NodesSetupStub) InitialNodesPubKeys() map[uint32][]string {
	if n.InitialNodesPubKeysCalled != nil {
		return n.InitialNodesPubKeysCalled()
	}

	return map[uint32][]string{0: {"val1", "val2"}}
}

// MinNumberOfMetaNodes -
func (n *NodesSetupStub) MinNumberOfMetaNodes() uint32 {
	if n.MinNumberOfMetaNodesCalled != nil {
		return n.MinNumberOfMetaNodesCalled()
	}

	return 1
}

// MinNumberOfShardNodes -
func (n *NodesSetupStub) MinNumberOfShardNodes() uint32 {
	if n.MinNumberOfShardNodesCalled != nil {
		return n.MinNumberOfShardNodesCalled()
	}

	return 1
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
