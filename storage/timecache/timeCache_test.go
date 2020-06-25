package timecache

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//------- Add

func TestTimeCache_EmptyKeyShouldErr(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	key := ""

	err := tc.Add(key)

	_, ok := tc.Value(key)
	assert.Equal(t, storage.ErrEmptyKey, err)
	assert.False(t, ok)
}

func TestTimeCache_AddShouldWork(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	key := "key1"

	err := tc.Add(key)

	keys := tc.Keys()
	_, ok := tc.Value(key)
	assert.Nil(t, err)
	assert.Equal(t, key, keys[0])
	assert.True(t, ok)
}

func TestTimeCache_DoubleAddShouldWork(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	key := "key1"

	_ = tc.AddWithSpan(key, time.Second)
	newSpan := time.Second * 4
	err := tc.AddWithSpan(key, newSpan)
	assert.Nil(t, err)

	keys := tc.Keys()
	s, ok := tc.Value(key)
	assert.Equal(t, key, keys[0])
	assert.True(t, ok)
	assert.Equal(t, newSpan, s.span)
}

func TestTimeCache_DoubleAddAfterExpirationAndSweepShouldWork(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Millisecond)
	key := "key1"

	_ = tc.Add(key)
	time.Sleep(time.Second)
	tc.Sweep()
	err := tc.Add(key)

	keys := tc.Keys()
	_, ok := tc.Value(key)
	assert.Nil(t, err)
	assert.Equal(t, key, keys[0])
	assert.True(t, ok)
}

func TestTimeCache_AddWithSpanShouldWork(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	key := "key1"

	duration := time.Second * 1638
	err := tc.AddWithSpan(key, duration)

	keys := tc.Keys()
	_, ok := tc.Value(key)
	assert.Nil(t, err)
	assert.Equal(t, key, keys[0])
	assert.True(t, ok)

	spanRecovered, _ := tc.Value(key)
	assert.Equal(t, duration, spanRecovered.span)
}

//------- Has

func TestTimeCache_HasNotExistingShouldRetFalse(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	key := "key1"

	exists := tc.Has(key)

	assert.False(t, exists)
}

func TestTimeCache_HasExistsShouldRetTrue(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	key := "key1"
	_ = tc.Add(key)

	exists := tc.Has(key)

	assert.True(t, exists)
}

func TestTimeCache_HasCheckEvictionIsDoneProperly(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Millisecond)
	key1 := "key1"
	key2 := "key2"
	_ = tc.Add(key1)
	_ = tc.Add(key2)
	time.Sleep(time.Second)
	tc.Sweep()

	exists1 := tc.Has(key1)
	exists2 := tc.Has(key2)

	assert.False(t, exists1)
	assert.False(t, exists2)
	assert.Equal(t, 0, len(tc.Keys()))
}

func TestTimeCache_HasCheckHandlingInconsistency(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	key := "key1"
	_ = tc.Add(key)
	tc.ClearMap()
	tc.Sweep()

	exists := tc.Has(key)

	assert.False(t, exists)
	assert.Equal(t, 0, len(tc.Keys()))
}

//------- Upsert

func TestTimeCache_UpsertEmptyKeyShouldErr(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	err := tc.Upsert("", time.Second)

	assert.Equal(t, storage.ErrEmptyKey, err)
}

func TestTimeCache_UpsertShouldAddIfMissing(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	key := "key"
	s := time.Second * 45
	err := tc.Upsert(key, s)
	assert.Nil(t, err)

	recovered, ok := tc.Value(key)
	require.True(t, ok)
	assert.Equal(t, s, recovered.span)
}

func TestTimeCache_UpsertLessSpanShouldNotUpdate(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	key := "key"
	highSpan := time.Second * 45
	lowSpan := time.Second * 44
	err := tc.Upsert(key, highSpan)
	assert.Nil(t, err)

	err = tc.Upsert(key, lowSpan)
	assert.Nil(t, err)

	recovered, ok := tc.Value(key)
	require.True(t, ok)
	assert.Equal(t, highSpan, recovered.span)
}

func TestTimeCache_UpsertmoreSpanShouldUpdate(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)
	key := "key"
	highSpan := time.Second * 45
	lowSpan := time.Second * 44
	err := tc.Upsert(key, lowSpan)
	assert.Nil(t, err)

	err = tc.Upsert(key, highSpan)
	assert.Nil(t, err)

	recovered, ok := tc.Value(key)
	require.True(t, ok)
	assert.Equal(t, highSpan, recovered.span)
}

//------- IsInterfaceNil

func TestTimeCache_IsInterfaceNilNotNil(t *testing.T) {
	t.Parallel()

	tc := NewTimeCache(time.Second)

	assert.False(t, check.IfNil(tc))
}

func TestTimeCache_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var tc *TimeCache

	assert.True(t, check.IfNil(tc))
}
