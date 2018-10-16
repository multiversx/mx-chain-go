package p2p

import (
	"sync"

	"github.com/pkg/errors"
)

type TopicHolder struct {
	mutTopics sync.RWMutex
	topics    map[string]*Topic

	OnNeedToRegisterTopic   func(topicName string) error
	OnNeedToSendDataOnTopic func(topicName string, buff []byte) error
}

func NewTopicHolder() *TopicHolder {
	th := TopicHolder{}
	th.topics = make(map[string]*Topic, 0)

	return &th
}

func (th *TopicHolder) has(topic *Topic) bool {
	if topic == nil {
		return false
	}

	_, ok := th.topics[topic.Name]

	return ok
}

func (th *TopicHolder) Has(topic *Topic) bool {
	th.mutTopics.RLock()
	defer th.mutTopics.RUnlock()

	return th.has(topic)
}

func (th *TopicHolder) AddTopic(topic *Topic) error {
	//check nil
	if topic == nil {
		return errors.New("nil topic to add")
	}

	//do not add twice
	th.mutTopics.Lock()
	defer th.mutTopics.Unlock()

	if th.has(topic) {
		return errors.New("topic already added")
	}

	if th.OnNeedToRegisterTopic == nil {
		return errors.New("can not register topic as OnNeedToRegisterTopic is not set")
	}

	if th.OnNeedToSendDataOnTopic == nil {
		return errors.New("can not register topic as OnNeedToSendDataTopic is not set")
	}

	th.topics[topic.Name] = topic
	//wire-up topics to main output sending func
	topic.OnNeedToSendMessage = func(topic string, buff []byte) error {
		//recheck whether OnNeedToSenddataOnTopic was not set to nil meanwhile
		if th.OnNeedToSendDataOnTopic != nil {
			return th.OnNeedToSendDataOnTopic(topic, buff)
		}

		return errors.New("can not send data as OnNeedToSendDataTopic was set to nil")
	}

	return th.OnNeedToRegisterTopic(topic.Name)
}

func (th *TopicHolder) GetTopic(topicName string) *Topic {
	th.mutTopics.RLock()
	defer th.mutTopics.RUnlock()

	val, ok := th.topics[topicName]

	if !ok {
		return nil
	}

	return val
}

// GotNewMessage will handle any input data
// it will discard any malformed data and will route the buffer to the correct Topic
func (th *TopicHolder) GotNewMessage(topicName string, buff []byte) {
	th.mutTopics.RLock()
	val, ok := th.topics[topicName]
	th.mutTopics.RUnlock()

	if !ok {
		return
	}

	val.NewMessageReceived(buff)
}
