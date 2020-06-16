package floodPreventers

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgQuotaFloodPreventer defines the arguments for a quota flood preventer
type ArgQuotaFloodPreventer struct {
	Name                      string
	Cacher                    storage.Cacher
	StatusHandlers            []QuotaStatusHandler
	BaseMaxNumMessagesPerPeer uint32
	MaxTotalSizePerPeer       uint64
	PercentReserved           float32
	IncreaseThreshold         uint32
	IncreaseFactor            float32
	SelfPid                   core.PeerID
}

var _ process.FloodPreventer = (*quotaFloodPreventer)(nil)

const minMessages = 1
const minTotalSize = 1 //1Byte
const initNumMessages = 1
const maxPercentReserved = 90.0
const minPercentReserved = 0.0
const quotaStructSize = 24

type quota struct {
	numReceivedMessages   uint32
	numProcessedMessages  uint32
	sizeReceivedMessages  uint64
	sizeProcessedMessages uint64
}

// Size returns the size of a quota object
func (q *quota) Size() int {
	return quotaStructSize
}

// quotaFloodPreventer represents a cache of quotas per peer used in antiflooding mechanism
type quotaFloodPreventer struct {
	name                          string
	mutOperation                  sync.RWMutex
	cacher                        storage.Cacher
	statusHandlers                []QuotaStatusHandler
	computedMaxNumMessagesPerPeer uint32
	baseMaxNumMessagesPerPeer     uint32
	maxTotalSizePerPeer           uint64
	percentReserved               float32
	increaseThreshold             uint32
	increaseFactor                float32
	selfPid                       core.PeerID
}

// NewQuotaFloodPreventer creates a new flood preventer based on quota / peer
func NewQuotaFloodPreventer(arg ArgQuotaFloodPreventer) (*quotaFloodPreventer, error) {

	if check.IfNil(arg.Cacher) {
		return nil, process.ErrNilCacher
	}
	for _, statusHandler := range arg.StatusHandlers {
		if check.IfNil(statusHandler) {
			return nil, process.ErrNilQuotaStatusHandler
		}
	}
	if arg.BaseMaxNumMessagesPerPeer < minMessages {
		return nil, fmt.Errorf("%w, maxMessagesPerPeer: provided %d, minimum %d",
			process.ErrInvalidValue,
			arg.BaseMaxNumMessagesPerPeer,
			minMessages,
		)
	}
	if arg.MaxTotalSizePerPeer < minTotalSize {
		return nil, fmt.Errorf("%w, maxTotalSizePerPeer: provided %d, minimum %d",
			process.ErrInvalidValue,
			arg.MaxTotalSizePerPeer,
			minTotalSize,
		)
	}
	if arg.PercentReserved > maxPercentReserved {
		return nil, fmt.Errorf("%w, percentReserved: provided %0.3f, maximum %0.3f",
			process.ErrInvalidValue,
			arg.PercentReserved,
			maxPercentReserved,
		)
	}
	if arg.PercentReserved < minPercentReserved {
		return nil, fmt.Errorf("%w, percentReserved: provided %0.3f, minimum %0.3f",
			process.ErrInvalidValue,
			arg.PercentReserved,
			minPercentReserved,
		)
	}
	if arg.IncreaseFactor < 0 {
		return nil, fmt.Errorf("%w, increaseFactor is negative: provided %0.3f",
			process.ErrInvalidValue,
			arg.IncreaseFactor,
		)
	}
	if len(arg.SelfPid) == 0 {
		return nil, process.ErrEmptyPeerId
	}

	return &quotaFloodPreventer{
		name:                          arg.Name,
		cacher:                        arg.Cacher,
		statusHandlers:                arg.StatusHandlers,
		computedMaxNumMessagesPerPeer: arg.BaseMaxNumMessagesPerPeer,
		baseMaxNumMessagesPerPeer:     arg.BaseMaxNumMessagesPerPeer,
		maxTotalSizePerPeer:           arg.MaxTotalSizePerPeer,
		percentReserved:               arg.PercentReserved,
		increaseThreshold:             arg.IncreaseThreshold,
		increaseFactor:                arg.IncreaseFactor,
		selfPid:                       arg.SelfPid,
	}, nil
}

// IncreaseLoad tries to increment the counter values held at "pid" position
// It returns true if it had succeeded incrementing (existing counter value is lower or equal with provided maxOperations)
// We need the mutOperation here as the get and put should be done atomically.
// Otherwise we might yield a slightly higher number of false valid increments
// This method also checks the global sum quota but does not increment its values
func (qfp *quotaFloodPreventer) IncreaseLoad(fromConnectedPid core.PeerID, originator core.PeerID, size uint64) error {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	if originator != qfp.selfPid {
		return qfp.increaseLoadExternalMessage(fromConnectedPid, originator, size)
	}

	return qfp.increaseLoadSelfMessage(size)
}

// increaseLoadExternalMessage is called by a message broadcast by other peer and this call will increase the values
// for the message originator, the peer that broadcast it to self (if different from the originator) and for the self peer id.
// The purpose of incrementing also the self peer id is that, in case of multiple mild-flooding connected peers, the current
// node should not propagate all messages because it will get blacklisted by other peers (it will become a proxy for the flooders)
// This call will increase the load for current peer taking into account the reserved percentage as it is an external message
// and not a message sent by self
func (qfp *quotaFloodPreventer) increaseLoadExternalMessage(fromConnectedPid core.PeerID, originator core.PeerID, size uint64) error {
	errOriginator := qfp.increaseLoad(originator, size, qfp.percentReserved)
	var errFromConnected error
	if fromConnectedPid != originator {
		errFromConnected = qfp.increaseLoad(fromConnectedPid, size, qfp.percentReserved)
	}
	errSelf := qfp.increaseLoad(qfp.selfPid, size, qfp.percentReserved)

	if errFromConnected != nil {
		return errFromConnected
	}
	if errOriginator != nil {
		return errOriginator
	}
	if errSelf != nil {
		return errSelf
	}

	return nil
}

