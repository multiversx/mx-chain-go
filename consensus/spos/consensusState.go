package spos

import (
	"bytes"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
)

var log = logger.DefaultLogger()

// ConsensusState defines the data needed by spos to do the consensus in each round
type ConsensusState struct {
	// hold the data on which validators do the consensus (could be for example a hash of the block header
	// proposed by the leader)
	Data      []byte
	BlockBody data.BodyHandler
	Header    data.HeaderHandler

	RoundIndex     int32
	RoundTimeStamp time.Time
	RoundCanceled  bool

	processingBlock    bool
	mutProcessingBlock sync.RWMutex

	*roundConsensus
	*roundThreshold
	*roundStatus
}

// NewConsensusState creates a new ConsensusState object
func NewConsensusState(
	roundConsensus *roundConsensus,
	roundThreshold *roundThreshold,
	roundStatus *roundStatus,
) *ConsensusState {

	cns := ConsensusState{
		roundConsensus: roundConsensus,
		roundThreshold: roundThreshold,
		roundStatus:    roundStatus,
	}

	cns.ResetConsensusState()

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
func (cns *ConsensusState) GetNextConsensusGroup(randomSource string, vgs consensus.ValidatorGroupSelector) ([]string,
	error) {
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

// IsConsensusDataSet method returns true if the consensus data for the current round is set and false otherwise
func (cns *ConsensusState) IsConsensusDataSet() bool {
	isConsensusDataSet := cns.Data != nil

	return isConsensusDataSet
}

// IsConsensusDataEqual method returns true if the consensus data for the current round is the same with the given
// one and false otherwise
func (cns *ConsensusState) IsConsensusDataEqual(data []byte) bool {
	isConsensusDataEqual := bytes.Equal(cns.Data, data)

	return isConsensusDataEqual
}

// IsJobDone method returns true if the node job for the current subround is done and false otherwise
func (cns *ConsensusState) IsJobDone(node string, currentSubroundId int) bool {
	jobDone, err := cns.JobDone(node, currentSubroundId)

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

// IsNodeSelf method returns true if the message is received from itself and false otherwise
func (cns *ConsensusState) IsNodeSelf(node string) bool {
	isNodeSelf := node == cns.SelfPubKey()

	return isNodeSelf
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
	if !cns.IsConsensusDataSet() {
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
func (cns *ConsensusState) CanProcessReceivedMessage(cnsDta *consensus.Message, currentRoundIndex int32,
	currentSubroundId int) bool {
	if cns.IsNodeSelf(string(cnsDta.PubKey)) {
		return false
	}

	if currentRoundIndex != cnsDta.RoundIndex {
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
		isJobDone, err := cns.JobDone(pubKey, subroundId)

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

// ProcessingBlock gets the state of block processing
func (cns *ConsensusState) ProcessingBlock() bool {
	cns.mutProcessingBlock.RLock()
	processingBlock := cns.processingBlock
	cns.mutProcessingBlock.RUnlock()
	return processingBlock
}

// SetProcessingBlock sets the state of block processing
func (cns *ConsensusState) SetProcessingBlock(processingBlock bool) {
	cns.mutProcessingBlock.Lock()
	cns.processingBlock = processingBlock
	cns.mutProcessingBlock.Unlock()
}

// GetData gets the Data of the consensusState
func (cns *ConsensusState) GetData() []byte {
	return cns.Data
}
