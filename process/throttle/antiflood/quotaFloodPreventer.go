package antiflood

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const minMessages = 1
const minTotalSize = 1 //1Byte
const initNumMessages = 1

type quota struct {
	numReceivedMessages   uint32
	sizeReceivedMessages  uint64
	numProcessedMessages  uint32
	sizeProcessedMessages uint64
}

// quotaFloodPreventer represents a cache of quotas per peer used in antiflooding mechanism
type quotaFloodPreventer struct {
	mutOperation  sync.RWMutex
	cacher        storage.Cacher
	statusHandler QuotaStatusHandler
	maxMessages   uint32
	maxSize       uint64
}

// NewQuotaFloodPreventer creates a new flood preventer based on quota / peer
func NewQuotaFloodPreventer(
	cacher storage.Cacher,
	statusHandler QuotaStatusHandler,
	maxMessagesPerPeer uint32,
	maxTotalSizePerPeer uint64,
) (*quotaFloodPreventer, error) {

	if check.IfNil(cacher) {
		return nil, process.ErrNilCacher
	}
	if check.IfNil(statusHandler) {
		return nil, process.ErrNilQuotaStatusHandler
	}
	if maxMessagesPerPeer < minMessages {
		return nil, fmt.Errorf("%w raised in NewCountersMap, maxMessages: provided %d, minimum %d",
			process.ErrInvalidValue,
			maxMessagesPerPeer,
			minMessages,
		)
	}
	if maxTotalSizePerPeer < minTotalSize {
		return nil, fmt.Errorf("%w raised in NewCountersMap, maxTotalSize: provided %d, minimum %d",
			process.ErrInvalidValue,
			maxTotalSizePerPeer,
			minTotalSize,
		)
	}

	return &quotaFloodPreventer{
		cacher:        cacher,
		statusHandler: statusHandler,
		maxMessages:   maxMessagesPerPeer,
		maxSize:       maxTotalSizePerPeer,
	}, nil
}

// Increment tries to increment the counter values held at "identifier" position
// It returns true if it had succeeded incrementing (existing counter value is lower or equal with provided maxOperations)
// We need the mutOperation here as the get and put should be done atomically.
// Otherwise we might yield a slightly higher number of false valid increments
func (qfp *quotaFloodPreventer) Increment(identifier string, size uint64) bool {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	valueQuota, ok := qfp.cacher.Get([]byte(identifier))
	if !ok {
		qfp.putDefaultQuota(qfp.cacher, identifier, size)

		return true
	}

	q, isQuota := valueQuota.(*quota)
	if !isQuota {
		qfp.putDefaultQuota(qfp.cacher, identifier, size)

		return true
	}

	q.numReceivedMessages++
	q.sizeReceivedMessages += size
	isQuotaReached := q.numReceivedMessages > qfp.maxMessages || q.sizeReceivedMessages > qfp.maxSize
	if !isQuotaReached {
		qfp.cacher.Put([]byte(identifier), q)
		q.numProcessedMessages++
		q.sizeProcessedMessages += size

		return true
	}

	return false
}

func (qfp *quotaFloodPreventer) putDefaultQuota(cacher storage.Cacher, identifier string, size uint64) {
	q := &quota{
		numReceivedMessages:   initNumMessages,
		sizeReceivedMessages:  size,
		numProcessedMessages:  initNumMessages,
		sizeProcessedMessages: size,
	}
	qfp.cacher.Put([]byte(identifier), q)
}

// Reset clears all map values
func (qfp *quotaFloodPreventer) Reset() {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	qfp.createStatistics()

	//TODO change this if cacher.Clear() is time consuming
	qfp.cacher.Clear()
}

// createStatistics is useful to benchmark the system when running
func (qfp quotaFloodPreventer) createStatistics() {
	qfp.statusHandler.ResetStatistics()

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

		qfp.statusHandler.AddQuota(
			string(k),
			q.numReceivedMessages,
			q.sizeReceivedMessages,
			q.numProcessedMessages,
			q.sizeProcessedMessages,
		)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (qfp *quotaFloodPreventer) IsInterfaceNil() bool {
	return qfp == nil
}
