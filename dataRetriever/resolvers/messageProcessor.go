package resolvers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// messageProcessor is used for basic message validity and parsing
type messageProcessor struct {
	marshalizer      marshal.Marshalizer
	antifloodHandler dataRetriever.P2PAntifloodHandler
	throttler        dataRetriever.ResolverThrottler
	topic            string
}

func (mp *messageProcessor) canProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	err := mp.antifloodHandler.CanProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return fmt.Errorf("%w on resolver topic %s", err, mp.topic)
	}
	err = mp.antifloodHandler.CanProcessMessageOnTopic(fromConnectedPeer, mp.topic)
	if err != nil {
		return fmt.Errorf("%w on resolver topic %s", err, mp.topic)
	}
	if !mp.throttler.CanProcess() {
		return fmt.Errorf("%w on resolver topic %s", dataRetriever.ErrSystemBusy, mp.topic)
	}

	return nil
}

// parseReceivedMessage will transform the received p2p.Message in a RequestData object.
func (mp *messageProcessor) parseReceivedMessage(message p2p.MessageP2P) (*dataRetriever.RequestData, error) {
	rd := &dataRetriever.RequestData{}
	err := rd.Unmarshal(mp.marshalizer, message)
	if err != nil {
		return nil, err
	}
	if rd.Value == nil {
		return nil, dataRetriever.ErrNilValue
	}

	return rd, nil
}
