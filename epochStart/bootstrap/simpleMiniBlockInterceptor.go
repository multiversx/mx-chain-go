package bootstrap

import (
	"bytes"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type simpleMiniBlockInterceptor struct {
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	mutReceivedMiniBlocks  sync.RWMutex
	mapReceivedMiniBlocks  map[string]*block.MiniBlock
	mapMiniBlocksFromPeers map[string][]p2p.PeerID
}

// NewSimpleMiniBlockInterceptor will return a new instance of simpleShardHeaderInterceptor
func NewSimpleMiniBlockInterceptor(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*simpleMiniBlockInterceptor, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	return &simpleMiniBlockInterceptor{
		marshalizer:            marshalizer,
		hasher:                 hasher,
		mutReceivedMiniBlocks:  sync.RWMutex{},
		mapReceivedMiniBlocks:  make(map[string]*block.MiniBlock),
		mapMiniBlocksFromPeers: make(map[string][]p2p.PeerID),
	}, nil
}

// ProcessReceivedMessage will receive the metablocks and will add them to the maps
func (s *simpleMiniBlockInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	var mb block.MiniBlock
	err := s.marshalizer.Unmarshal(&mb, message.Data())
	if err != nil {
		return err
	}

	s.mutReceivedMiniBlocks.Lock()
	mbHash, err := core.CalculateHash(s.marshalizer, s.hasher, &mb)
	if err != nil {
		s.mutReceivedMiniBlocks.Unlock()
		return err
	}

	s.mapReceivedMiniBlocks[string(mbHash)] = &mb
	s.addToPeerList(string(mbHash), message.Peer())
	s.mutReceivedMiniBlocks.Unlock()

	return nil
}

// this func should be called under mutex protection
func (s *simpleMiniBlockInterceptor) addToPeerList(hash string, id p2p.PeerID) {
	peersListForHash, ok := s.mapMiniBlocksFromPeers[hash]
	if !ok {
		s.mapMiniBlocksFromPeers[hash] = append(s.mapMiniBlocksFromPeers[hash], id)
		return
	}

	for _, peer := range peersListForHash {
		if peer == id {
			return
		}
	}

	s.mapMiniBlocksFromPeers[hash] = append(s.mapMiniBlocksFromPeers[hash], id)
}

// GetMiniBlock will return the miniblock with the given hash
func (s *simpleMiniBlockInterceptor) GetMiniBlock(hash []byte, target int) (*block.MiniBlock, error) {
	// TODO : replace this with a channel which will be written in when data is ready
	for count := 0; count < numTriesUntilExit; count++ {
		time.Sleep(timeToWaitBeforeCheckingReceivedHeaders)
		s.mutReceivedMiniBlocks.RLock()
		for hashInMap, peersList := range s.mapMiniBlocksFromPeers {
			isOk := s.isMapEntryOk(hash, peersList, hashInMap, target)
			if isOk {
				s.mutReceivedMiniBlocks.RUnlock()
				return s.mapReceivedMiniBlocks[hashInMap], nil
			}
		}
		s.mutReceivedMiniBlocks.RUnlock()
	}

	return nil, ErrNumTriesExceeded
}

func (s *simpleMiniBlockInterceptor) isMapEntryOk(
	expectedHash []byte,
	peersList []p2p.PeerID,
	hash string,
	target int,
) bool {
	mb, ok := s.mapReceivedMiniBlocks[string(expectedHash)]
	if !ok {
		return false
	}

	mbHash, err := core.CalculateHash(s.marshalizer, s.hasher, mb)
	if err != nil {
		return false
	}
	if bytes.Equal(expectedHash, mbHash) && len(peersList) >= target {
		log.Info("got consensus for mini block", "len", len(peersList))
		return true
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *simpleMiniBlockInterceptor) IsInterfaceNil() bool {
	return s == nil
}
