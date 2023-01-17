package resolvers

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
)

// messageProcessor is used for basic message validity and parsing
type messageProcessor struct {
	marshalizer      marshal.Marshalizer
	antifloodHandler dataRetriever.P2PAntifloodHandler
	throttler        dataRetriever.ResolverThrottler
	topic            string
}

func (mp *messageProcessor) canProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if check.IfNil(message) {
		return dataRetriever.ErrNilMessage
	}
	err := mp.antifloodHandler.CanProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return fmt.Errorf("%w on resolver topic %s", err, mp.topic)
	}
	err = mp.antifloodHandler.CanProcessMessagesOnTopic(fromConnectedPeer, mp.topic, 1, uint64(len(message.Data())), message.SeqNo())
	if err != nil {
		return fmt.Errorf("%w on resolver topic %s", err, mp.topic)
	}
	if !mp.throttler.CanProcess() {
		return fmt.Errorf("%w on resolver topic %s", dataRetriever.ErrSystemBusy, mp.topic)
	}

	return nil
}

// parseReceivedMessage will transform the received p2p.Message in a RequestData object.
func (mp *messageProcessor) parseReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) (*dataRetriever.RequestData, error) {
	rd := &dataRetriever.RequestData{}
	err := rd.UnmarshalWith(mp.marshalizer, message)
	if err != nil {
		//this situation is so severe that we need to black list the peers
		reason := "unmarshalable data got on request topic " + mp.topic
		mp.antifloodHandler.BlacklistPeer(message.Peer(), reason, common.InvalidMessageBlacklistDuration)
		mp.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, common.InvalidMessageBlacklistDuration)

		return nil, err
	}
	if rd.Value == nil {
		return nil, dataRetriever.ErrNilValue
	}

	return rd, nil
}
