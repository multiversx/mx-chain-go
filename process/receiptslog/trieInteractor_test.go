package receiptslog

import (
	"encoding/hex"
	"testing"

	hasherMock "github.com/multiversx/mx-chain-core-go/data/mock"
	"github.com/multiversx/mx-chain-core-go/data/state"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/require"
)

func createArgsTrieInteractor() ArgsTrieInteractor {
	return ArgsTrieInteractor{
		ReceiptDataStorer:   &mock.StorerMock{},
		Marshaller:          &mock.MarshalizerMock{},
		Hasher:              &hasherMock.HasherMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
}

func TestNewTrieInteractor(t *testing.T) {
	t.Parallel()

	t.Run("nil storer should err", func(t *testing.T) {
		t.Parallel()

		args := createArgsTrieInteractor()
		args.ReceiptDataStorer = nil

		_, err := NewTrieInteractor(args)
		require.Equal(t, dataRetriever.ErrNilReceiptsStorage, err)
	})

	t.Run("nil marshaller should err", func(t *testing.T) {
		t.Parallel()

		args := createArgsTrieInteractor()
		args.Marshaller = nil

		_, err := NewTrieInteractor(args)
		require.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("nil hasher should err", func(t *testing.T) {
		t.Parallel()

		args := createArgsTrieInteractor()
		args.Hasher = nil

		_, err := NewTrieInteractor(args)
		require.Equal(t, process.ErrNilHasher, err)
	})

	t.Run("nil enable epoch handler should err", func(t *testing.T) {
		t.Parallel()

		args := createArgsTrieInteractor()
		args.EnableEpochsHandler = nil

		_, err := NewTrieInteractor(args)
		require.Equal(t, process.ErrNilEnableEpochsHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgsTrieInteractor()

		interactor, err := NewTrieInteractor(args)
		require.Nil(t, err)
		require.NotNil(t, interactor)
	})
}

func TestTrieInteractor_CreateNewTrieAndSave(t *testing.T) {
	t.Parallel()

	args := createArgsTrieInteractor()
	args.ReceiptDataStorer = mock.NewStorerMock()
	interactor, err := NewTrieInteractor(args)
	require.Nil(t, err)

	err = interactor.CreateNewTrie()
	require.Nil(t, err)

	h1 := []byte("hash1")
	rec1 := state.Receipt{
		TxHash: h1,
	}
	err = interactor.AddReceiptData(rec1)
	require.Nil(t, err)

	h2 := []byte("hash2")
	rec2 := state.Receipt{
		TxHash: h2,
	}
	err = interactor.AddReceiptData(rec2)
	require.Nil(t, err)

	receiptTrieRootHash, err := interactor.Save()
	require.Nil(t, err)
	require.Equal(t, "e7a152fdd3be9fdbdd26cc0f915ac2b950fb2d5f91a503baefb6280084b98cee", hex.EncodeToString(receiptTrieRootHash))

	// is a key under receipt trie root hash
	res, err := args.ReceiptDataStorer.Get(receiptTrieRootHash)
	require.Nil(t, err)
	require.NotNil(t, res)

	// is data under tx hash1
	resLeafHash1, err := args.ReceiptDataStorer.Get(h1)
	require.Nil(t, err)
	require.NotNil(t, resLeafHash1)

	// is data under tx hash2
	resLeafHash2, err := args.ReceiptDataStorer.Get(h2)
	require.Nil(t, err)
	require.NotNil(t, resLeafHash2)

	res, err = args.ReceiptDataStorer.Get(resLeafHash1)
	require.Nil(t, err)
	require.NotNil(t, res)

	res, err = args.ReceiptDataStorer.Get(resLeafHash2)
	require.Nil(t, err)
	require.NotNil(t, res)
}
