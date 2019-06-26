package mock

import "github.com/ElrondNetwork/elrond-go/consensus"

type ConsensusStateMock struct {
	ResetConsensusStateCalled        func()
	IsNodeLeaderInCurrentRoundCalled func(node string) bool
	IsSelfLeaderInCurrentRoundCalled func() bool
	GetLeaderCalled                  func() (string, error)
	GetNextConsensusGroupCalled      func(randomSource string, vgs consensus.NodesCoordinator) ([]string, error)
	IsConsensusDataSetCalled         func() bool
	IsConsensusDataEqualCalled       func(data []byte) bool
	IsJobDoneCalled                  func(node string, currentSubroundId int) bool
	IsSelfJobDoneCalled              func(currentSubroundId int) bool
	IsCurrentSubroundFinishedCalled  func(currentSubroundId int) bool
	IsNodeSelfCalled                 func(node string) bool
	IsBlockBodyAlreadyReceivedCalled func() bool
	IsHeaderAlreadyReceivedCalled    func() bool
	CanDoSubroundJobCalled           func(currentSubroundId int) bool
	CanProcessReceivedMessageCalled  func(cnsDta consensus.Message, currentRoundIndex int32,
		currentSubroundId int) bool
	GenerateBitmapCalled     func(subroundId int) []byte
	ProcessingBlockCalled    func() bool
	SetProcessingBlockCalled func(processingBlock bool)
	ConsensusGroupSizeCalled func() int
	SetThresholdCalled       func(subroundId int, threshold int)
}

func (cnsm *ConsensusStateMock) ResetConsensusState() {
	cnsm.ResetConsensusStateCalled()
}

func (cnsm *ConsensusStateMock) IsNodeLeaderInCurrentRound(node string) bool {
	return cnsm.IsNodeLeaderInCurrentRoundCalled(node)
}

func (cnsm *ConsensusStateMock) IsSelfLeaderInCurrentRound() bool {
	return cnsm.IsSelfLeaderInCurrentRoundCalled()
}

func (cnsm *ConsensusStateMock) GetLeader() (string, error) {
	return cnsm.GetLeaderCalled()
}

func (cnsm *ConsensusStateMock) GetNextConsensusGroup(randomSource string,
	vgs consensus.NodesCoordinator) ([]string,
	error) {
	return cnsm.GetNextConsensusGroupCalled(randomSource, vgs)
}

func (cnsm *ConsensusStateMock) IsConsensusDataSet() bool {
	return cnsm.IsConsensusDataSetCalled()
}

func (cnsm *ConsensusStateMock) IsConsensusDataEqual(data []byte) bool {
	return cnsm.IsConsensusDataEqualCalled(data)
}

func (cnsm *ConsensusStateMock) IsJobDone(node string, currentSubroundId int) bool {
	return cnsm.IsJobDoneCalled(node, currentSubroundId)
}

func (cnsm *ConsensusStateMock) IsSelfJobDone(currentSubroundId int) bool {
	return cnsm.IsSelfJobDoneCalled(currentSubroundId)
}

func (cnsm *ConsensusStateMock) IsCurrentSubroundFinished(currentSubroundId int) bool {
	return cnsm.IsCurrentSubroundFinishedCalled(currentSubroundId)
}

func (cnsm *ConsensusStateMock) IsNodeSelf(node string) bool {
	return cnsm.IsNodeSelfCalled(node)
}

func (cnsm *ConsensusStateMock) IsBlockBodyAlreadyReceived() bool {
	return cnsm.IsBlockBodyAlreadyReceivedCalled()
}

func (cnsm *ConsensusStateMock) IsHeaderAlreadyReceived() bool {
	return cnsm.IsHeaderAlreadyReceivedCalled()
}

func (cnsm *ConsensusStateMock) CanDoSubroundJob(currentSubroundId int) bool {
	return cnsm.CanDoSubroundJobCalled(currentSubroundId)
}

func (cnsm *ConsensusStateMock) CanProcessReceivedMessage(cnsDta consensus.Message, currentRoundIndex int32,
	currentSubroundId int) bool {
	return cnsm.CanProcessReceivedMessageCalled(cnsDta, currentRoundIndex, currentSubroundId)
}

func (cnsm *ConsensusStateMock) GenerateBitmap(subroundId int) []byte {
	return cnsm.GenerateBitmapCalled(subroundId)
}

func (cnsm *ConsensusStateMock) ProcessingBlock() bool {
	return cnsm.ProcessingBlockCalled()
}

func (cnsm *ConsensusStateMock) SetProcessingBlock(processingBlock bool) {
	cnsm.SetProcessingBlockCalled(processingBlock)
}

func (cnsm *ConsensusStateMock) ConsensusGroupSize() int {
	return cnsm.ConsensusGroupSizeCalled()
}

func (cnsm *ConsensusStateMock) SetThreshold(subroundId int, threshold int) {
	cnsm.SetThresholdCalled(subroundId, threshold)
}
