package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

func TestDisabledAntiFlood_ShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		assert.Nil(t, r, "this shouldn't panic")
	}()

	daf := &disabledAntiFlood{}
	assert.False(t, check.IfNil(daf))

	daf.SetMaxMessagesForTopic("test", 10)
	daf.ResetForTopic("test")
	_ = daf.CanProcessMessageOnTopic(p2p.PeerID(1), "test")
	_ = daf.CanProcessMessage(nil, p2p.PeerID(2))
}
