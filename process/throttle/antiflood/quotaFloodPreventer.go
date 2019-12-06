package antiflood

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const minMessages = 1
const minTotalSize = 1 //1Byte

type quota struct {
	numMessages uint32
	totalSize   uint64
}

// qoutaFloodPreventer represents a cache of quotas per peer used in antiflooding mechanism
type quotaFloodPreventer struct {
	mutOperation sync.Mutex
	cacher       storage.Cacher
	maxMessages  uint32
	maxSize      uint64
}

// NewQuotaFloodPreventer creates a new flood preventer based on quota / peer
func NewQuotaFloodPreventer(
	cacher storage.Cacher,
	maxMessagesPerPeer uint32,
	maxTotalSizePerPeer uint64,
) (*quotaFloodPreventer, error) {

	if cacher == nil {
		return nil, process.ErrNilCacher
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
		cacher:      cacher,
		maxMessages: maxMessagesPerPeer,
		maxSize:     maxTotalSizePerPeer,
	}, nil
}

// TryIncrement tries to increment the counter values held at "identifier" position
// It returns true if it had succeeded incrementing (existing counter value is lower or equal with provided maxOperations)
func (qfp *quotaFloodPreventer) TryIncrement(identifier string, size uint64) bool {
	//we need the mutOperation here as the get and put should be done atomically.
	// Otherwise we might yield a slightly higher number of false valid increments
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

	q.numMessages++
	q.totalSize += size
	isQuotaReached := q.numMessages > qfp.maxMessages || q.totalSize > qfp.maxSize
	if !isQuotaReached {
		qfp.cacher.Put([]byte(identifier), q)

		return true
	}

	return false
}

func (qfp *quotaFloodPreventer) putDefaultQuota(cacher storage.Cacher, identifier string, size uint64) {
	q := &quota{
		numMessages: 1,
		totalSize:   size,
	}
	qfp.cacher.Put([]byte(identifier), q)
}

// Reset clears all map values
func (qfp *quotaFloodPreventer) Reset() {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	//TODO change this if cacher.Clear() is time consuming
	qfp.cacher.Clear()
}

// IsInterfaceNil returns true if there is no value under the interface
func (qfp *quotaFloodPreventer) IsInterfaceNil() bool {
	return qfp == nil
}
