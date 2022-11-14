package resolvers

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

var _ dataRetriever.Resolver = (*trieNodeResolver)(nil)
var logTrieNodes = logger.GetOrCreate("dataretriever/resolvers/trienoderesolver")

// ArgTrieNodeResolver is the argument structure used to create new trieNodeResolver instance
type ArgTrieNodeResolver struct {
	ArgBaseResolver
	TrieDataGetter dataRetriever.TrieDataGetter
}

// trieNodeResolver is a wrapper over Resolver that is specialized in resolving trie node requests
type trieNodeResolver struct {
	*baseResolver
	messageProcessor
	trieDataGetter dataRetriever.TrieDataGetter
}

// NewTrieNodeResolver creates a new trie node resolver
func NewTrieNodeResolver(arg ArgTrieNodeResolver) (*trieNodeResolver, error) {
	err := checkArgTrieNodeResolver(arg)
	if err != nil {
		return nil, err
	}

	return &trieNodeResolver{
		baseResolver: &baseResolver{
			TopicResolverSender: arg.SenderResolver,
		},
		trieDataGetter: arg.TrieDataGetter,
		messageProcessor: messageProcessor{
			marshalizer:      arg.Marshaller,
			antifloodHandler: arg.AntifloodHandler,
			topic:            arg.SenderResolver.RequestTopic(),
			throttler:        arg.Throttler,
		},
	}, nil
}

