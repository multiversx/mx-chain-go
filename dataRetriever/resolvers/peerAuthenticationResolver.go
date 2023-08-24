package resolvers

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// maxBuffToSendPeerAuthentications represents max buffer size to send in bytes
const maxBuffToSendPeerAuthentications = 1 << 18 // 256KB

// ArgPeerAuthenticationResolver is the argument structure used to create a new peer authentication resolver instance
type ArgPeerAuthenticationResolver struct {
	ArgBaseResolver
	PeerAuthenticationPool storage.Cacher
	DataPacker             dataRetriever.DataPacker
	PayloadValidator       dataRetriever.PeerAuthenticationPayloadValidator
}

// peerAuthenticationResolver is a wrapper over Resolver that is specialized in resolving peer authentication requests
type peerAuthenticationResolver struct {
	*baseResolver
	messageProcessor
	peerAuthenticationPool storage.Cacher
	dataPacker             dataRetriever.DataPacker
	payloadValidator       dataRetriever.PeerAuthenticationPayloadValidator
}

// NewPeerAuthenticationResolver creates a peer authentication resolver
func NewPeerAuthenticationResolver(arg ArgPeerAuthenticationResolver) (*peerAuthenticationResolver, error) {
	err := checkArgPeerAuthenticationResolver(arg)
	if err != nil {
		return nil, err
	}

	return &peerAuthenticationResolver{
		baseResolver: &baseResolver{
			TopicResolverSender: arg.SenderResolver,
		},
		messageProcessor: messageProcessor{
			marshalizer:      arg.Marshaller,
			antifloodHandler: arg.AntifloodHandler,
			throttler:        arg.Throttler,
			topic:            arg.SenderResolver.RequestTopic(),
		},
		peerAuthenticationPool: arg.PeerAuthenticationPool,
		dataPacker:             arg.DataPacker,
		payloadValidator:       arg.PayloadValidator,
	}, nil
}

func checkArgPeerAuthenticationResolver(arg ArgPeerAuthenticationResolver) error {
	err := checkArgBase(arg.ArgBaseResolver)
	if err != nil {
		return err
	}
	if check.IfNil(arg.PeerAuthenticationPool) {
		return dataRetriever.ErrNilPeerAuthenticationPool
	}
	if check.IfNil(arg.DataPacker) {
		return dataRetriever.ErrNilDataPacker
	}
	if check.IfNil(arg.PayloadValidator) {
		return dataRetriever.ErrNilPayloadValidator
	}
	return nil
}

// ProcessReceivedMessage represents the callback func from the p2p.Messenger that is called each time a new message is received
// (for the topic this validator was registered to, usually a request topic)
func (res *peerAuthenticationResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error {
	err := res.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	res.throttler.StartProcessing()
	defer res.throttler.EndProcessing()

	rd, err := res.parseReceivedMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	switch rd.Type {
	case dataRetriever.HashArrayType:
		return res.resolveMultipleHashesRequest(rd.Value, message.Peer(), source)
	default:
		err = dataRetriever.ErrRequestTypeNotImplemented
	}
	if err != nil {
		err = fmt.Errorf("%w for value %s", err, logger.DisplayByteSlice(rd.Value))
	}

	return err
}

// resolveMultipleHashesRequest sends the response for multiple hashes request
func (res *peerAuthenticationResolver) resolveMultipleHashesRequest(hashesBuff []byte, pid core.PeerID, source p2p.MessageHandler) error {
	b := batch.Batch{}
	err := res.marshalizer.Unmarshal(&b, hashesBuff)
	if err != nil {
		return err
	}
	hashes := b.Data

	peerAuthsForHashes, err := res.fetchPeerAuthenticationSlicesForPublicKeys(hashes)
	if err != nil {
		return fmt.Errorf("resolveMultipleHashesRequest error %w from buff %x", err, hashesBuff)
	}

	return res.sendPeerAuthsForHashes(peerAuthsForHashes, pid, source)
}

// sendPeerAuthsForHashes sends multiple peer authentication messages for specific hashes
func (res *peerAuthenticationResolver) sendPeerAuthsForHashes(dataBuff [][]byte, pid core.PeerID, source p2p.MessageHandler) error {
	buffsToSend, err := res.dataPacker.PackDataInChunks(dataBuff, maxBuffToSendPeerAuthentications)
	if err != nil {
		return err
	}

	for _, buff := range buffsToSend {
		err = res.Send(buff, pid, source)
		if err != nil {
			return err
		}
	}

	return nil
}

// fetchPeerAuthenticationSlicesForPublicKeys fetches all peer authentications for all pks
func (res *peerAuthenticationResolver) fetchPeerAuthenticationSlicesForPublicKeys(pks [][]byte) ([][]byte, error) {
	peerAuths := make([][]byte, 0)
	for _, pk := range pks {
		peerAuthForHash, _ := res.fetchPeerAuthenticationAsByteSlice(pk)
		if peerAuthForHash != nil {
			peerAuths = append(peerAuths, peerAuthForHash)
		}
	}

	if len(peerAuths) == 0 {
		return nil, dataRetriever.ErrPeerAuthNotFound
	}

	return peerAuths, nil
}

// fetchPeerAuthenticationAsByteSlice returns the value from authentication pool if exists
func (res *peerAuthenticationResolver) fetchPeerAuthenticationAsByteSlice(pk []byte) ([]byte, error) {
	value, ok := res.peerAuthenticationPool.Peek(pk)
	if !ok {
		return nil, dataRetriever.ErrPeerAuthNotFound
	}

	interceptedPeerAuthenticationData, ok := value.(*heartbeat.PeerAuthentication)
	if !ok {
		return nil, dataRetriever.ErrWrongTypeAssertion
	}

	payloadBuff := interceptedPeerAuthenticationData.Payload
	payload := &heartbeat.Payload{}
	err := res.marshalizer.Unmarshal(payload, payloadBuff)
	if err != nil {
		return nil, err
	}

	err = res.payloadValidator.ValidateTimestamp(payload.Timestamp)
	if err != nil {
		log.Trace("found peer authentication payload with invalid value, will not send it", "error", err)
		return nil, err
	}

	return res.marshalizer.Marshal(value)
}

// IsInterfaceNil returns true if there is no value under the interface
func (res *peerAuthenticationResolver) IsInterfaceNil() bool {
	return res == nil
}
