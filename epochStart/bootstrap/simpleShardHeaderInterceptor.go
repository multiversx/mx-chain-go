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

type simpleShardHeaderInterceptor struct {
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	mutReceivedShardHeaders  sync.RWMutex
	mapReceivedShardHeaders  map[string]*block.Header
	mapShardHeadersFromPeers map[string][]p2p.PeerID
}

// NewSimpleShardHeaderInterceptor will return a new instance of simpleShardHeaderInterceptor
func NewSimpleShardHeaderInterceptor(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*simpleShardHeaderInterceptor, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	return &simpleShardHeaderInterceptor{
		marshalizer:              marshalizer,
		hasher:                   hasher,
		mutReceivedShardHeaders:  sync.RWMutex{},
		mapReceivedShardHeaders:  make(map[string]*block.Header),
		mapShardHeadersFromPeers: make(map[string][]p2p.PeerID),
	}, nil
}

// ProcessReceivedMessage will receive the metablocks and will add them to the maps
func (s *simpleShardHeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	var mb block.Header
	err := s.marshalizer.Unmarshal(&mb, message.Data())
	if err != nil {
		return err
	}

	s.mutReceivedShardHeaders.Lock()
	mbHash, err := core.CalculateHash(s.marshalizer, s.hasher, &mb)
	if err != nil {
		s.mutReceivedShardHeaders.Unlock()
		return err
	}

	s.mapReceivedShardHeaders[string(mbHash)] = &mb
	s.addToPeerList(string(mbHash), message.Peer())
	s.mutReceivedShardHeaders.Unlock()

	return nil
}

// this func should be called under mutex protection
func (s *simpleShardHeaderInterceptor) addToPeerList(hash string, id p2p.PeerID) {
	peersListForHash, ok := s.mapShardHeadersFromPeers[hash]
	if !ok {
		s.mapShardHeadersFromPeers[hash] = append(s.mapShardHeadersFromPeers[hash], id)
		return
	}

	for _, peer := range peersListForHash {
		if peer == id {
			return
		}
	}

	s.mapShardHeadersFromPeers[hash] = append(s.mapShardHeadersFromPeers[hash], id)
}

// GetShardHeader will return the shard header
func (s *simpleShardHeaderInterceptor) GetShardHeader(hash []byte, target int) (*block.Header, error) {
	// TODO : replace this with a channel which will be written in when data is ready
	for count := 0; count < numTriesUntilExit; count++ {
		time.Sleep(timeToWaitBeforeCheckingReceivedHeaders)
		s.mutReceivedShardHeaders.RLock()
		for hashInMap, peersList := range s.mapShardHeadersFromPeers {
			isOk := s.isMapEntryOk(hash, peersList, hashInMap, target)
			if isOk {
				s.mutReceivedShardHeaders.RUnlock()
				return s.mapReceivedShardHeaders[hashInMap], nil
			}
		}
		s.mutReceivedShardHeaders.RUnlock()
	}

	return nil, ErrNumTriesExceeded
}

func (s *simpleShardHeaderInterceptor) isMapEntryOk(
	expectedHash []byte,
	peersList []p2p.PeerID,
	hashInMap string,
	target int,
) bool {
	mb, ok := s.mapReceivedShardHeaders[string(expectedHash)]
	if !ok {
		return false
	}

	hdrHash, err := core.CalculateHash(s.marshalizer, s.hasher, mb)
	if err != nil {
		return false
	}
	if bytes.Equal(expectedHash, hdrHash) && len(peersList) >= target {
		log.Info("got consensus for shard block", "len", len(peersList))
		return true
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *simpleShardHeaderInterceptor) IsInterfaceNil() bool {
	return s == nil
}
