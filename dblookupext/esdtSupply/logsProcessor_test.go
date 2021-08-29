package esdtSupply

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestProcessLogsSaveSupplyNothingInStorage(t *testing.T) {
	t.Parallel()

	token := []byte("nft-0001")
	logs := map[string]data.LogHandler{
		"txLog": &transaction.Log{
			Events: []*transaction.Event{
				{
					Identifier: []byte("something"),
				},
				{
					Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
					Topics: [][]byte{
						token, big.NewInt(2).Bytes(), big.NewInt(10).Bytes(),
					},
				},
				{
					Identifier: []byte(core.BuiltInFunctionESDTNFTAddQuantity),
					Topics: [][]byte{
						token, big.NewInt(2).Bytes(), big.NewInt(50).Bytes(),
					},
				},
				{
					Identifier: []byte(core.BuiltInFunctionESDTNFTBurn),
					Topics: [][]byte{
						token, big.NewInt(2).Bytes(), big.NewInt(30).Bytes(),
					},
				},
			},
		},
		"log": nil,
	}

	marshalizer := testscommon.MarshalizerMock{}
	storer := &testscommon.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return nil, storage.ErrKeyNotFound
		},
		PutCalled: func(key, data []byte) error {
			supplyKey := string(token) + "-" + string(big.NewInt(2).Bytes())
			require.Equal(t, supplyKey, string(key))

			var supplyESDT SupplyESDT
			_ = marshalizer.Unmarshal(&supplyESDT, data)
			require.Equal(t, big.NewInt(30), supplyESDT.Supply)

			return nil
		},
	}

	logsProc := newLogsProcessor(marshalizer, storer)

	err := logsProc.processLogs(0, logs, false)
	require.Nil(t, err)
}

func TestTestProcessLogsSaveSupplyExistsInStorage(t *testing.T) {
	t.Parallel()

	token := []byte("esdt-miiu")

	logs := map[string]data.LogHandler{
		"txLog": &transaction.Log{
			Events: []*transaction.Event{
				{
					Identifier: []byte(core.BuiltInFunctionESDTLocalBurn),
					Topics: [][]byte{
						token, big.NewInt(0).Bytes(), big.NewInt(20).Bytes(),
					},
				},
				{
					Identifier: []byte(core.BuiltInFunctionESDTLocalMint),
					Topics: [][]byte{
						token, big.NewInt(0).Bytes(), big.NewInt(25).Bytes(),
					},
				},
				nil,
			},
		},
	}

	marshalizer := testscommon.MarshalizerMock{}
	storer := &testscommon.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			supplyESDT := &SupplyESDT{
				Supply: big.NewInt(1000),
			}
			return marshalizer.Marshal(supplyESDT)
		},
		PutCalled: func(key, data []byte) error {
			supplyKey := string(token)
			require.Equal(t, supplyKey, string(key))

			var supplyESDT SupplyESDT
			_ = marshalizer.Unmarshal(&supplyESDT, data)
			require.Equal(t, big.NewInt(1005), supplyESDT.Supply)

			return nil
		},
	}

	logsProc := newLogsProcessor(marshalizer, storer)

	err := logsProc.processLogs(0, logs, false)
	require.Nil(t, err)
}
