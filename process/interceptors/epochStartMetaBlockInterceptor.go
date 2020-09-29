package interceptors

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgsEpochStartMetaBlockInterceptor holds the arguments needed for creating a new epochStartMetaBlockInterceptor
type ArgsEpochStartMetaBlockInterceptor struct {
	Marshalizer               marshal.Marshalizer
	Hasher                    hashing.Hasher
	NumConnectedPeersProvider process.NumConnectedPeersProvider
	CurrentNetworkEpochSetter process.CurrentNetworkEpochSetter
	ConsensusPercentage       int
}

type epochStartMetaBlockInterceptor struct {
	marshalizer               marshal.Marshalizer
	hasher                    hashing.Hasher
	numConnectedPeersProvider process.NumConnectedPeersProvider
	currentNetworkEpochSetter process.CurrentNetworkEpochSetter
	consensusPercentage       float32
	mutReceivedMetaBlocks     sync.RWMutex
	mapReceivedMetaBlocks     map[string]*block.MetaBlock
	mapMetaBlocksFromPeers    map[string][]core.PeerID
}

// NewEpochStartMetaBlockInterceptor returns a new instance of epochStartMetaBlockInterceptor
func NewEpochStartMetaBlockInterceptor(args ArgsEpochStartMetaBlockInterceptor) (*epochStartMetaBlockInterceptor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	consensusPercentageFloat := float32(args.ConsensusPercentage) / 100.0
	return &epochStartMetaBlockInterceptor{
		marshalizer:               args.Marshalizer,
		hasher:                    args.Hasher,
		numConnectedPeersProvider: args.NumConnectedPeersProvider,
		currentNetworkEpochSetter: args.CurrentNetworkEpochSetter,
		consensusPercentage:       consensusPercentageFloat,
		mapReceivedMetaBlocks:     make(map[string]*block.MetaBlock),
		mapMetaBlocksFromPeers:    make(map[string][]core.PeerID),
	}, nil
}

// ProcessReceivedMessage will handle received messages containing epoch start meta blocks
func (e *epochStartMetaBlockInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	var epochStartMb block.MetaBlock
	err := e.marshalizer.Unmarshal(&epochStartMb, message.Data())
	if err != nil {
		return err
	}

	mbHash, err := core.CalculateHash(e.marshalizer, e.hasher, epochStartMb)
	if err != nil {
		return err
	}

	log.Debug("received epoch start meta", "epoch", epochStartMb.GetEpoch(), "from peer", fromConnectedPeer.Pretty())
	e.mutReceivedMetaBlocks.Lock()
	e.mapReceivedMetaBlocks[string(mbHash)] = &epochStartMb
	e.addToPeerList(string(mbHash), fromConnectedPeer)
	e.mutReceivedMetaBlocks.Unlock()

	metaBlock, found := e.checkMaps()
	if !found {
		return nil
	}

	e.handleFoundEpochStartMetaBlock(metaBlock)
	return nil
}

// this func should be called under mutex protection
func (e *epochStartMetaBlockInterceptor) addToPeerList(hash string, peer core.PeerID) {
	peersListForHash := e.mapMetaBlocksFromPeers[hash]
	for _, pid := range peersListForHash {
		if pid == peer {
			return
		}
	}
	e.mapMetaBlocksFromPeers[hash] = append(e.mapMetaBlocksFromPeers[hash], peer)
}

func (e *epochStartMetaBlockInterceptor) checkMaps() (*block.MetaBlock, bool) {
	e.mutReceivedMetaBlocks.RLock()
	defer e.mutReceivedMetaBlocks.RUnlock()

	numConnectedPeers := len(e.numConnectedPeersProvider.ConnectedPeers())
	numPeersTarget := int(float32(numConnectedPeers) * e.consensusPercentage)
	for hash, peersList := range e.mapMetaBlocksFromPeers {
		log.Debug("metablock from peers", "num peers", len(peersList), "hash", []byte(hash))
		metaBlock, found := e.processEntry(peersList, hash, numPeersTarget)
		if found {
			return metaBlock, true
		}
	}

	return nil, false
}

func (e *epochStartMetaBlockInterceptor) processEntry(
	peersList []core.PeerID,
	hash string,
	numPeersTarget int,
) (*block.MetaBlock, bool) {
	if len(peersList) >= numPeersTarget {
		log.Info("got consensus for epoch start metablock", "len", len(peersList))
		return e.mapReceivedMetaBlocks[hash], true
	}

	return nil, false
}

// SetInterceptedDebugHandler won't do anything
func (e *epochStartMetaBlockInterceptor) SetInterceptedDebugHandler(_ process.InterceptedDebugger) error {
	return nil
}

// RegisterHandler won't do anything as the logic is handled inside the interceptor
func (e *epochStartMetaBlockInterceptor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
}

func (e *epochStartMetaBlockInterceptor) handleFoundEpochStartMetaBlock(metaBlock *block.MetaBlock) {
	e.currentNetworkEpochSetter.SetCurrentEpoch(metaBlock.Epoch)
	e.resetMaps()
}

func (e *epochStartMetaBlockInterceptor) resetMaps() {
	e.mutReceivedMetaBlocks.Lock()
	e.mapMetaBlocksFromPeers = make(map[string][]core.PeerID)
	e.mapReceivedMetaBlocks = make(map[string]*block.MetaBlock)
	e.mutReceivedMetaBlocks.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *epochStartMetaBlockInterceptor) IsInterfaceNil() bool {
	return e == nil
}

func checkArgs(args ArgsEpochStartMetaBlockInterceptor) error {
	if check.IfNil(args.Marshalizer) {
		return wrapArgsError(process.ErrNilMarshalizer)
	}
	if check.IfNil(args.Hasher) {
		return wrapArgsError(process.ErrNilHasher)
	}
	if check.IfNil(args.NumConnectedPeersProvider) {
		return wrapArgsError(process.ErrNilNumConnectedPeersProvider)
	}
	if check.IfNil(args.CurrentNetworkEpochSetter) {
		return wrapArgsError(process.ErrNilCurrentNetworkEpochSetter)
	}
	if !(args.ConsensusPercentage >= 0 && args.ConsensusPercentage <= 100) {
		return wrapArgsError(process.ErrInvalidEpochStartMetaBlockConsensusPercentage)
	}

	return nil
}

func wrapArgsError(err error) error {
	return fmt.Errorf("%w when creating epochStartMetaBlockInterceptor", err)
}
