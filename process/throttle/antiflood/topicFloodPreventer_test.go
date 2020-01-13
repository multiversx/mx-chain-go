package antiflood_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood"
	"github.com/stretchr/testify/assert"
)

func TestNewTopicFloodPreventer_InvalidMaxNumOfMessagesShouldErr(t *testing.T) {
	t.Parallel()

	tfp, err := antiflood.NewTopicFloodPreventer(0)

	assert.Nil(t, tfp)
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewTopicFloodPreventer_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	tfp, err := antiflood.NewTopicFloodPreventer(10)

	assert.Nil(t, err)
	assert.NotNil(t, tfp)
	assert.False(t, tfp.IsInterfaceNil())
}

func TestTopicFloodPreventer_AccumulateOnceShouldWork(t *testing.T) {
	t.Parallel()

	tfp, _ := antiflood.NewTopicFloodPreventer(10)

	ok := tfp.Accumulate("ion", "topic_1")
	assert.True(t, ok)
}

func TestTopicFloodPreventer_AccumulateShouldReturnFalseIfNumberIsExceeded(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(2)
	tfp, _ := antiflood.NewTopicFloodPreventer(defaultMaxMessages)

	// no max limit is set for the topic, so the default value given as a parameter on the constructor will be used

	id := "identifier"
	topic := "topic_1"
	// call Accumulate 2 times so this will return true
	ok := tfp.Accumulate(id, topic)
	assert.True(t, ok)

	ok = tfp.Accumulate(id, topic)
	assert.True(t, ok)

	// this time, it should fail
	ok = tfp.Accumulate(id, topic)
	assert.False(t, ok)
}

func TestTopicFloodPreventer_AccumulateShouldReturnFalseIfNumberIsExceededUsingCustomLimitForTopic(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(2)
	customMaxMessages := uint32(3)
	tfp, _ := antiflood.NewTopicFloodPreventer(defaultMaxMessages)

	id := "identifier"
	topic := "topic_1"

	// set the limit for the topic, so it should use the custom value instead of the default
	tfp.SetMaxMessagesForTopic(topic, customMaxMessages)
	// call Accumulate 2 times so this will return true
	ok := tfp.Accumulate(id, topic)
	assert.True(t, ok)

	ok = tfp.Accumulate(id, topic)
	assert.True(t, ok)

	// this time, it should still pass because the max number of messages for the given topic was rewritten
	ok = tfp.Accumulate(id, topic)
	assert.True(t, ok)

	// now the custom value is reached, so the Accumulate should return false
	ok = tfp.Accumulate(id, topic)
	assert.False(t, ok)
}

func TestTopicFloodPreventer_ResetForTopic(t *testing.T) {
	t.Parallel()

	maxMessages := uint32(2)
	tfp, _ := antiflood.NewTopicFloodPreventer(maxMessages)

	id := "identifier"
	topic := "topic_1"

	// call Accumulate 2 times. it should work
	ok := tfp.Accumulate(id, topic)
	assert.True(t, ok)

	ok = tfp.Accumulate(id, topic)
	assert.True(t, ok)

	// now call Reset so we should be able to call Accumulate again with result true
	// If the Reset wouldn't have been called, then the method would have returned false
	tfp.ResetForTopic(topic)

	ok = tfp.Accumulate(id, topic)
	assert.True(t, ok)
}
