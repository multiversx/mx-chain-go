package consensus

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/consensus"
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
	StatusCalled                                 func(subroundId int) int
	SetStatusCalled                              func(subroundId int, subroundStatus int)
	ResetRoundStatusCalled                       func()
	ThresholdCalled                              func(subroundId int) int
	FallbackThresholdCalled                      func(subroundId int) int
	SetFallbackThresholdCalled                   func(subroundId int, threshold int)
	ResetConsensusRoundStateCalled               func()
}

// AddReceivedHeader -
func (cnsm *ConsensusStateMock) AddReceivedHeader(headerHandler data.HeaderHandler) {
	if cnsm.AddReceivedHeaderCalled != nil {
		cnsm.AddReceivedHeaderCalled(headerHandler)
	}
}

// GetReceivedHeaders -
func (cnsm *ConsensusStateMock) GetReceivedHeaders() []data.HeaderHandler {
	if cnsm.GetReceivedHeadersCalled != nil {
		return cnsm.GetReceivedHeadersCalled()
	}
	return nil
}

// AddMessageWithSignature -
func (cnsm *ConsensusStateMock) AddMessageWithSignature(key string, message p2p.MessageP2P) {
	if cnsm.AddMessageWithSignatureCalled != nil {
		cnsm.AddMessageWithSignatureCalled(key, message)
	}
}

// GetMessageWithSignature -
func (cnsm *ConsensusStateMock) GetMessageWithSignature(key string) (p2p.MessageP2P, bool) {
	if cnsm.GetMessageWithSignatureCalled != nil {
		return cnsm.GetMessageWithSignatureCalled(key)
	}
	return nil, false
}

// IsSubroundFinished -
func (cnsm *ConsensusStateMock) IsSubroundFinished(subroundID int) bool {
	if cnsm.IsSubroundFinishedCalled != nil {
		return cnsm.IsSubroundFinishedCalled(subroundID)
	}
	return false
}

// GetData -
func (cnsm *ConsensusStateMock) GetData() []byte {
	if cnsm.GetDataCalled != nil {
		return cnsm.GetDataCalled()
	}
	return nil
}

// SetData -
func (cnsm *ConsensusStateMock) SetData(data []byte) {
	if cnsm.SetDataCalled != nil {
		cnsm.SetDataCalled(data)
	}
}

// IsMultiKeyLeaderInCurrentRound -
func (cnsm *ConsensusStateMock) IsMultiKeyLeaderInCurrentRound() bool {
	if cnsm.IsMultiKeyLeaderInCurrentRoundCalled != nil {
		return cnsm.IsMultiKeyLeaderInCurrentRoundCalled()
	}
	return false
}

// IsLeaderJobDone -
func (cnsm *ConsensusStateMock) IsLeaderJobDone(currentSubroundId int) bool {
	if cnsm.IsLeaderJobDoneCalled != nil {
		return cnsm.IsLeaderJobDoneCalled(currentSubroundId)
	}
	return false
}

// IsMultiKeyJobDone -
func (cnsm *ConsensusStateMock) IsMultiKeyJobDone(currentSubroundId int) bool {
	if cnsm.IsMultiKeyJobDoneCalled != nil {
		return cnsm.IsMultiKeyJobDoneCalled(currentSubroundId)
	}
	return false
}

// GetMultikeyRedundancyStepInReason -
func (cnsm *ConsensusStateMock) GetMultikeyRedundancyStepInReason() string {
	if cnsm.GetMultikeyRedundancyStepInReasonCalled != nil {
		return cnsm.GetMultikeyRedundancyStepInReasonCalled()
	}
	return ""
}

// ResetRoundsWithoutReceivedMessages -
func (cnsm *ConsensusStateMock) ResetRoundsWithoutReceivedMessages(pkBytes []byte, pid core.PeerID) {
	if cnsm.ResetRoundsWithoutReceivedMessagesCalled != nil {
		cnsm.ResetRoundsWithoutReceivedMessagesCalled(pkBytes, pid)
	}
}

// GetRoundCanceled -
func (cnsm *ConsensusStateMock) GetRoundCanceled() bool {
	if cnsm.GetRoundCanceledCalled != nil {
		return cnsm.GetRoundCanceledCalled()
	}
	return false
}

