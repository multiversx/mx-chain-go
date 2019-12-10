package track

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("process/track")

type headerInfo struct {
	hash   []byte
	header data.HeaderHandler
}

type baseBlockTrack struct {
	rounder          consensus.Rounder
	shardCoordinator sharding.Coordinator

	mutHeaders sync.RWMutex
	headers    map[uint32]map[uint64][]*headerInfo

	mutNotarizedHeaders sync.RWMutex
	notarizedHeaders    map[uint32][]data.HeaderHandler
}

// AddHeader adds the given header to the received headers list
func (bbt *baseBlockTrack) AddHeader(header data.HeaderHandler, hash []byte) {
	if check.IfNil(header) {
		return
	}

	shardID := header.GetShardID()
	nonce := header.GetNonce()

	bbt.mutHeaders.Lock()
	defer bbt.mutHeaders.Unlock()

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		headersForShard = make(map[uint64][]*headerInfo)
		bbt.headers[shardID] = headersForShard
	}

	for _, headerInfo := range headersForShard[nonce] {
		if bytes.Equal(headerInfo.hash, hash) {
			return
		}
	}

	headersForShard[nonce] = append(headersForShard[nonce], &headerInfo{hash: hash, header: header})
	//bbt.displayHeadersForShard(shardID)
}

// LastHeaderForShard returns the last header received (highest round) for the given shard
func (bbt *baseBlockTrack) LastHeaderForShard(shardID uint32) data.HeaderHandler {
	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	var lastHeaderForShard data.HeaderHandler

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		return lastHeaderForShard
	}

	maxRound := uint64(0)
	for _, headersInfo := range headersForShard {
		for _, headerInfo := range headersInfo {
			if headerInfo.header.GetRound() > maxRound {
				maxRound = headerInfo.header.GetRound()
				lastHeaderForShard = headerInfo.header
			}
		}
	}

	return lastHeaderForShard
}

// IsShardStuck returns true if the given shard is stuck
func (bbt *baseBlockTrack) IsShardStuck(shardId uint32) bool {
	header := bbt.LastHeaderForShard(shardId)
	if check.IfNil(header) {
		return false
	}

	isShardStuck := bbt.rounder.Index()-int64(header.GetRound()) >= process.MaxRoundsWithoutCommittedBlock
	return isShardStuck
}

// IsInterfaceNil returns true if there is no value under the interface
func (bbt *baseBlockTrack) IsInterfaceNil() bool {
	return bbt == nil
}

func (bbt *baseBlockTrack) displayHeaders() {
	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	for shardID := range bbt.headers {
		bbt.displayHeadersForShard(shardID)
	}
}

func (bbt *baseBlockTrack) displayHeadersForShard(shardID uint32) {
	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		return
	}

	log.Debug("headers tracked", "shard", shardID)

	headers := make([]data.HeaderHandler, 0)
	for _, headersInfo := range headersForShard {
		for _, headerInfo := range headersInfo {
			headers = append(headers, headerInfo.header)
		}
	}

	process.SortHeadersByNonce(headers)

	for _, header := range headers {
		log.Debug("header info",
			"round", header.GetRound(),
			"nonce", header.GetNonce())
	}
}

func (bbt *baseBlockTrack) setNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bbt.mutNotarizedHeaders.Lock()
	defer bbt.mutNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrNotarizedHdrsSliceIsNil
	}

	bbt.notarizedHeaders = make(map[uint32][]data.HeaderHandler, bbt.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < bbt.shardCoordinator.NumberOfShards(); i++ {
		header, ok := startHeaders[i].(*block.Header)
		if !ok {
			return process.ErrWrongTypeAssertion
		}
		bbt.notarizedHeaders[i] = append(bbt.notarizedHeaders[i], header)
	}

	metaBlock, ok := startHeaders[sharding.MetachainShardId].(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}
	bbt.notarizedHeaders[sharding.MetachainShardId] = append(bbt.notarizedHeaders[sharding.MetachainShardId], metaBlock)

	return nil
}
