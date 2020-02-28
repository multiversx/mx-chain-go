package bootstrap

import (
	"errors"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
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
func NewSimpleMetaBlockInterceptor(marshalizer marshal.Marshalizer, hasher hashing.Hasher) *simpleMetaBlockInterceptor {
	return &simpleMetaBlockInterceptor{
		marshalizer:            marshalizer,
		hasher:                 hasher,
		mutReceivedMetaBlocks:  sync.RWMutex{},
		mapReceivedMetaBlocks:  make(map[string]*block.MetaBlock),
		mapMetaBlocksFromPeers: make(map[string][]p2p.PeerID),
	}
}

// ProcessReceivedMessage will receive the metablocks and will add them to the maps
func (s *simpleMetaBlockInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	var mb block.MetaBlock
	err := s.marshalizer.Unmarshal(&mb, message.Data())
	if err == nil {
		s.mutReceivedMetaBlocks.Lock()
		mbHash, err := core.CalculateHash(s.marshalizer, s.hasher, mb)
		if err != nil {
			s.mutReceivedMetaBlocks.Unlock()
			return nil
		}
		s.mapReceivedMetaBlocks[string(mbHash)] = &mb
		s.addToPeerList(string(mbHash), message.Peer())
		s.mutReceivedMetaBlocks.Unlock()
	}

	return nil
}

// this func should be called under mutex protection
func (s *simpleMetaBlockInterceptor) addToPeerList(hash string, id p2p.PeerID) {
	peersListForHash, ok := s.mapMetaBlocksFromPeers[hash]

	// no entry for this hash. add it directly
	if !ok {
		s.mapMetaBlocksFromPeers[hash] = append(s.mapMetaBlocksFromPeers[hash], id)
		return
	}

	// entries exist for this hash. search so we don't have duplicates
	for _, peer := range peersListForHash {
		if peer == id {
			return
		}
	}

	// entry not found so add it
	s.mapMetaBlocksFromPeers[hash] = append(s.mapMetaBlocksFromPeers[hash], id)
}

// GetMetaBlock will return the metablock after it is confirmed or an error if the number of tries was exceeded
func (s *simpleMetaBlockInterceptor) GetMetaBlock(target int) (*block.MetaBlock, error) {
	for count := 0; count < numTriesUntilExit; count++ {
		time.Sleep(timeToWaitBeforeCheckingReceivedMetaBlocks)
		s.mutReceivedMetaBlocks.RLock()
		for hash, peersList := range s.mapMetaBlocksFromPeers {
			if len(peersList) >= target {
				s.mutReceivedMetaBlocks.RUnlock()
				log.Info("got consensus for metablock", "len", len(peersList))
				return s.mapReceivedMetaBlocks[hash], nil
			}
		}
		s.mutReceivedMetaBlocks.RUnlock()
	}

	return nil, errors.New("num of tries exceeded. try re-request")
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *simpleMetaBlockInterceptor) IsInterfaceNil() bool {
	return s == nil
}
