package mock

import (
	"time"

	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// GenesisNodesSetupHandlerStub -
type GenesisNodesSetupHandlerStub struct {
	InitialNodesInfoForShardCalled            func(shardId uint32) ([]nodesCoordinator.GenesisNodeInfoHandler, []nodesCoordinator.GenesisNodeInfoHandler, error)
	InitialNodesInfoCalled                    func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
	GetStartTimeCalled                        func() int64
	GetRoundDurationCalled                    func() uint64
	GetChainIdCalled                          func() string
	GetMinTransactionVersionCalled            func() uint32
	GetShardConsensusGroupSizeCalled          func() uint32
	GetMetaConsensusGroupSizeCalled           func() uint32
	MinNumberOfShardNodesCalled               func() uint32
	MinNumberOfMetaNodesCalled                func() uint32
	GetHysteresisCalled                       func() float32
	GetAdaptivityCalled                       func() bool
	NumberOfShardsCalled                      func() uint32
	MinNumberOfNodesCalled                    func() uint32
	AllInitialNodesCalled                     func() []nodesCoordinator.GenesisNodeInfoHandler
	InitialNodesPubKeysCalled                 func() map[uint32][]string
	GetShardIDForPubKeyCalled                 func(pubkey []byte) (uint32, error)
	InitialEligibleNodesPubKeysForShardCalled func(shardId uint32) ([]string, error)
	SetStartTimeCalled                        func(startTime int64)
	MinNumberOfNodesWithHysteresisCalled      func() uint32
	MinShardHysteresisNodesCalled             func() uint32
	MinMetaHysteresisNodesCalled              func() uint32
}

// InitialNodesInfoForShard -
func (g *GenesisNodesSetupHandlerStub) InitialNodesInfoForShard(shardId uint32) ([]nodesCoordinator.GenesisNodeInfoHandler, []nodesCoordinator.GenesisNodeInfoHandler, error) {
	if g.InitialNodesInfoForShardCalled != nil {
		return g.InitialNodesInfoForShardCalled(shardId)
	}

	return nil, nil, nil
}

// InitialNodesInfo -
func (g *GenesisNodesSetupHandlerStub) InitialNodesInfo() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
	if g.InitialNodesInfoCalled != nil {
		return g.InitialNodesInfoCalled()
	}

	return nil, nil
}

// GetStartTime -
func (g *GenesisNodesSetupHandlerStub) GetStartTime() int64 {
	if g.GetStartTimeCalled != nil {
		return g.GetStartTimeCalled()
	}

	return time.Now().Unix()
}

// GetRoundDuration -
func (g *GenesisNodesSetupHandlerStub) GetRoundDuration() uint64 {
	if g.GetRoundDurationCalled != nil {
		return g.GetRoundDurationCalled()
	}

	return 4500
}

// GetChainId -
func (g *GenesisNodesSetupHandlerStub) GetChainId() string {
	if g.GetChainIdCalled != nil {
		return g.GetChainIdCalled()
	}

	return "chainID"
}

// GetMinTransactionVersion -
func (g *GenesisNodesSetupHandlerStub) GetMinTransactionVersion() uint32 {
	if g.GetMinTransactionVersionCalled != nil {
		return g.GetMinTransactionVersionCalled()
	}

	return 1
}

// GetShardConsensusGroupSize -
func (g *GenesisNodesSetupHandlerStub) GetShardConsensusGroupSize() uint32 {
	if g.GetShardConsensusGroupSizeCalled != nil {
		return g.GetShardConsensusGroupSizeCalled()
	}

	return 1
}

// GetMetaConsensusGroupSize -
func (g *GenesisNodesSetupHandlerStub) GetMetaConsensusGroupSize() uint32 {
	if g.GetMetaConsensusGroupSizeCalled != nil {
		return g.GetMetaConsensusGroupSizeCalled()
	}

	return 1
}

