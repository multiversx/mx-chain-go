package peerDiscovery

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type MessageProcesssor struct {
	RequiredValue   []byte
	chanDone        chan struct{}
	mutDataReceived sync.Mutex
	wasDataReceived bool
}

func NewMessageProcessor(chanDone chan struct{}, requiredVal []byte) *MessageProcesssor {
	return &MessageProcesssor{
		RequiredValue: requiredVal,
		chanDone:      chanDone,
	}
}

func (mp *MessageProcesssor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	if bytes.Equal(mp.RequiredValue, message.Data()) {
		mp.mutDataReceived.Lock()
		mp.wasDataReceived = true
		mp.mutDataReceived.Unlock()

		mp.chanDone <- struct{}{}
	}

	return nil
}

func (mp *MessageProcesssor) WasDataReceived() bool {
	mp.mutDataReceived.Lock()
	defer mp.mutDataReceived.Unlock()

	return mp.wasDataReceived
}
