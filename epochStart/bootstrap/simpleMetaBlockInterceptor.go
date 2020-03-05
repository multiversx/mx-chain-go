package bootstrap

import (
	"math"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

const timeToWaitBeforeCheckingReceivedMetaBlocks = 500 * time.Millisecond
const numTriesUntilExit = 5

type simpleMetaBlockInterceptor struct {
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	mutReceivedMetaBlocks  sync.RWMutex
	mapReceivedMetaBlocks  map[string]*block.MetaBlock
	mapMetaBlocksFromPeers map[string][]p2p.PeerID
}

// NewSimpleMetaBlockInterceptor will return a new instance of simpleMetaBlockInterceptor
func NewSimpleMetaBlockInterceptor(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*simpleMetaBlockInterceptor, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	return &simpleMetaBlockInterceptor{
		marshalizer:            marshalizer,
		hasher:                 hasher,
		mutReceivedMetaBlocks:  sync.RWMutex{},
		mapReceivedMetaBlocks:  make(map[string]*block.MetaBlock),
		mapMetaBlocksFromPeers: make(map[string][]p2p.PeerID),
	}, nil
}

// ProcessReceivedMessage will receive the metablocks and will add them to the maps
func (s *simpleMetaBlockInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	var mb block.MetaBlock
	err := s.marshalizer.Unmarshal(&mb, message.Data())
	if err != nil {
		return err
	}
	s.mutReceivedMetaBlocks.Lock()
	mbHash, err := core.CalculateHash(s.marshalizer, s.hasher, mb)
	if err != nil {
		s.mutReceivedMetaBlocks.Unlock()
		return err
	}
	s.mapReceivedMetaBlocks[string(mbHash)] = &mb
	s.addToPeerList(string(mbHash), message.Peer())
	s.mutReceivedMetaBlocks.Unlock()

	return nil
}

// this func should be called under mutex protection
func (s *simpleMetaBlockInterceptor) addToPeerList(hash string, id p2p.PeerID) {
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

// GetMetaBlock will return the metablock after it is confirmed or an error if the number of tries was exceeded
func (s *simpleMetaBlockInterceptor) GetMetaBlock(target int, epoch uint32) (*block.MetaBlock, error) {
	for count := 0; count < numTriesUntilExit; count++ {
		time.Sleep(timeToWaitBeforeCheckingReceivedMetaBlocks)
		s.mutReceivedMetaBlocks.RLock()
		for hash, peersList := range s.mapMetaBlocksFromPeers {
			isOk := s.isMapEntryOk(peersList, hash, target, epoch)
			if isOk {
				s.mutReceivedMetaBlocks.RUnlock()
				return s.mapReceivedMetaBlocks[hash], nil
			}
		}
		s.mutReceivedMetaBlocks.RUnlock()
	}

	return nil, ErrNumTriesExceeded
}

func (s *simpleMetaBlockInterceptor) isMapEntryOk(
	peersList []p2p.PeerID,
	hash string,
	target int,
	epoch uint32,
) bool {
	mb := s.mapReceivedMetaBlocks[hash]
	epochCheckNotRequired := epoch == math.MaxUint32
	isEpochOk := epochCheckNotRequired || mb.Epoch == epoch
	if len(peersList) >= target && isEpochOk {
		log.Info("got consensus for metablock", "len", len(peersList))
		return true
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *simpleMetaBlockInterceptor) IsInterfaceNil() bool {
	return s == nil
}
