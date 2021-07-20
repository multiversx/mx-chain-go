package track

import (
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type blockNotarizer struct {
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator

	mutNotarizedHeaders sync.RWMutex
	notarizedHeaders    map[uint32][]*HeaderInfo
}

// NewBlockNotarizer creates a block notarizer object which implements blockNotarizerHandler interface
func NewBlockNotarizer(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
) (*blockNotarizer, error) {
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	bn := blockNotarizer{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: shardCoordinator,
	}

	bn.notarizedHeaders = make(map[uint32][]*HeaderInfo)

	return &bn, nil
}

// AddNotarizedHeader adds a notarized header to the given shard
func (bn *blockNotarizer) AddNotarizedHeader(
	shardID uint32,
	notarizedHeader data.HeaderHandler,
	notarizedHeaderHash []byte,
) {
	if check.IfNil(notarizedHeader) {
		return
	}

	bn.mutNotarizedHeaders.Lock()
	bn.notarizedHeaders[shardID] = append(bn.notarizedHeaders[shardID], &HeaderInfo{Header: notarizedHeader, Hash: notarizedHeaderHash})
	sort.Slice(bn.notarizedHeaders[shardID], func(i, j int) bool {
		return bn.notarizedHeaders[shardID][i].Header.GetNonce() < bn.notarizedHeaders[shardID][j].Header.GetNonce()
	})
	bn.mutNotarizedHeaders.Unlock()
}

// CleanupNotarizedHeadersBehindNonce cleanups notarized headers for a given shard behind a given nonce
func (bn *blockNotarizer) CleanupNotarizedHeadersBehindNonce(shardID uint32, nonce uint64) {
	if nonce == 0 {
		return
	}

	bn.mutNotarizedHeaders.Lock()
	defer bn.mutNotarizedHeaders.Unlock()

	notarizedHeaders, ok := bn.notarizedHeaders[shardID]
	if !ok {
		return
	}

	headersInfo := make([]*HeaderInfo, 0)
	for _, hdrInfo := range notarizedHeaders {
		if hdrInfo.Header.GetNonce() < nonce {
			continue
		}

		headersInfo = append(headersInfo, hdrInfo)
	}

	if len(headersInfo) == 0 {
		hdrInfo := bn.lastNotarizedHeaderInfo(shardID)
		if hdrInfo == nil {
			return
		}

		headersInfo = append(headersInfo, hdrInfo)
	}

	bn.notarizedHeaders[shardID] = headersInfo
}

// DisplayNotarizedHeaders displays notarized headers for a given shard
func (bn *blockNotarizer) DisplayNotarizedHeaders(shardID uint32, message string) {
	bn.mutNotarizedHeaders.RLock()
	defer bn.mutNotarizedHeaders.RUnlock()

	notarizedHeaders, ok := bn.notarizedHeaders[shardID]
	if !ok {
		return
	}

	if len(notarizedHeaders) > 1 {
		sort.Slice(notarizedHeaders, func(i, j int) bool {
			return notarizedHeaders[i].Header.GetNonce() < notarizedHeaders[j].Header.GetNonce()
		})
	}

	shouldNotDisplay := len(notarizedHeaders) == 0 ||
		len(notarizedHeaders) == 1 && notarizedHeaders[0].Header.GetNonce() == 0
	if shouldNotDisplay {
		return
	}

	log.Debug(message,
		"shard", shardID,
		"nb", len(notarizedHeaders))

	for _, hdrInfo := range notarizedHeaders {
		log.Trace("notarized header info",
			"round", hdrInfo.Header.GetRound(),
			"nonce", hdrInfo.Header.GetNonce(),
			"hash", hdrInfo.Hash)
	}
}

// GetFirstNotarizedHeader returns the first notarized header for a given shard
func (bn *blockNotarizer) GetFirstNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	bn.mutNotarizedHeaders.RLock()
	defer bn.mutNotarizedHeaders.RUnlock()

	hdrInfo := bn.firstNotarizedHeaderInfo(shardID)
	if hdrInfo == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceForShardIsNil
	}

	return hdrInfo.Header, hdrInfo.Hash, nil
}

