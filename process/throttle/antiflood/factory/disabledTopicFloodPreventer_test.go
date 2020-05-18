package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledTopicFloodPreventer_ShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		assert.Nil(t, r, "this shouldn't panic")
	}()

	daf := &disabledTopicFloodPreventer{}
	assert.False(t, check.IfNil(daf))

	_ = daf.IncreaseLoad("test", "test", 10)
	daf.SetMaxMessagesForTopic("test", 10)
	daf.ResetForNotRegisteredTopics()
	daf.ResetForTopic("topic")
}