func checkArgTrieNodeResolver(arg ArgTrieNodeResolver) error {
	err := checkArgBase(arg.ArgBaseResolver)
	if err != nil {
		return err
	}
	if check.IfNil(arg.TrieDataGetter) {
		return dataRetriever.ErrNilTrieDataGetter
	}
	return nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (tnRes *trieNodeResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	err := tnRes.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	tnRes.throttler.StartProcessing()
	defer tnRes.throttler.EndProcessing()

	rd, err := tnRes.parseReceivedMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	switch rd.Type {
	case dataRetriever.HashType:
		return tnRes.resolveOneHash(rd.Value, rd.ChunkIndex, message)
	case dataRetriever.HashArrayType:
		return tnRes.resolveMultipleHashes(rd.Value, message)
	default:
		return dataRetriever.ErrRequestTypeNotImplemented
	}
}

func (tnRes *trieNodeResolver) resolveMultipleHashes(hashesBuff []byte, message p2p.MessageP2P) error {
	b := batch.Batch{}
	err := tnRes.marshalizer.Unmarshal(&b, hashesBuff)
	if err != nil {
		return err
	}
	hashes := b.Data

	supportedChunkIndex := uint32(0)
	nodes := make(map[string]struct{})
	spaceUsed, usedAllSpace := tnRes.resolveOnlyRequestedHashes(hashes, nodes)
	if usedAllSpace {
		return tnRes.sendResponse(convertMapToSlice(nodes), hashes, supportedChunkIndex, message)
	}

	tnRes.resolveSubTries(hashes, nodes, spaceUsed)

	return tnRes.sendResponse(convertMapToSlice(nodes), hashes, supportedChunkIndex, message)
}

func (tnRes *trieNodeResolver) resolveOnlyRequestedHashes(hashes [][]byte, nodes map[string]struct{}) (int, bool) {
	spaceUsed := 0
	usedAllSpace := false
	remainingSpace := core.MaxBufferSizeToSendTrieNodes
	for _, hash := range hashes {
		serializedNode, err := tnRes.trieDataGetter.GetSerializedNode(hash)
		if err != nil {
			continue
		}

		isNotFirstLargeElement := spaceUsed > 0 && (remainingSpace-len(serializedNode)) < 0
		if isNotFirstLargeElement {
			usedAllSpace = true
			break
		}

		spaceUsed += len(serializedNode)
		nodes[string(serializedNode)] = struct{}{}
		remainingSpace -= len(serializedNode)
	}

	usedAllSpace = usedAllSpace || remainingSpace == 0
	return spaceUsed, usedAllSpace
}

func (tnRes *trieNodeResolver) resolveSubTries(hashes [][]byte, nodes map[string]struct{}, spaceUsedAlready int) {
	var serializedNodes [][]byte
	var err error
	var serializedNode []byte
	for _, hash := range hashes {
		remainingForSubtries := core.MaxBufferSizeToSendTrieNodes - spaceUsedAlready
		if remainingForSubtries < 0 {
			return
		}

		serializedNodes, _, err = tnRes.getSubTrie(hash, uint64(remainingForSubtries))
		if err != nil {
			continue
		}

		for _, serializedNode = range serializedNodes {
			_, exists := nodes[string(serializedNode)]
			if exists {
				continue
			}

			if remainingForSubtries-len(serializedNode) < 0 {
				return
			}
			spaceUsedAlready += len(serializedNode)
			nodes[string(serializedNode)] = struct{}{}
		}
	}
}

func convertMapToSlice(m map[string]struct{}) [][]byte {
	buff := make([][]byte, 0, len(m))
	for data := range m {
		buff = append(buff, []byte(data))
	}

	return buff
}

func (tnRes *trieNodeResolver) resolveOneHash(hash []byte, chunkIndex uint32, message p2p.MessageP2P) error {
	serializedNode, err := tnRes.trieDataGetter.GetSerializedNode(hash)
	if err != nil {
		return err
	}

	return tnRes.sendResponse([][]byte{serializedNode}, [][]byte{hash}, chunkIndex, message)
}

func (tnRes *trieNodeResolver) getSubTrie(hash []byte, remainingSpace uint64) ([][]byte, uint64, error) {
	serializedNodes, remainingSpace, err := tnRes.trieDataGetter.GetSerializedNodes(hash, remainingSpace)
	if err != nil {
		tnRes.ResolverDebugHandler().LogFailedToResolveData(
			tnRes.topic,
			hash,
			err,
		)

		return nil, remainingSpace, err
	}

	tnRes.ResolverDebugHandler().LogSucceededToResolveData(tnRes.topic, hash)

	return serializedNodes, remainingSpace, nil
}

func (tnRes *trieNodeResolver) sendResponse(
	serializedNodes [][]byte,
	hashes [][]byte,
	chunkIndex uint32,
	message p2p.MessageP2P,
) error {

	if len(serializedNodes) == 0 {
		//do not send useless message
		return nil
	}

	if len(serializedNodes) == 1 && len(serializedNodes[0]) > core.MaxBufferSizeToSendTrieNodes {
		return tnRes.sendLargeMessage(serializedNodes[0], hashes[0], int(chunkIndex), message)
	}

	buff, err := tnRes.marshalizer.Marshal(&batch.Batch{Data: serializedNodes})
	if err != nil {
		return err
	}

	return tnRes.Send(buff, message.Peer())
}

func (tnRes *trieNodeResolver) sendLargeMessage(
	largeBuff []byte,
	reference []byte,
	chunkIndex int,
	message p2p.MessageP2P,
) error {

	logTrieNodes.Trace("assembling chunk", "reference", reference, "len", len(largeBuff))
	maxChunks := len(largeBuff) / core.MaxBufferSizeToSendTrieNodes
	if len(largeBuff)%core.MaxBufferSizeToSendTrieNodes != 0 {
		maxChunks++
	}
	chunkIndexOutOfBounds := chunkIndex < 0 || chunkIndex > maxChunks
	if chunkIndexOutOfBounds {
		return nil
	}

	startIndex := chunkIndex * core.MaxBufferSizeToSendTrieNodes
	endIndex := startIndex + core.MaxBufferSizeToSendTrieNodes
	if endIndex > len(largeBuff) {
		endIndex = len(largeBuff)
	}
	chunkBuffer := largeBuff[startIndex:endIndex]
	chunk := batch.NewChunk(uint32(chunkIndex), reference, uint32(maxChunks), chunkBuffer)
	logTrieNodes.Trace("assembled chunk", "index", chunkIndex, "reference", reference, "max chunks", maxChunks, "len", len(chunkBuffer))

	buff, err := tnRes.marshalizer.Marshal(chunk)
	if err != nil {
		return err
	}

	return tnRes.Send(buff, message.Peer())
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnRes *trieNodeResolver) IsInterfaceNil() bool {
	return tnRes == nil
}
