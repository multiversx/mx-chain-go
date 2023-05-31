package disabled

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/stretchr/testify/assert"
)

func TestStorer_MethodsDoNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	s := NewStorer()
	assert.False(t, s.IsInterfaceNil())
	assert.Nil(t, s.Put(nil, nil))
	assert.Nil(t, s.PutInEpoch(nil, nil, 0))
	assert.Nil(t, s.Has(nil))
	assert.Nil(t, s.RemoveFromCurrentEpoch(nil))
	assert.Nil(t, s.Remove(nil))
	assert.Nil(t, s.DestroyUnit())
	assert.Nil(t, s.Close())

	val, err := s.Get(nil)
	assert.Nil(t, val)
	assert.Nil(t, err)

	val, err = s.SearchFirst(nil)
	assert.Nil(t, val)
	assert.Nil(t, err)

	val, err = s.GetFromEpoch(nil, 0)
	assert.Nil(t, val)
	assert.Nil(t, err)

	sliceVal, err := s.GetBulkFromEpoch(nil, 0)
	assert.Nil(t, sliceVal)
	assert.Nil(t, err)

	uintVal, err := s.GetOldestEpoch()
	assert.Equal(t, uint32(0), uintVal)
	assert.Equal(t, common.ErrOldestEpochNotAvailable, err)

	s.ClearCache()
	s.RangeKeys(nil)
}
