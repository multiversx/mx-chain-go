package blockchain

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewApiBlockchain(t *testing.T) {
	t.Parallel()

	t.Run("nil main chain handler should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewApiBlockchain(nil)
		require.Equal(t, ErrNilBlockChain, err)
		require.Nil(t, instance)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		instance, err := NewApiBlockchain(&testscommon.ChainHandlerStub{})
		require.NoError(t, err)
		require.NotNil(t, instance)
	})
}

func TestApiBlockchain_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *apiBlockchain
	require.True(t, instance.IsInterfaceNil())

	instance, _ = NewApiBlockchain(&testscommon.ChainHandlerStub{})
	require.False(t, instance.IsInterfaceNil())
}

func TestApiBlockchain_GetGenesisHeader(t *testing.T) {
	t.Parallel()

	providedHeader := &testscommon.HeaderHandlerStub{}
	instance, err := NewApiBlockchain(&testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return providedHeader
		},
	})
	require.NoError(t, err)

	header := instance.GetGenesisHeader()
	require.Equal(t, providedHeader, header)
}

func TestApiBlockchain_SetGenesisHeader(t *testing.T) {
	t.Parallel()

	instance, err := NewApiBlockchain(&testscommon.ChainHandlerStub{})
	require.NoError(t, err)

	err = instance.SetGenesisHeader(&testscommon.HeaderHandlerStub{})
	require.Equal(t, ErrOperationNotPermitted, err)
}

func TestApiBlockchain_GetGenesisHeaderHash(t *testing.T) {
	t.Parallel()

	providedHash := []byte("provided hash")
	instance, err := NewApiBlockchain(&testscommon.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return providedHash
		},
	})
	require.NoError(t, err)

	hash := instance.GetGenesisHeaderHash()
	require.Equal(t, providedHash, hash)
}

func TestApiBlockchain_GetCurrentBlockHeader(t *testing.T) {
	t.Parallel()

	providedHeader := &testscommon.HeaderHandlerStub{}
	instance, err := NewApiBlockchain(&testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return providedHeader
		},
	})
	require.NoError(t, err)

	header := instance.GetCurrentBlockHeader()
	require.Equal(t, providedHeader, header)
}

func TestApiBlockchain_SetCurrentBlockHeaderAndRootHash(t *testing.T) {
	t.Parallel()

	instance, err := NewApiBlockchain(&testscommon.ChainHandlerStub{})
	require.NoError(t, err)

	providedRootHash := []byte("provided root hash")
	err = instance.SetCurrentBlockHeaderAndRootHash(&testscommon.HeaderHandlerStub{}, providedRootHash)
	require.NoError(t, err)

	instance.mut.RLock()
	rootHash := instance.currentRootHash
	instance.mut.RUnlock()
	require.Equal(t, providedRootHash, rootHash)
}

func TestApiBlockchain_GetCurrentBlockHeaderHash(t *testing.T) {
	t.Parallel()

	providedHash := []byte("provided hash")
	instance, err := NewApiBlockchain(&testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderHashCalled: func() []byte {
			return providedHash
		},
	})
	require.NoError(t, err)

	hash := instance.GetCurrentBlockHeaderHash()
	require.Equal(t, providedHash, hash)
}

func TestApiBlockchain_GetCurrentBlockRootHash(t *testing.T) {
	t.Parallel()

	providedRootHashFromMain := []byte("provided root hash from main")
	instance, err := NewApiBlockchain(&testscommon.ChainHandlerStub{
		GetCurrentBlockRootHashCalled: func() []byte {
			return providedRootHashFromMain
		},
	})
	require.NoError(t, err)

	rootHash := instance.GetCurrentBlockRootHash()
	require.Equal(t, providedRootHashFromMain, rootHash)

	// set a local root hash which should be returned instead
	providedRootHash := []byte("provided root hash")
	_ = instance.SetCurrentBlockHeaderAndRootHash(&testscommon.HeaderHandlerStub{}, providedRootHash)
	rootHash = instance.GetCurrentBlockRootHash()
	require.Equal(t, providedRootHash, rootHash)

	// resetting the local root hash should return the one from main block chain
	_ = instance.SetCurrentBlockHeaderAndRootHash(&testscommon.HeaderHandlerStub{}, make([]byte, 0))
	rootHash = instance.GetCurrentBlockRootHash()
	require.Equal(t, providedRootHashFromMain, rootHash)
}

func TestApiBlockchain_GetFinalBlockInfo(t *testing.T) {
	t.Parallel()

	providedNonce := uint64(123)
	providedHash := []byte("hash")
	providedRootHash := []byte("root hash")
	instance, err := NewApiBlockchain(&testscommon.ChainHandlerStub{
		GetFinalBlockInfoCalled: func() (nonce uint64, blockHash []byte, rootHash []byte) {
			return providedNonce, providedHash, providedRootHash
		},
	})
	require.NoError(t, err)

	nonce, hash, rootHash := instance.GetFinalBlockInfo()
	require.Equal(t, providedNonce, nonce)
	require.Equal(t, providedHash, hash)
	require.Equal(t, providedRootHash, rootHash)
}

func TestApiBlockchain_EmptyMethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	instance, err := NewApiBlockchain(&testscommon.ChainHandlerStub{})
	require.NoError(t, err)

	instance.SetGenesisHeaderHash(nil)
	instance.SetCurrentBlockHeaderHash(nil)
	instance.SetFinalBlockInfo(0, nil, nil)
}
