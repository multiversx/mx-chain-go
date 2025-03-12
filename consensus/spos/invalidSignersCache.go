package spos

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/p2p"
)

// ArgInvalidSignersCache defines the DTO used to create a new instance of invalidSignersCache
type ArgInvalidSignersCache struct {
	Hasher         hashing.Hasher
	SigningHandler p2p.P2PSigningHandler
	Marshaller     marshal.Marshalizer
}

type invalidSignersCache struct {
	sync.RWMutex
	invalidSignersHashesMap    map[string]struct{}
	invalidSignersForHeaderMap map[string]map[string]struct{}
	hasher                     hashing.Hasher
	signingHandler             p2p.P2PSigningHandler
	marshaller                 marshal.Marshalizer
}

// NewInvalidSignersCache returns a new instance of invalidSignersCache
func NewInvalidSignersCache(args ArgInvalidSignersCache) (*invalidSignersCache, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &invalidSignersCache{
		invalidSignersHashesMap:    make(map[string]struct{}),
		invalidSignersForHeaderMap: make(map[string]map[string]struct{}),
		hasher:                     args.Hasher,
		signingHandler:             args.SigningHandler,
		marshaller:                 args.Marshaller,
	}, nil
}

func checkArgs(args ArgInvalidSignersCache) error {
	if check.IfNil(args.Hasher) {
		return ErrNilHasher
	}
	if check.IfNil(args.SigningHandler) {
		return ErrNilSigningHandler
	}
	if check.IfNil(args.Marshaller) {
		return ErrNilMarshalizer
	}

	return nil
}

// AddInvalidSigners adds the provided hash into the internal map if it does not exist
func (cache *invalidSignersCache) AddInvalidSigners(headerHash []byte, invalidSigners []byte, invalidPublicKeys []string) {
	if len(invalidPublicKeys) == 0 || len(invalidSigners) == 0 {
		return
	}

	cache.Lock()
	defer cache.Unlock()

	invalidSignersHash := cache.hasher.Compute(string(invalidSigners))
	cache.invalidSignersHashesMap[string(invalidSignersHash)] = struct{}{}

	_, ok := cache.invalidSignersForHeaderMap[string(headerHash)]
	if !ok {
		cache.invalidSignersForHeaderMap[string(headerHash)] = make(map[string]struct{})
	}

	for _, pk := range invalidPublicKeys {
		cache.invalidSignersForHeaderMap[string(headerHash)][pk] = struct{}{}
	}
}

// HasInvalidSigners check whether the provided hash exists in the internal map or not
func (cache *invalidSignersCache) HasInvalidSigners(headerHash []byte, invalidSigners []byte) bool {
	cache.RLock()
	defer cache.RUnlock()

	invalidSignersHash := cache.hasher.Compute(string(invalidSigners))
	_, hasSameInvalidSigners := cache.invalidSignersHashesMap[string(invalidSignersHash)]
	if hasSameInvalidSigners {
		return true
	}

	_, isHeaderKnown := cache.invalidSignersForHeaderMap[string(headerHash)]
	if !isHeaderKnown {
		return false
	}

	messages, err := cache.signingHandler.Deserialize(invalidSigners)
	if err != nil {
		return false
	}

	knownInvalidSigners := 0
	for _, msg := range messages {
		cnsMsg := &consensus.Message{}
		err = cache.marshaller.Unmarshal(cnsMsg, msg.Data())
		if err != nil {
			return false
		}

		_, isKnownInvalidSigner := cache.invalidSignersForHeaderMap[string(headerHash)][string(cnsMsg.PubKey)]
		if isKnownInvalidSigner {
			knownInvalidSigners++
		}
	}

	return knownInvalidSigners == len(messages)
}

// Reset clears the internal map
func (cache *invalidSignersCache) Reset() {
	cache.Lock()
	defer cache.Unlock()

	cache.invalidSignersHashesMap = make(map[string]struct{})
	cache.invalidSignersForHeaderMap = make(map[string]map[string]struct{})
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *invalidSignersCache) IsInterfaceNil() bool {
	return cache == nil
}
