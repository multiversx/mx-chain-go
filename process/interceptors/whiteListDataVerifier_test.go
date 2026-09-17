package interceptors

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTTL = time.Minute

func TestNewWhiteListDataVerifier_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	wldv, err := NewWhiteListDataVerifier(nil, testTTL)

	assert.True(t, check.IfNil(wldv))
	assert.True(t, errors.Is(err, process.ErrNilCacher))
}

func TestNewWhiteListDataVerifier_NegativeTTLShouldErr(t *testing.T) {
	t.Parallel()

	wldv, err := NewWhiteListDataVerifier(cache.NewCacherStub(), -time.Second)

	assert.True(t, check.IfNil(wldv))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewWhiteListDataVerifier_ShouldWork(t *testing.T) {
	t.Parallel()

	wldv, err := NewWhiteListDataVerifier(cache.NewCacherStub(), testTTL)

	assert.False(t, check.IfNil(wldv))
	assert.Nil(t, err)
}

func TestWhiteListRequestTTL(t *testing.T) {
	t.Parallel()

	roundDuration := 6 * time.Second
	ttl := WhiteListRequestTTL(roundDuration)

	assert.Equal(t, roundDuration*numRoundsWhiteListRequestTTL, ttl)
	// the shard epoch start watchdog re-requests after 5 rounds without a metablock,
	// the entry must outlive that cadence
	assert.Greater(t, ttl, roundDuration*5)

	// an invalid round duration must not silently disable the expiry
	assert.Equal(t, defaultRoundDuration*numRoundsWhiteListRequestTTL, WhiteListRequestTTL(0))
	assert.Equal(t, defaultRoundDuration*numRoundsWhiteListRequestTTL, WhiteListRequestTTL(-time.Second))
}

func TestWhiteListDataVerifier_Add(t *testing.T) {
	t.Parallel()

	keys := [][]byte{[]byte("key1"), []byte("key2")}
	added := map[string]interface{}{}
	cacher := &cache.CacherStub{
		PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
			added[string(key)] = value
			return false
		},
	}
	wldv, _ := NewWhiteListDataVerifier(cacher, testTTL)
	now := time.Now()
	wldv.getTimeHandler = func() time.Time {
		return now
	}

	wldv.Add(keys)

	for _, key := range keys {
		value, ok := added[string(key)]
		assert.True(t, ok)
		assert.Equal(t, now.Add(testTTL), value)
	}
}

func TestWhiteListDataVerifier_AddWithoutTTLShouldStoreZeroExpiry(t *testing.T) {
	t.Parallel()

	added := map[string]interface{}{}
	cacher := &cache.CacherStub{
		PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
			added[string(key)] = value
			return false
		},
	}
	wldv, _ := NewWhiteListDataVerifier(cacher, 0)

	wldv.Add([][]byte{[]byte("key1")})

	value, ok := added["key1"]
	require.True(t, ok)
	expiry, ok := value.(time.Time)
	require.True(t, ok)
	assert.True(t, expiry.IsZero())
}

func TestWhiteListDataVerifier_AddKeyTooLargeShouldSkip(t *testing.T) {
	t.Parallel()

	numPutCalls := 0
	cacher := &cache.CacherStub{
		PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
			numPutCalls++
			return false
		},
	}
	wldv, _ := NewWhiteListDataVerifier(cacher, testTTL)

	wldv.Add([][]byte{make([]byte, maxElementSize+1)})

	assert.Equal(t, 0, numPutCalls)
}

func TestWhiteListDataVerifier_Remove(t *testing.T) {
	t.Parallel()

	keys := [][]byte{[]byte("key1"), []byte("key2")}
	removed := map[string]struct{}{}
	cacher := &cache.CacherStub{
		RemoveCalled: func(key []byte) {
			removed[string(key)] = struct{}{}
		},
	}
	wldv, _ := NewWhiteListDataVerifier(cacher, testTTL)

	wldv.Remove(keys)

	for _, key := range keys {
		_, ok := removed[string(key)]
		assert.True(t, ok)
	}
}

func TestWhiteListDataVerifier_IsWhiteListedNilInterceptedDataShouldRetFalse(t *testing.T) {
	t.Parallel()

	wldv, _ := NewWhiteListDataVerifier(cache.NewCacherStub(), testTTL)

	assert.False(t, wldv.IsWhiteListed(nil))
}