// SetRoundCanceled -
func (cnsm *ConsensusStateMock) SetRoundCanceled(state bool) {
	if cnsm.SetRoundCanceledCalled != nil {
		cnsm.SetRoundCanceledCalled(state)
	}
}

// GetRoundIndex -
func (cnsm *ConsensusStateMock) GetRoundIndex() int64 {
	if cnsm.GetRoundIndexCalled != nil {
		return cnsm.GetRoundIndexCalled()
	}
	return 0
}

// SetRoundIndex -
func (cnsm *ConsensusStateMock) SetRoundIndex(roundIndex int64) {
	if cnsm.SetRoundIndexCalled != nil {
		cnsm.SetRoundIndexCalled(roundIndex)
	}
}

// GetRoundTimeStamp -
func (cnsm *ConsensusStateMock) GetRoundTimeStamp() time.Time {
	if cnsm.GetRoundTimeStampCalled != nil {
		return cnsm.GetRoundTimeStampCalled()
	}
	return time.Time{}
}

// SetRoundTimeStamp -
func (cnsm *ConsensusStateMock) SetRoundTimeStamp(roundTimeStamp time.Time) {
	if cnsm.SetRoundTimeStampCalled != nil {
		cnsm.SetRoundTimeStampCalled(roundTimeStamp)
	}
}

// GetExtendedCalled -
func (cnsm *ConsensusStateMock) GetExtendedCalled() bool {
	if cnsm.GetExtendedCalledCalled != nil {
		return cnsm.GetExtendedCalledCalled()
	}
	return false
}

// GetBody -
func (cnsm *ConsensusStateMock) GetBody() data.BodyHandler {
	if cnsm.GetBodyCalled != nil {
		return cnsm.GetBodyCalled()
	}
	return nil
}

// SetBody -
func (cnsm *ConsensusStateMock) SetBody(body data.BodyHandler) {
	if cnsm.SetBodyCalled != nil {
		cnsm.SetBodyCalled(body)
	}
}

// GetHeader -
func (cnsm *ConsensusStateMock) GetHeader() data.HeaderHandler {
	if cnsm.GetHeaderCalled != nil {
		return cnsm.GetHeaderCalled()
	}
	return nil
}

// SetHeader -
func (cnsm *ConsensusStateMock) SetHeader(header data.HeaderHandler) {
	if cnsm.SetHeaderCalled != nil {
		cnsm.SetHeaderCalled(header)
	}
}

// GetWaitingAllSignaturesTimeOut -
func (cnsm *ConsensusStateMock) GetWaitingAllSignaturesTimeOut() bool {
	if cnsm.GetWaitingAllSignaturesTimeOutCalled != nil {
		return cnsm.GetWaitingAllSignaturesTimeOutCalled()
	}
	return false
}

// SetWaitingAllSignaturesTimeOut -
func (cnsm *ConsensusStateMock) SetWaitingAllSignaturesTimeOut(b bool) {
	if cnsm.SetWaitingAllSignaturesTimeOutCalled != nil {
		cnsm.SetWaitingAllSignaturesTimeOutCalled(b)
	}
}

// ConsensusGroupIndex -
func (cnsm *ConsensusStateMock) ConsensusGroupIndex(pubKey string) (int, error) {
	if cnsm.ConsensusGroupIndexCalled != nil {
		return cnsm.ConsensusGroupIndexCalled(pubKey)
	}
	return 0, nil
}

// SelfConsensusGroupIndex -
func (cnsm *ConsensusStateMock) SelfConsensusGroupIndex() (int, error) {
	if cnsm.SelfConsensusGroupIndexCalled != nil {
		return cnsm.SelfConsensusGroupIndexCalled()
	}
	return 0, nil
}

// SetEligibleList -
func (cnsm *ConsensusStateMock) SetEligibleList(eligibleList map[string]struct{}) {
	if cnsm.SetEligibleListCalled != nil {
		cnsm.SetEligibleListCalled(eligibleList)
	}
}

// ConsensusGroup -
func (cnsm *ConsensusStateMock) ConsensusGroup() []string {
	if cnsm.ConsensusGroupCalled != nil {
		return cnsm.ConsensusGroupCalled()
	}
	return nil
}

// SetConsensusGroup -
func (cnsm *ConsensusStateMock) SetConsensusGroup(consensusGroup []string) {
	if cnsm.SetConsensusGroupCalled != nil {
		cnsm.SetConsensusGroupCalled(consensusGroup)
	}
}

