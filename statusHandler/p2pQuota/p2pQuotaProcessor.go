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

// p2pQuotaProcessor implements process.QuotaStatusHandler and is able to periodically sends to a
// statusHandler the processed p2p quota information
type p2pQuotaProcessor struct {
	mutStatistics    sync.Mutex
	statistics       map[string]*quota
	peakNetworkQuota *quota
	peakPeerQuota    *quota
	peakNumReceivers uint64
	handler          core.AppStatusHandler
}

// NewP2pQuotaProcessor creates a new p2pQuotaProcessor instance
func NewP2pQuotaProcessor(handler core.AppStatusHandler) (*p2pQuotaProcessor, error) {
	if check.IfNil(handler) {
		return nil, statusHandler.ErrNilAppStatusHandler
	}

	return &p2pQuotaProcessor{
		statistics:       make(map[string]*quota),
		peakNetworkQuota: &quota{},
		peakPeerQuota:    &quota{},
		handler:          handler,
	}, nil
}

// ResetStatistics output gathered statistics, process and prints them. After that it empties the statistics map
func (pqp *p2pQuotaProcessor) ResetStatistics() {
	networkQuota := &quota{}
	peakPeerQuota := &quota{}

	pqp.mutStatistics.Lock()
	defer pqp.mutStatistics.Unlock()

	for _, q := range pqp.statistics {
		networkQuota.numReceivedMessages += q.numReceivedMessages
		networkQuota.sizeReceivedMessages += q.sizeReceivedMessages
		networkQuota.numProcessedMessages += q.numProcessedMessages
		networkQuota.sizeProcessedMessages += q.sizeProcessedMessages

		peakPeerQuota.numReceivedMessages = core.MaxUint32(peakPeerQuota.numReceivedMessages, q.numReceivedMessages)
		peakPeerQuota.sizeReceivedMessages = core.MaxUint64(peakPeerQuota.sizeReceivedMessages, q.sizeReceivedMessages)
		peakPeerQuota.numProcessedMessages = core.MaxUint32(peakPeerQuota.numProcessedMessages, q.numProcessedMessages)
		peakPeerQuota.sizeProcessedMessages = core.MaxUint64(peakPeerQuota.sizeProcessedMessages, q.sizeProcessedMessages)
	}

	pqp.peakPeerQuota.numReceivedMessages = core.MaxUint32(peakPeerQuota.numReceivedMessages, pqp.peakPeerQuota.numReceivedMessages)
	pqp.peakPeerQuota.sizeReceivedMessages = core.MaxUint64(peakPeerQuota.sizeReceivedMessages, pqp.peakPeerQuota.sizeReceivedMessages)
	pqp.peakPeerQuota.numProcessedMessages = core.MaxUint32(peakPeerQuota.numProcessedMessages, pqp.peakPeerQuota.numProcessedMessages)
	pqp.peakPeerQuota.sizeProcessedMessages = core.MaxUint64(peakPeerQuota.sizeProcessedMessages, pqp.peakPeerQuota.sizeProcessedMessages)

	pqp.peakNetworkQuota.numReceivedMessages = core.MaxUint32(networkQuota.numReceivedMessages, pqp.peakNetworkQuota.numReceivedMessages)
	pqp.peakNetworkQuota.sizeReceivedMessages = core.MaxUint64(networkQuota.sizeReceivedMessages, pqp.peakNetworkQuota.sizeReceivedMessages)
	pqp.peakNetworkQuota.numProcessedMessages = core.MaxUint32(networkQuota.numProcessedMessages, pqp.peakNetworkQuota.numProcessedMessages)
	pqp.peakNetworkQuota.sizeProcessedMessages = core.MaxUint64(networkQuota.sizeProcessedMessages, pqp.peakNetworkQuota.sizeProcessedMessages)

	numPeers := uint64(len(pqp.statistics))
	pqp.peakNumReceivers = core.MaxUint64(numPeers, pqp.peakNumReceivers)

	pqp.moveStatisticsInAppStatusHandler(peakPeerQuota, networkQuota, numPeers, pqp.peakNumReceivers)

	pqp.statistics = make(map[string]*quota)
}

func (pqp *p2pQuotaProcessor) moveStatisticsInAppStatusHandler(
	peerQuota *quota,
	networkQuota *quota,
	numReceiverPeers uint64,
	peakNumReceiverPeers uint64,
) {

	pqp.handler.SetUInt64Value(core.MetricP2pNetworkNumReceivedMessages, uint64(networkQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pNetworkSizeReceivedMessages, networkQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2pNetworkNumProcessedMessages, uint64(networkQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pNetworkSizeProcessedMessages, networkQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2pPeakNetworkNumReceivedMessages, uint64(pqp.peakNetworkQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pPeakNetworkSizeReceivedMessages, pqp.peakNetworkQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2pPeakNetworkNumProcessedMessages, uint64(pqp.peakNetworkQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pPeakNetworkSizeProcessedMessages, pqp.peakNetworkQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2pPeerNumReceivedMessages, uint64(peerQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pPeerSizeReceivedMessages, peerQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2pPeerNumProcessedMessages, uint64(peerQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pPeerSizeProcessedMessages, peerQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2pPeakPeerNumReceivedMessages, uint64(pqp.peakPeerQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pPeakPeerSizeReceivedMessages, pqp.peakPeerQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2pPeakPeerxNumProcessedMessages, uint64(pqp.peakPeerQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pPeakPeerSizeProcessedMessages, pqp.peakPeerQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2pNumReceiverPeers, numReceiverPeers)
	pqp.handler.SetUInt64Value(core.MetricP2pPeakNumReceiverPeers, peakNumReceiverPeers)
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

// IsInterfaceNil returns true if there is no value under the interface
func (pqp *p2pQuotaProcessor) IsInterfaceNil() bool {
	return pqp == nil
}