func TestWhiteListDataVerifier_IsWhiteListedNotFoundShouldRetFalse(t *testing.T) {
	t.Parallel()

	keyCheck := []byte("key")
	now := time.Now()
	wldv, _ := NewWhiteListDataVerifier(
		&cache.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				if bytes.Equal(key, keyCheck) {
					return nil, false
				}

				return now.Add(testTTL), true
			},
		},
		testTTL,
	)

	ids := &testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return keyCheck
		},
	}

	assert.False(t, wldv.IsWhiteListed(ids))
}

func TestWhiteListDataVerifier_IsWhiteListedFoundShouldRetTrue(t *testing.T) {
	t.Parallel()

	keyCheck := []byte("key")
	now := time.Now()
	wldv, _ := NewWhiteListDataVerifier(
		&cache.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				if !bytes.Equal(key, keyCheck) {
					return nil, false
				}

				return now.Add(testTTL), true
			},
		},
		testTTL,
	)
	wldv.getTimeHandler = func() time.Time {
		return now
	}

	ids := &testscommon.InterceptedDataStub{
		IdentifiersCalled: func() [][]byte {
			return [][]byte{keyCheck}
		},
	}

	assert.True(t, wldv.IsWhiteListed(ids))
}

func TestWhiteListDataVerifier_IsWhiteListedExpiryBehaviour(t *testing.T) {
	t.Parallel()

	keyCheck := []byte("key")
	ids := &testscommon.InterceptedDataStub{
		IdentifiersCalled: func() [][]byte {
			return [][]byte{keyCheck}
		},
	}

	createVerifier := func(ttl time.Duration) (*whiteListDataVerifier, *time.Time, *[][]byte) {
		entries := map[string]interface{}{}
		removedKeys := make([][]byte, 0)
		cacher := &cache.CacherStub{
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				entries[string(key)] = value
				return false
			},
			GetCalled: func(key []byte) (interface{}, bool) {
				value, ok := entries[string(key)]
				return value, ok
			},
			RemoveCalled: func(key []byte) {
				removedKeys = append(removedKeys, key)
				delete(entries, string(key))
			},
		}

		wldv, _ := NewWhiteListDataVerifier(cacher, ttl)
		currentTime := time.Now()
		wldv.getTimeHandler = func() time.Time {
			return currentTime
		}

		return wldv, &currentTime, &removedKeys
	}

	t.Run("entry is valid before the ttl elapses", func(t *testing.T) {
		t.Parallel()

		wldv, currentTime, removedKeys := createVerifier(testTTL)
		wldv.Add([][]byte{keyCheck})

		*currentTime = currentTime.Add(testTTL - time.Second)

		assert.True(t, wldv.IsWhiteListed(ids))
		assert.Empty(t, *removedKeys)
	})
	t.Run("entry is rejected and evicted once the ttl elapsed", func(t *testing.T) {
		t.Parallel()

		wldv, currentTime, removedKeys := createVerifier(testTTL)
		wldv.Add([][]byte{keyCheck})

		*currentTime = currentTime.Add(testTTL)

		assert.False(t, wldv.IsWhiteListed(ids))
		require.Equal(t, 1, len(*removedKeys))
		assert.Equal(t, keyCheck, (*removedKeys)[0])

		// a second check does not find it anymore
		assert.False(t, wldv.IsWhiteListed(ids))
	})
	t.Run("re-adding the key refreshes its expiry", func(t *testing.T) {
		t.Parallel()

		wldv, currentTime, _ := createVerifier(testTTL)
		wldv.Add([][]byte{keyCheck})

		*currentTime = currentTime.Add(testTTL - time.Second)
		wldv.Add([][]byte{keyCheck})

		*currentTime = currentTime.Add(testTTL - time.Second)

		assert.True(t, wldv.IsWhiteListed(ids))
	})
	t.Run("a zero ttl keeps the entry forever", func(t *testing.T) {
		t.Parallel()

		wldv, currentTime, removedKeys := createVerifier(0)
		wldv.Add([][]byte{keyCheck})

		*currentTime = currentTime.Add(100 * 24 * time.Hour)

		assert.True(t, wldv.IsWhiteListed(ids))
		assert.Empty(t, *removedKeys)
	})
	t.Run("an unexpected entry type is not honoured", func(t *testing.T) {
		t.Parallel()

		removedKeys := make([][]byte, 0)
		wldv, _ := NewWhiteListDataVerifier(
			&cache.CacherStub{
				GetCalled: func(key []byte) (interface{}, bool) {
					return struct{}{}, true
				},
				RemoveCalled: func(key []byte) {
					removedKeys = append(removedKeys, key)
				},
			},
			testTTL,
		)

		assert.False(t, wldv.IsWhiteListed(ids))
		assert.Equal(t, 1, len(removedKeys))
	})
}
