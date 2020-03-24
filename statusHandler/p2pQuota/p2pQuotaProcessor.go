package p2pQuota

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

type quota struct {
	numReceivedMessages   uint32
	sizeReceivedMessages  uint64
	numProcessedMessages  uint32
	sizeProcessedMessages uint64
}

// p2pQuotaProcessor implements process.QuotaStatusHandler and is able to periodically send to a
// statusHandler the processed p2p quota information
type p2pQuotaProcessor struct {
	mutStatistics    sync.Mutex
	statistics       map[string]*quota
	peakNetworkQuota *quota
	networkQuota     *quota
	peakPeerQuota    *quota
	peakNumReceivers uint64
	handler          core.AppStatusHandler
}

// NewP2PQuotaProcessor creates a new p2pQuotaProcessor instance
func NewP2PQuotaProcessor(handler core.AppStatusHandler) (*p2pQuotaProcessor, error) {
	if check.IfNil(handler) {
		return nil, statusHandler.ErrNilAppStatusHandler
	}

	return &p2pQuotaProcessor{
		statistics:       make(map[string]*quota),
		peakNetworkQuota: &quota{},
		networkQuota:     &quota{},
		peakPeerQuota:    &quota{},
		handler:          handler,
	}, nil
}

// ResetStatistics output gathered statistics, process and prints them. After that it empties the statistics map
func (pqp *p2pQuotaProcessor) ResetStatistics() {
	pqp.mutStatistics.Lock()
	defer pqp.mutStatistics.Unlock()

	peakPeerQuota := pqp.computePeerStatistics()
	numPeers := uint64(len(pqp.statistics))
	pqp.setPeakStatistics(peakPeerQuota, numPeers)

	pqp.moveStatisticsInAppStatusHandler(peakPeerQuota, pqp.networkQuota, numPeers, pqp.peakNumReceivers)

	pqp.statistics = make(map[string]*quota)
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

	pqp.peakNetworkQuota.numReceivedMessages = core.MaxUint32(pqp.networkQuota.numReceivedMessages, pqp.peakNetworkQuota.numReceivedMessages)
	pqp.peakNetworkQuota.sizeReceivedMessages = core.MaxUint64(pqp.networkQuota.sizeReceivedMessages, pqp.peakNetworkQuota.sizeReceivedMessages)
	pqp.peakNetworkQuota.numProcessedMessages = core.MaxUint32(pqp.networkQuota.numProcessedMessages, pqp.peakNetworkQuota.numProcessedMessages)
	pqp.peakNetworkQuota.sizeProcessedMessages = core.MaxUint64(pqp.networkQuota.sizeProcessedMessages, pqp.peakNetworkQuota.sizeProcessedMessages)

	pqp.peakNumReceivers = core.MaxUint64(numPeers, pqp.peakNumReceivers)
}

func (pqp *p2pQuotaProcessor) moveStatisticsInAppStatusHandler(
	peerQuota *quota,
	networkQuota *quota,
	numReceiverPeers uint64,
	peakNumReceiverPeers uint64,
) {

	pqp.handler.SetUInt64Value(core.MetricP2PNetworkNumReceivedMessages, uint64(networkQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2PNetworkSizeReceivedMessages, networkQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2PNetworkNumProcessedMessages, uint64(networkQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2PNetworkSizeProcessedMessages, networkQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2PPeakNetworkNumReceivedMessages, uint64(pqp.peakNetworkQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2PPeakNetworkSizeReceivedMessages, pqp.peakNetworkQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2PPeakNetworkNumProcessedMessages, uint64(pqp.peakNetworkQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2PPeakNetworkSizeProcessedMessages, pqp.peakNetworkQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2PPeerNumReceivedMessages, uint64(peerQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2PPeerSizeReceivedMessages, peerQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2PPeerNumProcessedMessages, uint64(peerQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2PPeerSizeProcessedMessages, peerQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2PPeakPeerNumReceivedMessages, uint64(pqp.peakPeerQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2PPeakPeerSizeReceivedMessages, pqp.peakPeerQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2PPeakPeerxNumProcessedMessages, uint64(pqp.peakPeerQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2PPeakPeerSizeProcessedMessages, pqp.peakPeerQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2PNumReceiverPeers, numReceiverPeers)
	pqp.handler.SetUInt64Value(core.MetricP2PPeakNumReceiverPeers, peakNumReceiverPeers)
}

// AddQuota adds a quota statistics
func (pqp *p2pQuotaProcessor) AddQuota(
	identifier string,
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
	pqp.statistics[identifier] = q
	pqp.mutStatistics.Unlock()
}

// SetGlobalQuota sets the global quota statistics
func (pqp *p2pQuotaProcessor) SetGlobalQuota(
	numReceived uint32,
	sizeReceived uint64,
	numProcessed uint32,
	sizeProcessed uint64,
) {
	pqp.mutStatistics.Lock()
	pqp.networkQuota = &quota{
		numReceivedMessages:   numReceived,
		sizeReceivedMessages:  sizeReceived,
		numProcessedMessages:  numProcessed,
		sizeProcessedMessages: sizeProcessed,
	}
	pqp.mutStatistics.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (pqp *p2pQuotaProcessor) IsInterfaceNil() bool {
	return pqp == nil
}
