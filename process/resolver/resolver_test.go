package resolver_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/resolver"
	"github.com/stretchr/testify/assert"
)

//------- NewResolver

func TestNewResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	_, err := resolver.NewResolver("test", nil, &mock.MarshalizerMock{})

	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	_, err := resolver.NewResolver("test", &mock.MessengerStub{}, nil)

	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewResolver_NilTopicShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	_, err := resolver.NewResolver("test", mes, &mock.MarshalizerMock{})

	assert.Equal(t, process.ErrNilTopic, err)
}

func TestNewResolver_TopicWithResolveRequestAssignedShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	topic := p2p.NewTopic("test", &mock.StringNewer{}, &mock.MarshalizerMock{})
	topic.ResolveRequest = func(hash []byte) []byte {
		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	_, err := resolver.NewResolver("test", mes, &mock.MarshalizerMock{})

	assert.Equal(t, process.ErrResolveRequestAlreadyAssigned, err)
}

func TestNewResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	topic := p2p.NewTopic("test", &mock.StringNewer{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, err := resolver.NewResolver("test", mes, &mock.MarshalizerMock{})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

//------- ResolveRequest

func TestResolver_ResolveRequestMarshalizerFailsShouldReturnNil(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}
	resMarshalizer.Fail = true

	topic := p2p.NewTopic("test", &mock.StringNewer{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, err := resolver.NewResolver("test", mes, resMarshalizer)

	assert.Nil(t, err)
	assert.NotNil(t, res)

	assert.Nil(t, topic.ResolveRequest([]byte("a")))
}

func TestResolver_ResolveRequestNilShouldReturnNil(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &mock.StringNewer{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, err := resolver.NewResolver("test", mes, resMarshalizer)

	assert.Nil(t, err)
	assert.NotNil(t, res)

	assert.Nil(t, topic.ResolveRequest(nil))
}

func TestResolver_ResolveRequestNilFuncShouldReturnNil(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &mock.StringNewer{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, err := resolver.NewResolver("test", mes, resMarshalizer)

	assert.Nil(t, err)
	assert.NotNil(t, res)

	rd := resolver.RequestData{
		Type:  resolver.HashType,
		Value: []byte("aaa"),
	}
	buffRd, err := resMarshalizer.Marshal(&rd)
	assert.Nil(t, err)

	assert.Nil(t, topic.ResolveRequest(buffRd))
}

func TestResolver_ResolveRequestShouldWork(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &mock.StringNewer{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, err := resolver.NewResolver("test", mes, resMarshalizer)

	res.ResolveRequest = func(rd *resolver.RequestData) []byte {
		return []byte("aaa")
	}

	assert.Nil(t, err)
	assert.NotNil(t, res)

	rd := resolver.RequestData{
		Type:  resolver.HashType,
		Value: []byte("aaa"),
	}
	buffRd, err := resMarshalizer.Marshal(&rd)
	assert.Nil(t, err)

	assert.Equal(t, []byte("aaa"), topic.ResolveRequest(buffRd))
}

//------- RequestData
