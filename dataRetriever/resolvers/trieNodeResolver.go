package resolvers

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

var _ dataRetriever.Resolver = (*TrieNodeResolver)(nil)

// maxBuffToSendTrieNodes represents max buffer size to send in bytes
var maxBuffToSendTrieNodes = 1 << 18 //256KB

// ArgTrieNodeResolver is the argument structure used to create new TrieNodeResolver instance
type ArgTrieNodeResolver struct {
	SenderResolver   dataRetriever.TopicResolverSender
	TrieDataGetter   dataRetriever.TrieDataGetter
	Marshalizer      marshal.Marshalizer
	AntifloodHandler dataRetriever.P2PAntifloodHandler
	Throttler        dataRetriever.ResolverThrottler
}

// TrieNodeResolver is a wrapper over Resolver that is specialized in resolving trie node requests
type TrieNodeResolver struct {
	dataRetriever.TopicResolverSender
	messageProcessor
	trieDataGetter dataRetriever.TrieDataGetter
}

// NewTrieNodeResolver creates a new trie node resolver
func NewTrieNodeResolver(arg ArgTrieNodeResolver) (*TrieNodeResolver, error) {
	if check.IfNil(arg.SenderResolver) {
		return nil, dataRetriever.ErrNilResolverSender
	}
	if check.IfNil(arg.TrieDataGetter) {
		return nil, dataRetriever.ErrNilTrieDataGetter
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(arg.AntifloodHandler) {
		return nil, dataRetriever.ErrNilAntifloodHandler
	}
	if check.IfNil(arg.Throttler) {
		return nil, dataRetriever.ErrNilThrottler
	}

	return &TrieNodeResolver{
		TopicResolverSender: arg.SenderResolver,
		trieDataGetter:      arg.TrieDataGetter,
		messageProcessor: messageProcessor{
			marshalizer:      arg.Marshalizer,
			antifloodHandler: arg.AntifloodHandler,
			topic:            arg.SenderResolver.RequestTopic(),
			throttler:        arg.Throttler,
		},
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (tnRes *TrieNodeResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
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
		return tnRes.resolveOneHash(rd.Value, message)
	case dataRetriever.HashArrayType:
		return tnRes.resolveMultipleHashes(rd.Value, message)
	default:
		return dataRetriever.ErrRequestTypeNotImplemented
	}
}

func (tnRes *TrieNodeResolver) resolveMultipleHashes(hashesBuff []byte, message p2p.MessageP2P) error {
	b := batch.Batch{}
	err := tnRes.marshalizer.Unmarshal(&b, hashesBuff)
	if err != nil {
		return err
	}
	hashes := b.Data

	nodes := make(map[string]struct{})
	spaceUsed, usedAllSpace := tnRes.resolveOnlyRequestedHashes(hashes, nodes)

	if usedAllSpace {
		return tnRes.sendResponse(convertMapToSlice(nodes), message)
	}

	tnRes.resolveSubTries(hashes, nodes, spaceUsed)

	return tnRes.sendResponse(convertMapToSlice(nodes), message)
}

func (tnRes *TrieNodeResolver) resolveOnlyRequestedHashes(hashes [][]byte, nodes map[string]struct{}) (int, bool) {
	spaceUsed := 0
	usedAllSpace := false
	remainingSpace := maxBuffToSendTrieNodes
	for _, hash := range hashes {
		serializedNode, err := tnRes.trieDataGetter.GetSerializedNode(hash)
		if err != nil {
			continue
		}

		if remainingSpace-len(serializedNode) < 0 {
			usedAllSpace = true
			break
		}

		spaceUsed += len(serializedNode)
		nodes[string(serializedNode)] = struct{}{}
		remainingSpace -= len(serializedNode)
	}

	return spaceUsed, usedAllSpace
}

func (tnRes *TrieNodeResolver) resolveSubTries(hashes [][]byte, nodes map[string]struct{}, spaceUsedAlready int) {
	var serialiazedNodes [][]byte
	var err error
	var serializedNode []byte
	for _, hash := range hashes {
		remainingForSubtries := maxBuffToSendTrieNodes - spaceUsedAlready
		if remainingForSubtries < 0 {
			return
		}

		serialiazedNodes, _, err = tnRes.getSubTrie(hash, uint64(remainingForSubtries))
		if err != nil {
			continue
		}

		for _, serializedNode = range serialiazedNodes {
			_, exists := nodes[string(serializedNode)]
			if exists {
				continue
			}

			if spaceUsedAlready-len(serializedNode) < 0 {
				break
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

func (tnRes *TrieNodeResolver) resolveOneHash(hash []byte, message p2p.MessageP2P) error {
	serializedNode, err := tnRes.trieDataGetter.GetSerializedNode(hash)
	if err != nil {
		return err
	}

	return tnRes.sendResponse([][]byte{serializedNode}, message)
}

func (tnRes *TrieNodeResolver) getSubTrie(hash []byte, remainingSpace uint64) ([][]byte, uint64, error) {
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

func (tnRes *TrieNodeResolver) sendResponse(serializedNodes [][]byte, message p2p.MessageP2P) error {
	buff, err := tnRes.marshalizer.Marshal(&batch.Batch{Data: serializedNodes})
	if err != nil {
		return err
	}

	return tnRes.Send(buff, message.Peer())
}

// RequestDataFromHash requests trie nodes from other peers having input a trie node hash
func (tnRes *TrieNodeResolver) RequestDataFromHash(hash []byte, _ uint32) error {
	return tnRes.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashType,
			Value: hash,
		},
		[][]byte{hash},
	)
}

// RequestDataFromHashArray requests trie nodes from other peers having input multiple trie node hashes
func (tnRes *TrieNodeResolver) RequestDataFromHashArray(hashes [][]byte, _ uint32) error {
	b := &batch.Batch{
		Data: hashes,
	}
	buffHashes, err := tnRes.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return tnRes.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffHashes,
		},
		hashes,
	)
}

// SetNumPeersToQuery will set the number of intra shard and cross shard number of peer to query
func (tnRes *TrieNodeResolver) SetNumPeersToQuery(intra int, cross int) {
	tnRes.TopicResolverSender.SetNumPeersToQuery(intra, cross)
}

// NumPeersToQuery will return the number of intra shard and cross shard number of peer to query
func (tnRes *TrieNodeResolver) NumPeersToQuery() (int, int) {
	return tnRes.TopicResolverSender.NumPeersToQuery()
}

// SetResolverDebugHandler will set a resolver debug handler
func (tnRes *TrieNodeResolver) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	return tnRes.TopicResolverSender.SetResolverDebugHandler(handler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnRes *TrieNodeResolver) IsInterfaceNil() bool {
	return tnRes == nil
}
