package esdtSupply

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"

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

	marshalizer := testscommon.MarshalizerMock{}
	suppliesStorer := &storageStubs.StorerStub{
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

	suppliesProc, err := NewSuppliesProcessor(marshalizer, suppliesStorer, &storageStubs.StorerStub{})
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(0, logs)
	require.Nil(t, err)
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
	require.Equal(t, &SupplyESDT{
		Supply: big.NewInt(123456),
	}, res)
}
