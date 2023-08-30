package resolvers

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func TestBaseFullHistoryResolver_SearchFirst(t *testing.T) {
	t.Parallel()

	testKey := []byte("key")
	testValue := []byte("value")
	resolver := &baseFullHistoryResolver{
		storer: &storage.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				assert.Equal(t, testKey, key)

				return testValue, nil
			},
		},
	}

	val, err := resolver.searchFirst(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testValue, val)
}

func TestBaseFullHistoryResolver_GetFromStorage(t *testing.T) {
	t.Parallel()

	testKey := []byte("key")
	testValue := []byte("value")
	testEpoch := uint32(37)

	t.Run("get from epoch returned nil error and not empty buffer", func(t *testing.T) {
		resolver := &baseFullHistoryResolver{
			storer: &storage.StorerStub{
				SearchFirstCalled: func(key []byte) ([]byte, error) {
					assert.Fail(t, "should have not called SearchFirst")

					return nil, nil
				},
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					assert.Equal(t, testKey, key)
					assert.Equal(t, testEpoch, epoch)

					return testValue, nil
				},
			},
		}

		val, err := resolver.getFromStorage(testKey, testEpoch)
		assert.Nil(t, err)
		assert.Equal(t, testValue, val)
	})
	t.Run("get from epoch returned nil error and nil buffer", func(t *testing.T) {
		resolver := &baseFullHistoryResolver{
			storer: &storage.StorerStub{
				SearchFirstCalled: func(key []byte) ([]byte, error) {
					assert.Fail(t, "should have not called SearchFirst")

					return nil, nil
				},
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					assert.Equal(t, testKey, key)
					assert.Equal(t, testEpoch, epoch)

					return nil, nil
				},
			},
		}

		val, err := resolver.getFromStorage(testKey, testEpoch)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(val))
	})
	t.Run("get from epoch returned error and will default to search first", func(t *testing.T) {
		resolver := &baseFullHistoryResolver{
			storer: &storage.StorerStub{
				SearchFirstCalled: func(key []byte) ([]byte, error) {
					assert.Equal(t, testKey, key)

					return testValue, nil
				},
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					assert.Equal(t, testKey, key)
					assert.Equal(t, testEpoch, epoch)

					return nil, errors.New("not found")
				},
			},
		}

		val, err := resolver.getFromStorage(testKey, testEpoch)
		assert.Nil(t, err)
		assert.Equal(t, testValue, val)
	})
}
