package mock

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// ConsensusStateMock -
type ConsensusStateMock struct {
	ResetConsensusStateCalled        func()
	IsNodeLeaderInCurrentRoundCalled func(node string) bool
	IsSelfLeaderInCurrentRoundCalled func() bool
	GetLeaderCalled                  func() (string, error)
	GetNextConsensusGroupCalled      func(randomSource string, vgs nodesCoordinator.NodesCoordinator) ([]string, error)
	IsConsensusDataSetCalled         func() bool
	IsConsensusDataEqualCalled       func(data []byte) bool
	IsJobDoneCalled                  func(node string, currentSubroundId int) bool
	IsSelfJobDoneCalled              func(currentSubroundId int) bool
	IsCurrentSubroundFinishedCalled  func(currentSubroundId int) bool
	IsNodeSelfCalled                 func(node string) bool
	IsBlockBodyAlreadyReceivedCalled func() bool
	IsHeaderAlreadyReceivedCalled    func() bool
	CanDoSubroundJobCalled           func(currentSubroundId int) bool
	CanProcessReceivedMessageCalled  func(cnsDta consensus.Message, currentRoundIndex int32, currentSubroundId int) bool
	GenerateBitmapCalled             func(subroundId int) []byte
	ProcessingBlockCalled            func() bool
	SetProcessingBlockCalled         func(processingBlock bool)
	ConsensusGroupSizeCalled         func() int
	SetThresholdCalled               func(subroundId int, threshold int)
}

// ResetConsensusState -
func (cnsm *ConsensusStateMock) ResetConsensusState() {
	cnsm.ResetConsensusStateCalled()
}

// IsNodeLeaderInCurrentRound -
func (cnsm *ConsensusStateMock) IsNodeLeaderInCurrentRound(node string) bool {
	return cnsm.IsNodeLeaderInCurrentRoundCalled(node)
}

// IsSelfLeaderInCurrentRound -
func (cnsm *ConsensusStateMock) IsSelfLeaderInCurrentRound() bool {
	return cnsm.IsSelfLeaderInCurrentRoundCalled()
}

// GetLeader -
func (cnsm *ConsensusStateMock) GetLeader() (string, error) {
	return cnsm.GetLeaderCalled()
}

// GetNextConsensusGroup -
func (cnsm *ConsensusStateMock) GetNextConsensusGroup(
	randomSource string,
	vgs nodesCoordinator.NodesCoordinator,
) ([]string, error) {
	return cnsm.GetNextConsensusGroupCalled(randomSource, vgs)
}

// IsConsensusDataSet -
func (cnsm *ConsensusStateMock) IsConsensusDataSet() bool {
	return cnsm.IsConsensusDataSetCalled()
}

// IsConsensusDataEqual -
func (cnsm *ConsensusStateMock) IsConsensusDataEqual(data []byte) bool {
	return cnsm.IsConsensusDataEqualCalled(data)
}

// IsJobDone -
func (cnsm *ConsensusStateMock) IsJobDone(node string, currentSubroundId int) bool {
	return cnsm.IsJobDoneCalled(node, currentSubroundId)
}

// IsSelfJobDone -
func (cnsm *ConsensusStateMock) IsSelfJobDone(currentSubroundId int) bool {
	return cnsm.IsSelfJobDoneCalled(currentSubroundId)
}

// IsCurrentSubroundFinished -
func (cnsm *ConsensusStateMock) IsCurrentSubroundFinished(currentSubroundId int) bool {
	return cnsm.IsCurrentSubroundFinishedCalled(currentSubroundId)
}

// IsNodeSelf -
func (cnsm *ConsensusStateMock) IsNodeSelf(node string) bool {
	return cnsm.IsNodeSelfCalled(node)
}

// IsBlockBodyAlreadyReceived -
func (cnsm *ConsensusStateMock) IsBlockBodyAlreadyReceived() bool {
	return cnsm.IsBlockBodyAlreadyReceivedCalled()
}

// IsHeaderAlreadyReceived -
func (cnsm *ConsensusStateMock) IsHeaderAlreadyReceived() bool {
	return cnsm.IsHeaderAlreadyReceivedCalled()
}

// CanDoSubroundJob -
func (cnsm *ConsensusStateMock) CanDoSubroundJob(currentSubroundId int) bool {
	return cnsm.CanDoSubroundJobCalled(currentSubroundId)
}

// CanProcessReceivedMessage -
func (cnsm *ConsensusStateMock) CanProcessReceivedMessage(
	cnsDta consensus.Message,
	currentRoundIndex int32,
	currentSubroundId int,
) bool {
	return cnsm.CanProcessReceivedMessageCalled(cnsDta, currentRoundIndex, currentSubroundId)
}

// GenerateBitmap -
func (cnsm *ConsensusStateMock) GenerateBitmap(subroundId int) []byte {
	return cnsm.GenerateBitmapCalled(subroundId)
}

// ProcessingBlock -
func (cnsm *ConsensusStateMock) ProcessingBlock() bool {
	return cnsm.ProcessingBlockCalled()
}

// SetProcessingBlock -
func (cnsm *ConsensusStateMock) SetProcessingBlock(processingBlock bool) {
	cnsm.SetProcessingBlockCalled(processingBlock)
}

// ConsensusGroupSize -
func (cnsm *ConsensusStateMock) ConsensusGroupSize() int {
	return cnsm.ConsensusGroupSizeCalled()
}

// SetThreshold -
func (cnsm *ConsensusStateMock) SetThreshold(subroundId int, threshold int) {
	cnsm.SetThresholdCalled(subroundId, threshold)
}