// SetLeader -
func (cnsm *ConsensusStateMock) SetLeader(leader string) {
	if cnsm.SetLeaderCalled != nil {
		cnsm.SetLeaderCalled(leader)
	}
}

// SetConsensusGroupSize -
func (cnsm *ConsensusStateMock) SetConsensusGroupSize(consensusGroupSize int) {
	if cnsm.SetConsensusGroupSizeCalled != nil {
		cnsm.SetConsensusGroupSizeCalled(consensusGroupSize)
	}
}

// SelfPubKey -
func (cnsm *ConsensusStateMock) SelfPubKey() string {
	if cnsm.SelfPubKeyCalled != nil {
		return cnsm.SelfPubKeyCalled()
	}
	return ""
}

// SetSelfPubKey -
func (cnsm *ConsensusStateMock) SetSelfPubKey(selfPubKey string) {
	if cnsm.SetSelfPubKeyCalled != nil {
		cnsm.SetSelfPubKeyCalled(selfPubKey)
	}
}

// JobDone -
func (cnsm *ConsensusStateMock) JobDone(key string, subroundId int) (bool, error) {
	if cnsm.JobDoneCalled != nil {
		return cnsm.JobDoneCalled(key, subroundId)
	}
	return false, nil
}

// SetJobDone -
func (cnsm *ConsensusStateMock) SetJobDone(key string, subroundId int, value bool) error {
	if cnsm.SetJobDoneCalled != nil {
		return cnsm.SetJobDoneCalled(key, subroundId, value)
	}
	return nil
}

// SelfJobDone -
func (cnsm *ConsensusStateMock) SelfJobDone(subroundId int) (bool, error) {
	if cnsm.SelfJobDoneCalled != nil {
		return cnsm.SelfJobDoneCalled(subroundId)
	}
	return false, nil
}

// IsNodeInConsensusGroup -
func (cnsm *ConsensusStateMock) IsNodeInConsensusGroup(node string) bool {
	if cnsm.IsNodeInConsensusGroupCalled != nil {
		return cnsm.IsNodeInConsensusGroupCalled(node)
	}
	return false
}

// IsNodeInEligibleList -
func (cnsm *ConsensusStateMock) IsNodeInEligibleList(node string) bool {
	if cnsm.IsNodeInEligibleListCalled != nil {
		return cnsm.IsNodeInEligibleListCalled(node)
	}
	return false
}

// ComputeSize -
func (cnsm *ConsensusStateMock) ComputeSize(subroundId int) int {
	if cnsm.ComputeSizeCalled != nil {
		return cnsm.ComputeSizeCalled(subroundId)
	}
	return 0
}

// ResetRoundState -
func (cnsm *ConsensusStateMock) ResetRoundState() {
	if cnsm.ResetRoundStateCalled != nil {
		cnsm.ResetRoundStateCalled()
	}
}

// IsMultiKeyInConsensusGroup -
func (cnsm *ConsensusStateMock) IsMultiKeyInConsensusGroup() bool {
	if cnsm.IsMultiKeyInConsensusGroupCalled != nil {
		return cnsm.IsMultiKeyInConsensusGroupCalled()
	}
	return false
}

// IsKeyManagedBySelf -
func (cnsm *ConsensusStateMock) IsKeyManagedBySelf(pkBytes []byte) bool {
	if cnsm.IsKeyManagedBySelfCalled != nil {
		return cnsm.IsKeyManagedBySelfCalled(pkBytes)
	}
	return false
}

// IncrementRoundsWithoutReceivedMessages -
func (cnsm *ConsensusStateMock) IncrementRoundsWithoutReceivedMessages(pkBytes []byte) {
	if cnsm.IncrementRoundsWithoutReceivedMessagesCalled != nil {
		cnsm.IncrementRoundsWithoutReceivedMessagesCalled(pkBytes)
	}
}

// GetKeysHandler -
func (cnsm *ConsensusStateMock) GetKeysHandler() consensus.KeysHandler {
	if cnsm.GetKeysHandlerCalled != nil {
		return cnsm.GetKeysHandlerCalled()
	}
	return nil
}

