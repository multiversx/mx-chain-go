package floodPreventers_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/floodPreventers"
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

func TestTopicFloodPreventer_IncreaseLoadOnceShouldWork(t *testing.T) {
	t.Parallel()

	tfp, _ := floodPreventers.NewTopicFloodPreventer(10)

	err := tfp.IncreaseLoad("pid", "topic_1", 1)
	assert.Nil(t, err)
}

func TestTopicFloodPreventer_IncreaseLoadShouldReturnFalseIfNumberIsExceeded(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(2)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(defaultMaxMessages)

	// no max limit is set for the topic, so the default value given as a parameter on the constructor will be used

	id := core.PeerID("identifier")
	topic := "topic_1"
	// call Accumulate 2 times so this will return true
	err := tfp.IncreaseLoad(id, topic, 1)
	assert.Nil(t, err)

	err = tfp.IncreaseLoad(id, topic, 1)
	assert.Nil(t, err)

	// this time, it should fail
	err = tfp.IncreaseLoad(id, topic, 1)
	assert.Equal(t, process.ErrSystemBusy, err)
}

func TestTopicFloodPreventer_IncreaseLoadShouldReturnFalseIfNumberIsExceededUsingCustomLimitForTopic(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(2)
	customMaxMessages := uint32(3)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(defaultMaxMessages)

	id := core.PeerID("identifier")
	topic := "topic_1"

	// set the limit for the topic, so it should use the custom value instead of the default
	tfp.SetMaxMessagesForTopic(topic, customMaxMessages)
	// call Accumulate 2 times so this will return nil
	err := tfp.IncreaseLoad(id, topic, 1)
	assert.Nil(t, err)

	err = tfp.IncreaseLoad(id, topic, 1)
	assert.Nil(t, err)

	// this time, it should still pass because the max number of messages for the given topic was rewritten
	err = tfp.IncreaseLoad(id, topic, 1)
	assert.Nil(t, err)

	// now the custom value is reached, so the Accumulate should return error
	err = tfp.IncreaseLoad(id, topic, 1)
	assert.Equal(t, process.ErrSystemBusy, err)
}

func TestTopicFloodPreventer_IncreaseLoadShouldReturnFalseIfNumberIsExceededWithNumMessages(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(20)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(defaultMaxMessages)

	id := core.PeerID("identifier")
	topic := "topic_1"

	err := tfp.IncreaseLoad(id, topic, defaultMaxMessages-1)
	assert.Nil(t, err)

	tfp.ResetForTopic(topic)

	err = tfp.IncreaseLoad(id, topic, defaultMaxMessages)
	assert.Nil(t, err)

	tfp.ResetForTopic(topic)

	err = tfp.IncreaseLoad(id, topic, defaultMaxMessages+1)
	assert.Equal(t, process.ErrSystemBusy, err)

	tfp.ResetForTopic(topic)

	err = tfp.IncreaseLoad(id, topic, 5)
	assert.Nil(t, err)
	err = tfp.IncreaseLoad(id, topic, defaultMaxMessages-4)
	assert.Equal(t, process.ErrSystemBusy, err)
}

func TestTopicFloodPreventer_ResetForTopic(t *testing.T) {
	t.Parallel()

	maxMessages := uint32(2)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(maxMessages)

	id := core.PeerID("identifier")
	topic := "topic_1"

	// call IncreaseLoad 2 times. it should work
	err := tfp.IncreaseLoad(id, topic, 1)
	assert.Nil(t, err)

	err = tfp.IncreaseLoad(id, topic, 1)
	assert.Nil(t, err)

	assert.Equal(t, uint32(2), tfp.CountForTopicAndIdentifier(topic, id))

	// now call Reset, so we should be able to call Accumulate again with result nil
	// If the Reset hadn't been called, then the method would have errored
	tfp.ResetForTopic(topic)

	assert.Equal(t, uint32(0), tfp.CountForTopicAndIdentifier(topic, id))
}

func TestTopicFloodPreventer_ResetForTopicWithBadWildcardNothingShouldHappen(t *testing.T) {
	t.Parallel()

	maxMessages := uint32(2)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(maxMessages)

	id := core.PeerID("identifier")
	topic1 := "topic_1"
	topic2 := "topic_2"

	// call Accumulate 2 times. it should work
	err := tfp.IncreaseLoad(id, topic1, 1)
	assert.Nil(t, err)
	err = tfp.IncreaseLoad(id, topic2, 1)
	assert.Nil(t, err)

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

	id := core.PeerID("identifier")
	topic1 := "topic_1"
	topic2 := "topic_2"

	// call IncreaseLoad for both topics. it should work
	err := tfp.IncreaseLoad(id, topic1, 1)
	assert.Nil(t, err)
	err = tfp.IncreaseLoad(id, topic2, 1)
	assert.Nil(t, err)

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

	//testing for the wildcard topic
	assert.Equal(t, headersMaxMessages, tfp.MaxMessagesForTopic(headersTopic))
	assert.Equal(t, headersMaxMessages, tfp.MaxMessagesForTopic(headersTopic+"suffix"))
	assert.Equal(t, headersMaxMessages, tfp.MaxMessagesForTopic("prefix"+headersTopic))
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

func TestTopicFloodPreventer_ResetForNotRegisteredTopics(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(2)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(defaultMaxMessages)

	identifier := core.PeerID("pid")
	headersTopic := "headers"
	headersMaxMessages := uint32(100)
	tfp.SetMaxMessagesForTopic(headersTopic, headersMaxMessages)

	unregisteredTopic := "unregistered topic"

	err := tfp.IncreaseLoad(identifier, headersTopic, defaultMaxMessages)
	assert.Nil(t, err)

	err = tfp.IncreaseLoad(identifier, unregisteredTopic, defaultMaxMessages)
	assert.Nil(t, err)

	tfp.ResetForNotRegisteredTopics()

	//registered topic should not have been reset
	assert.Equal(t, defaultMaxMessages, tfp.CountForTopicAndIdentifier(headersTopic, identifier))
	//unregistered topic should have been reset
	assert.Equal(t, uint32(0), tfp.CountForTopicAndIdentifier(unregisteredTopic, identifier))

	err = tfp.IncreaseLoad(identifier, unregisteredTopic, defaultMaxMessages)
	assert.Nil(t, err)
}

func TestTopicFloodPreventer_ResetForNotRegisteredTopicsWithWildcardShuldWork(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(2)
	tfp, _ := floodPreventers.NewTopicFloodPreventer(defaultMaxMessages)

	identifier := core.PeerID("pid")
	headersTopic := "headers"
	headersMaxMessages := uint32(100)
	tfp.SetMaxMessagesForTopic(headersTopic+floodPreventers.WildcardCharacter, headersMaxMessages)

	unregisteredTopic := "unregistered topic"

	err := tfp.IncreaseLoad(identifier, headersTopic, defaultMaxMessages)
	assert.Nil(t, err)

	err = tfp.IncreaseLoad(identifier, unregisteredTopic, defaultMaxMessages)
	assert.Nil(t, err)

	tfp.ResetForNotRegisteredTopics()

	//registered topic should not have been reset
	assert.Equal(t, defaultMaxMessages, tfp.CountForTopicAndIdentifier(headersTopic, identifier))
	//unregistered topic should have been reset
	assert.Equal(t, uint32(0), tfp.CountForTopicAndIdentifier(unregisteredTopic, identifier))

	err = tfp.IncreaseLoad(identifier, unregisteredTopic, defaultMaxMessages)
	assert.Nil(t, err)
}
