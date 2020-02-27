package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type simpleShardHeaderInterceptor struct {
	marshalizer      marshal.Marshalizer
	receivedHandlers []block.ShardData
}

func NewSimpleShardHeaderInterceptor(marshalizer marshal.Marshalizer) *simpleShardHeaderInterceptor {
	return &simpleShardHeaderInterceptor{
		marshalizer:      marshalizer,
		receivedHandlers: make([]block.ShardData, 0),
	}
}

func (s *simpleShardHeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
	var hdr block.ShardData
	err := s.marshalizer.Unmarshal(&hdr, message.Data())
	if err == nil {
		s.receivedHandlers = append(s.receivedHandlers, hdr)
	}

	return nil
}

func (s *simpleShardHeaderInterceptor) GetAllReceivedShardHeaders() []block.ShardData {
	return s.receivedHandlers
}

func (s *simpleShardHeaderInterceptor) IsInterfaceNil() bool {
	return s == nil
}
