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
	mutOperation       *sync.RWMutex
	cacher             storage.Cacher
	statusHandler      QuotaStatusHandler
	maxMessagesPerPeer uint32
	maxSizePerPeer     uint64
	maxMessages        uint32
	maxSize            uint64
	globalQuota        *quota
}

// NewQuotaFloodPreventer creates a new flood preventer based on quota / peer
func NewQuotaFloodPreventer(
	cacher storage.Cacher,
	statusHandler QuotaStatusHandler,
	maxMessagesPerPeer uint32,
	maxTotalSizePerPeer uint64,
	maxMessages uint32,
	maxTotalSize uint64,
) (*quotaFloodPreventer, error) {

	if check.IfNil(cacher) {
		return nil, process.ErrNilCacher
	}
	if check.IfNil(statusHandler) {
		return nil, process.ErrNilQuotaStatusHandler
	}
	if maxMessagesPerPeer < minMessages {
		return nil, fmt.Errorf("%w raised in NewCountersMap, maxMessagesPerPeer: provided %d, minimum %d",
			process.ErrInvalidValue,
			maxMessagesPerPeer,
			minMessages,
		)
	}
	if maxTotalSizePerPeer < minTotalSize {
		return nil, fmt.Errorf("%w raised in NewCountersMap, maxTotalSizePerPeer: provided %d, minimum %d",
			process.ErrInvalidValue,
			maxTotalSize,
			minTotalSize,
		)
	}
	if maxMessages < minMessages {
		return nil, fmt.Errorf("%w raised in NewCountersMap, maxMessages: provided %d, minimum %d",
			process.ErrInvalidValue,
			maxMessagesPerPeer,
			minMessages,
		)
	}
	if maxTotalSize < minTotalSize {
		return nil, fmt.Errorf("%w raised in NewCountersMap, maxTotalSize: provided %d, minimum %d",
			process.ErrInvalidValue,
			maxTotalSize,
			minTotalSize,
		)
	}

	return &quotaFloodPreventer{
		mutOperation:       &sync.RWMutex{},
		cacher:             cacher,
		statusHandler:      statusHandler,
		maxMessagesPerPeer: maxMessagesPerPeer,
		maxSizePerPeer:     maxTotalSizePerPeer,
		maxMessages:        maxMessages,
		maxSize:            maxTotalSize,
		globalQuota:        &quota{},
	}, nil
}

// IncrementAddingToSum tries to increment the counter values held at "identifier" position
// It returns true if it had succeeded incrementing (existing counter value is lower or equal with provided maxOperations)
// We need the mutOperation here as the get and put should be done atomically.
// Otherwise we might yield a slightly higher number of false valid increments
// This method also checks the global sum quota and increment its values
func (qfp *quotaFloodPreventer) IncrementAddingToSum(identifier string, size uint64) bool {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	qfp.globalQuota.numReceivedMessages++
	qfp.globalQuota.sizeReceivedMessages += size

	result := qfp.increment(identifier, size)
	if result {
		qfp.globalQuota.numProcessedMessages++
		qfp.globalQuota.sizeProcessedMessages += size
	}

	return result
}

// Increment tries to increment the counter values held at "identifier" position
// It returns true if it had succeeded incrementing (existing counter value is lower or equal with provided maxOperations)
// We need the mutOperation here as the get and put should be done atomically.
// Otherwise we might yield a slightly higher number of false valid increments
// This method also checks the global sum quota but does not increment its values
func (qfp *quotaFloodPreventer) Increment(identifier string, size uint64) bool {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	return qfp.increment(identifier, size)
}

func (qfp *quotaFloodPreventer) increment(identifier string, size uint64) bool {
	isGlobalQuotaReached := qfp.globalQuota.numReceivedMessages > qfp.maxMessages ||
		qfp.globalQuota.sizeReceivedMessages > qfp.maxSize
	if isGlobalQuotaReached {
		return false
	}

	valueQuota, ok := qfp.cacher.Get([]byte(identifier))
	if !ok {
		qfp.putDefaultQuota(identifier, size)

		return true
	}

	q, isQuota := valueQuota.(*quota)
	if !isQuota {
		qfp.putDefaultQuota(identifier, size)

		return true
	}

	q.numReceivedMessages++
	q.sizeReceivedMessages += size

	isPeerQuotaReached := q.numReceivedMessages > qfp.maxMessagesPerPeer ||
		q.sizeReceivedMessages > qfp.maxSizePerPeer
	if isPeerQuotaReached {
		return false
	}

	qfp.cacher.Put([]byte(identifier), q)
	q.numProcessedMessages++
	q.sizeProcessedMessages += size

	return true
}

func (qfp *quotaFloodPreventer) putDefaultQuota(identifier string, size uint64) {
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
	qfp.globalQuota = &quota{}
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

	qfp.statusHandler.SetGlobalQuota(
		qfp.globalQuota.numReceivedMessages,
		qfp.globalQuota.sizeReceivedMessages,
		qfp.globalQuota.numProcessedMessages,
		qfp.globalQuota.sizeProcessedMessages,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (qfp *quotaFloodPreventer) IsInterfaceNil() bool {
	return qfp == nil
}
