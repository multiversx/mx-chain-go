package errors_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-storage/common/commonErrors"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/stretchr/testify/assert"
)

func TestIsClosingError(t *testing.T) {
	t.Parallel()

	t.Run("nil error should return false", func(t *testing.T) {
		t.Parallel()

		assert.False(t, errors.IsClosingError(nil))
	})
	t.Run("context closing error should return true", func(t *testing.T) {
		t.Parallel()

		assert.True(t, errors.IsClosingError(fmt.Errorf("%w random string", errors.ErrContextClosing)))
	})
	t.Run("DB closed error should return true", func(t *testing.T) {
		t.Parallel()

		assert.True(t, errors.IsClosingError(fmt.Errorf("%w random string", commonErrors.ErrDBIsClosed)))
	})
	t.Run("contains 'DB is closed' should return true", func(t *testing.T) {
		t.Parallel()

		assert.True(t, errors.IsClosingError(fmt.Errorf("random string DB is closed random string")))
	})
	t.Run("contains 'DB is closed' should return true", func(t *testing.T) {
		t.Parallel()

		assert.True(t, errors.IsClosingError(fmt.Errorf("random string context closing random string")))
	})
	t.Run("random error should return false", func(t *testing.T) {
		t.Parallel()

		assert.False(t, errors.IsClosingError(fmt.Errorf("random error")))
	})
}
