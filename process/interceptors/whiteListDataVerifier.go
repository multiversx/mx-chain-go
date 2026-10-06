package interceptors

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

const maxElementSize = 32 + 4 // hash + chunk_index_as_uint32

// numRoundsWhiteListRequestTTL is the number of rounds a whitelist entry added by a data request stays valid.
// It has to comfortably exceed the cadence of the routine re-request paths (the shard epoch start watchdog
// re-requests the epoch start metablock after 5 rounds without a metablock), so that data which is still
// legitimately awaited is not rejected, while keeping a past request from standing as a permanent authorisation.
const numRoundsWhiteListRequestTTL = 10

// defaultRoundDuration is used to compute the whitelist entries' lifespan when the chain's round
// duration is not available, so that a misconfiguration can not silently disable the expiry
const defaultRoundDuration = 6 * time.Second

var _ process.WhiteListHandler = (*whiteListDataVerifier)(nil)

type whiteListDataVerifier struct {
	cache          storage.Cacher
	ttl            time.Duration
	getTimeHandler func() time.Time
}

// NewWhiteListDataVerifier returns a default data verifier.
// The entries it holds expire after the provided ttl, so that a past request can not be used as a standing
// authorisation. A ttl of 0 disables the expiry, for the caches whose entries do not become stale.
func NewWhiteListDataVerifier(cacher storage.Cacher, ttl time.Duration) (*whiteListDataVerifier, error) {
	if check.IfNil(cacher) {
		return nil, fmt.Errorf("%w in NewWhiteListDataVerifier", process.ErrNilCacher)
	}
	if ttl < 0 {
		return nil, fmt.Errorf("%w in NewWhiteListDataVerifier, negative ttl", process.ErrInvalidValue)
	}

	return &whiteListDataVerifier{
		cache:          cacher,
		ttl:            ttl,
		getTimeHandler: time.Now,
	}, nil
}

// WhiteListRequestTTL returns the lifespan of a whitelist entry created by a data request,
// computed from the chain's round duration
func WhiteListRequestTTL(roundDuration time.Duration) time.Duration {
	if roundDuration <= 0 {
		log.Warn("WhiteListRequestTTL: invalid round duration, using the default one",
			"provided", roundDuration,
			"used", defaultRoundDuration,
		)
		roundDuration = defaultRoundDuration
	}

	return roundDuration * numRoundsWhiteListRequestTTL
}

// IsWhiteListed return true if intercepted data is accepted
func (w *whiteListDataVerifier) IsWhiteListed(interceptedData process.InterceptedData) bool {
	if check.IfNil(interceptedData) {
		return false
	}

	return w.IsWhiteListedAtLeastOne(interceptedData.Identifiers())
}

// IsWhiteListedAtLeastOne return true if at least one identifier from the slice is whitelisted
func (w *whiteListDataVerifier) IsWhiteListedAtLeastOne(identifiers [][]byte) bool {
	for _, identifier := range identifiers {
		if w.isKeyValid(identifier) {
			return true
		}
	}

	return false
}

func (w *whiteListDataVerifier) isKeyValid(key []byte) bool {
	value, found := w.cache.Get(key)
	if !found {
		return false
	}

	expiry, ok := value.(time.Time)
	if !ok {
		log.Warn("whitelist isKeyValid", "error", "unexpected entry type")
		return false
	}

	if expiry.IsZero() {
		return true
	}

	// Leave eviction to the cache; deletion could remove a concurrent refresh.
	return w.getTimeHandler().Before(expiry)
}

// Add adds all the list to the cache
func (w *whiteListDataVerifier) Add(keys [][]byte) {
	expiry := w.computeExpiry()
	for _, key := range keys {
		if len(key) > maxElementSize {
			log.Warn("whitelist add", "error", "key too large",
				"len", len(key),
			)
			continue
		}

		_ = w.cache.Put(key, expiry, 0)
	}
}

func (w *whiteListDataVerifier) computeExpiry() time.Time {
	if w.ttl == 0 {
		// the zero time marks an entry which never expires
		return time.Time{}
	}

	return w.getTimeHandler().Add(w.ttl)
}

// Remove removes all the keys from the cache
func (w *whiteListDataVerifier) Remove(keys [][]byte) {
	for _, key := range keys {
		w.cache.Remove(key)
	}
}

// IsInterfaceNil returns true if underlying object is nil
func (w *whiteListDataVerifier) IsInterfaceNil() bool {
	return w == nil
}
