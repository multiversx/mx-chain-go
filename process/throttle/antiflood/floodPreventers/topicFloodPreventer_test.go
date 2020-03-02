package floodPreventers_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/floodPreventers"
	"github.com/stretchr/testify/assert"
)

func TestNewTopicFloodPreventer_InvalidMaxNumOfMessagesShouldErr(t *testing.T) {
	t.Parallel()

	tfp, err := floodPreventers.NewTopicFloodPreventer(0)

	assert.Nil(t, tfp)
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewTopicFloodPreventer_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	tfp, err := floodPreventers.NewTopicFloodPreventer(10)

	assert.Nil(t, err)
	assert.NotNil(t, tfp)
	assert.False(t, tfp.IsInterfaceNil())
}

func TestTopicFloodPreventer_AccumulateOnceShouldWork(t *testing.T) {
	t.Parallel()

	tfp, _ := floodPreventers.NewTopicFloodPreventer(10)

	ok := tfp.Accumulate("ion", "topic_1")
	assert.True(t, ok)
}

func TestTopicFloodPreventer_AccumulateShouldReturnFalseIfNumberIsExceeded(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(2)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(defaultMaxMessages)

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
	tfp, _ := floodPreventers.NewTopicFloodPreventer(defaultMaxMessages)

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
	tfp, _ := floodPreventers.NewTopicFloodPreventer(maxMessages)

	id := "identifier"
	topic := "topic_1"

	// call Accumulate 2 times. it should work
	ok := tfp.Accumulate(id, topic)
	assert.True(t, ok)

	ok = tfp.Accumulate(id, topic)
	assert.True(t, ok)

	assert.Equal(t, uint32(2), tfp.CountForTopicAndIdentifier(topic, id))

	// now call Reset so we should be able to call Accumulate again with result true
	// If the Reset wouldn't have been called, then the method would have returned false
	tfp.ResetForTopic(topic)

	assert.Equal(t, uint32(0), tfp.CountForTopicAndIdentifier(topic, id))
}

func TestTopicFloodPreventer_ResetForTopicWithBadWildcardNothingShouldHappen(t *testing.T) {
	t.Parallel()

	maxMessages := uint32(2)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(maxMessages)

	id := "identifier"
	topic1 := "topic_1"
	topic2 := "topic_2"

	// call Accumulate 2 times. it should work
	ok := tfp.Accumulate(id, topic1)
	assert.True(t, ok)
	ok = tfp.Accumulate(id, topic2)
	assert.True(t, ok)

	// check the values
	assert.Equal(t, uint32(1), tfp.CountForTopicAndIdentifier(topic1, id))
	assert.Equal(t, uint32(1), tfp.CountForTopicAndIdentifier(topic2, id))

	// try to call Reset with a bad wildcard. nothing should happen
	tfp.ResetForTopic("wrong*")

	// check the values again
	assert.Equal(t, uint32(1), tfp.CountForTopicAndIdentifier(topic1, id))
	assert.Equal(t, uint32(1), tfp.CountForTopicAndIdentifier(topic2, id))

	// now call Reset with a wildcarded topic. both of topics should have been reset
	tfp.ResetForTopic("topic*")

	// check the values again
	assert.Equal(t, uint32(0), tfp.CountForTopicAndIdentifier(topic1, id))
	assert.Equal(t, uint32(0), tfp.CountForTopicAndIdentifier(topic2, id))
}

func TestTopicFloodPreventer_ResetForTopicWithOkWildcardShouldReset(t *testing.T) {
	t.Parallel()

	maxMessages := uint32(2)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(maxMessages)

	id := "identifier"
	topic1 := "topic_1"
	topic2 := "topic_2"

	// call Accumulate for both topics. it should work
	ok := tfp.Accumulate(id, topic1)
	assert.True(t, ok)
	ok = tfp.Accumulate(id, topic2)
	assert.True(t, ok)

	// check the values
	assert.Equal(t, uint32(1), tfp.CountForTopicAndIdentifier(topic1, id))
	assert.Equal(t, uint32(1), tfp.CountForTopicAndIdentifier(topic2, id))

	// now call Reset with a wildcarded topic. both of topics should have been reset
	tfp.ResetForTopic("topic*")

	// check the values again
	assert.Equal(t, uint32(0), tfp.CountForTopicAndIdentifier(topic1, id))
	assert.Equal(t, uint32(0), tfp.CountForTopicAndIdentifier(topic2, id))
}

func TestTopicFloodPreventer_MaxMessagesOnWildcardTopicWorks(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(2)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(defaultMaxMessages)

	headersTopic := "headers"
	headersMaxMessages := uint32(100)
	tfp.SetMaxMessagesForTopic(headersTopic+floodPreventers.WildcardCharacter, headersMaxMessages)

	heartbeatTopic := "heartbeat"
	heartbeatMaxMessages := uint32(200)
	tfp.SetMaxMessagesForTopic(heartbeatTopic, heartbeatMaxMessages)

	//testing for the the wildcard topic
	assert.Equal(t, headersMaxMessages, tfp.MaxMessagesForTopic(headersTopic))
	assert.Equal(t, headersMaxMessages, tfp.MaxMessagesForTopic(headersTopic+"suffix"))
	assert.Equal(t, headersMaxMessages, tfp.MaxMessagesForTopic("prefix"+headersTopic))

	//testing for the topic without wildcard
	assert.Equal(t, heartbeatMaxMessages, tfp.MaxMessagesForTopic(heartbeatTopic))
	assert.Equal(t, defaultMaxMessages, tfp.MaxMessagesForTopic(heartbeatTopic+"suffix"))
}

func TestTopicFloodPreventer_MaxMessagesOnWildcardTopicCachesTheValue(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(2)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(defaultMaxMessages)

	headersTopic := "headers"
	headersMaxMessages := uint32(100)
	tfp.SetMaxMessagesForTopic(headersTopic+floodPreventers.WildcardCharacter, headersMaxMessages)

	maxMessagesMap := tfp.TopicMaxMessages()
	_, ok := maxMessagesMap[headersTopic]

	assert.False(t, ok)

	_ = tfp.MaxMessagesForTopic(headersTopic)

	maxMessagesMap = tfp.TopicMaxMessages()
	_, ok = maxMessagesMap[headersTopic]

	assert.True(t, ok)
}
