package floodPreventers

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood"
	"github.com/multiversx/mx-chain-go/storage"
)

// ArgQuotaFloodPreventer defines the arguments for a quota flood preventer
type ArgQuotaFloodPreventer struct {
	Name             string
	Cacher           storage.Cacher
	StatusHandlers   []QuotaStatusHandler
	AntifloodConfigs common.AntifloodConfigsHandler
	ConfigFetcher    antiflood.FloodPreventerConfigFetcher
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
	antifloodConfigs              common.AntifloodConfigsHandler
	configFetcher                 antiflood.FloodPreventerConfigFetcher
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
	if check.IfNil(arg.AntifloodConfigs) {
		return nil, process.ErrNilAntifloodConfigsHandler
	}

	qfp := &quotaFloodPreventer{
		name:             arg.Name,
		cacher:           arg.Cacher,
		statusHandlers:   arg.StatusHandlers,
		antifloodConfigs: arg.AntifloodConfigs,
		configFetcher:    arg.ConfigFetcher,
	}
	qfp.computedMaxNumMessagesPerPeer = qfp.getBbaseMaxNumMessagesPerPeer()

	return qfp, nil
}

// IncreaseLoad tries to increment the counter values held at "pid" position
// It returns true if it had succeeded incrementing (existing counter value is lower or equal with provided maxOperations)
// We need the mutOperation here as the get and put should be done atomically.
// Otherwise, we might yield a slightly higher number of false valid increments
// This method also checks the global sum quota but does not increment its values
func (qfp *quotaFloodPreventer) IncreaseLoad(pid core.PeerID, size uint64) error {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	return qfp.increaseLoad(pid, size)
}

func (qfp *quotaFloodPreventer) increaseLoad(pid core.PeerID, size uint64) error {
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

	maxNumMessagesReached := qfp.isMaximumReached(uint64(qfp.computedMaxNumMessagesPerPeer), uint64(q.numReceivedMessages))
	maxSizeMessagesReached := qfp.isMaximumReached(qfp.getMaxTotalSizePerInternal(), q.sizeReceivedMessages)
	isPeerQuotaReached := maxNumMessagesReached || maxSizeMessagesReached
	if isPeerQuotaReached {
		return fmt.Errorf("%w for pid %s", process.ErrSystemBusy, pid.Pretty())
	}

	q.numProcessedMessages++
	q.sizeProcessedMessages += size

	return nil
}

func (qfp *quotaFloodPreventer) isMaximumReached(absoluteMax uint64, counted uint64) bool {
	max := uint64(100-qfp.getReservedPercent()) * absoluteMax / 100

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
	if qfp.getIncreaseThreshold() > uint32(size) {
		log.Debug("consensus size did not reach the threshold for quota flood preventer",
			"name", qfp.name,
			"provided", size,
			"threshold", qfp.getIncreaseThreshold(),
		)
		return
	}

	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	numNodesOverThreshold := float32(uint32(size) - qfp.getIncreaseThreshold())
	value := numNodesOverThreshold * qfp.getIncreaseFactor()
	oldComputed := qfp.computedMaxNumMessagesPerPeer
	qfp.computedMaxNumMessagesPerPeer = qfp.getBbaseMaxNumMessagesPerPeer() + uint32(value)

	log.Debug("quotaFloodPreventer.ApplyConsensusSize",
		"name", qfp.name,
		"provided", size,
		"threshold", qfp.getIncreaseThreshold(),
		"factor", qfp.getIncreaseFactor(),
		"base", qfp.getBbaseMaxNumMessagesPerPeer(),
		"old computed", oldComputed,
		"new computed", qfp.computedMaxNumMessagesPerPeer,
	)
}

func (qfp *quotaFloodPreventer) getBbaseMaxNumMessagesPerPeer() uint32 {
	if qfp.name == antiflood.OutputIdentifier {
		currentConfig := qfp.antifloodConfigs.GetCurrentConfig()
		return currentConfig.PeerMaxOutput.BaseMessagesPerInterval
	}

	currentConfig := qfp.configFetcher(qfp.antifloodConfigs, qfp.name)
	return currentConfig.PeerMaxInput.BaseMessagesPerInterval
}

func (qfp *quotaFloodPreventer) getMaxTotalSizePerInternal() uint64 {
	if qfp.name == antiflood.OutputIdentifier {
		currentConfig := qfp.antifloodConfigs.GetCurrentConfig()
		return currentConfig.PeerMaxOutput.TotalSizePerInterval
	}

	currentConfig := qfp.configFetcher(qfp.antifloodConfigs, qfp.name)
	return currentConfig.PeerMaxInput.TotalSizePerInterval
}

func (qfp *quotaFloodPreventer) getReservedPercent() float32 {
	if qfp.name == antiflood.OutputIdentifier {
		// this is not handled on this flow
		return 0
	}

	currentConfig := qfp.configFetcher(qfp.antifloodConfigs, qfp.name)
	return currentConfig.ReservedPercent
}

func (qfp *quotaFloodPreventer) getIncreaseThreshold() uint32 {
	if qfp.name == antiflood.OutputIdentifier {
		currentConfig := qfp.antifloodConfigs.GetCurrentConfig()
		return currentConfig.PeerMaxOutput.IncreaseFactor.Threshold
	}

	currentConfig := qfp.configFetcher(qfp.antifloodConfigs, qfp.name)
	return currentConfig.PeerMaxInput.IncreaseFactor.Threshold
}

func (qfp *quotaFloodPreventer) getIncreaseFactor() float32 {
	if qfp.name == antiflood.OutputIdentifier {
		currentConfig := qfp.antifloodConfigs.GetCurrentConfig()
		return currentConfig.PeerMaxOutput.IncreaseFactor.Factor
	}

	currentConfig := qfp.configFetcher(qfp.antifloodConfigs, qfp.name)
	return currentConfig.PeerMaxInput.IncreaseFactor.Factor
}

// IsInterfaceNil returns true if there is no value under the interface
func (qfp *quotaFloodPreventer) IsInterfaceNil() bool {
	return qfp == nil
}
