package spos

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators/groupSelectors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

var log = logger.NewDefaultLogger()

// ConsensusState defines the data needed by spos to do the consensus in each round
type ConsensusState struct {
	// hold the data on which validators do the consensus (could be for example a hash of the block header
	// proposed by the leader)
	Data      []byte
	BlockBody *block.TxBlockBody
	Header    *block.Header

	RoundCanceled bool

	*RoundConsensus
	*RoundThreshold
	*RoundStatus
}

// NewConsensusState creates a new ConsensusState object
func NewConsensusState(
	roundConsensus *RoundConsensus,
	roundThreshold *RoundThreshold,
	roundStatus *RoundStatus,
) *ConsensusState {

	cns := ConsensusState{
		RoundConsensus: roundConsensus,
		RoundThreshold: roundThreshold,
		RoundStatus:    roundStatus,
	}

	return &cns
}

// ResetConsensusState method resets all the consensus data
func (cns *ConsensusState) ResetConsensusState() {
	cns.BlockBody = nil
	cns.Header = nil
	cns.Data = nil

	cns.RoundCanceled = false

	cns.ResetRoundStatus()
	cns.ResetRoundState()
}

// IsNodeLeaderInCurrentRound method checks if the given node is leader in the current round
func (cns *ConsensusState) IsNodeLeaderInCurrentRound(node string) bool {
	leader, err := cns.GetLeader()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return leader == node
}

// IsSelfLeaderInCurrentRound method checks if the current node is leader in the current round
func (cns *ConsensusState) IsSelfLeaderInCurrentRound() bool {
	return cns.IsNodeLeaderInCurrentRound(cns.selfPubKey)
}

// GetLeader method gets the leader of the current round
func (cns *ConsensusState) GetLeader() (string, error) {
	if cns.consensusGroup == nil {
		return "", ErrNilConsensusGroup
	}

	if len(cns.consensusGroup) == 0 {
		return "", ErrEmptyConsensusGroup
	}

	return cns.consensusGroup[0], nil
}

// GetNextConsensusGroup gets the new consensus group for the current round based on current eligible list and a random
// source for the new selection
func (cns *ConsensusState) GetNextConsensusGroup(randomSource string, vgs groupSelectors.ValidatorGroupSelector) ([]string, error) {
	validatorsGroup, err := vgs.ComputeValidatorsGroup([]byte(randomSource))

	if err != nil {
		return nil, err
	}

	newConsensusGroup := make([]string, 0)

	for i := 0; i < len(validatorsGroup); i++ {
		newConsensusGroup = append(newConsensusGroup, string(validatorsGroup[i].PubKey()))
	}

	return newConsensusGroup, nil
}

// IsConsensusDataNotSet method returns true if consensus data for the current subround is not set and false otherwise
func (cns *ConsensusState) IsConsensusDataNotSet() bool {
	isConsensusDataNotSet := cns.Data == nil

	return isConsensusDataNotSet
}

// IsConsensusDataAlreadySet method returns true if consensus data for the current subround is already set and false otherwise
func (cns *ConsensusState) IsConsensusDataAlreadySet() bool {
	isConsensusDataAlreadySet := cns.Data != nil

	return isConsensusDataAlreadySet
}

// IsJobDone method returns true if the node job for the current subround is done and false otherwise
func (cns *ConsensusState) IsJobDone(node string, currentSubroundId int) bool {
	jobDone, err := cns.GetJobDone(node, currentSubroundId)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return jobDone
}

// IsSelfJobDone method returns true if self job for the current subround is done and false otherwise
func (cns *ConsensusState) IsSelfJobDone(currentSubroundId int) bool {
	return cns.IsJobDone(cns.selfPubKey, currentSubroundId)
}

// IsCurrentSubroundFinished method returns true if the current subround is finished and false otherwise
func (cns *ConsensusState) IsCurrentSubroundFinished(currentSubroundId int) bool {
	isCurrentSubroundFinished := cns.Status(currentSubroundId) == SsFinished

	return isCurrentSubroundFinished
}

// IsMessageReceivedFromItself method returns true if the message is received from itself and false otherwise
func (cns *ConsensusState) IsMessageReceivedFromItself(node string) bool {
	isMessageReceivedFromItself := node == cns.SelfPubKey()

	return isMessageReceivedFromItself
}

// IsMessageReceivedTooLate method returns true if the message is received too late and false otherwise
func (cns *ConsensusState) IsMessageReceivedTooLate() bool {
	// TODO: This method should implement, if necessary, a mechanism of detecting the late messages received
	return false
}

// IsMessageReceivedForOtherRound method returns true if the message received is for other round and false otherwise
func (cns *ConsensusState) IsMessageReceivedForOtherRound(currentRoundIndex int32, dataRoundIndex int32) bool {
	isMessageReceivedForOtherRound := currentRoundIndex != dataRoundIndex

	return isMessageReceivedForOtherRound
}

// IsBlockBodyAlreadyReceived method returns true if block body is already received and false otherwise
func (cns *ConsensusState) IsBlockBodyAlreadyReceived() bool {
	isBlockBodyAlreadyReceived := cns.BlockBody != nil

	return isBlockBodyAlreadyReceived
}

// IsHeaderAlreadyReceived method returns true if header is already received and false otherwise
func (cns *ConsensusState) IsHeaderAlreadyReceived() bool {
	isHeaderAlreadyReceived := cns.Header != nil

	return isHeaderAlreadyReceived
}

// CanDoSubroundJob method returns true if the job of the subround can be done and false otherwise
func (cns *ConsensusState) CanDoSubroundJob(currentSubroundId int) bool {
	if cns.IsConsensusDataNotSet() {
		return false
	}

	if cns.IsSelfJobDone(currentSubroundId) {
		return false
	}

	if cns.IsCurrentSubroundFinished(currentSubroundId) {
		return false
	}

	return true
}

// CanProcessReceivedMessage method returns true if the message received can be processed and false otherwise
func (cns *ConsensusState) CanProcessReceivedMessage(cnsDta *ConsensusData, currentRoundIndex int32, currentSubroundId int) bool {
	if cns.IsMessageReceivedFromItself(string(cnsDta.PubKey)) {
		return false
	}

	if cns.IsMessageReceivedTooLate() {
		return false
	}

	if cns.IsMessageReceivedForOtherRound(currentRoundIndex, cnsDta.RoundIndex) {
		return false
	}

	if cns.IsJobDone(string(cnsDta.PubKey), currentSubroundId) {
		return false
	}

	if cns.IsCurrentSubroundFinished(currentSubroundId) {
		return false
	}

	return true
}

// GenerateBitmap method generates a bitmap, for a given subround, in which each node will be marked with 1
// if its job has been done
func (cns *ConsensusState) GenerateBitmap(subroundId int) []byte {
	// generate bitmap according to set commitment hashes
	sizeConsensus := len(cns.ConsensusGroup())

	bitmap := make([]byte, sizeConsensus/8+1)

	for i := 0; i < sizeConsensus; i++ {
		pubKey := cns.ConsensusGroup()[i]
		isJobDone, err := cns.GetJobDone(pubKey, subroundId)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobDone {
			bitmap[i/8] |= 1 << (uint16(i) % 8)
		}
	}

	return bitmap
}
