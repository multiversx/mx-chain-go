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
	mutStatistics   sync.Mutex
	statistics      map[string]*quota
	topSumQuota     *quota
	topMaxQuota     *quota
	topNumReceivers uint64
	handler         core.AppStatusHandler
}

// NewP2pQuotaProcessor creates a new p2pQuotaProcessor instance
func NewP2pQuotaProcessor(handler core.AppStatusHandler) (*p2pQuotaProcessor, error) {
	if check.IfNil(handler) {
		return nil, statusHandler.ErrNilAppStatusHandler
	}

	return &p2pQuotaProcessor{
		statistics:  make(map[string]*quota),
		topSumQuota: &quota{},
		topMaxQuota: &quota{},
		handler:     handler,
	}, nil
}

// ResetStatistics output gathered statistics, process and prints them. After that it empties the statistics map
func (pqp *p2pQuotaProcessor) ResetStatistics() {
	sumQuota := &quota{}
	maxQuota := &quota{}

	pqp.mutStatistics.Lock()
	defer pqp.mutStatistics.Unlock()

	for _, q := range pqp.statistics {
		sumQuota.numReceivedMessages += q.numReceivedMessages
		sumQuota.sizeReceivedMessages += q.sizeReceivedMessages
		sumQuota.numProcessedMessages += q.numProcessedMessages
		sumQuota.sizeProcessedMessages += q.sizeProcessedMessages

		maxQuota.numReceivedMessages = core.MaxUint32(maxQuota.numReceivedMessages, q.numReceivedMessages)
		maxQuota.sizeReceivedMessages = core.MaxUint64(maxQuota.sizeReceivedMessages, q.sizeReceivedMessages)
		maxQuota.numProcessedMessages = core.MaxUint32(maxQuota.numProcessedMessages, q.numProcessedMessages)
		maxQuota.sizeProcessedMessages = core.MaxUint64(maxQuota.sizeProcessedMessages, q.sizeProcessedMessages)
	}

	pqp.topMaxQuota.numReceivedMessages = core.MaxUint32(maxQuota.numReceivedMessages, pqp.topMaxQuota.numReceivedMessages)
	pqp.topMaxQuota.sizeReceivedMessages = core.MaxUint64(maxQuota.sizeReceivedMessages, pqp.topMaxQuota.sizeReceivedMessages)
	pqp.topMaxQuota.numProcessedMessages = core.MaxUint32(maxQuota.numProcessedMessages, pqp.topMaxQuota.numProcessedMessages)
	pqp.topMaxQuota.sizeProcessedMessages = core.MaxUint64(maxQuota.sizeProcessedMessages, pqp.topMaxQuota.sizeProcessedMessages)

	pqp.topSumQuota.numReceivedMessages = core.MaxUint32(sumQuota.numReceivedMessages, pqp.topSumQuota.numReceivedMessages)
	pqp.topSumQuota.sizeReceivedMessages = core.MaxUint64(sumQuota.sizeReceivedMessages, pqp.topSumQuota.sizeReceivedMessages)
	pqp.topSumQuota.numProcessedMessages = core.MaxUint32(sumQuota.numProcessedMessages, pqp.topSumQuota.numProcessedMessages)
	pqp.topSumQuota.sizeProcessedMessages = core.MaxUint64(sumQuota.sizeProcessedMessages, pqp.topSumQuota.sizeProcessedMessages)

	numPeers := uint64(len(pqp.statistics))
	pqp.topNumReceivers = core.MaxUint64(numPeers, pqp.topNumReceivers)

	pqp.moveStatisticsInAppStatusHandler(maxQuota, sumQuota, numPeers, pqp.topNumReceivers)

	pqp.statistics = make(map[string]*quota)
}

func (pqp *p2pQuotaProcessor) moveStatisticsInAppStatusHandler(
	maxQuota *quota,
	sumQuota *quota,
	numReceiverPeers uint64,
	topNumReceiverPeers uint64,
) {

	pqp.handler.SetUInt64Value(core.MetricP2pSumNumReceivedMessages, uint64(sumQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pSumSizeReceivedMessages, sumQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2pSumNumProcessedMessages, uint64(sumQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pSumSizeProcessedMessages, sumQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2pTopSumNumReceivedMessages, uint64(pqp.topSumQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pTopSumSizeReceivedMessages, pqp.topSumQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2pTopSumNumProcessedMessages, uint64(pqp.topSumQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pTopSumSizeProcessedMessages, pqp.topSumQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2pMaxNumReceivedMessages, uint64(maxQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pMaxSizeReceivedMessages, maxQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2pMaxNumProcessedMessages, uint64(maxQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pMaxSizeProcessedMessages, maxQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2pTopMaxNumReceivedMessages, uint64(pqp.topMaxQuota.numReceivedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pTopMaxSizeReceivedMessages, pqp.topMaxQuota.sizeReceivedMessages)
	pqp.handler.SetUInt64Value(core.MetricP2pTopMaxNumProcessedMessages, uint64(pqp.topMaxQuota.numProcessedMessages))
	pqp.handler.SetUInt64Value(core.MetricP2pTopMaxSizeProcessedMessages, pqp.topMaxQuota.sizeProcessedMessages)

	pqp.handler.SetUInt64Value(core.MetricP2pNumReceiverPeers, numReceiverPeers)
	pqp.handler.SetUInt64Value(core.MetricP2pTopNumReceiverPeers, topNumReceiverPeers)
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
