package storageResolvers

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestTrieNodeResolver_ImplementedMethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	tnr := NewTrieNodeResolver()

	assert.False(t, check.IfNil(tnr))
	assert.Nil(t, tnr.RequestDataFromHash(nil, 0))
	assert.Nil(t, tnr.RequestDataFromHashArray(nil, 0))
}
