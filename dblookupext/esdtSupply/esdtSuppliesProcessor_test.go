package esdtSupply

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestNewSuppliesProcessor(t *testing.T) {
	t.Parallel()

	_, err := NewSuppliesProcessor(nil, &storageStubs.StorerStub{}, &storageStubs.StorerStub{})
	require.Equal(t, core.ErrNilMarshalizer, err)

	_, err = NewSuppliesProcessor(&testscommon.MarshalizerMock{}, nil, &storageStubs.StorerStub{})
	require.Equal(t, core.ErrNilStore, err)

	_, err = NewSuppliesProcessor(&testscommon.MarshalizerMock{}, &storageStubs.StorerStub{}, nil)
	require.Equal(t, core.ErrNilStore, err)

	proc, err := NewSuppliesProcessor(&testscommon.MarshalizerMock{}, &storageStubs.StorerStub{}, &storageStubs.StorerStub{})
	require.Nil(t, err)
	require.NotNil(t, proc)
	require.False(t, proc.IsInterfaceNil())
}

func TestProcessLogsSaveSupply(t *testing.T) {
	t.Parallel()

	token := []byte("nft-0001")
	logs := []*data.LogData{
		{
			TxHash: "txLog",
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
		{
			TxHash: "log",
		},
	}

	wasPutCalled := false
	marshalizer := testscommon.MarshalizerMock{}
	suppliesStorer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if string(key) == "processed-block" {
				pbn := ProcessedBlockNonce{Nonce: 5}
				pbnB, _ := marshalizer.Marshal(pbn)
				return pbnB, nil
			}

			return nil, storage.ErrKeyNotFound
		},
		PutCalled: func(key, data []byte) error {
			if string(key) == "processed-block" {
				return nil
			}

			supplyKey := string(token) + "-" + hex.EncodeToString(big.NewInt(2).Bytes())
			require.Equal(t, supplyKey, string(key))

			var supplyESDT SupplyESDT
			_ = marshalizer.Unmarshal(&supplyESDT, data)
			require.Equal(t, big.NewInt(30), supplyESDT.Supply)

			wasPutCalled = true

			return nil
		},
	}

	suppliesProc, err := NewSuppliesProcessor(marshalizer, suppliesStorer, &storageStubs.StorerStub{})
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(6, logs)
	require.Nil(t, err)

	require.True(t, wasPutCalled)
}

func TestProcessLogsSaveSupplyShouldUpdateSupplyMintedAndBurned(t *testing.T) {
	t.Parallel()

	token := []byte("nft-0001")
	logsCreate := []*data.LogData{
		{
			TxHash: "txLog",
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
				},
			},
		},
		{
			TxHash: "log",
		},
	}
	logsAddQuantity := []*data.LogData{
		{
			TxHash: "txLog",
			LogHandler: &transaction.Log{
				Events: []*transaction.Event{
					{
						Identifier: []byte("something"),
					},
					{
						Identifier: []byte(core.BuiltInFunctionESDTNFTAddQuantity),
						Topics: [][]byte{
							token, big.NewInt(2).Bytes(), big.NewInt(50).Bytes(),
						},
					},
				},
			},
		},
		{
			TxHash: "log",
		},
	}

	logsBurn := []*data.LogData{
		{
			TxHash: "txLog",
			LogHandler: &transaction.Log{
				Events: []*transaction.Event{
					{
						Identifier: []byte("something"),
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
		{
			TxHash: "log",
		},
	}

	membDB := testscommon.NewMemDbMock()
	marshalizer := testscommon.MarshalizerMock{}
	numTimesCalled := 0
	suppliesStorer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if string(key) == "processed-block" {
				pbn := ProcessedBlockNonce{Nonce: 5}
				pbnB, _ := marshalizer.Marshal(pbn)
				return pbnB, nil
			}
			supplyKey := string(token) + "-" + hex.EncodeToString(big.NewInt(2).Bytes())
			if string(key) == supplyKey {
				val, err := membDB.Get(key)
				if err != nil {
					return nil, storage.ErrKeyNotFound
				}
				return val, nil
			}
			return nil, storage.ErrKeyNotFound
		},
		PutCalled: func(key, data []byte) error {
			supplyKey := string(token) + "-" + hex.EncodeToString(big.NewInt(2).Bytes())
			if string(key) == supplyKey {
				switch numTimesCalled {
				case 0:
					supplyEsdt := getSupplyESDT(marshalizer, data)
					require.Equal(t, big.NewInt(10), supplyEsdt.Supply)
					require.Equal(t, big.NewInt(0), supplyEsdt.Burned)
					require.Equal(t, big.NewInt(10), supplyEsdt.Minted)
				case 1:
					supplyEsdt := getSupplyESDT(marshalizer, data)
					require.Equal(t, big.NewInt(60), supplyEsdt.Supply)
					require.Equal(t, big.NewInt(0), supplyEsdt.Burned)
					require.Equal(t, big.NewInt(60), supplyEsdt.Minted)
				case 2:
					supplyEsdt := getSupplyESDT(marshalizer, data)
					require.Equal(t, big.NewInt(30), supplyEsdt.Supply)
					require.Equal(t, big.NewInt(30), supplyEsdt.Burned)
					require.Equal(t, big.NewInt(60), supplyEsdt.Minted)
				}

				_ = membDB.Put(key, data)
				numTimesCalled++

				return nil
			}

			return nil
		},
	}

	suppliesProc, err := NewSuppliesProcessor(marshalizer, suppliesStorer, &storageStubs.StorerStub{})
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(6, logsCreate)
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(7, logsAddQuantity)
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(8, logsBurn)
	require.Nil(t, err)

	require.Equal(t, 3, numTimesCalled)
}

func getSupplyESDT(marshalizer marshal.Marshalizer, data []byte) SupplyESDT {
	var supplyESDT SupplyESDT
	_ = marshalizer.Unmarshal(&supplyESDT, data)

	makePropertiesNotNil(&supplyESDT)
	return supplyESDT
}

func TestSupplyESDT_GetSupply(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.MarshalizerMock{}
	proc, _ := NewSuppliesProcessor(marshalizer, &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if string(key) == "my-token" {
				supply := &SupplyESDT{Supply: big.NewInt(123456)}
				return marshalizer.Marshal(supply)
			}
			return nil, errors.New("local err")
		},
	}, &storageStubs.StorerStub{})

	res, err := proc.GetESDTSupply("my-token")
	require.Nil(t, err)
	expectedESDTSupply := &SupplyESDT{
		Supply: big.NewInt(123456),
		Burned: big.NewInt(0),
		Minted: big.NewInt(0),
	}

	require.Equal(t, expectedESDTSupply, res)
}
