package timecache

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type peerTimeCache struct {
	timeCache storage.TimeCacher
}

// NewPeerTimeCache creates a new peer time cache data structure instance
func NewPeerTimeCache(timeCache storage.TimeCacher) (*peerTimeCache, error) {
	if check.IfNil(timeCache) {
		return nil, storage.ErrNilTimeCache
	}

	return &peerTimeCache{
		timeCache: timeCache,
	}, nil
}

// Upsert will add the pid and provided duration if not exists
// If the record exists, will update the duration if the provided duration is larger than existing
// Also, it will reset the contained timestamp to time.Now
func (ptc *peerTimeCache) Upsert(pid core.PeerID, duration time.Duration) error {
	return ptc.timeCache.Upsert(string(pid), duration)
}

// Sweep will call the inner time cache method
func (ptc *peerTimeCache) Sweep() {
	ptc.timeCache.Sweep()
}

// Has will call the inner time cache method with the provided pid as string
func (ptc *peerTimeCache) Has(pid core.PeerID) bool {
	return ptc.timeCache.Has(string(pid))
}

// IsInterfaceNil returns true if there is no value under the interface
func (ptc *peerTimeCache) IsInterfaceNil() bool {
	return ptc == nil
}
