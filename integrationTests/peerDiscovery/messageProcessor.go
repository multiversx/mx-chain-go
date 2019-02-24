package peerDiscovery

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type MessageProcesssor struct {
	RequiredValue []byte
	chanDone      chan struct{}
}

func NewMessageProcessor(chanDone chan struct{}, requiredVal []byte) *MessageProcesssor {
	return &MessageProcesssor{
		RequiredValue: requiredVal,
		chanDone:      chanDone,
	}
}

func (mp *MessageProcesssor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	if bytes.Equal(mp.RequiredValue, message.Data()) {
		mp.chanDone <- struct{}{}
	}

	return nil
}
