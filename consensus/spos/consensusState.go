package spos

import (
	"bytes"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("consensus/spos")

// ConsensusState defines the data needed by spos to do the consensus in each round
type ConsensusState struct {
	// hold the data on which validators do the consensus (could be for example a hash of the block header
	// proposed by the leader)
	Data   []byte
	Body   data.BodyHandler
	Header data.HeaderHandler

	receivedHeaders    []data.HeaderHandler
	mutReceivedHeaders sync.RWMutex

	RoundIndex                  int64
	RoundTimeStamp              time.Time
	RoundCanceled               bool
	ExtendedCalled              bool
	WaitingAllSignaturesTimeOut bool

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
	cns.Body = nil
	cns.Header = nil
	cns.Data = nil

	cns.initReceivedHeaders()

	cns.RoundCanceled = false
	cns.ExtendedCalled = false
	cns.WaitingAllSignaturesTimeOut = false

	cns.ResetRoundStatus()
	cns.ResetRoundState()
}

func (cns *ConsensusState) initReceivedHeaders() {
	cns.mutReceivedHeaders.Lock()
	cns.receivedHeaders = make([]data.HeaderHandler, 0)
	cns.mutReceivedHeaders.Unlock()
}

// AddReceivedHeader append the provided header to the inner received headers list
func (cns *ConsensusState) AddReceivedHeader(headerHandler data.HeaderHandler) {
	cns.mutReceivedHeaders.Lock()
	cns.receivedHeaders = append(cns.receivedHeaders, headerHandler)
	cns.mutReceivedHeaders.Unlock()
}

// GetReceivedHeaders returns the received headers list
func (cns *ConsensusState) GetReceivedHeaders() []data.HeaderHandler {
	cns.mutReceivedHeaders.RLock()
	receivedHeaders := cns.receivedHeaders
	cns.mutReceivedHeaders.RUnlock()

	return receivedHeaders
}

// IsNodeLeaderInCurrentRound method checks if the given node is leader in the current round
func (cns *ConsensusState) IsNodeLeaderInCurrentRound(node string) bool {
	leader, err := cns.GetLeader()
	if err != nil {
		log.Debug("GetLeader", "error", err.Error())
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
func (cns *ConsensusState) GetNextConsensusGroup(
	randomSource []byte,
	round uint64,
	shardId uint32,
	nodesCoordinator sharding.NodesCoordinator,
	epoch uint32,
) ([]string, error) {
	validatorsGroup, err := nodesCoordinator.ComputeConsensusGroup(randomSource, round, shardId, epoch)
	if err != nil {
		log.Debug(
			"compute consensus group",
			"error", err.Error(),
			"randomSource", randomSource,
			"round", round,
			"shardId", shardId,
			"epoch", epoch,
		)
		return nil, err
	}

	consensusSize := len(validatorsGroup)
	newConsensusGroup := make([]string, consensusSize)

	for i := 0; i < consensusSize; i++ {
		newConsensusGroup[i] = string(validatorsGroup[i].PubKey())
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
		log.Debug("JobDone", "error", err.Error())
		return false
	}

	return jobDone
}

// IsSelfJobDone method returns true if self job for the current subround is done and false otherwise
func (cns *ConsensusState) IsSelfJobDone(currentSubroundId int) bool {
	return cns.IsJobDone(cns.selfPubKey, currentSubroundId)
}

// IsSubroundFinished method returns true if the current subround is finished and false otherwise
func (cns *ConsensusState) IsSubroundFinished(subroundID int) bool {
	isSubroundFinished := cns.Status(subroundID) == SsFinished

	return isSubroundFinished
}

// IsNodeSelf method returns true if the message is received from itself and false otherwise
func (cns *ConsensusState) IsNodeSelf(node string) bool {
	isNodeSelf := node == cns.SelfPubKey()

	return isNodeSelf
}

// IsBlockBodyAlreadyReceived method returns true if block body is already received and false otherwise
func (cns *ConsensusState) IsBlockBodyAlreadyReceived() bool {
	isBlockBodyAlreadyReceived := cns.Body != nil

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

	if cns.IsSubroundFinished(currentSubroundId) {
		return false
	}

	return true
}

// CanProcessReceivedMessage method returns true if the message received can be processed and false otherwise
func (cns *ConsensusState) CanProcessReceivedMessage(cnsDta *consensus.Message, currentRoundIndex int64,
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

	if cns.IsSubroundFinished(currentSubroundId) {
		return false
	}

	return true
}

// GenerateBitmap method generates a bitmap, for a given subround, in which each node will be marked with 1
// if its job has been done
func (cns *ConsensusState) GenerateBitmap(subroundId int) []byte {
	// generate bitmap according to set commitment hashes
	sizeConsensus := len(cns.ConsensusGroup())

	bitmapSize := sizeConsensus / 8
	if sizeConsensus%8 != 0 {
		bitmapSize++
	}
	bitmap := make([]byte, bitmapSize)

	for i := 0; i < sizeConsensus; i++ {
		pubKey := cns.ConsensusGroup()[i]
		isJobDone, err := cns.JobDone(pubKey, subroundId)
		if err != nil {
			log.Debug("JobDone", "error", err.Error())
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
