package receiptslog

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/state"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestReceiptsManager_SyncReceiptsTrie(t *testing.T) {
	rHash1 := []byte("hash1")
	rHash2 := []byte("hash2")

	rec1 := state.Receipt{
		TxHash: rHash1,
	}
	rec2 := state.Receipt{
		TxHash: rHash2,
	}

	storer := mock.NewStorerMock()
	args := createArgsTrieInteractor()
	args.Marshaller = &marshal.GogoProtoMarshalizer{}
	args.ReceiptDataStorer = storer

	interactor, err := NewTrieInteractor(args)
	require.Nil(t, err)

	err = interactor.CreateNewTrie()
	require.Nil(t, err)
	err = interactor.AddReceiptData(rec1)
	require.Nil(t, err)
	err = interactor.AddReceiptData(rec2)
	require.Nil(t, err)

	receiptsTrieHash, err := interactor.Save()
	require.Nil(t, err)

	resultsMap := make(map[string][]byte)

	manager, err := NewReceiptsManager(ArgsReceiptsManager{
		TrieHandler: interactor,
		ReceiptsDataSyncer: &testscommon.ReceiptsDataSyncerStub{
			SyncReceiptsDataForCalled: func(hashes [][]byte, ctx context.Context) error {
				resultsMap = make(map[string][]byte)
				for _, hash := range hashes {
					res, errGet := storer.Get(hash)
					require.Nil(t, errGet)

					resultsMap[string(hash)] = res
				}

				return nil
			},
			GetReceiptsDataCalled: func() (map[string][]byte, error) {
				return resultsMap, nil
			},
		},
		Marshaller: args.Marshaller,
		Hasher:     args.Hasher,
	})
	require.Nil(t, err)

	err = manager.SyncReceiptsTrie(receiptsTrieHash)
	require.Nil(t, err)
}
