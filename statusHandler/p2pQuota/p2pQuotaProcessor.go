package p2pQuota

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/statusHandler"
)

type quota struct {
	numReceivedMessages   uint32
	numProcessedMessages  uint32
	sizeReceivedMessages  uint64
	sizeProcessedMessages uint64
}

// p2pQuotaProcessor implements process.QuotaStatusHandler and is able to periodically send to a
// statusHandler the processed p2p quota information
type p2pQuotaProcessor struct {
	mutStatistics    sync.Mutex
	statistics       map[core.PeerID]*quota
	peakPeerQuota    *quota
	peakNumReceivers uint64
	handler          core.AppStatusHandler
	quotaIdentifier  string
}

// NewP2PQuotaProcessor creates a new p2pQuotaProcessor instance
func NewP2PQuotaProcessor(handler core.AppStatusHandler, quotaIdentifier string) (*p2pQuotaProcessor, error) {
	if check.IfNil(handler) {
		return nil, statusHandler.ErrNilAppStatusHandler
	}

	return &p2pQuotaProcessor{
		statistics:      make(map[core.PeerID]*quota),
		peakPeerQuota:   &quota{},
		handler:         handler,
		quotaIdentifier: quotaIdentifier,
	}, nil
}

// ResetStatistics output gathered statistics, process and prints them. After that it empties the statistics map
func (pqp *p2pQuotaProcessor) ResetStatistics() {
	pqp.mutStatistics.Lock()
	defer pqp.mutStatistics.Unlock()

	peakPeerQuota := pqp.computePeerStatistics()
	numPeers := uint64(len(pqp.statistics))
	pqp.setPeakStatistics(peakPeerQuota, numPeers)

	pqp.moveStatisticsInAppStatusHandler(peakPeerQuota, numPeers, pqp.peakNumReceivers)

	pqp.statistics = make(map[core.PeerID]*quota)
}

func (pqp *p2pQuotaProcessor) computePeerStatistics() *quota {
	peakPeerQuota := &quota{}

	for _, q := range pqp.statistics {
		peakPeerQuota.numReceivedMessages = core.MaxUint32(peakPeerQuota.numReceivedMessages, q.numReceivedMessages)
		peakPeerQuota.sizeReceivedMessages = core.MaxUint64(peakPeerQuota.sizeReceivedMessages, q.sizeReceivedMessages)
		peakPeerQuota.numProcessedMessages = core.MaxUint32(peakPeerQuota.numProcessedMessages, q.numProcessedMessages)
		peakPeerQuota.sizeProcessedMessages = core.MaxUint64(peakPeerQuota.sizeProcessedMessages, q.sizeProcessedMessages)
	}

	return peakPeerQuota
}

func (pqp *p2pQuotaProcessor) setPeakStatistics(peakPeerQuota *quota, numPeers uint64) {
	pqp.peakPeerQuota.numReceivedMessages = core.MaxUint32(peakPeerQuota.numReceivedMessages, pqp.peakPeerQuota.numReceivedMessages)
	pqp.peakPeerQuota.sizeReceivedMessages = core.MaxUint64(peakPeerQuota.sizeReceivedMessages, pqp.peakPeerQuota.sizeReceivedMessages)
	pqp.peakPeerQuota.numProcessedMessages = core.MaxUint32(peakPeerQuota.numProcessedMessages, pqp.peakPeerQuota.numProcessedMessages)
	pqp.peakPeerQuota.sizeProcessedMessages = core.MaxUint64(peakPeerQuota.sizeProcessedMessages, pqp.peakPeerQuota.sizeProcessedMessages)

	pqp.peakNumReceivers = core.MaxUint64(numPeers, pqp.peakNumReceivers)
}

func (pqp *p2pQuotaProcessor) moveStatisticsInAppStatusHandler(
	peerQuota *quota,
	numReceiverPeers uint64,
	peakNumReceiverPeers uint64,
) {

	pqp.handler.SetUInt64Value(pqp.getMetric(common.MetricP2PPeerNumReceivedMessages), uint64(peerQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(pqp.getMetric(common.MetricP2PPeerSizeReceivedMessages), peerQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(pqp.getMetric(common.MetricP2PPeerNumProcessedMessages), uint64(peerQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(pqp.getMetric(common.MetricP2PPeerSizeProcessedMessages), peerQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(pqp.getMetric(common.MetricP2PPeakPeerNumReceivedMessages), uint64(pqp.peakPeerQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(pqp.getMetric(common.MetricP2PPeakPeerSizeReceivedMessages), pqp.peakPeerQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(pqp.getMetric(common.MetricP2PPeakPeerNumProcessedMessages), uint64(pqp.peakPeerQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(pqp.getMetric(common.MetricP2PPeakPeerSizeProcessedMessages), pqp.peakPeerQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(pqp.getMetric(common.MetricP2PNumReceiverPeers), numReceiverPeers)
	pqp.handler.SetUInt64Value(pqp.getMetric(common.MetricP2PPeakNumReceiverPeers), peakNumReceiverPeers)
}

func (pqp *p2pQuotaProcessor) getMetric(metric string) string {
	return metric + "_" + pqp.quotaIdentifier
}

// AddQuota adds a quota statistics
func (pqp *p2pQuotaProcessor) AddQuota(
	pid core.PeerID,
	numReceived uint32,
	sizeReceived uint64,
	numProcessed uint32,
	sizeProcessed uint64,
) {
	q := &quota{
		numReceivedMessages:   numReceived,
		sizeReceivedMessages:  sizeReceived,
		numProcessedMessages:  numProcessed,
		sizeProcessedMessages: sizeProcessed,
	}

	pqp.mutStatistics.Lock()
	pqp.statistics[pid] = q
	pqp.mutStatistics.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (pqp *p2pQuotaProcessor) IsInterfaceNil() bool {
	return pqp == nil
}
