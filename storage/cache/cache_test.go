package cache

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewTimeCache_ShouldWork(t *testing.T) {
	t.Parallel()

	instance := NewTimeCache(0)
	assert.NotNil(t, instance)
}

func TestNewTimeCacher(t *testing.T) {
	t.Parallel()

	t.Run("invalid argument should error", func(t *testing.T) {
		t.Parallel()

		args := ArgTimeCacher{
			DefaultSpan: time.Second - time.Nanosecond,
			CacheExpiry: time.Second,
		}

		instance, err := NewTimeCacher(args)
		assert.Nil(t, instance)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := ArgTimeCacher{
			DefaultSpan: time.Second,
			CacheExpiry: time.Second,
		}

		instance, err := NewTimeCacher(args)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}

func TestNewLRUCache(t *testing.T) {
	t.Parallel()

	t.Run("invalid argument should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewLRUCache(0)
		assert.Nil(t, instance)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		instance, err := NewLRUCache(1)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}

func TestNewPeerTimeCache(t *testing.T) {
	t.Parallel()

	t.Run("invalid argument should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewPeerTimeCache(nil)
		assert.Nil(t, instance)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		instance, err := NewPeerTimeCache(&testscommon.TimeCacheStub{})
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}

func TestNewCapacityLRU(t *testing.T) {
	t.Parallel()

	t.Run("invalid argument should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewCapacityLRU(0, 1)
		assert.Nil(t, instance)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		instance, err := NewCapacityLRU(1, 1)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}

func TestNewLRUCacheWithEviction(t *testing.T) {
	t.Parallel()

	t.Run("invalid argument should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewLRUCacheWithEviction(0, nil)
		assert.Nil(t, instance)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		t.Run("nil handler should work", func(t *testing.T) {
			t.Parallel()

			instance, err := NewLRUCacheWithEviction(1, nil)
			assert.NotNil(t, instance)
			assert.Nil(t, err)
		})
		t.Run("with handler should work", func(t *testing.T) {
			t.Parallel()

			instance, err := NewLRUCacheWithEviction(1, func(key interface{}, value interface{}) {})
			assert.NotNil(t, instance)
			assert.Nil(t, err)
		})
	})
}

func TestNewImmunityCache(t *testing.T) {
	t.Parallel()

	t.Run("invalid argument should error", func(t *testing.T) {
		t.Parallel()

		config := CacheConfig{
			MaxNumBytes:                 0,
			MaxNumItems:                 0,
			NumChunks:                   0,
			Name:                        "test",
			NumItemsToPreemptivelyEvict: 0,
		}
		instance, err := NewImmunityCache(config)
		assert.Nil(t, instance)
		assert.NotNil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		config := CacheConfig{
			MaxNumBytes:                 4,
			MaxNumItems:                 4,
			NumChunks:                   1,
			Name:                        "test",
			NumItemsToPreemptivelyEvict: 1,
		}
		instance, err := NewImmunityCache(config)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}
