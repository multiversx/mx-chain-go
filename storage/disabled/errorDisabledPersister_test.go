package disabled

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewErrorDisabledPersister(t *testing.T) {
	t.Parallel()

	disabled := NewErrorDisabledPersister()
	assert.NotNil(t, disabled)
}

func TestErrorDisabledPersister_MethodsShouldError(t *testing.T) {
	t.Parallel()

	disabled := NewErrorDisabledPersister()
	t.Run("Put should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "disabledPersister.Put"
		err := disabled.Put(nil, nil)
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("Get should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "disabledPersister.Get"
		value, err := disabled.Get(nil)
		assert.Equal(t, expectedErrorString, err.Error())
		assert.Nil(t, value)
	})
	t.Run("Has should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "disabledPersister.Has"
		err := disabled.Has(nil)
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("Close should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "disabledPersister.Close"
		err := disabled.Close()
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("Remove should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "disabledPersister.Remove"
		err := disabled.Remove(nil)
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("Destroy should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "disabledPersister.Destroy"
		err := disabled.Destroy()
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("DestroyClosed should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "disabledPersister.DestroyClosed"
		err := disabled.DestroyClosed()
		assert.Equal(t, expectedErrorString, err.Error())
	})
}

func TestErrorDisabledPersister_RangeKeys(t *testing.T) {
	t.Parallel()

	disabled := NewErrorDisabledPersister()
	t.Run("nil handler should not panic", func(t *testing.T) {
		t.Parallel()

		assert.NotPanics(t, func() {
			disabled.RangeKeys(nil)
		})
	})
	t.Run("handler should not be called", func(t *testing.T) {
		t.Parallel()

		disabled.RangeKeys(func(key []byte, val []byte) bool {
			assert.Fail(t, "should have not called the handler")
			return false
		})
	})
}

func TestErrorDisabledPersister_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var edp *errorDisabledPersister
	require.True(t, edp.IsInterfaceNil())

	edp = NewErrorDisabledPersister()
	require.False(t, edp.IsInterfaceNil())
}
