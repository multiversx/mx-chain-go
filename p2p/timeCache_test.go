package p2p

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeCacheAddHasAddingAfterDurShouldRetFalse(t *testing.T) {
	tc := NewTimeCache(time.Second)

	tc.Add("AAA")

	time.Sleep(time.Second * 2)

	assert.False(t, tc.Has("AAA"))
}

func TestTimeCacheAddHasAddingBeforeDurShouldRetTrue(t *testing.T) {
	tc := NewTimeCache(time.Second)

	tc.Add("AAA")

	time.Sleep(time.Millisecond * 500)

	assert.True(t, tc.Has("AAA"))
}

func TestTimeCacheAddHasReaddingAfterDurShouldWork(t *testing.T) {
	tc := NewTimeCache(time.Second)

	tc.Add("AAA")

	time.Sleep(time.Second * 2)

	if !tc.Has("AAA") {
		tc.Add("AAA")
	} else {
		assert.Fail(t, "Should have not had the object!")
	}

	time.Sleep(time.Millisecond * 500)

	assert.True(t, tc.Has("AAA"))
}
