package disabled

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestCurrentPayloadProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	provider := &CurrentPayloadProvider{}
	assert.False(t, check.IfNil(provider))
	buff, isValid := provider.BytesToSendToNewPeers()
	assert.Empty(t, buff)
	assert.False(t, isValid)
}
