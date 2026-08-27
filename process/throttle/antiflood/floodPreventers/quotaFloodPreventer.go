package floodPreventers

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

// ArgQuotaFloodPreventer defines the arguments for a quota flood preventer
type ArgQuotaFloodPreventer struct {
	Name             common.FloodPreventerType
	Cacher           storage.Cacher
	StatusHandlers   []QuotaStatusHandler
	AntifloodConfigs common.AntifloodConfigsHandler
}

var _ process.FloodPreventer = (*quotaFloodPreventer)(nil)

const minMessages = 1
const minTotalSize = 1 // 1Byte
const initNumMessages = 1
const quotaStructSize = 32

type quota struct {
	numReceivedMessages   uint32
	numProcessedMessages  uint32
	sizeReceivedMessages  uint64
	sizeProcessedMessages uint64
	// quotaReachedLogged is used to log the quota reached event only once per peer per interval, so a
	// flooding peer can not spam the log with one line per rejected message
	quotaReachedLogged bool
}

// Size returns the size of a quota object
func (q *quota) Size() int {
	return quotaStructSize
}

// quotaFloodPreventer represents a cache of quotas per peer used in antiflooding mechanism
type quotaFloodPreventer struct {
	name                          common.FloodPreventerType
	mutOperation                  sync.RWMutex
	cacher                        storage.Cacher
	statusHandlers                []QuotaStatusHandler
	computedMaxNumMessagesPerPeer uint32
	antifloodConfigs              common.AntifloodConfigsHandler
	stats                         *quotaStats
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
		stats:            newQuotaStats(quotaStatsPrintInterval),
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

	maxNumMessages := qfp.computeMaxAllowed(uint64(qfp.computedMaxNumMessagesPerPeer))
	maxSize := qfp.computeMaxAllowed(qfp.getMaxTotalSizePerInternal())
	maxNumMessagesReached := uint64(q.numReceivedMessages) > maxNumMessages
	maxSizeMessagesReached := q.sizeReceivedMessages > maxSize
	isPeerQuotaReached := maxNumMessagesReached || maxSizeMessagesReached
	if isPeerQuotaReached {
		qfp.logQuotaReached(pid, q, size, maxNumMessages, maxSize, maxNumMessagesReached, maxSizeMessagesReached)

		return fmt.Errorf("%w for pid %s", process.ErrSystemBusy, pid.Pretty())
	}

	q.numProcessedMessages++
	q.sizeProcessedMessages += size

	return nil
}

// computeMaxAllowed returns the effective maximum value that a peer can reach in an interval, that is the
// configured absolute maximum diminished by the configured reserved percent
func (qfp *quotaFloodPreventer) computeMaxAllowed(absoluteMax uint64) uint64 {
	return uint64(100-qfp.getReservedPercent()) * absoluteMax / 100
}

// logQuotaReached must be called with the mutOperation locked. It prints the whole context that led to the
// rejection, but only once per peer per interval as to not flood the log while the peer floods the node
func (qfp *quotaFloodPreventer) logQuotaReached(
	pid core.PeerID,
	q *quota,
	size uint64,
	maxNumMessages uint64,
	maxSize uint64,
	maxNumMessagesReached bool,
	maxSizeMessagesReached bool,
) {
	firstForPeerInInterval := !q.quotaReachedLogged
	q.quotaReachedLogged = true
	qfp.stats.addRejectedMessage(firstForPeerInInterval)

	if !firstForPeerInInterval {
		return
	}

	log.Debug("quotaFloodPreventer peer quota reached",
		"name", qfp.name,
		"pid", pid.Pretty(),
		"num messages limit reached", maxNumMessagesReached,
		"size limit reached", maxSizeMessagesReached,
		"num messages", q.numReceivedMessages,
		"max num messages", maxNumMessages,
		"size", core.ConvertBytes(q.sizeReceivedMessages),
		"max size", core.ConvertBytes(maxSize),
		"rejected message size", core.ConvertBytes(size),
		"num processed messages", q.numProcessedMessages,
		"size processed messages", core.ConvertBytes(q.sizeProcessedMessages),
		"reserved percent", qfp.getReservedPercent(),
	)
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
	shouldPrint, statsArgs := qfp.resetAndGatherStatistics()
	if shouldPrint {
		log.Debug("quotaFloodPreventer statistics", statsArgs...)
	}
}

