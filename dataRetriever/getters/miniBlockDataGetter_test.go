package getters_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/getters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockMiniblockGetterArgs(
	dataPoolExistingHashes [][]byte,
	storerExistingHashes [][]byte,
) getters.ArgMiniBlockDataGetter {
	marshalizer := &mock.MarshalizerMock{}

	return getters.ArgMiniBlockDataGetter{
		Marshalizer: marshalizer,
		MiniBlockStorage: &mock.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				if isByteSliceInSlice(key, storerExistingHashes) {
					buff, _ := marshalizer.Marshal(&dataBlock.MiniBlock{})

					return buff, nil
				}

				return nil, fmt.Errorf("not found")
			},
		},
		MiniBlockPool: &mock.CacherStub{
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

func TestNewMiniBlockDataGetter_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockGetterArgs(nil, nil)
	arg.Marshalizer = nil

	mbdg, err := getters.NewMiniBlockDataGetter(arg)

	assert.True(t, check.IfNil(mbdg))
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewMiniBlockDataGetter_NilPoolShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockGetterArgs(nil, nil)
	arg.MiniBlockPool = nil

	mbdg, err := getters.NewMiniBlockDataGetter(arg)

	assert.True(t, check.IfNil(mbdg))
	assert.Equal(t, dataRetriever.ErrNilMiniblocksPool, err)
}

func TestNewMiniBlockDataGetter_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockGetterArgs(nil, nil)
	arg.MiniBlockStorage = nil

	mbdg, err := getters.NewMiniBlockDataGetter(arg)

	assert.True(t, check.IfNil(mbdg))
	assert.Equal(t, dataRetriever.ErrNilMiniblocksStorage, err)
}

func TestNewMiniBlockDataGetter_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockGetterArgs(nil, nil)
	mbdg, err := getters.NewMiniBlockDataGetter(arg)

	assert.False(t, check.IfNil(mbdg))
	assert.Nil(t, err)
}

//------- GetMiniBlocksFromPool

func TestMiniBlockDataGetter_GetMiniBlocksFromPoolFoundInPoolShouldReturn(t *testing.T) {
	t.Parallel()

	existingHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	arg := createMockMiniblockGetterArgs(existingHashes, nil)
	mbdg, _ := getters.NewMiniBlockDataGetter(arg)

	miniblocks, missingHashes := mbdg.GetMiniBlocksFromPool(existingHashes)

	assert.Equal(t, 2, len(miniblocks))
	assert.Equal(t, 0, len(missingHashes))
}

func TestMiniBlockDataGetter_GetMiniBlocksFromPoolTwoFoundInPoolShouldReturn(t *testing.T) {
	t.Parallel()

	existingHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	requestedHashes := append(existingHashes, []byte("hash3"))
	arg := createMockMiniblockGetterArgs(existingHashes, nil)
	mbdg, _ := getters.NewMiniBlockDataGetter(arg)

	miniblocks, missingHashes := mbdg.GetMiniBlocksFromPool(requestedHashes)

	assert.Equal(t, 2, len(miniblocks))
	require.Equal(t, 1, len(missingHashes))
	assert.Equal(t, requestedHashes[len(requestedHashes)-1], missingHashes[0])
}

func TestMiniBlockDataGetter_GetMiniBlocksFromPoolWrongTypeInPoolShouldNotReturn(t *testing.T) {
	t.Parallel()

	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	arg := createMockMiniblockGetterArgs(hashes, nil)
	arg.MiniBlockPool = &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return "not a miniblock", true
		},
	}
	mbdg, _ := getters.NewMiniBlockDataGetter(arg)

	miniblocks, missingHashes := mbdg.GetMiniBlocksFromPool(hashes)

	assert.Equal(t, 0, len(miniblocks))
	assert.Equal(t, hashes, missingHashes)
}

//------- GetMiniBlocks

func TestMiniBlockDataGetter_GetMiniBlocksFoundInPoolShouldReturn(t *testing.T) {
	t.Parallel()

	existingHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	arg := createMockMiniblockGetterArgs(existingHashes, nil)
	mbdg, _ := getters.NewMiniBlockDataGetter(arg)

	miniblocks, missingHashes := mbdg.GetMiniBlocks(existingHashes)

	assert.Equal(t, 2, len(miniblocks))
	assert.Equal(t, 0, len(missingHashes))
}

func TestMiniBlockDataGetter_GetMiniBlocksNotFoundInPoolButFoundInStorageShouldReturn(t *testing.T) {
	t.Parallel()

	existingHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	arg := createMockMiniblockGetterArgs(nil, existingHashes)
	mbdg, _ := getters.NewMiniBlockDataGetter(arg)

	miniblocks, missingHashes := mbdg.GetMiniBlocks(existingHashes)

	assert.Equal(t, 2, len(miniblocks))
	assert.Equal(t, 0, len(missingHashes))
}

func TestMiniBlockDataGetter_GetMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	existingInPool := [][]byte{[]byte("hash1")}
	existingInStorage := [][]byte{[]byte("hash2")}
	missingHash := []byte("hash3")
	requestedHashes := append(existingInPool, existingInStorage...)
	requestedHashes = append(requestedHashes, missingHash)
	arg := createMockMiniblockGetterArgs(existingInPool, existingInStorage)
	mbdg, _ := getters.NewMiniBlockDataGetter(arg)

	miniblocks, missingHashes := mbdg.GetMiniBlocks(requestedHashes)

	assert.Equal(t, 2, len(miniblocks))
	require.Equal(t, 1, len(missingHashes))
	assert.Equal(t, missingHash, missingHashes[0])
}
