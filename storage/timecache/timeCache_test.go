package timecache_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/stretchr/testify/assert"
)

//------- Add

func TestTimeCache_AddShouldWork(t *testing.T) {
	t.Parallel()

	tc := timecache.NewTimeCache(time.Second)
	key := "key1"

	err := tc.Add(key, true)

	keys := tc.Keys()
	_, ok := tc.KeyTime(key)
	assert.Nil(t, err)
	assert.Equal(t, key, keys[0])
	assert.True(t, ok)
}

func TestTimeCache_DoubleAddShouldErrAndRetainTheKey(t *testing.T) {
	t.Parallel()

	tc := timecache.NewTimeCache(time.Second)
	key := "key1"

	_ = tc.Add(key, true)
	err := tc.Add(key, true)

	keys := tc.Keys()
	_, ok := tc.KeyTime(key)
	assert.Equal(t, storage.ErrDuplicateKeyToAdd, err)
	assert.Equal(t, key, keys[0])
	assert.True(t, ok)
}

func TestTimeCache_DoubleAddShouldAfterExpirationShouldWork(t *testing.T) {
	t.Parallel()

	tc := timecache.NewTimeCache(time.Millisecond)
	key := "key1"

	_ = tc.Add(key, true)
	time.Sleep(time.Second)
	err := tc.Add(key, true)

	keys := tc.Keys()
	_, ok := tc.KeyTime(key)
	assert.Nil(t, err)
	assert.Equal(t, key, keys[0])
	assert.True(t, ok)
}

//------- Has

func TestTimeCache_HasNotExistingShouldRetFalse(t *testing.T) {
	t.Parallel()

	tc := timecache.NewTimeCache(time.Second)
	key := "key1"

	exists := tc.Has(key, true)

	assert.False(t, exists)
}

func TestTimeCache_HasExistsShouldRetTrue(t *testing.T) {
	t.Parallel()

	tc := timecache.NewTimeCache(time.Second)
	key := "key1"
	_ = tc.Add(key, true)

	exists := tc.Has(key, true)

	assert.True(t, exists)
}

func TestTimeCache_HasCheckEvictionIsDoneProperly(t *testing.T) {
	t.Parallel()

	tc := timecache.NewTimeCache(time.Millisecond)
	key1 := "key1"
	key2 := "key2"
	_ = tc.Add(key1, true)
	_ = tc.Add(key2, true)
	time.Sleep(time.Second)

	exists1 := tc.Has(key1, true)
	exists2 := tc.Has(key2, true)

	assert.False(t, exists1)
	assert.False(t, exists2)
	assert.Equal(t, 0, len(tc.Keys()))
}

func TestTimeCache_HasCheckHandlingInconsistency(t *testing.T) {
	t.Parallel()

	tc := timecache.NewTimeCache(time.Second)
	key := "key1"
	_ = tc.Add(key, true)
	tc.ClearMap()

	exists := tc.Has(key, true)

	assert.False(t, exists)
	assert.Equal(t, 0, len(tc.Keys()))
}

//------- IsInterfaceNil

func TestTimeCache_IsInterfaceNilNotNil(t *testing.T) {
	t.Parallel()

	tc := timecache.NewTimeCache(time.Second)

	assert.False(t, check.IfNil(tc))
}

func TestTimeCache_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var tc *timecache.TimeCache

	assert.True(t, check.IfNil(tc))
}
