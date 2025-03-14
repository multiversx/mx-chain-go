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

// CheckKnownInvalidSigners checks whether all the provided invalid signers are known for the header hash
func (cache *invalidSignersCache) CheckKnownInvalidSigners(headerHash []byte, serializedInvalidSigners []byte) bool {
	cache.RLock()
	defer cache.RUnlock()

	invalidSignersHash := cache.hasher.Compute(string(serializedInvalidSigners))
	_, hasSameInvalidSigners := cache.invalidSignersHashesMap[string(invalidSignersHash)]
	if hasSameInvalidSigners {
		return true
	}

	_, isHeaderKnown := cache.invalidSignersForHeaderMap[string(headerHash)]
	if !isHeaderKnown {
		return false
	}

	invalidSignersP2PMessages, err := cache.signingHandler.Deserialize(serializedInvalidSigners)
	if err != nil {
		return false
	}

	for _, msg := range invalidSignersP2PMessages {
		cnsMsg := &consensus.Message{}
		err = cache.marshaller.Unmarshal(cnsMsg, msg.Data())
		if err != nil {
			return false
		}

		_, isKnownInvalidSigner := cache.invalidSignersForHeaderMap[string(headerHash)][string(cnsMsg.PubKey)]
		if !isKnownInvalidSigner {
			return false
		}
	}

	return true
}

// Reset clears the internal maps
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
