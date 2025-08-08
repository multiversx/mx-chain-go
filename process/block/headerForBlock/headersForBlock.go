package headerForBlock

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type headersForBlock struct {
	missingHdrs                  uint32
	missingFinalityAttestingHdrs uint32
	missingProofs                uint32
	highestHdrNonce              map[uint32]uint64
	mutHdrsForBlock              sync.RWMutex
	hdrHashAndInfo               map[string]HeaderInfo
	lastNotarizedShardHeaders    map[uint32]LastNotarizedHeaderInfoHandler
}

// NewHeadersForBlock returns a new instance of headersForBlock
func NewHeadersForBlock() *headersForBlock {
	return &headersForBlock{
		hdrHashAndInfo:            make(map[string]HeaderInfo),
		highestHdrNonce:           make(map[uint32]uint64),
		lastNotarizedShardHeaders: make(map[uint32]LastNotarizedHeaderInfoHandler),
	}
}

// Reset resets the internal structures
func (hfb *headersForBlock) Reset() {
	hfb.ResetMissingHeaders()
	hfb.initMaps()
}

func (hfb *headersForBlock) initMaps() {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	hfb.hdrHashAndInfo = make(map[string]HeaderInfo)
	hfb.highestHdrNonce = make(map[uint32]uint64)
	hfb.lastNotarizedShardHeaders = make(map[uint32]LastNotarizedHeaderInfoHandler)
}

// ResetMissingHeaders resets the missing headers
func (hfb *headersForBlock) ResetMissingHeaders() {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	hfb.missingHdrs = 0
	hfb.missingFinalityAttestingHdrs = 0
	hfb.missingProofs = 0
}

// AddHeaderInfo adds the provided header info for the provided hash
func (hfb *headersForBlock) AddHeaderInfo(hash string, headerInfo HeaderInfo) {
	if check.IfNil(headerInfo) {
		return
	}

	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	hfb.hdrHashAndInfo[hash] = headerInfo
}

// GetMisingData returns the missing headers and proofs
func (hfb *headersForBlock) GetMisingData() (uint32, uint32, uint32) {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	return hfb.missingHdrs, hfb.missingProofs, hfb.missingFinalityAttestingHdrs
}

// GetHeadersInfoMap returns the internal header hash map
func (hfb *headersForBlock) GetHeadersInfoMap() map[string]HeaderInfo {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	mapCopy := make(map[string]HeaderInfo, len(hfb.hdrHashAndInfo))
	for hash, hi := range hfb.hdrHashAndInfo {
		mapCopy[hash] = hi
	}

	return mapCopy
}

// GetHeadersMap returns a map of headers with their hashes
func (hfb *headersForBlock) GetHeadersMap() map[string]data.HeaderHandler {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	headersMap := make(map[string]data.HeaderHandler, len(hfb.hdrHashAndInfo))
	for hash, hi := range hfb.hdrHashAndInfo {
		headersMap[hash] = hi.GetHeader()
	}

	return headersMap
}

// GetHeaderInfo returns the header info for the provided hash
func (hfb *headersForBlock) GetHeaderInfo(hash string) (HeaderInfo, bool) {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	hi, found := hfb.hdrHashAndInfo[hash]
	return hi, found
}

// GetHighestHeaderNonceForShard returns the highest header nonce for the provided shard
func (hfb *headersForBlock) GetHighestHeaderNonceForShard(shardID uint32) uint64 {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	return hfb.highestHdrNonce[shardID]
}

// SetHighestHeaderNonceForShard sets the highest header nonce for the provided shard
func (hfb *headersForBlock) SetHighestHeaderNonceForShard(shardID uint32, nonce uint64) {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	hfb.highestHdrNonce[shardID] = nonce
}

// GetLastNotarizedHeaderForShard returns the last notarized header for the provided shard
func (hfb *headersForBlock) GetLastNotarizedHeaderForShard(shardID uint32) (LastNotarizedHeaderInfoHandler, bool) {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	lastNotarizedShardHeader, found := hfb.lastNotarizedShardHeaders[shardID]
	return lastNotarizedShardHeader, found
}

// SetLastNotarizedHeaderForShard sets the last notarized header for the provided shard
func (hfb *headersForBlock) SetLastNotarizedHeaderForShard(shardID uint32, lastNotarizedHeader LastNotarizedHeaderInfoHandler) {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	hfb.lastNotarizedShardHeaders[shardID] = lastNotarizedHeader
}

// SetHeader sets the header for the provided header hash
func (hfb *headersForBlock) SetHeader(hash string, header data.HeaderHandler) {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	_, ok := hfb.hdrHashAndInfo[hash]
	if !ok {
		hfb.hdrHashAndInfo[hash] = NewEmptyHeaderInfo()
	}

	hfb.hdrHashAndInfo[hash].SetHeader(header)
}

// SetHasProof sets hasProof for the provided header hash
func (hfb *headersForBlock) SetHasProof(hash string) {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	_, ok := hfb.hdrHashAndInfo[hash]
	if !ok {
		hfb.hdrHashAndInfo[hash] = NewEmptyHeaderInfo()
	}

	hfb.hdrHashAndInfo[hash].SetHasProof(true)
}

// HasProofRequested returns true if the header has been requested
func (hfb *headersForBlock) HasProofRequested(hash string) bool {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()
	hi, found := hfb.hdrHashAndInfo[hash]
	if !found {
		return false
	}

	return hi.HasProofRequested()
}

// SetHasProofRequested sets hasProofRequested for the provided header hash
func (hfb *headersForBlock) SetHasProofRequested(hash string) {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	_, ok := hfb.hdrHashAndInfo[hash]
	if !ok {
		hfb.hdrHashAndInfo[hash] = NewEmptyHeaderInfo()
	}

	hfb.hdrHashAndInfo[hash].SetHasProofRequested(true)
}

// IncreaseMissingProofs increases the missing proofs counter
func (hfb *headersForBlock) IncreaseMissingProofs() {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()
	hfb.missingProofs++
}

// DecreaseMissingProofs decreases the missing proofs counter
func (hfb *headersForBlock) DecreaseMissingProofs() {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()
	hfb.missingProofs--
}

// IncreaseMissingHeaders increases the missing headers counter
func (hfb *headersForBlock) IncreaseMissingHeaders() {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()
	hfb.missingHdrs++
}

// DecreaseMissingHeaders decreases the missing headers counter
func (hfb *headersForBlock) DecreaseMissingHeaders() {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()
	hfb.missingHdrs--
}

// IncreaseMissingFinalityAttestingHeaders increases the missing finality attesting headers counter
func (hfb *headersForBlock) IncreaseMissingFinalityAttestingHeaders() {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()
	hfb.missingFinalityAttestingHdrs++
}

// DecreaseMissingFinalityAttestingHeaders decreases the missing finality attesting headers counter
func (hfb *headersForBlock) DecreaseMissingFinalityAttestingHeaders() {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()
	hfb.missingFinalityAttestingHdrs--
}

// SetMissingFinalityAttestingHeaders sets the missing finality attesting headers counter to the provided value
func (hfb *headersForBlock) SetMissingFinalityAttestingHeaders(missing uint32) {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()
	hfb.missingFinalityAttestingHdrs = missing
}
