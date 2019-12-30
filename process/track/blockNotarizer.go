package track

import (
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

type blockNotarizer struct {
	hasher              hashing.Hasher
	marshalizer         marshal.Marshalizer
	startHeaders        map[uint32]data.HeaderHandler
	mutNotarizedHeaders sync.RWMutex
	notarizedHeaders    map[uint32][]*headerInfo
}

// NewBlockNotarizer creates a block notarizer object which implements blockNotarizerHandler interface
func NewBlockNotarizer(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) (*blockNotarizer, error) {

	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	bn := blockNotarizer{
		hasher:      hasher,
		marshalizer: marshalizer,
	}

	bn.notarizedHeaders = make(map[uint32][]*headerInfo)

	return &bn, nil
}

func (bn *blockNotarizer) addNotarizedHeader(
	shardID uint32,
	notarizedHeader data.HeaderHandler,
	notarizedHeaderHash []byte,
) {
	if check.IfNil(notarizedHeader) {
		return
	}

	bn.mutNotarizedHeaders.Lock()
	bn.notarizedHeaders[shardID] = append(bn.notarizedHeaders[shardID], &headerInfo{header: notarizedHeader, hash: notarizedHeaderHash})
	if len(bn.notarizedHeaders[shardID]) > 1 {
		sort.Slice(bn.notarizedHeaders[shardID], func(i, j int) bool {
			return bn.notarizedHeaders[shardID][i].header.GetNonce() < bn.notarizedHeaders[shardID][j].header.GetNonce()
		})
	}
	bn.mutNotarizedHeaders.Unlock()
}

func (bn *blockNotarizer) cleanupNotarizedHeadersBehindNonce(shardID uint32, nonce uint64) {
	bn.mutNotarizedHeaders.Lock()
	defer bn.mutNotarizedHeaders.Unlock()

	if nonce == 0 {
		return
	}

	notarizedHeaders, ok := bn.notarizedHeaders[shardID]
	if !ok {
		return
	}

	headersInfo := make([]*headerInfo, 0)
	for _, headerInfo := range notarizedHeaders {
		if headerInfo.header.GetNonce() < nonce {
			continue
		}

		headersInfo = append(headersInfo, headerInfo)
	}

	bn.notarizedHeaders[shardID] = headersInfo
}

func (bn *blockNotarizer) displayNotarizedHeaders(shardID uint32) {
	bn.mutNotarizedHeaders.RLock()
	defer bn.mutNotarizedHeaders.RUnlock()

	notarizedHeaders, ok := bn.notarizedHeaders[shardID]
	if ok {
		if len(notarizedHeaders) > 1 {
			sort.Slice(notarizedHeaders, func(i, j int) bool {
				return notarizedHeaders[i].header.GetNonce() < notarizedHeaders[j].header.GetNonce()
			})
		}

		shouldNotDisplay := len(notarizedHeaders) == 0 ||
			len(notarizedHeaders) == 1 && notarizedHeaders[0].header.GetNonce() == 0
		if shouldNotDisplay {
			return
		}

		for _, headerInfo := range notarizedHeaders {
			log.Trace("notarized header info",
				"round", headerInfo.header.GetRound(),
				"nonce", headerInfo.header.GetNonce(),
				"hash", headerInfo.hash)
		}
	}
}

func (bn *blockNotarizer) getLastNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	bn.mutNotarizedHeaders.RLock()
	defer bn.mutNotarizedHeaders.RUnlock()

	if bn.notarizedHeaders == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceIsNil
	}

	headerInfo := bn.lastNotarizedHeaderInfo(shardID)
	if headerInfo == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceForShardIsNil
	}

	return headerInfo.header, headerInfo.hash, nil
}

func (bn *blockNotarizer) getLastNotarizedHeaderNonce(shardID uint32) uint64 {
	bn.mutNotarizedHeaders.RLock()
	defer bn.mutNotarizedHeaders.RUnlock()

	if bn.notarizedHeaders == nil {
		return 0
	}

	lastNotarizedHeaderInfo := bn.lastNotarizedHeaderInfo(shardID)
	if lastNotarizedHeaderInfo == nil {
		return 0
	}

	return lastNotarizedHeaderInfo.header.GetNonce()
}

func (bn *blockNotarizer) getNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	bn.mutNotarizedHeaders.RLock()
	defer bn.mutNotarizedHeaders.RUnlock()

	if bn.notarizedHeaders == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceIsNil
	}
	if bn.notarizedHeaders[shardID] == nil {
		return nil, nil, process.ErrNotarizedHeadersSliceForShardIsNil
	}

	headerInfo := bn.notarizedHeaders[shardID][0]

	notarizedHeadersCount := uint64(len(bn.notarizedHeaders[shardID]))
	if notarizedHeadersCount > offset {
		headerInfo = bn.notarizedHeaders[shardID][notarizedHeadersCount-offset-1]
	}

	return headerInfo.header, headerInfo.hash, nil
}

func (bn *blockNotarizer) initNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bn.mutNotarizedHeaders.Lock()
	defer bn.mutNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrNotarizedHeadersSliceIsNil
	}

	bn.notarizedHeaders = make(map[uint32][]*headerInfo)

	for _, startHeader := range startHeaders {
		shardID := startHeader.GetShardID()
		startHeaderHash, err := core.CalculateHash(bn.marshalizer, bn.hasher, startHeader)
		if err != nil {
			return err
		}

		bn.notarizedHeaders[shardID] = append(bn.notarizedHeaders[shardID], &headerInfo{header: startHeader, hash: startHeaderHash})
	}

	return nil
}

func (bn *blockNotarizer) lastNotarizedHeaderInfo(shardID uint32) *headerInfo {
	notarizedHeadersCount := len(bn.notarizedHeaders[shardID])
	if notarizedHeadersCount > 0 {
		return bn.notarizedHeaders[shardID][notarizedHeadersCount-1]
	}

	return nil
}

func (bn *blockNotarizer) removeLastNotarizedHeader() {
	bn.mutNotarizedHeaders.Lock()
	for shardID := range bn.notarizedHeaders {
		notarizedHeadersCount := len(bn.notarizedHeaders[shardID])
		if notarizedHeadersCount > 1 {
			bn.notarizedHeaders[shardID] = bn.notarizedHeaders[shardID][:notarizedHeadersCount-1]
		}
	}
	bn.mutNotarizedHeaders.Unlock()
}

func (bn *blockNotarizer) restoreNotarizedHeadersToGenesis() {
	bn.mutNotarizedHeaders.Lock()
	for shardID := range bn.notarizedHeaders {
		notarizedHeadersCount := len(bn.notarizedHeaders[shardID])
		if notarizedHeadersCount > 1 {
			bn.notarizedHeaders[shardID] = bn.notarizedHeaders[shardID][:1]
		}
	}
	bn.mutNotarizedHeaders.Unlock()
}
