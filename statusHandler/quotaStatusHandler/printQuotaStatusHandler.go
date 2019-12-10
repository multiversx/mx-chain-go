package quotaStatusHandler

import (
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("statushandler/quotastatushandler")

type quota struct {
	numReceivedMessages   uint32
	sizeReceivedMessages  uint64
	numProcessedMessages  uint32
	sizeProcessedMessages uint64
}

//TODO replace this structure with a new one that can output on a route the statistics measured
//TODO add tests
// printQuotaStatusHandler implements process.QuotaStatusHandler and is able to periodically print peer connection statistics
type printQuotaStatusHandler struct {
	mutStatistics sync.Mutex
	statistics    map[string]*quota
}

// NewPrintQuotaStatusHandler creates a new NewPrintQuotaStatusHandler instance
func NewPrintQuotaStatusHandler() *printQuotaStatusHandler {
	return &printQuotaStatusHandler{
		statistics: make(map[string]*quota),
	}
}

// ResetStatistics output gathered statistics, process and prints them. After that it empties the statistics map
func (pqsh *printQuotaStatusHandler) ResetStatistics() {
	minQuota := &quota{
		numReceivedMessages:   math.MaxUint32,
		sizeReceivedMessages:  math.MaxUint64,
		numProcessedMessages:  math.MaxUint32,
		sizeProcessedMessages: math.MaxUint64,
	}
	avgQuota := &quota{}
	maxQuota := &quota{}

	pqsh.mutStatistics.Lock()
	defer pqsh.mutStatistics.Unlock()

	numStatistics := len(pqsh.statistics)

	if numStatistics == 0 {
		log.Trace("empty quota statistics")
		return
	}

	for name, q := range pqsh.statistics {
		avgQuota.numReceivedMessages += q.numReceivedMessages
		avgQuota.sizeReceivedMessages += q.sizeReceivedMessages
		avgQuota.numProcessedMessages += q.numProcessedMessages
		avgQuota.sizeProcessedMessages += q.sizeProcessedMessages

		minQuota.numReceivedMessages = pickMinUint32(minQuota.numReceivedMessages, q.numReceivedMessages)
		minQuota.sizeReceivedMessages = pickMinUint64(minQuota.sizeReceivedMessages, q.sizeReceivedMessages)
		minQuota.numProcessedMessages = pickMinUint32(minQuota.numProcessedMessages, q.numProcessedMessages)
		minQuota.sizeProcessedMessages = pickMinUint64(minQuota.sizeProcessedMessages, q.sizeProcessedMessages)

		maxQuota.numReceivedMessages = pickMaxUint32(maxQuota.numReceivedMessages, q.numReceivedMessages)
		maxQuota.sizeReceivedMessages = pickMaxUint64(maxQuota.sizeReceivedMessages, q.sizeReceivedMessages)
		maxQuota.numProcessedMessages = pickMaxUint32(maxQuota.numProcessedMessages, q.numProcessedMessages)
		maxQuota.sizeProcessedMessages = pickMaxUint64(maxQuota.sizeProcessedMessages, q.sizeProcessedMessages)
		log.Trace("peer quota statistics",
			"peer", name,
			"num received msg", q.numReceivedMessages,
			"size received", core.ConvertBytes(q.sizeReceivedMessages),
			"num processed msg", q.numProcessedMessages,
			"size processed", core.ConvertBytes(q.sizeProcessedMessages),
		)
	}

	log.Trace("quota statistics", "num peers", numStatistics)
	log.Trace("min quota statistics / peer",
		"num received msg", minQuota.numReceivedMessages,
		"size received", core.ConvertBytes(minQuota.sizeReceivedMessages),
		"num processed msg", minQuota.numProcessedMessages,
		"size processed", core.ConvertBytes(minQuota.sizeProcessedMessages),
	)
	log.Trace("avg quota statistics / peer",
		"num received msg", avgQuota.numReceivedMessages/uint32(numStatistics),
		"size received", core.ConvertBytes(avgQuota.sizeReceivedMessages/uint64(numStatistics)),
		"num processed msg", avgQuota.numProcessedMessages/uint32(numStatistics),
		"size processed", core.ConvertBytes(avgQuota.sizeProcessedMessages/uint64(numStatistics)),
	)
	log.Trace("max quota statistics / peer",
		"num received msg", maxQuota.numReceivedMessages,
		"size received", core.ConvertBytes(maxQuota.sizeReceivedMessages),
		"num processed msg", maxQuota.numProcessedMessages,
		"size processed", core.ConvertBytes(maxQuota.sizeProcessedMessages),
	)
	log.Trace("total quota statistics / network",
		"num received msg", avgQuota.numReceivedMessages,
		"size received", core.ConvertBytes(avgQuota.sizeReceivedMessages),
		"num processed msg", avgQuota.numProcessedMessages,
		"size processed", core.ConvertBytes(avgQuota.sizeProcessedMessages),
	)
}

func pickMinUint32(val1 uint32, val2 uint32) uint32 {
	if val1 > val2 {
		return val2
	}

	return val1
}

func pickMinUint64(val1 uint64, val2 uint64) uint64 {
	if val1 > val2 {
		return val2
	}

	return val1
}

func pickMaxUint32(val1 uint32, val2 uint32) uint32 {
	if val1 < val2 {
		return val2
	}

	return val1
}

func pickMaxUint64(val1 uint64, val2 uint64) uint64 {
	if val1 < val2 {
		return val2
	}

	return val1
}

// AddQuota adds a quota statistics
func (pqsh *printQuotaStatusHandler) AddQuota(identifier string, numReceivedMessages uint32, sizeReceivedMessages uint64,
	numProcessedMessages uint32, sizeProcessedMessages uint64) {

	q := &quota{
		numReceivedMessages:   numReceivedMessages,
		sizeReceivedMessages:  sizeReceivedMessages,
		numProcessedMessages:  numProcessedMessages,
		sizeProcessedMessages: sizeProcessedMessages,
	}

	pqsh.mutStatistics.Lock()
	pqsh.statistics[identifier] = q
	pqsh.mutStatistics.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (pqsh *printQuotaStatusHandler) IsInterfaceNil() bool {
	return pqsh == nil
}
