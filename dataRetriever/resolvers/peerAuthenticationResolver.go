package resolvers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
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

// RequestDataFromHash requests peer authentication data from other peers having input a public key hash
func (res *peerAuthenticationResolver) RequestDataFromHash(hash []byte, epoch uint32) error {
	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashType,
			Value: hash,
			Epoch: epoch,
		},
		[][]byte{hash},
	)
}

// RequestDataFromHashArray requests peer authentication data from other peers having input multiple public key hashes
func (res *peerAuthenticationResolver) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	b := &batch.Batch{
		Data: hashes,
	}
	buffHashes, err := res.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffHashes,
			Epoch: epoch,
		},
		hashes,
	)
}

// ProcessReceivedMessage represents the callback func from the p2p.Messenger that is called each time a new message is received
// (for the topic this validator was registered to, usually a request topic)
func (res *peerAuthenticationResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
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
		return res.resolveMultipleHashesRequest(rd.Value, message.Peer())
	default:
		err = dataRetriever.ErrRequestTypeNotImplemented
	}
	if err != nil {
		err = fmt.Errorf("%w for value %s", err, logger.DisplayByteSlice(rd.Value))
	}

	return err
}

// resolveMultipleHashesRequest sends the response for multiple hashes request
func (res *peerAuthenticationResolver) resolveMultipleHashesRequest(hashesBuff []byte, pid core.PeerID) error {
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

	return res.sendPeerAuthsForHashes(peerAuthsForHashes, pid)
}

// sendPeerAuthsForHashes sends multiple peer authentication messages for specific hashes
func (res *peerAuthenticationResolver) sendPeerAuthsForHashes(dataBuff [][]byte, pid core.PeerID) error {
	buffsToSend, err := res.dataPacker.PackDataInChunks(dataBuff, maxBuffToSendPeerAuthentications)
	if err != nil {
		return err
	}

	for _, buff := range buffsToSend {
		err = res.Send(buff, pid)
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