func (qfp *quotaFloodPreventer) resetAndGatherStatistics() (bool, []interface{}) {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	qfp.resetStatusHandlers()
	numPeers, peakNumMessages, peakSize := qfp.createStatistics()
	qfp.stats.addIntervalPeaks(numPeers, peakNumMessages, peakSize)

	// TODO change this if cacher.Clear() is time consuming
	qfp.cacher.Clear()

	if !qfp.stats.shouldPrint() {
		return false, nil
	}

	statsArgs := qfp.createStatisticsLogArgs()
	qfp.stats.reset()

	return true, statsArgs
}

// createStatisticsLogArgs must be called with the mutOperation locked. It reports the peak load generated by a
// single peer over the elapsed window, relative to the currently effective limits, so the antiflood
// configuration can be tuned against the real traffic present on the network
func (qfp *quotaFloodPreventer) createStatisticsLogArgs() []interface{} {
	stats := qfp.stats
	maxNumMessages := qfp.computeMaxAllowed(uint64(qfp.computedMaxNumMessagesPerPeer))
	maxSize := qfp.computeMaxAllowed(qfp.getMaxTotalSizePerInternal())

	return []interface{}{
		"name", qfp.name,
		"window", stats.window(),
		"num intervals", stats.numIntervals,
		"peak num peers", stats.peakNumPeers,
		"peak num messages/peer", stats.peakNumMessages.numMessages,
		"max num messages/peer", maxNumMessages,
		"num messages usage", computeUsagePercent(uint64(stats.peakNumMessages.numMessages), maxNumMessages),
		"peak num messages pid", stats.peakNumMessages.pid.Pretty(),
		"peak size/peer", core.ConvertBytes(stats.peakSize.size),
		"max size/peer", core.ConvertBytes(maxSize),
		"size usage", computeUsagePercent(stats.peakSize.size, maxSize),
		"peak size pid", stats.peakSize.pid.Pretty(),
		"num rejected messages", stats.numRejectedMessages,
		"num peers reaching quota", stats.numPeersReachingQuota,
	}
}

func (qfp *quotaFloodPreventer) resetStatusHandlers() {
	for _, statusHandler := range qfp.statusHandlers {
		statusHandler.ResetStatistics()
	}
}

// createStatistics is useful to benchmark the system when running. It also returns the number of peers seen in
// the elapsed interval and the peak values recorded for a single peer
func (qfp *quotaFloodPreventer) createStatistics() (int, peerPeak, peerPeak) {
	numPeers := 0
	peakNumMessages := peerPeak{}
	peakSize := peerPeak{}

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

		pid := core.PeerID(k)
		numPeers++
		if q.numReceivedMessages > peakNumMessages.numMessages {
			peakNumMessages = peerPeak{numMessages: q.numReceivedMessages, size: q.sizeReceivedMessages, pid: pid}
		}
		if q.sizeReceivedMessages > peakSize.size {
			peakSize = peerPeak{numMessages: q.numReceivedMessages, size: q.sizeReceivedMessages, pid: pid}
		}

		qfp.addQuota(
			pid,
			q.numReceivedMessages,
			q.sizeReceivedMessages,
			q.numProcessedMessages,
			q.sizeProcessedMessages,
		)
	}

	return numPeers, peakNumMessages, peakSize
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
	currentConfig := qfp.antifloodConfigs.GetFloodPreventerConfigByType(qfp.name)
	return currentConfig.PeerMaxInput.BaseMessagesPerInterval
}

func (qfp *quotaFloodPreventer) getMaxTotalSizePerInternal() uint64 {
	currentConfig := qfp.antifloodConfigs.GetFloodPreventerConfigByType(qfp.name)
	return currentConfig.PeerMaxInput.TotalSizePerInterval
}

func (qfp *quotaFloodPreventer) getReservedPercent() float32 {
	currentConfig := qfp.antifloodConfigs.GetFloodPreventerConfigByType(qfp.name)
	return currentConfig.ReservedPercent
}

func (qfp *quotaFloodPreventer) getIncreaseThreshold() uint32 {
	currentConfig := qfp.antifloodConfigs.GetFloodPreventerConfigByType(qfp.name)
	return currentConfig.PeerMaxInput.IncreaseFactor.Threshold
}

func (qfp *quotaFloodPreventer) getIncreaseFactor() float32 {
	currentConfig := qfp.antifloodConfigs.GetFloodPreventerConfigByType(qfp.name)
	return currentConfig.PeerMaxInput.IncreaseFactor.Factor
}

// IsInterfaceNil returns true if there is no value under the interface
func (qfp *quotaFloodPreventer) IsInterfaceNil() bool {
	return qfp == nil
}
