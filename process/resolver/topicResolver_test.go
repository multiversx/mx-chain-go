package resolver_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/resolver"
	"github.com/stretchr/testify/assert"
)

//-------NewTopicResolver

func TestNewTopicResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	tr, err := resolver.NewTopicResolver("test", nil, &mock.MarshalizerMock{})

	assert.Equal(t, process.ErrNilMessenger, err)
	assert.Nil(t, tr)
}

func TestNewTopicResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	tr, err := resolver.NewTopicResolver("test", &mock.MessengerStub{}, nil)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, tr)
}

func TestNewTopicResolver_NilTopicShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	tr, err := resolver.NewTopicResolver("test", mes, &mock.MarshalizerMock{})

	assert.Equal(t, process.ErrNilTopic, err)
	assert.Nil(t, tr)
}

func TestNewTopicResolver_TopicWithResolveRequestAssignedShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	topic := p2p.NewTopic("test", &mock.StringCreator{}, &mock.MarshalizerMock{})
	topic.ResolveRequest = func(hash []byte) []byte {
		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	tr, err := resolver.NewTopicResolver("test", mes, &mock.MarshalizerMock{})

	assert.Equal(t, process.ErrResolveRequestAlreadyAssigned, err)
	assert.Nil(t, tr)
}

func TestNewTopicResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	topic := p2p.NewTopic("test", &mock.StringCreator{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, err := resolver.NewTopicResolver("test", mes, &mock.MarshalizerMock{})

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

//------- ResolveRequest

func TestTopicResolver_ResolveRequestMarshalizerFailsShouldReturnNil(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}
	resMarshalizer.Fail = true

	topic := p2p.NewTopic("test", &mock.StringCreator{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	_, _ = resolver.NewTopicResolver("test", mes, resMarshalizer)

	assert.Nil(t, topic.ResolveRequest([]byte("a")))
}

func TestTopicResolver_ResolveRequestNilShouldReturnNil(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &mock.StringCreator{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	_, _ = resolver.NewTopicResolver("test", mes, resMarshalizer)

	assert.Nil(t, topic.ResolveRequest(nil))
}

func TestTopicResolver_ResolveRequestNilFuncShouldReturnNil(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &mock.StringCreator{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	_, _ = resolver.NewTopicResolver("test", mes, resMarshalizer)

	rd := process.RequestData{
		Type:  process.HashType,
		Value: []byte("aaa"),
	}
	buffRd, err := resMarshalizer.Marshal(&rd)
	assert.Nil(t, err)

	assert.Nil(t, topic.ResolveRequest(buffRd))
}

func TestTopicResolver_ResolveRequestShouldWork(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &mock.StringCreator{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, _ := resolver.NewTopicResolver("test", mes, resMarshalizer)

	res.SetResolverHandler(func(rd process.RequestData) ([]byte, error) {
		return []byte("aaa"), nil
	})

	rd := process.RequestData{
		Type:  process.HashType,
		Value: []byte("aaa"),
	}
	buffRd, err := resMarshalizer.Marshal(&rd)
	assert.Nil(t, err)

	assert.Equal(t, []byte("aaa"), topic.ResolveRequest(buffRd))
}

//------- RequestData

func TestTopicResolver_RequestDataMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}
	resMarshalizer.Fail = true

	topic := p2p.NewTopic("test", &mock.StringCreator{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, _ := resolver.NewTopicResolver("test", mes, resMarshalizer)

	assert.Equal(t, "MarshalizerMock generic error", res.RequestData(
		process.RequestData{
			Type:  process.HashType,
			Value: []byte("aaa"),
		}).Error())
}

func TestTopicResolver_RequestDataTopicNotWiredShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &mock.StringCreator{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, _ := resolver.NewTopicResolver("test", mes, resMarshalizer)

	assert.Equal(t, process.ErrTopicNotWiredToMessenger, res.RequestData(
		process.RequestData{
			Type:  process.HashType,
			Value: []byte("aaa"),
		}))
}

func TestTopicResolver_RequestDataShouldWork(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &mock.StringCreator{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, _ := resolver.NewTopicResolver("test", mes, resMarshalizer)

	wasCalled := false

	topic.Request = func(hash []byte) error {
		wasCalled = true
		return nil
	}

	assert.Nil(t, res.RequestData(
		process.RequestData{
			Type:  process.HashType,
			Value: []byte("aaa"),
		}))
	assert.True(t, wasCalled)
}

func TestTopicResolver_ResolverHandler(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	resMarshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &mock.StringCreator{}, &mock.MarshalizerMock{})

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	res, _ := resolver.NewTopicResolver("test", mes, resMarshalizer)

	//first test for nil
	assert.Nil(t, res.ResolverHandler())

	res.SetResolverHandler(func(rd process.RequestData) (bytes []byte, e error) {
		return nil, nil
	})

	//second, test is not nil
	assert.NotNil(t, res.ResolverHandler())
}
