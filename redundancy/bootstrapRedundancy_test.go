package redundancy

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/redundancy/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewBootstrapNodeRedundancy(t *testing.T) {
	t.Parallel()

	t.Run("nil key should error", func(t *testing.T) {
		t.Parallel()

		bnr, err := NewBootstrapNodeRedundancy(nil)
		assert.Equal(t, ErrNilObserverPrivateKey, err)
		assert.True(t, check.IfNil(bnr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedKey := &mock.PrivateKeyStub{}
		bnr, err := NewBootstrapNodeRedundancy(providedKey)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(bnr))

		assert.False(t, bnr.IsRedundancyNode())
		assert.True(t, bnr.IsMainMachineActive())
		assert.Equal(t, providedKey, bnr.ObserverPrivateKey())
	})
}
