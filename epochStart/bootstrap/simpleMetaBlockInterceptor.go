package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type simpleMetaBlockInterceptor struct {
	marshalizer      marshal.Marshalizer
	receivedHandlers []block.MetaBlock
}

func NewSimpleMetaBlockInterceptor(marshalizer marshal.Marshalizer) *simpleMetaBlockInterceptor {
	return &simpleMetaBlockInterceptor{
		marshalizer:      marshalizer,
		receivedHandlers: make([]block.MetaBlock, 0),
	}
}

func (s *simpleMetaBlockInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	var hdr block.MetaBlock
	err := s.marshalizer.Unmarshal(&hdr, message.Data())
	if err == nil {
		s.receivedHandlers = append(s.receivedHandlers, hdr)
	}

	return nil
}

func (s *simpleMetaBlockInterceptor) GetAllReceivedMetaBlocks() []block.MetaBlock {
	return s.receivedHandlers
}

func (s *simpleMetaBlockInterceptor) IsInterfaceNil() bool {
	return s == nil
}
