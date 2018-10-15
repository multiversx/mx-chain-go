package topics

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/stretchr/testify/assert"
)

var mockMarshalizer = mock.MarshalizerMock{}

func TestTopic_AddEventHandler_Nil_ShouldNotAddHandler(t *testing.T) {
	topic := NewTopic("test", "", &mockMarshalizer)

	topic.AddEventHandler(nil)

	assert.Equal(t, len(topic.eventBus), 0)
}

func TestTopic_AddEventHandler_WithARealFunc_ShouldWork(t *testing.T) {
	topic := NewTopic("test", "", &mockMarshalizer)

	topic.AddEventHandler(func(name string, data interface{}) {

	})

	assert.Equal(t, len(topic.eventBus), 1)
}