// increaseLoadSelfMessage is called for a message that this node broadcast it. The function will increase just the
// current peer id limits taking into account the reserved percent as to utilize the maximum limit allowed by the flood
// preventer
func (qfp *quotaFloodPreventer) increaseLoadSelfMessage(size uint64) error {
	return qfp.increaseLoad(qfp.selfPid, size, 0)
}

func (qfp *quotaFloodPreventer) increaseLoad(pid core.PeerID, size uint64, percentReserved float32) error {
	valueQuota, ok := qfp.cacher.Get(pid.Bytes())
	if !ok {
		qfp.putDefaultQuota(pid, size)

		return nil
	}

	q, isQuota := valueQuota.(*quota)
	if !isQuota {
		qfp.putDefaultQuota(pid, size)

		return nil
	}

	q.numReceivedMessages++
	q.sizeReceivedMessages += size

	maxNumMessagesReached := qfp.isMaximumReached(
		uint64(qfp.computedMaxNumMessagesPerPeer),
		uint64(q.numReceivedMessages),
		percentReserved,
	)
	maxSizeMessagesReached := qfp.isMaximumReached(
		qfp.maxTotalSizePerPeer,
		q.sizeReceivedMessages,
		percentReserved,
	)
	isPeerQuotaReached := maxNumMessagesReached || maxSizeMessagesReached
	if isPeerQuotaReached {
		return fmt.Errorf("%w for pid %s", process.ErrSystemBusy, pid.Pretty())
	}

	q.numProcessedMessages++
	q.sizeProcessedMessages += size

	return nil
}

func (qfp *quotaFloodPreventer) isMaximumReached(absoluteMax uint64, counted uint64, percentReserved float32) bool {
	max := uint64(100-percentReserved) * absoluteMax / 100

	return counted > max
}

func (qfp *quotaFloodPreventer) putDefaultQuota(pid core.PeerID, size uint64) {
	q := &quota{
		numReceivedMessages:   initNumMessages,
		sizeReceivedMessages:  size,
		numProcessedMessages:  initNumMessages,
		sizeProcessedMessages: size,
	}
	qfp.cacher.Put(pid.Bytes(), q, q.Size())
}

// Reset clears all map values
func (qfp *quotaFloodPreventer) Reset() {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	qfp.resetStatusHandlers()
	qfp.createStatistics()

	//TODO change this if cacher.Clear() is time consuming
	qfp.cacher.Clear()
}

func (qfp *quotaFloodPreventer) resetStatusHandlers() {
	for _, statusHandler := range qfp.statusHandlers {
		statusHandler.ResetStatistics()
	}
}

// createStatistics is useful to benchmark the system when running
func (qfp *quotaFloodPreventer) createStatistics() {
	keys := qfp.cacher.Keys()
	for _, k := range keys {
		val, ok := qfp.cacher.Get(k)
		if !ok {
			continue
		}

		q, isQuota := val.(*quota)
		if !isQuota {
			continue
		}

		qfp.addQuota(
			core.PeerID(k),
			q.numReceivedMessages,
			q.sizeReceivedMessages,
			q.numProcessedMessages,
			q.sizeProcessedMessages,
		)
	}
}

func (qfp *quotaFloodPreventer) addQuota(
	pid core.PeerID,
	numReceived uint32,
	sizeReceived uint64,
	numProcessed uint32,
	sizeProcessed uint64,
) {
	for _, statusHandler := range qfp.statusHandlers {
		statusHandler.AddQuota(pid, numReceived, sizeReceived, numProcessed, sizeProcessed)
	}
}

// ApplyConsensusSize will set the maximum number of messages that can be received from a peer
func (qfp *quotaFloodPreventer) ApplyConsensusSize(size int) {
	if size < 1 {
		log.Warn("invalid consensus size in quota flood preventer",
			"name", qfp.name,
			"provided value", size,
		)
		return
	}
	if qfp.increaseThreshold > uint32(size) {
		log.Debug("consensus size did not reach the threshold for quota flood preventer",
			"name", qfp.name,
			"provided", size,
			"threshold", qfp.increaseThreshold,
		)
		return
	}

	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	numNodesOverThreshold := float32(uint32(size) - qfp.increaseThreshold)
	value := numNodesOverThreshold * qfp.increaseFactor
	oldComputed := qfp.computedMaxNumMessagesPerPeer
	qfp.computedMaxNumMessagesPerPeer = qfp.baseMaxNumMessagesPerPeer + uint32(value)

	log.Debug("quotaFloodPreventer.ApplyConsensusSize",
		"name", qfp.name,
		"provided", size,
		"threshold", qfp.increaseThreshold,
		"factor", qfp.increaseFactor,
		"base", qfp.baseMaxNumMessagesPerPeer,
		"old computed", oldComputed,
		"new computed", qfp.computedMaxNumMessagesPerPeer,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (qfp *quotaFloodPreventer) IsInterfaceNil() bool {
	return qfp == nil
}
