package interceptor

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/libp2p/go-libp2p-pubsub"
)

var log = logger.NewDefaultLogger()

// TopicInterceptor is a struct coupled with a p2p.Topic that calls CheckReceivedObject whenever Messenger needs to validate
// the data
type topicInterceptor struct {
	messenger      p2p.Messenger
	templateObject p2p.Creator
	name           string

	checkReceivedObject func(newer p2p.Creator, rawData []byte) error
}

// NewTopicInterceptor returns a new data interceptor that runs coupled with p2p.Topics
func NewTopicInterceptor(
	name string,
	messenger p2p.Messenger,
	templateObject p2p.Creator,
) (*topicInterceptor, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if templateObject == nil {
		return nil, process.ErrNilNewer
	}

	topic, err := getOrCreateTopic(name, templateObject, messenger)
	if err != nil {
		return nil, err
	}

	intercept := topicInterceptor{
		messenger:      messenger,
		templateObject: templateObject,
		name:           name,
	}

	err = topic.RegisterValidator(intercept.validator)
	if err != nil {
		return nil, process.ErrRegisteringValidator
	}

	return &intercept, nil
}

func getOrCreateTopic(name string, templateObject p2p.Creator, messenger p2p.Messenger) (*p2p.Topic, error) {
	existingTopic := messenger.GetTopic(name)

	if existingTopic != nil {
		return existingTopic, nil
	}

	topic := p2p.NewTopic(name, templateObject, messenger.Marshalizer())
	return topic, messenger.AddTopic(topic)
}

func (ti *topicInterceptor) validator(ctx context.Context, message *pubsub.Message) bool {
	obj := ti.templateObject.Create()

	marshalizer := ti.messenger.Marshalizer()
	err := marshalizer.Unmarshal(obj, message.GetData())

	if err != nil {
		return false
	}

	if ti.checkReceivedObject == nil {
		return false
	}

	err = ti.checkReceivedObject(obj, message.GetData())
	if err != nil {
		log.Debug(err.Error())
		return false
	}

	return true
}

// Name returns the name of the interceptor
func (ti *topicInterceptor) Name() string {
	return ti.name
}

// SetCheckReceivedObjectHandler sets the handler that gets called each time new data arrives in a form of
// a newer object
func (ti *topicInterceptor) SetCheckReceivedObjectHandler(handler func(newer p2p.Creator, rawData []byte) error) {
	ti.checkReceivedObject = handler
}

// CheckReceivedObjectHandler returns the handler that gets called each time new data arrives in a form of
// a newer object
func (ti *topicInterceptor) CheckReceivedObjectHandler() func(newer p2p.Creator, rawData []byte) error {
	return ti.checkReceivedObject
}
