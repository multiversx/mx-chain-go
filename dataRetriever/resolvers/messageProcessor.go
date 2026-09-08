package resolvers

import (
	"fmt"
	"math/rand/v2"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
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

func (mp *messageProcessor) parseRequestedHashes(
	hashesBuff []byte,
	fromConnectedPeer core.PeerID,
	sequence []byte,
) ([][]byte, error) {
	return mp.parseRequestedHashesWithCompatibility(hashesBuff, fromConnectedPeer, sequence, false)
}

func (mp *messageProcessor) parseRequestedHashesWithPartialResponse(
	hashesBuff []byte,
	fromConnectedPeer core.PeerID,
	sequence []byte,
) ([][]byte, error) {
	return mp.parseRequestedHashesWithCompatibility(hashesBuff, fromConnectedPeer, sequence, true)
}

func (mp *messageProcessor) parseRequestedHashesWithCompatibility(
	hashesBuff []byte,
	fromConnectedPeer core.PeerID,
	sequence []byte,
	allowPartialResponse bool,
) ([][]byte, error) {
	b := batch.Batch{}
	err := mp.marshalizer.Unmarshal(&b, hashesBuff)
	if err != nil {
		return nil, err
	}

	hashes := b.Data
	if len(hashes) > common.MaxHashesInRequest {
		if !allowPartialResponse {
			return nil, fmt.Errorf("%w: received %d hashes, maximum is %d", dataRetriever.ErrBadRequest, len(hashes), common.MaxHashesInRequest)
		}

		hashes = selectRequestedHashes(hashes)
	}
	numHashes := len(hashes)

	err = mp.antifloodHandler.CanProcessMessagesOnTopic(
		fromConnectedPeer,
		mp.topic,
		uint32(numHashes),
		uint64(len(hashesBuff)),
		sequence,
	)
	if err != nil {
		return nil, fmt.Errorf("%w on resolver topic %s", err, mp.topic)
	}

	return deduplicateHashes(hashes), nil
}

func selectRequestedHashes(hashes [][]byte) [][]byte {
	if len(hashes) <= common.MaxHashesInRequest {
		return hashes
	}

	numBatches := (len(hashes) + common.MaxHashesInRequest - 1) / common.MaxHashesInRequest
	batchIndex := rand.IntN(numBatches)
	startIndex := batchIndex * common.MaxHashesInRequest
	endIndex := core.MinInt(startIndex+common.MaxHashesInRequest, len(hashes))

	return hashes[startIndex:endIndex]
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
