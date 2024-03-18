package spos

import (
	"bytes"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// IndexOfLeaderInConsensusGroup represents the index of the leader in the consensus group
const IndexOfLeaderInConsensusGroup = 0

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

	receivedMessagesWithSignature    map[string]p2p.MessageP2P
	mutReceivedMessagesWithSignature sync.RWMutex

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
	cns.initReceivedMessagesWithSig()

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

func (cns *ConsensusState) initReceivedMessagesWithSig() {
	cns.mutReceivedMessagesWithSignature.Lock()
	cns.receivedMessagesWithSignature = make(map[string]p2p.MessageP2P)
	cns.mutReceivedMessagesWithSignature.Unlock()
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

// AddMessageWithSignature will add the p2p message to received list of messages
func (cns *ConsensusState) AddMessageWithSignature(key string, message p2p.MessageP2P) {
	cns.mutReceivedMessagesWithSignature.Lock()
	cns.receivedMessagesWithSignature[key] = message
	cns.mutReceivedMessagesWithSignature.Unlock()
}

// GetMessageWithSignature will get the p2p message based on key
func (cns *ConsensusState) GetMessageWithSignature(key string) (p2p.MessageP2P, bool) {
	cns.mutReceivedMessagesWithSignature.RLock()
	defer cns.mutReceivedMessagesWithSignature.RUnlock()

	val, ok := cns.receivedMessagesWithSignature[key]
	return val, ok
}

// IsNodeLeaderInCurrentRound method checks if the given node is leader in the current round
func (cns *ConsensusState) IsNodeLeaderInCurrentRound(node string) bool {
	leader, err := cns.GetLeader()
	if err != nil {
		log.Debug("IsNodeLeaderInCurrentRound.GetLeader", "error", err.Error())
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

	return cns.consensusGroup[IndexOfLeaderInConsensusGroup], nil
}

// GetNextConsensusGroup gets the new consensus group for the current round based on current eligible list and a random
// source for the new selection
func (cns *ConsensusState) GetNextConsensusGroup(
	randomSource []byte,
	round uint64,
	shardId uint32,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
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

	selfJobDone := true
	if cns.IsNodeInConsensusGroup(cns.SelfPubKey()) {
		selfJobDone = cns.IsSelfJobDone(currentSubroundId)
	}
	multiKeyJobDone := true
	if cns.IsMultiKeyInConsensusGroup() {
		multiKeyJobDone = cns.IsMultiKeyJobDone(currentSubroundId)
	}

	if selfJobDone && multiKeyJobDone {
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

// IsMultiKeyLeaderInCurrentRound method checks if one of the nodes which are controlled by this instance
// is leader in the current round
func (cns *ConsensusState) IsMultiKeyLeaderInCurrentRound() bool {
	leader, err := cns.GetLeader()
	if err != nil {
		log.Debug("IsMultiKeyLeaderInCurrentRound.GetLeader", "error", err.Error())
		return false
	}

	return cns.IsKeyManagedByCurrentNode([]byte(leader))
}

// IsLeaderJobDone method returns true if the leader job for the current subround is done and false otherwise
func (cns *ConsensusState) IsLeaderJobDone(currentSubroundId int) bool {
	leader, err := cns.GetLeader()
	if err != nil {
		log.Debug("GetLeader", "error", err.Error())
		return false
	}

	return cns.IsJobDone(leader, currentSubroundId)
}

// IsMultiKeyJobDone method returns true if all the nodes controlled by this instance finished the current job for
// the current subround and false otherwise
func (cns *ConsensusState) IsMultiKeyJobDone(currentSubroundId int) bool {
	for _, validator := range cns.consensusGroup {
		if !cns.keysHandler.IsKeyManagedByCurrentNode([]byte(validator)) {
			continue
		}

		if !cns.IsJobDone(validator, currentSubroundId) {
			return false
		}
	}

	return true
}

// GetMultikeyRedundancyStepInReason returns the reason if the current node stepped in as a multikey redundancy node
func (cns *ConsensusState) GetMultikeyRedundancyStepInReason() string {
	return cns.keysHandler.GetRedundancyStepInReason()
}

// ResetRoundsWithoutReceivedMessages will reset the rounds received without a message for a specified public key by
// providing also the peer ID from the received message
func (cns *ConsensusState) ResetRoundsWithoutReceivedMessages(pkBytes []byte, pid core.PeerID) {
	cns.keysHandler.ResetRoundsWithoutReceivedMessages(pkBytes, pid)
}
