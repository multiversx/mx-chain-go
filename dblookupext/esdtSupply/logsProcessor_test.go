package esdtSupply

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	storageRepo "github.com/ElrondNetwork/elrond-go-storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessLogsSaveSupplyNothingInStorage(t *testing.T) {
	t.Parallel()

	token := []byte("nft-0001")
	logs := map[string]*data.LogData{
		"txLog": {
			LogHandler: &transaction.Log{
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
		},
		"log": nil,
	}

	marshalizer := testscommon.MarshalizerMock{}
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return nil, storageRepo.ErrKeyNotFound
		},
		PutCalled: func(key, data []byte) error {
			if string(key) == processedBlockKey {
				return nil
			}

			supplyKey := string(token) + "-" + hex.EncodeToString(big.NewInt(2).Bytes())
			require.Equal(t, supplyKey, string(key))

			var supplyESDT SupplyESDT
			_ = marshalizer.Unmarshal(&supplyESDT, data)
			require.Equal(t, big.NewInt(30), supplyESDT.Supply)

			return nil
		},
	}

	logsProc := newLogsProcessor(marshalizer, storer)

	err := logsProc.processLogs(1, logs, false)
	require.Nil(t, err)
}

func TestTestProcessLogsSaveSupplyExistsInStorage(t *testing.T) {
	t.Parallel()

	token := []byte("esdt-miiu")

	logs := map[string]*data.LogData{
		"txLog": {
			LogHandler: &transaction.Log{
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
		},
	}

	marshalizer := testscommon.MarshalizerMock{}
	storer := &storageStubs.StorerStub{
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

func TestMakePropertiesNotNil(t *testing.T) {
	t.Parallel()

	t.Run("supply is nil", func(t *testing.T) {
		t.Parallel()

		provided := SupplyESDT{
			Supply: nil,
			Burned: big.NewInt(1),
			Minted: big.NewInt(2),
		}
		expected := SupplyESDT{
			Supply: big.NewInt(0),
			Burned: big.NewInt(1),
			Minted: big.NewInt(2),
		}
		makePropertiesNotNil(&provided)
		assert.Equal(t, expected, provided)
	})
	t.Run("burned is nil", func(t *testing.T) {
		t.Parallel()

		provided := SupplyESDT{
			Supply: big.NewInt(1),
			Burned: nil,
			Minted: big.NewInt(2),
		}
		expected := SupplyESDT{
			Supply: big.NewInt(1),
			Burned: big.NewInt(0),
			Minted: big.NewInt(2),
		}
		makePropertiesNotNil(&provided)
		assert.Equal(t, expected, provided)
	})
	t.Run("minted is nil", func(t *testing.T) {
		t.Parallel()

		provided := SupplyESDT{
			Supply: big.NewInt(1),
			Burned: big.NewInt(2),
			Minted: nil,
		}
		expected := SupplyESDT{
			Supply: big.NewInt(1),
			Burned: big.NewInt(2),
			Minted: big.NewInt(0),
		}
		makePropertiesNotNil(&provided)
		assert.Equal(t, expected, provided)
	})
	t.Run("all are nil", func(t *testing.T) {
		t.Parallel()

		provided := SupplyESDT{}
		expected := SupplyESDT{
			Supply: big.NewInt(0),
			Burned: big.NewInt(0),
			Minted: big.NewInt(0),
		}
		makePropertiesNotNil(&provided)
		assert.Equal(t, expected, provided)
	})
	t.Run("none is nil", func(t *testing.T) {
		t.Parallel()

		provided := SupplyESDT{
			Supply: big.NewInt(1),
			Burned: big.NewInt(2),
			Minted: big.NewInt(3),
		}
		expected := SupplyESDT{
			Supply: big.NewInt(1),
			Burned: big.NewInt(2),
			Minted: big.NewInt(3),
		}
		makePropertiesNotNil(&provided)
		assert.Equal(t, expected, provided)
	})

}
