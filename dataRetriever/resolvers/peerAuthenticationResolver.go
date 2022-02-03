package resolvers

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// maxNumOfPeerAuthenticationInResponse represents max num of peer authentication messages to send
const maxNumOfPeerAuthenticationInResponse = 50

// ArgPeerAuthenticationResolver is the argument structure used to create a new peer authentication resolver instance
type ArgPeerAuthenticationResolver struct {
	ArgBaseResolver
	PeerAuthenticationPool storage.Cacher
	DataPacker             dataRetriever.DataPacker
}

// peerAuthenticationResolver is a wrapper over Resolver that is specialized in resolving peer authentication requests
type peerAuthenticationResolver struct {
	*baseResolver
	messageProcessor
	peerAuthenticationPool storage.Cacher
	dataPacker             dataRetriever.DataPacker
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
			marshalizer:      arg.Marshalizer,
			antifloodHandler: arg.AntifloodHandler,
			throttler:        arg.Throttler,
			topic:            arg.SenderResolver.RequestTopic(),
		},
		peerAuthenticationPool: arg.PeerAuthenticationPool,
		dataPacker:             arg.DataPacker,
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
	return nil
}

// RequestDataFromHash requests peer authentication data from other peers having input a public key hash
func (res *peerAuthenticationResolver) RequestDataFromHash(hash []byte, _ uint32) error {
	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashType,
			Value: hash,
		},
		[][]byte{hash},
	)
}

// RequestDataFromHashArray requests peer authentication data from other peers having input multiple public key hashes
func (res *peerAuthenticationResolver) RequestDataFromHashArray(hashes [][]byte, _ uint32) error {
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
		},
		hashes,
	)
}

// RequestDataFromReferenceAndChunk requests a peer authentication chunk by specifying the reference and the chunk index
func (res *peerAuthenticationResolver) RequestDataFromReferenceAndChunk(hash []byte, chunkIndex uint32) error {
	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:       dataRetriever.HashType,
			Value:      hash,
			ChunkIndex: chunkIndex,
		},
		[][]byte{hash},
	)
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
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
	case dataRetriever.HashType:
		return res.resolveOneHash(rd.Value, int(rd.ChunkIndex), message.Peer())
	case dataRetriever.HashArrayType:
		// Todo add implementation
		err = dataRetriever.ErrRequestTypeNotImplemented
	default:
		err = dataRetriever.ErrRequestTypeNotImplemented
	}
	if err != nil {
		err = fmt.Errorf("%w for value %s", err, logger.DisplayByteSlice(rd.Value))
	}

	return err
}

func (res *peerAuthenticationResolver) resolveOneHash(hash []byte, chunkIndex int, pid core.PeerID) error {
	peerAuthMsgs := res.fetchPeerAuthenticationMessagesForHash(hash)
	if len(peerAuthMsgs) == 0 {
		return nil
	}

	if len(peerAuthMsgs) > maxNumOfPeerAuthenticationInResponse {
		return res.sendMessageFromChunk(hash, peerAuthMsgs, chunkIndex, pid)
	}

	return res.marshalAndSend(&batch.Batch{Data: peerAuthMsgs}, pid)
}

func (res *peerAuthenticationResolver) sendMessageFromChunk(hash []byte, peerAuthMsgs [][]byte, chunkIndex int, pid core.PeerID) error {
	maxChunks := len(peerAuthMsgs) / maxNumOfPeerAuthenticationInResponse
	if len(peerAuthMsgs)%maxNumOfPeerAuthenticationInResponse != 0 {
		maxChunks++
	}

	chunkIndexOutOfBounds := chunkIndex < 0 || chunkIndex > maxChunks
	if chunkIndexOutOfBounds {
		return nil
	}

	startingIndex := chunkIndex * maxNumOfPeerAuthenticationInResponse
	endIndex := startingIndex + maxNumOfPeerAuthenticationInResponse
	if endIndex > len(peerAuthMsgs) {
		endIndex = len(peerAuthMsgs)
	}
	messagesBuff := peerAuthMsgs[startingIndex:endIndex]
	chunk := batch.NewChunk(uint32(chunkIndex), hash, uint32(maxChunks), messagesBuff...)
	return res.marshalAndSend(chunk, pid)
}

func (res *peerAuthenticationResolver) marshalAndSend(message *batch.Batch, pid core.PeerID) error {
	buffToSend, err := res.marshalizer.Marshal(message)
	if err != nil {
		return err
	}

	return res.Send(buffToSend, pid)
}

func (res *peerAuthenticationResolver) fetchPeerAuthenticationMessagesForHash(hash []byte) [][]byte {
	var messages [][]byte

	keys := res.peerAuthenticationPool.Keys()
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i], keys[j]) < 0
	})

	for _, key := range keys {
		if bytes.Compare(hash, key[:len(hash)]) == 0 {
			peerAuth, _ := res.fetchPeerAuthenticationAsByteSlice(key)
			messages = append(messages, peerAuth)
		}
	}

	return messages
}

func (res *peerAuthenticationResolver) fetchPeerAuthenticationAsByteSlice(pk []byte) ([]byte, error) {
	value, ok := res.peerAuthenticationPool.Peek(pk)
	if ok {
		return res.marshalizer.Marshal(value)
	}

	return nil, dataRetriever.ErrNotFound
}

// IsInterfaceNil returns true if there is no value under the interface
func (res *peerAuthenticationResolver) IsInterfaceNil() bool {
	return res == nil
}
