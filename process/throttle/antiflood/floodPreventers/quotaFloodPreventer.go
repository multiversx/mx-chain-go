package floodPreventers

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.FloodPreventer = (*quotaFloodPreventer)(nil)

const minMessages = 1
const minTotalSize = 1 //1Byte
const initNumMessages = 1
const maxPercentReserved = 90
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
	maxMessagesPerPeer uint32
	mutOperation       *sync.RWMutex
	cacher             storage.Cacher
	statusHandlers     []QuotaStatusHandler
	maxSizePerPeer     uint64
	percentReserved    uint32
}

// NewQuotaFloodPreventer creates a new flood preventer based on quota / peer
func NewQuotaFloodPreventer(
	cacher storage.Cacher,
	statusHandlers []QuotaStatusHandler,
	maxMessagesPerPeer uint32,
	maxTotalSizePerPeer uint64,
	percentReserved uint32,
) (*quotaFloodPreventer, error) {

	if check.IfNil(cacher) {
		return nil, process.ErrNilCacher
	}
	for _, statusHandler := range statusHandlers {
		if check.IfNil(statusHandler) {
			return nil, process.ErrNilQuotaStatusHandler
		}
	}
	if maxMessagesPerPeer < minMessages {
		return nil, fmt.Errorf("%w, maxMessagesPerPeer: provided %d, minimum %d",
			process.ErrInvalidValue,
			maxMessagesPerPeer,
			minMessages,
		)
	}
	if maxTotalSizePerPeer < minTotalSize {
		return nil, fmt.Errorf("%w, maxTotalSizePerPeer: provided %d, minimum %d",
			process.ErrInvalidValue,
			maxMessagesPerPeer,
			minTotalSize,
		)
	}
	if percentReserved > maxPercentReserved {
		return nil, fmt.Errorf("%w, percentReserved: provided %d, maximum %d",
			process.ErrInvalidValue,
			percentReserved,
			maxPercentReserved,
		)
	}

	return &quotaFloodPreventer{
		mutOperation:       &sync.RWMutex{},
		cacher:             cacher,
		statusHandlers:     statusHandlers,
		maxMessagesPerPeer: maxMessagesPerPeer,
		maxSizePerPeer:     maxTotalSizePerPeer,
		percentReserved:    percentReserved,
	}, nil
}

// IncreaseLoad tries to increment the counter values held at "identifier" position
// It returns true if it had succeeded incrementing (existing counter value is lower or equal with provided maxOperations)
// We need the mutOperation here as the get and put should be done atomically.
// Otherwise we might yield a slightly higher number of false valid increments
// This method also checks the global sum quota but does not increment its values
func (qfp *quotaFloodPreventer) IncreaseLoad(identifier string, size uint64) error {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	return qfp.increaseLoad(identifier, size)
}

func (qfp *quotaFloodPreventer) increaseLoad(identifier string, size uint64) error {
	valueQuota, ok := qfp.cacher.Get([]byte(identifier))
	if !ok {
		qfp.putDefaultQuota(identifier, size)

		return nil
	}

	q, isQuota := valueQuota.(*quota)
	if !isQuota {
		qfp.putDefaultQuota(identifier, size)

		return nil
	}

	q.numReceivedMessages++
	q.sizeReceivedMessages += size

	maxNumMessagesReached := qfp.isMaximumReached(uint64(qfp.maxMessagesPerPeer), uint64(q.numReceivedMessages))
	maxSizeMessagesReached := qfp.isMaximumReached(qfp.maxSizePerPeer, q.sizeReceivedMessages)
	isPeerQuotaReached := maxNumMessagesReached || maxSizeMessagesReached
	if isPeerQuotaReached {
		return fmt.Errorf("%w for pid %s", process.ErrSystemBusy, identifier)
	}

	q.numProcessedMessages++
	q.sizeProcessedMessages += size

	return nil
}

func (qfp *quotaFloodPreventer) isMaximumReached(absoluteMax uint64, counted uint64) bool {
	max := uint64(100-qfp.percentReserved) * absoluteMax / 100

	return counted > max
}

func (qfp *quotaFloodPreventer) putDefaultQuota(identifier string, size uint64) {
	q := &quota{
		numReceivedMessages:   initNumMessages,
		sizeReceivedMessages:  size,
		numProcessedMessages:  initNumMessages,
		sizeProcessedMessages: size,
	}
	qfp.cacher.Put([]byte(identifier), q, q.Size())
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
func (qfp quotaFloodPreventer) createStatistics() {
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
			string(k),
			q.numReceivedMessages,
			q.sizeReceivedMessages,
			q.numProcessedMessages,
			q.sizeProcessedMessages,
		)
	}
}

func (qfp *quotaFloodPreventer) addQuota(
	identifier string,
	numReceived uint32,
	sizeReceived uint64,
	numProcessed uint32,
	sizeProcessed uint64,
) {
	for _, statusHandler := range qfp.statusHandlers {
		statusHandler.AddQuota(identifier, numReceived, sizeReceived, numProcessed, sizeProcessed)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (qfp *quotaFloodPreventer) IsInterfaceNil() bool {
	return qfp == nil
}