// GetLastNotarizedHeader gets the last notarized header for a given shard
func (bn *blockNotarizer) GetLastNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	bn.mutNotarizedHeaders.RLock()
	defer bn.mutNotarizedHeaders.RUnlock()

	hdrInfo := bn.lastNotarizedHeaderInfo(shardID)
	if hdrInfo == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceForShardIsNil
	}

	return hdrInfo.Header, hdrInfo.Hash, nil
}

// GetLastNotarizedHeaderNonce gets the nonce of the last notarized header for a given shard
func (bn *blockNotarizer) GetLastNotarizedHeaderNonce(shardID uint32) uint64 {
	bn.mutNotarizedHeaders.RLock()
	defer bn.mutNotarizedHeaders.RUnlock()

	hdrInfo := bn.lastNotarizedHeaderInfo(shardID)
	if hdrInfo == nil {
		return 0
	}

	return hdrInfo.Header.GetNonce()
}

func (bn *blockNotarizer) firstNotarizedHeaderInfo(shardID uint32) *HeaderInfo {
	notarizedHeadersCount := len(bn.notarizedHeaders[shardID])
	if notarizedHeadersCount > 0 {
		return bn.notarizedHeaders[shardID][0]
	}

	return nil
}

func (bn *blockNotarizer) lastNotarizedHeaderInfo(shardID uint32) *HeaderInfo {
	notarizedHeadersCount := len(bn.notarizedHeaders[shardID])
	if notarizedHeadersCount > 0 {
		return bn.notarizedHeaders[shardID][notarizedHeadersCount-1]
	}

	return nil
}

// GetNotarizedHeader gets notarized header for a given shard with a given offset
func (bn *blockNotarizer) GetNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	bn.mutNotarizedHeaders.RLock()
	defer bn.mutNotarizedHeaders.RUnlock()

	headersInfo := bn.notarizedHeaders[shardID]
	if headersInfo == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceForShardIsNil
	}

	notarizedHeadersCount := uint64(len(headersInfo))
	if notarizedHeadersCount <= offset {
		return nil, nil, ErrNotarizedHeaderOffsetIsOutOfBound
	}

	hdrInfo := headersInfo[notarizedHeadersCount-offset-1]

	return hdrInfo.Header, hdrInfo.Hash, nil
}

// InitNotarizedHeaders initializes all notarized headers for each shard with the genesis value (nonce 0)
func (bn *blockNotarizer) InitNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	if startHeaders == nil {
		return process.ErrNotarizedHeadersSliceIsNil
	}

	bn.mutNotarizedHeaders.Lock()
	defer bn.mutNotarizedHeaders.Unlock()

	bn.notarizedHeaders = make(map[uint32][]*HeaderInfo)

	for shardID, startHeader := range startHeaders {
		startHeaderHash, err := core.CalculateHash(bn.marshalizer, bn.hasher, startHeader)
		if err != nil {
			return err
		}

		bn.notarizedHeaders[shardID] = append(bn.notarizedHeaders[shardID], &HeaderInfo{Header: startHeader, Hash: startHeaderHash})
	}

	return nil
}

// RemoveLastNotarizedHeader removes last notarized header from each shard
func (bn *blockNotarizer) RemoveLastNotarizedHeader() {
	bn.mutNotarizedHeaders.Lock()
	for shardID := range bn.notarizedHeaders {
		notarizedHeadersCount := len(bn.notarizedHeaders[shardID])
		if notarizedHeadersCount > 1 {
			bn.notarizedHeaders[shardID] = bn.notarizedHeaders[shardID][:notarizedHeadersCount-1]
		}
	}
	bn.mutNotarizedHeaders.Unlock()
}

// RestoreNotarizedHeadersToGenesis restores all notarized headers from each shard to the genesis value (nonce 0)
func (bn *blockNotarizer) RestoreNotarizedHeadersToGenesis() {
	bn.mutNotarizedHeaders.Lock()
	for shardID := range bn.notarizedHeaders {
		notarizedHeadersCount := len(bn.notarizedHeaders[shardID])
		if notarizedHeadersCount > 1 {
			bn.notarizedHeaders[shardID] = bn.notarizedHeaders[shardID][:1]
		}
	}
	bn.mutNotarizedHeaders.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (bn *blockNotarizer) IsInterfaceNil() bool {
	return bn == nil
}
