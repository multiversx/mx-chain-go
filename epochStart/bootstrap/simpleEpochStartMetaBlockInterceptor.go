package bootstrap

import (
	"math"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

const timeToWaitBeforeCheckingReceivedHeaders = 1 * time.Second
const numTriesUntilExit = 5

type simpleEpochStartMetaBlockInterceptor struct {
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	mutReceivedMetaBlocks  sync.RWMutex
	mapReceivedMetaBlocks  map[string]*block.MetaBlock
	mapMetaBlocksFromPeers map[string][]p2p.PeerID
}

// NewSimpleEpochStartMetaBlockInterceptor will return a new instance of simpleEpochStartMetaBlockInterceptor
func NewSimpleEpochStartMetaBlockInterceptor(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*simpleEpochStartMetaBlockInterceptor, error) {
	if check.IfNil(marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, epochStart.ErrNilHasher
	}

	return &simpleEpochStartMetaBlockInterceptor{
		marshalizer:            marshalizer,
		hasher:                 hasher,
		mutReceivedMetaBlocks:  sync.RWMutex{},
		mapReceivedMetaBlocks:  make(map[string]*block.MetaBlock),
		mapMetaBlocksFromPeers: make(map[string][]p2p.PeerID),
	}, nil
}

// SetIsDataForCurrentShardVerifier -
func (s *simpleEpochStartMetaBlockInterceptor) SetIsDataForCurrentShardVerifier(_ process.InterceptedDataVerifier) error {
	return nil
}

// ProcessReceivedMessage will receive the metablocks and will add them to the maps
func (s *simpleEpochStartMetaBlockInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	metaBlock := &block.MetaBlock{}
	err := s.marshalizer.Unmarshal(metaBlock, message.Data())
	if err != nil {
		return err
	}

	if !metaBlock.IsStartOfEpochBlock() {
		return epochStart.ErrNotEpochStartBlock
	}

	s.mutReceivedMetaBlocks.Lock()
	mbHash, err := core.CalculateHash(s.marshalizer, s.hasher, metaBlock)
	if err != nil {
		s.mutReceivedMetaBlocks.Unlock()
		return err
	}

	s.mapReceivedMetaBlocks[string(mbHash)] = metaBlock
	s.addToPeerList(string(mbHash), message.Peer())
	s.mutReceivedMetaBlocks.Unlock()

	return nil
}

// this func should be called under mutex protection
func (s *simpleEpochStartMetaBlockInterceptor) addToPeerList(hash string, id p2p.PeerID) {
	peersListForHash, ok := s.mapMetaBlocksFromPeers[hash]
	if !ok {
		s.mapMetaBlocksFromPeers[hash] = append(s.mapMetaBlocksFromPeers[hash], id)
		return
	}

	for _, peer := range peersListForHash {
		if peer == id {
			return
		}
	}

	s.mapMetaBlocksFromPeers[hash] = append(s.mapMetaBlocksFromPeers[hash], id)
}

// GetEpochStartMetaBlock will return the metablock after it is confirmed or an error if the number of tries was exceeded
func (s *simpleEpochStartMetaBlockInterceptor) GetEpochStartMetaBlock(target int, epoch uint32) (*block.MetaBlock, error) {
	// TODO : replace this with a channel which will be written in when data is ready
	for count := 0; count < numTriesUntilExit; count++ {
		time.Sleep(timeToWaitBeforeCheckingReceivedHeaders)
		s.mutReceivedMetaBlocks.RLock()
		for hash, peersList := range s.mapMetaBlocksFromPeers {
			log.Debug("metablock from peers", "num peers", len(peersList), "target", target, "hash", []byte(hash))
			isOk := s.isMapEntryOk(peersList, hash, target, epoch)
			if isOk {
				s.mutReceivedMetaBlocks.RUnlock()
				metaBlockToReturn := s.mapReceivedMetaBlocks[hash]
				s.clearFields()
				return metaBlockToReturn, nil
			}
		}
		s.mutReceivedMetaBlocks.RUnlock()
	}

	return nil, epochStart.ErrNumTriesExceeded
}

func (s *simpleEpochStartMetaBlockInterceptor) isMapEntryOk(
	peersList []p2p.PeerID,
	hash string,
	target int,
	epoch uint32,
) bool {
	mb := s.mapReceivedMetaBlocks[hash]
	epochCheckNotRequired := epoch == math.MaxUint32
	isEpochOk := epochCheckNotRequired || mb.Epoch == epoch
	if len(peersList) >= target && isEpochOk {
		log.Info("got consensus for epoch start metablock", "len", len(peersList))
		return true
	}

	return false
}

func (s *simpleEpochStartMetaBlockInterceptor) clearFields() {
	s.mutReceivedMetaBlocks.Lock()
	s.mapReceivedMetaBlocks = make(map[string]*block.MetaBlock)
	s.mapMetaBlocksFromPeers = make(map[string][]p2p.PeerID)
	s.mutReceivedMetaBlocks.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *simpleEpochStartMetaBlockInterceptor) IsInterfaceNil() bool {
	return s == nil
}
