package headerForBlock_test

//
// import (
// 	"testing"
//
// 	"github.com/multiversx/mx-chain-core-go/data/block"
// 	headerForBlock "github.com/multiversx/mx-chain-go/process/block/headerForBlock"
// 	"github.com/stretchr/testify/require"
// )
//
// TODO[Sorin-nextPR]: move and adapt these tests for headersForBlock
// func TestNewHeadersForBlock(t *testing.T) {
// 	t.Parallel()
//
// 	hfb := headerForBlock.NewHeadersForBlock()
// 	require.NotNil(t, hfb)
// }
//
// func TestHeadersForBlock_IsInterfaceNil(t *testing.T) {
// 	t.Parallel()
//
// 	hfb := headerForBlock.NewHeadersForBlock()
// 	require.False(t, hfb.IsInterfaceNil())
// }
//
// func TestHeadersForBlock_Reset(t *testing.T) {
// 	t.Parallel()
//
// 	hfb := headerForBlock.NewHeadersForBlock()
//
// 	// Set some initial state
// 	hfb.IncreaseMissingHeaders()
// 	hfb.IncreaseMissingFinalityAttestingHeaders()
// 	hfb.IncreaseMissingProofs()
// 	hfb.SetHighestHeaderNonceForShard(0, 1)
//
// 	hfb.Reset()
//
// 	// Check if all counters are reset
// 	missingHdrs, missingFinalityHdrs, missingProofs := hfb.GetMissingData()
// 	require.Equal(t, uint32(0), missingHdrs)
// 	require.Equal(t, uint32(0), missingFinalityHdrs)
// 	require.Equal(t, uint32(0), missingProofs)
//
// 	// Check if maps are empty after reset
// 	highestNonce := hfb.GetHighestHeaderNonceForShard(0)
// 	require.Equal(t, uint64(0), highestNonce)
//
// 	_, exists := hfb.GetHeaderInfo("test")
// 	require.False(t, exists)
// }
//
// func TestHeadersForBlock_HeaderInfoOperations(t *testing.T) {
// 	t.Parallel()
//
// 	hfb := headerForBlock.NewHeadersForBlock()
// 	header := &block.Header{Nonce: 1}
// 	hash := "hash1"
//
// 	// Test AddHeaderInfo and GetHeaderInfo
// 	hfb.AddHeaderInfo(hash, nil) // coverage only
// 	headerInfo := headerForBlock.NewHeaderInfo(header, true, false, false)
// 	hfb.AddHeaderInfo(hash, headerInfo)
//
// 	retrievedHeaderInfo, exists := hfb.GetHeaderInfo(hash)
// 	require.True(t, exists)
// 	require.Equal(t, header, retrievedHeaderInfo.GetHeader())
//
// 	// Test GetHeadersInfoMap
// 	headersMap := hfb.GetHeadersInfoMap()
// 	require.Len(t, headersMap, 1)
// 	require.Equal(t, header, headersMap[hash].GetHeader())
//
// 	// Test GetHeadersMap
// 	headersOnlyMap := hfb.GetHeadersMap()
// 	require.Len(t, headersOnlyMap, 1)
// 	require.Equal(t, header, headersOnlyMap[hash])
// }
//
// func TestHeadersForBlock_HeaderOperations(t *testing.T) {
// 	t.Parallel()
//
// 	hfb := headerForBlock.NewHeadersForBlock()
// 	header := &block.Header{Nonce: 1}
// 	hash := "hash1"
//
// 	// Test SetHeader
// 	hfb.SetHeader(hash, header)
//
// 	retrievedHeaderInfo, exists := hfb.GetHeaderInfo(hash)
// 	require.True(t, exists)
// 	require.Equal(t, header, retrievedHeaderInfo.GetHeader())
// }
//
// func TestHeadersForBlock_ProofOperations(t *testing.T) {
// 	t.Parallel()
//
// 	hfb := headerForBlock.NewHeadersForBlock()
// 	hash := "hash1"
//
// 	// Test SetHasProof and HasProof
// 	hfb.SetHasProof(hash)
// 	headerInfo, exists := hfb.GetHeaderInfo(hash)
// 	require.True(t, exists)
// 	require.True(t, headerInfo.HasProof())
//
// 	// Test HasProofRequested
// 	require.False(t, hfb.HasProofRequested("missing hash")) // coverage only
// 	require.False(t, hfb.HasProofRequested(hash))
//
// 	// Test SetHasProofRequested
// 	hfb.SetHasProofRequested("missing hash") // coverage only
// 	hfb.SetHasProofRequested(hash)
// 	require.True(t, hfb.HasProofRequested(hash))
// }
//
// func TestHeadersForBlock_CounterOperations(t *testing.T) {
// 	t.Parallel()
//
// 	hfb := headerForBlock.NewHeadersForBlock()
//
// 	// Test missing headers counter
// 	hfb.IncreaseMissingHeaders()
// 	hfb.IncreaseMissingHeaders()
// 	hfb.DecreaseMissingHeaders()
//
// 	// Test missing finality attesting headers counter
// 	hfb.IncreaseMissingFinalityAttestingHeaders()
// 	hfb.DecreaseMissingFinalityAttestingHeaders()
// 	hfb.SetMissingFinalityAttestingHeaders(5)
//
// 	// Test missing proofs counter
// 	hfb.IncreaseMissingProofs()
// 	hfb.IncreaseMissingProofs()
// 	hfb.DecreaseMissingProofs()
//
// 	// Verify all counters
// 	missingHdrs, missingProofs, missingFinalityHdrs := hfb.GetMissingData()
// 	require.Equal(t, uint32(1), missingHdrs)
// 	require.Equal(t, uint32(5), missingFinalityHdrs)
// 	require.Equal(t, uint32(1), missingProofs)
// }
//
// func TestHeadersForBlock_HighestHeaderNonce(t *testing.T) {
// 	t.Parallel()
//
// 	hfb := headerForBlock.NewHeadersForBlock()
// 	shardID := uint32(0)
// 	nonce := uint64(42)
//
// 	// Test Set and Get HighestHeaderNonceForShard
// 	hfb.SetHighestHeaderNonceForShard(shardID, nonce)
// 	retrievedNonce := hfb.GetHighestHeaderNonceForShard(shardID)
// 	require.Equal(t, nonce, retrievedNonce)
//
// 	// Test with non-existent shard
// 	nonExistentShard := uint32(1)
// 	retrievedNonce = hfb.GetHighestHeaderNonceForShard(nonExistentShard)
// 	require.Equal(t, uint64(0), retrievedNonce)
// }
//
// func TestHeadersForBlock_LastNotarizedHeader(t *testing.T) {
// 	hfb := headerForBlock.NewHeadersForBlock()
// 	shardID := uint32(0)
// 	header := &block.Header{Nonce: 1}
//
// 	// Test Set and Get LastNotarizedHeaderForShard
// 	lastNotarizedHeader := headerForBlock.NewLastNotarizedHeaderInfo(header, []byte("hash"), false, false)
// 	hfb.SetLastNotarizedHeaderForShard(shardID, lastNotarizedHeader)
// 	retrievedHeader, exists := hfb.GetLastNotarizedHeaderForShard(shardID)
// 	require.True(t, exists)
// 	require.Equal(t, header, retrievedHeader.GetHeader())
//
// 	// Test with non-existent shard
// 	nonExistentShard := uint32(1)
// 	_, exists = hfb.GetLastNotarizedHeaderForShard(nonExistentShard)
// 	require.False(t, exists)
// }
