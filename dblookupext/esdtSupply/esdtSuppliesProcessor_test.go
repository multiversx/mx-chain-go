package esdtSupply

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewSuppliesProcessor(t *testing.T) {
	t.Parallel()

	_, err := NewSuppliesProcessor(nil, &testscommon.StorerStub{}, &testscommon.StorerStub{})
	require.Equal(t, core.ErrNilMarshalizer, err)

	_, err = NewSuppliesProcessor(&testscommon.MarshalizerMock{}, nil, &testscommon.StorerStub{})
	require.Equal(t, core.ErrNilStore, err)

	_, err = NewSuppliesProcessor(&testscommon.MarshalizerMock{}, &testscommon.StorerStub{}, nil)
	require.Equal(t, core.ErrNilStore, err)

	proc, err := NewSuppliesProcessor(&testscommon.MarshalizerMock{}, &testscommon.StorerStub{}, &testscommon.StorerStub{})
	require.Nil(t, err)
	require.NotNil(t, proc)
	require.False(t, proc.IsInterfaceNil())
}

func TestProcessLogsSaveSupply(t *testing.T) {
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
	suppliesStorer := &testscommon.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return nil, errors.New("not found")
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

	suppliesProc, err := NewSuppliesProcessor(marshalizer, suppliesStorer, &testscommon.StorerStub{})
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(0, logs)
	require.Nil(t, err)
}

func TestSupplyESDT_GetSupply(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.MarshalizerMock{}
	proc, _ := NewSuppliesProcessor(marshalizer, &testscommon.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if string(key) == "my-token" {
				supply := &SupplyESDT{Supply: big.NewInt(123456)}
				return marshalizer.Marshal(supply)
			}
			return nil, errors.New("local err")
		},
	}, &testscommon.StorerStub{})

	res, err := proc.GetESDTSupply("my-token")
	require.Nil(t, err)
	require.Equal(t, "123456", res)
}
