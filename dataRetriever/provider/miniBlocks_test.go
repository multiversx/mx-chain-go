package provider_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/provider"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockMiniblockProviderArgs(
	dataPoolExistingHashes [][]byte,
	storerExistingHashes [][]byte,
) provider.ArgMiniBlockProvider {
	marshalizer := &mock.MarshalizerMock{}

	return provider.ArgMiniBlockProvider{
		Marshalizer: marshalizer,
		MiniBlockStorage: &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				if isByteSliceInSlice(key, storerExistingHashes) {
					buff, _ := marshalizer.Marshal(&dataBlock.MiniBlock{})

					return buff, nil
				}

				return nil, fmt.Errorf("not found")
			},
		},
		MiniBlockPool: &testscommon.CacherStub{
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				if isByteSliceInSlice(key, dataPoolExistingHashes) {
					return &dataBlock.MiniBlock{}, true
				}

				return
			},
		},
	}
}

func isByteSliceInSlice(key []byte, keys [][]byte) bool {
	for _, k := range keys {
		if bytes.Equal(k, key) {
			return true
		}
	}

	return false
}

func TestNewMiniBlockProvider_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockProviderArgs(nil, nil)
	arg.Marshalizer = nil

	mbp, err := provider.NewMiniBlockProvider(arg)

	assert.True(t, check.IfNil(mbp))
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewMiniBlockProvider_NilPoolShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockProviderArgs(nil, nil)
	arg.MiniBlockPool = nil

	mbp, err := provider.NewMiniBlockProvider(arg)

	assert.True(t, check.IfNil(mbp))
	assert.Equal(t, dataRetriever.ErrNilMiniblocksPool, err)
}

func TestNewMiniBlockProvider_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockProviderArgs(nil, nil)
	arg.MiniBlockStorage = nil

	mbp, err := provider.NewMiniBlockProvider(arg)

	assert.True(t, check.IfNil(mbp))
	assert.Equal(t, dataRetriever.ErrNilMiniblocksStorage, err)
}

func TestNewMiniBlockProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockProviderArgs(nil, nil)
	mbp, err := provider.NewMiniBlockProvider(arg)

	assert.False(t, check.IfNil(mbp))
	assert.Nil(t, err)
}

//------- GetMiniBlocksFromPool

func TestMiniBlockProvider_GetMiniBlocksFromPoolFoundInPoolShouldReturn(t *testing.T) {
	t.Parallel()

	existingHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	arg := createMockMiniblockProviderArgs(existingHashes, nil)
	mbp, _ := provider.NewMiniBlockProvider(arg)

	miniBlocksAndHashes, missingHashes := mbp.GetMiniBlocksFromPool(existingHashes)

	assert.Equal(t, 2, len(miniBlocksAndHashes))
	assert.Equal(t, 0, len(missingHashes))
}

func TestMiniBlockProvider_GetMiniBlocksFromPoolTwoFoundInPoolShouldReturn(t *testing.T) {
	t.Parallel()

	existingHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	requestedHashes := append(existingHashes, []byte("hash3"))
	arg := createMockMiniblockProviderArgs(existingHashes, nil)
	mbp, _ := provider.NewMiniBlockProvider(arg)

	miniBlocksAndHashes, missingHashes := mbp.GetMiniBlocksFromPool(requestedHashes)

	assert.Equal(t, 2, len(miniBlocksAndHashes))
	require.Equal(t, 1, len(missingHashes))
	assert.Equal(t, requestedHashes[len(requestedHashes)-1], missingHashes[0])
}

func TestMiniBlockProvider_GetMiniBlocksFromPoolWrongTypeInPoolShouldNotReturn(t *testing.T) {
	t.Parallel()

	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	arg := createMockMiniblockProviderArgs(hashes, nil)
	arg.MiniBlockPool = &testscommon.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return "not a miniblock", true
		},
	}
	mbp, _ := provider.NewMiniBlockProvider(arg)

	miniBlocksAndHashes, missingHashes := mbp.GetMiniBlocksFromPool(hashes)

	assert.Equal(t, 0, len(miniBlocksAndHashes))
	assert.Equal(t, hashes, missingHashes)
}

//------- GetMiniBlocks

func TestMiniBlockProvider_GetMiniBlocksFoundInPoolShouldReturn(t *testing.T) {
	t.Parallel()

	existingHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	arg := createMockMiniblockProviderArgs(existingHashes, nil)
	mbp, _ := provider.NewMiniBlockProvider(arg)

	miniBlocksAndHashes, missingHashes := mbp.GetMiniBlocks(existingHashes)

	assert.Equal(t, 2, len(miniBlocksAndHashes))
	assert.Equal(t, 0, len(missingHashes))
}

func TestMiniBlockProvider_GetMiniBlocksNotFoundInPoolButFoundInStorageShouldReturn(t *testing.T) {
	t.Parallel()

	existingHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	arg := createMockMiniblockProviderArgs(nil, existingHashes)
	mbp, _ := provider.NewMiniBlockProvider(arg)

	miniBlocksAndHashes, missingHashes := mbp.GetMiniBlocks(existingHashes)

	assert.Equal(t, 2, len(miniBlocksAndHashes))
	assert.Equal(t, 0, len(missingHashes))
}

func TestMiniBlockProvider_GetMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	existingInPool := [][]byte{[]byte("hash1")}
	existingInStorage := [][]byte{[]byte("hash2")}
	missingHash := []byte("hash3")
	requestedHashes := append(existingInPool, existingInStorage...)
	requestedHashes = append(requestedHashes, missingHash)
	arg := createMockMiniblockProviderArgs(existingInPool, existingInStorage)
	mbp, _ := provider.NewMiniBlockProvider(arg)

	miniBlocksAndHashes, missingHashes := mbp.GetMiniBlocks(requestedHashes)

	assert.Equal(t, 2, len(miniBlocksAndHashes))
	require.Equal(t, 1, len(missingHashes))
	assert.Equal(t, missingHash, missingHashes[0])
}
