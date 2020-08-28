package storageResolvers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageResolver_ImplementedMethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("%v", r))
		}
	}()

	sr := &storageResolver{}
	assert.Nil(t, sr.ProcessReceivedMessage(nil, ""))
	assert.Nil(t, sr.SetResolverDebugHandler(nil))
	sr.SetNumPeersToQuery(0, 0)
	v1, v2 := sr.NumPeersToQuery()
	assert.Equal(t, 0, v1)
	assert.Equal(t, 0, v2)
}
