package consensus

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// ConsensusStateMock -
type ConsensusStateMock struct {
	ResetConsensusStateCalled                    func()
	IsNodeLeaderInCurrentRoundCalled             func(node string) bool
	IsSelfLeaderInCurrentRoundCalled             func() bool
	GetLeaderCalled                              func() (string, error)
	GetNextConsensusGroupCalled                  func(randomSource []byte, round uint64, shardId uint32, nodesCoordinator nodesCoordinator.NodesCoordinator, epoch uint32) (string, []string, error)
	IsConsensusDataSetCalled                     func() bool
	IsConsensusDataEqualCalled                   func(data []byte) bool
	IsJobDoneCalled                              func(node string, currentSubroundId int) bool
	IsSelfJobDoneCalled                          func(currentSubroundId int) bool
	IsCurrentSubroundFinishedCalled              func(currentSubroundId int) bool
	IsNodeSelfCalled                             func(node string) bool
	IsBlockBodyAlreadyReceivedCalled             func() bool
	IsHeaderAlreadyReceivedCalled                func() bool
	CanDoSubroundJobCalled                       func(currentSubroundId int) bool
	CanProcessReceivedMessageCalled              func(cnsDta *consensus.Message, currentRoundIndex int64, currentSubroundId int) bool
	GenerateBitmapCalled                         func(subroundId int) []byte
	ProcessingBlockCalled                        func() bool
	SetProcessingBlockCalled                     func(processingBlock bool)
	ConsensusGroupSizeCalled                     func() int
	SetThresholdCalled                           func(subroundId int, threshold int)
	AddReceivedHeaderCalled                      func(headerHandler data.HeaderHandler)
	GetReceivedHeadersCalled                     func() []data.HeaderHandler
	AddMessageWithSignatureCalled                func(key string, message p2p.MessageP2P)
	GetMessageWithSignatureCalled                func(key string) (p2p.MessageP2P, bool)
	IsSubroundFinishedCalled                     func(subroundID int) bool
	GetDataCalled                                func() []byte
	SetDataCalled                                func(data []byte)
	IsMultiKeyLeaderInCurrentRoundCalled         func() bool
	IsLeaderJobDoneCalled                        func(currentSubroundId int) bool
	IsMultiKeyJobDoneCalled                      func(currentSubroundId int) bool
	GetMultikeyRedundancyStepInReasonCalled      func() string
	ResetRoundsWithoutReceivedMessagesCalled     func(pkBytes []byte, pid core.PeerID)
	GetRoundCanceledCalled                       func() bool
	SetRoundCanceledCalled                       func(state bool)
	GetRoundIndexCalled                          func() int64
	SetRoundIndexCalled                          func(roundIndex int64)
	GetRoundTimeStampCalled                      func() time.Time
	SetRoundTimeStampCalled                      func(roundTimeStamp time.Time)
	GetExtendedCalledCalled                      func() bool
	GetBodyCalled                                func() data.BodyHandler
	SetBodyCalled                                func(body data.BodyHandler)
	GetHeaderCalled                              func() data.HeaderHandler
	SetHeaderCalled                              func(header data.HeaderHandler)
	GetWaitingAllSignaturesTimeOutCalled         func() bool
	SetWaitingAllSignaturesTimeOutCalled         func(b bool)
	ConsensusGroupIndexCalled                    func(pubKey string) (int, error)
	SelfConsensusGroupIndexCalled                func() (int, error)
	SetEligibleListCalled                        func(eligibleList map[string]struct{})
	ConsensusGroupCalled                         func() []string
	SetConsensusGroupCalled                      func(consensusGroup []string)
	SetLeaderCalled                              func(leader string)
	SetConsensusGroupSizeCalled                  func(consensusGroupSize int)
	SelfPubKeyCalled                             func() string
	SetSelfPubKeyCalled                          func(selfPubKey string)
	JobDoneCalled                                func(key string, subroundId int) (bool, error)
	SetJobDoneCalled                             func(key string, subroundId int, value bool) error
	SelfJobDoneCalled                            func(subroundId int) (bool, error)
	IsNodeInConsensusGroupCalled                 func(node string) bool
	IsNodeInEligibleListCalled                   func(node string) bool
	ComputeSizeCalled                            func(subroundId int) int
	ResetRoundStateCalled                        func()
	IsMultiKeyInConsensusGroupCalled             func() bool
	IsKeyManagedBySelfCalled                     func(pkBytes []byte) bool
	IncrementRoundsWithoutReceivedMessagesCalled func(pkBytes []byte)
	GetKeysHandlerCalled                         func() consensus.KeysHandler
	LeaderCalled                                 func() string
	StatusCalled                                 func(subroundId int) spos.SubroundStatus
	SetStatusCalled                              func(subroundId int, subroundStatus spos.SubroundStatus)
	ResetRoundStatusCalled                       func()
	ThresholdCalled                              func(subroundId int) int
	FallbackThresholdCalled                      func(subroundId int) int
	SetFallbackThresholdCalled                   func(subroundId int, threshold int)
}

