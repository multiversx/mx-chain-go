package esdtSupply

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestGetLogsBasedOnBody(t *testing.T) {
	t.Parallel()

	marshalizer := &marshallerMock.MarshalizerMock{}
	txHash := []byte("txHash")
	scrHash := []byte("scrHash")

	logTx := &transaction.Log{}
	logSCR := &transaction.Log{}

	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if bytes.Equal(key, txHash) {
				return marshalizer.Marshal(logTx)
			}

			if bytes.Equal(key, scrHash) {
				return marshalizer.Marshal(logSCR)
			}

			return nil, errors.New("not found")
		},
	}

	getter := newLogsGetter(marshalizer, storer)

	blockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				Type: block.InvalidBlock,
			},
			{
				Type:     block.TxBlock,
				TxHashes: [][]byte{txHash, []byte("tx")},
			},
			{
				Type:     block.SmartContractResultBlock,
				TxHashes: [][]byte{scrHash},
			},
		},
	}

	res, err := getter.getLogsBasedOnBody(blockBody)
	require.Nil(t, err)
	require.Len(t, res, 2)
}

func TestGetLogsWrongBodyType(t *testing.T) {
	t.Parallel()

	getter := newLogsGetter(&marshallerMock.MarshalizerMock{}, &storageStubs.StorerStub{})

	_, err := getter.getLogsBasedOnBody(nil)
	require.Equal(t, errCannotCastToBlockBody, err)
}