// Leader -
func (cnsm *ConsensusStateMock) Leader() string {
	if cnsm.LeaderCalled != nil {
		return cnsm.LeaderCalled()
	}
	return ""
}

// Status -
func (cnsm *ConsensusStateMock) Status(subroundId int) int {
	if cnsm.StatusCalled != nil {
		return cnsm.StatusCalled(subroundId)
	}
	return 0
}

// SetStatus -
func (cnsm *ConsensusStateMock) SetStatus(subroundId int, subroundStatus int) {
	if cnsm.SetStatusCalled != nil {
		cnsm.SetStatusCalled(subroundId, subroundStatus)
	}
}

// ResetRoundStatus -
func (cnsm *ConsensusStateMock) ResetRoundStatus() {
	if cnsm.ResetRoundStatusCalled != nil {
		cnsm.ResetRoundStatusCalled()
	}
}

// Threshold -
func (cnsm *ConsensusStateMock) Threshold(subroundId int) int {
	if cnsm.ThresholdCalled != nil {
		return cnsm.ThresholdCalled(subroundId)
	}
	return 0
}

// FallbackThreshold -
func (cnsm *ConsensusStateMock) FallbackThreshold(subroundId int) int {
	if cnsm.FallbackThresholdCalled != nil {
		return cnsm.FallbackThresholdCalled(subroundId)
	}
	return 0
}

func (cnsm *ConsensusStateMock) SetFallbackThreshold(subroundId int, threshold int) {
	if cnsm.SetFallbackThresholdCalled != nil {
		cnsm.SetFallbackThresholdCalled(subroundId, threshold)
	}
}

// ResetConsensusState -
func (cnsm *ConsensusStateMock) ResetConsensusState() {
	if cnsm.ResetConsensusStateCalled != nil {
		cnsm.ResetConsensusStateCalled()
	}
}

// ResetConsensusRoundState -
func (cnsm *ConsensusStateMock) ResetConsensusRoundState() {
	if cnsm.ResetConsensusRoundStateCalled != nil {
		cnsm.ResetConsensusRoundStateCalled()
	}
}

// IsNodeLeaderInCurrentRound -
func (cnsm *ConsensusStateMock) IsNodeLeaderInCurrentRound(node string) bool {
	if cnsm.IsNodeLeaderInCurrentRoundCalled != nil {
		return cnsm.IsNodeLeaderInCurrentRoundCalled(node)
	}
	return false
}

// IsSelfLeaderInCurrentRound -
func (cnsm *ConsensusStateMock) IsSelfLeaderInCurrentRound() bool {
	if cnsm.IsSelfLeaderInCurrentRoundCalled != nil {
		return cnsm.IsSelfLeaderInCurrentRoundCalled()
	}
	return false
}

// GetLeader -
func (cnsm *ConsensusStateMock) GetLeader() (string, error) {
	if cnsm.GetLeaderCalled != nil {
		return cnsm.GetLeaderCalled()
	}
	return "", nil
}

// GetNextConsensusGroup -
func (cnsm *ConsensusStateMock) GetNextConsensusGroup(
	randomSource []byte,
	round uint64,
	shardId uint32,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	epoch uint32,
) (string, []string, error) {
	if cnsm.GetNextConsensusGroupCalled != nil {
		return cnsm.GetNextConsensusGroupCalled(randomSource, round, shardId, nodesCoordinator, epoch)
	}
	return "", nil, nil
}

// IsConsensusDataSet -
func (cnsm *ConsensusStateMock) IsConsensusDataSet() bool {
	if cnsm.IsConsensusDataSetCalled != nil {
		return cnsm.IsConsensusDataSetCalled()
	}
	return false
}

// IsConsensusDataEqual -
func (cnsm *ConsensusStateMock) IsConsensusDataEqual(data []byte) bool {
	if cnsm.IsConsensusDataEqualCalled != nil {
		return cnsm.IsConsensusDataEqualCalled(data)
	}
	return false
}

// IsJobDone -
func (cnsm *ConsensusStateMock) IsJobDone(node string, currentSubroundId int) bool {
	if cnsm.IsJobDoneCalled != nil {
		return cnsm.IsJobDoneCalled(node, currentSubroundId)
	}
	return false
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

// IsInterfaceNil returns true if there is no value under the interface
func (cnsm *ConsensusStateMock) IsInterfaceNil() bool {
	return cnsm == nil
}