// MinNumberOfShardNodes -
func (g *GenesisNodesSetupHandlerStub) MinNumberOfShardNodes() uint32 {
	if g.MinNumberOfShardNodesCalled != nil {
		return g.MinNumberOfShardNodesCalled()
	}

	return 1
}

// MinNumberOfMetaNodes -
func (g *GenesisNodesSetupHandlerStub) MinNumberOfMetaNodes() uint32 {
	if g.MinNumberOfMetaNodesCalled != nil {
		return g.MinNumberOfMetaNodesCalled()
	}

	return 1
}

// GetHysteresis -
func (g *GenesisNodesSetupHandlerStub) GetHysteresis() float32 {
	if g.GetHysteresisCalled != nil {
		return g.GetHysteresisCalled()
	}

	return 0
}

// GetAdaptivity -
func (g *GenesisNodesSetupHandlerStub) GetAdaptivity() bool {
	if g.GetAdaptivityCalled != nil {
		return g.GetAdaptivityCalled()
	}

	return false
}

// NumberOfShards -
func (g *GenesisNodesSetupHandlerStub) NumberOfShards() uint32 {
	if g.NumberOfShardsCalled != nil {
		return g.NumberOfShardsCalled()
	}

	return 2
}

// MinNumberOfNodes -
func (g *GenesisNodesSetupHandlerStub) MinNumberOfNodes() uint32 {
	if g.MinNumberOfNodesCalled != nil {
		return g.MinNumberOfNodesCalled()
	}

	return 1
}

// AllInitialNodes -
func (g *GenesisNodesSetupHandlerStub) AllInitialNodes() []nodesCoordinator.GenesisNodeInfoHandler {
	if g.AllInitialNodesCalled != nil {
		return g.AllInitialNodesCalled()
	}

	return nil
}

// InitialNodesPubKeys -
func (g *GenesisNodesSetupHandlerStub) InitialNodesPubKeys() map[uint32][]string {
	if g.InitialNodesPubKeysCalled != nil {
		return g.InitialNodesPubKeysCalled()
	}

	return nil
}

// GetShardIDForPubKey -
func (g *GenesisNodesSetupHandlerStub) GetShardIDForPubKey(pubkey []byte) (uint32, error) {
	if g.GetShardIDForPubKeyCalled != nil {
		return g.GetShardIDForPubKeyCalled(pubkey)
	}

	return 0, nil
}

// InitialEligibleNodesPubKeysForShard -
func (g *GenesisNodesSetupHandlerStub) InitialEligibleNodesPubKeysForShard(shardId uint32) ([]string, error) {
	if g.InitialEligibleNodesPubKeysForShardCalled != nil {
		return g.InitialEligibleNodesPubKeysForShardCalled(shardId)
	}

	return nil, nil
}

// SetStartTime -
func (g *GenesisNodesSetupHandlerStub) SetStartTime(startTime int64) {
	if g.SetStartTimeCalled != nil {
		g.SetStartTimeCalled(startTime)
	}
}

// MinNumberOfNodesWithHysteresis -
func (g *GenesisNodesSetupHandlerStub) MinNumberOfNodesWithHysteresis() uint32 {
	if g.MinNumberOfNodesWithHysteresisCalled != nil {
		return g.MinNumberOfNodesWithHysteresisCalled()
	}

	return 0
}

// MinShardHysteresisNodes -
func (g *GenesisNodesSetupHandlerStub) MinShardHysteresisNodes() uint32 {
	if g.MinShardHysteresisNodesCalled != nil {
		return g.MinShardHysteresisNodesCalled()
	}

	return 0
}

// MinMetaHysteresisNodes -
func (g *GenesisNodesSetupHandlerStub) MinMetaHysteresisNodes() uint32 {
	if g.MinMetaHysteresisNodesCalled != nil {
		return g.MinMetaHysteresisNodesCalled()
	}

	return 0
}

// IsInterfaceNil -
func (g *GenesisNodesSetupHandlerStub) IsInterfaceNil() bool {
	return g == nil
}
