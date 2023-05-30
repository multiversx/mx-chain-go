package disabled

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledDataTrieHandler(t *testing.T) {
	t.Parallel()

	t.Run("new disabledDataTrieHandler", func(t *testing.T) {
		t.Parallel()

		assert.False(t, check.IfNil(NewDisabledDataTrieHandler()))
	})

	t.Run("root hash", func(t *testing.T) {
		t.Parallel()

		ddth := NewDisabledDataTrieHandler()

		rootHash, err := ddth.RootHash()
		assert.Equal(t, 0, len(rootHash))
		assert.Nil(t, err)
	})

	t.Run("get all leaves on channel", func(t *testing.T) {
		t.Parallel()

		ddth := NewDisabledDataTrieHandler()

		chans := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, 1),
			ErrChan:    errChan.NewErrChanWrapper(),
		}

		err := ddth.GetAllLeavesOnChannel(chans, nil, nil, nil, nil)
		assert.Nil(t, err)
		_, ok := <-chans.LeavesChan
		assert.False(t, ok)
		err = chans.ErrChan.ReadFromChanNonBlocking()
		assert.Nil(t, err)
	})
}
