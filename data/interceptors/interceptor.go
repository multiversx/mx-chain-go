package interceptors

import (
	"context"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/libp2p/go-libp2p-pubsub"
)

type Interceptor struct {
	messenger      p2p.Messenger
	templateObject p2p.Newer
	name           string

	CheckReceivedObject func(newer p2p.Newer, rawData []byte) bool
}

func NewInterceptor(name string, messenger p2p.Messenger, templateObject p2p.Newer) (*Interceptor, error) {
	if messenger == nil {
		return nil, ErrNilMessenger
	}

	if templateObject == nil {
		return nil, ErrNilNewer
	}

	topic := p2p.NewTopic(name, templateObject, messenger.Marshalizer())
	err := messenger.AddTopic(topic)

	if err != nil {
		return nil, err
	}

	intercept := Interceptor{
		messenger:      messenger,
		templateObject: templateObject,
		name:           name,
	}

	err = topic.RegisterValidator(intercept.validator)
	if err != nil {
		return nil, ErrRegisteringValidator
	}

	return &intercept, nil
}

func (intercept *Interceptor) Messenger() p2p.Messenger {
	return intercept.messenger
}

func (intercept *Interceptor) validator(ctx context.Context, message *pubsub.Message) bool {
	obj := intercept.templateObject.New()

	marshalizer := intercept.messenger.Marshalizer()
	err := marshalizer.Unmarshal(obj, message.GetData())

	if err != nil {
		return false
	}

	if intercept.CheckReceivedObject == nil {
		return false
	}

	return intercept.CheckReceivedObject(obj, message.GetData())
}