func (cnsm *ConsensusStateMock) AddReceivedHeader(headerHandler data.HeaderHandler) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetReceivedHeaders() []data.HeaderHandler {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) AddMessageWithSignature(key string, message p2p.MessageP2P) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetMessageWithSignature(key string) (p2p.MessageP2P, bool) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) IsSubroundFinished(subroundID int) bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetData() []byte {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetData(data []byte) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) IsMultiKeyLeaderInCurrentRound() bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) IsLeaderJobDone(currentSubroundId int) bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) IsMultiKeyJobDone(currentSubroundId int) bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetMultikeyRedundancyStepInReason() string {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) ResetRoundsWithoutReceivedMessages(pkBytes []byte, pid core.PeerID) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetRoundCanceled() bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetRoundCanceled(state bool) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetRoundIndex() int64 {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetRoundIndex(roundIndex int64) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetRoundTimeStamp() time.Time {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetRoundTimeStamp(roundTimeStamp time.Time) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetExtendedCalled() bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetBody() data.BodyHandler {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetBody(body data.BodyHandler) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetHeader() data.HeaderHandler {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetHeader(header data.HeaderHandler) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetWaitingAllSignaturesTimeOut() bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetWaitingAllSignaturesTimeOut(b bool) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) ConsensusGroupIndex(pubKey string) (int, error) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SelfConsensusGroupIndex() (int, error) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetEligibleList(eligibleList map[string]struct{}) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) ConsensusGroup() []string {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetConsensusGroup(consensusGroup []string) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetLeader(leader string) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetConsensusGroupSize(consensusGroupSize int) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SelfPubKey() string {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetSelfPubKey(selfPubKey string) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) JobDone(key string, subroundId int) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetJobDone(key string, subroundId int, value bool) error {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SelfJobDone(subroundId int) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) IsNodeInConsensusGroup(node string) bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) IsNodeInEligibleList(node string) bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) ComputeSize(subroundId int) int {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) ResetRoundState() {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) IsMultiKeyInConsensusGroup() bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) IsKeyManagedBySelf(pkBytes []byte) bool {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) IncrementRoundsWithoutReceivedMessages(pkBytes []byte) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) GetKeysHandler() consensus.KeysHandler {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) Leader() string {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) Status(subroundId int) spos.SubroundStatus {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetStatus(subroundId int, subroundStatus spos.SubroundStatus) {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) ResetRoundStatus() {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) Threshold(subroundId int) int {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) FallbackThreshold(subroundId int) int {
	// TODO implement me
	panic("implement me")
}

func (cnsm *ConsensusStateMock) SetFallbackThreshold(subroundId int, threshold int) {
	// TODO implement me
	panic("implement me")
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
	randomSource []byte,
	round uint64,
	shardId uint32,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	epoch uint32,
) (string, []string, error) {
	return cnsm.GetNextConsensusGroupCalled(randomSource, round, shardId, nodesCoordinator, epoch)
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
	if cnsm.IsSelfJobDoneCalled != nil {
		return cnsm.IsSelfJobDoneCalled(currentSubroundId)
	}
	return false
}

// IsCurrentSubroundFinished -
func (cnsm *ConsensusStateMock) IsCurrentSubroundFinished(currentSubroundId int) bool {
	if cnsm.IsCurrentSubroundFinishedCalled != nil {
		return cnsm.IsCurrentSubroundFinishedCalled(currentSubroundId)
	}
	return false
}

// IsNodeSelf -
func (cnsm *ConsensusStateMock) IsNodeSelf(node string) bool {
	if cnsm.IsNodeSelfCalled != nil {
		return cnsm.IsNodeSelfCalled(node)
	}
	return false
}

// IsBlockBodyAlreadyReceived -
func (cnsm *ConsensusStateMock) IsBlockBodyAlreadyReceived() bool {
	if cnsm.IsBlockBodyAlreadyReceivedCalled != nil {
		return cnsm.IsBlockBodyAlreadyReceivedCalled()
	}
	return false
}

// IsHeaderAlreadyReceived -
func (cnsm *ConsensusStateMock) IsHeaderAlreadyReceived() bool {
	if cnsm.IsHeaderAlreadyReceivedCalled != nil {
		return cnsm.IsHeaderAlreadyReceivedCalled()
	}
	return false
}

// CanDoSubroundJob -
func (cnsm *ConsensusStateMock) CanDoSubroundJob(currentSubroundId int) bool {
	if cnsm.CanDoSubroundJobCalled != nil {
		return cnsm.CanDoSubroundJobCalled(currentSubroundId)
	}
	return false
}

// CanProcessReceivedMessage -
func (cnsm *ConsensusStateMock) CanProcessReceivedMessage(
	cnsDta *consensus.Message,
	currentRoundIndex int64,
	currentSubroundId int,
) bool {
	return cnsm.CanProcessReceivedMessageCalled(cnsDta, currentRoundIndex, currentSubroundId)
}

// GenerateBitmap -
func (cnsm *ConsensusStateMock) GenerateBitmap(subroundId int) []byte {
	if cnsm.GenerateBitmapCalled != nil {
		return cnsm.GenerateBitmapCalled(subroundId)
	}
	return nil
}

// ProcessingBlock -
func (cnsm *ConsensusStateMock) ProcessingBlock() bool {
	if cnsm.ProcessingBlockCalled != nil {
		return cnsm.ProcessingBlockCalled()
	}
	return false
}

// SetProcessingBlock -
func (cnsm *ConsensusStateMock) SetProcessingBlock(processingBlock bool) {
	if cnsm.SetProcessingBlockCalled != nil {
		cnsm.SetProcessingBlockCalled(processingBlock)
	}
}

// ConsensusGroupSize -
func (cnsm *ConsensusStateMock) ConsensusGroupSize() int {
	if cnsm.ConsensusGroupSizeCalled != nil {
		return cnsm.ConsensusGroupSizeCalled()
	}
	return 0
}

// SetThreshold -
func (cnsm *ConsensusStateMock) SetThreshold(subroundId int, threshold int) {
	if cnsm.SetThresholdCalled != nil {
		cnsm.SetThresholdCalled(subroundId, threshold)
	}
}

func (cnsm *ConsensusStateMock) IsInterfaceNil() bool {
	return cnsm == nil
}
