package accounts

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/require"
)

const baseTokenID = "WEGLD"

func TestNewESDTAsBalance(t *testing.T) {
	t.Parallel()

	t.Run("empty base token, should return error", func(t *testing.T) {
		esdtBalance, err := NewESDTAsBalance("", &marshallerMock.MarshalizerMock{})
		require.Equal(t, errorsMx.ErrEmptyBaseToken, err)
		require.Nil(t, esdtBalance)
	})
	t.Run("nil marshaller, should return error", func(t *testing.T) {
		esdtBalance, err := NewESDTAsBalance(baseTokenID, nil)
		require.Equal(t, errorsMx.ErrNilMarshalizer, err)
		require.Nil(t, esdtBalance)
	})
	t.Run("should work", func(t *testing.T) {
		esdtBalance, err := NewESDTAsBalance(baseTokenID, &marshallerMock.MarshalizerMock{})
		require.Nil(t, err)
		require.False(t, esdtBalance.IsInterfaceNil())
	})
}

func TestEsdtAsBalance_getESDTData(t *testing.T) {
	t.Parallel()

	t.Run("no value stored in account for esdt, should return empty token", func(t *testing.T) {
		esdtBalance, _ := NewESDTAsBalance(baseTokenID, &marshallerMock.MarshalizerMock{})
		accHandler := &trie.DataTrieTrackerStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				require.Equal(t, []byte(baseESDTKeyPrefix+baseTokenID), key)
				return nil, 0, errors.New("error retrieving value")
			},
		}

		esdtData, err := esdtBalance.getESDTData(accHandler)
		require.Nil(t, err)
		require.Equal(t, createEmptyESDT(), esdtData)
	})

	t.Run("empty buffer when retrieving data, should return empty token", func(t *testing.T) {
		esdtBalance, _ := NewESDTAsBalance(baseTokenID, &marshallerMock.MarshalizerMock{})
		accHandler := &trie.DataTrieTrackerStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, nil
			},
		}

		esdtData, err := esdtBalance.getESDTData(accHandler)
		require.Nil(t, err)
		require.Equal(t, createEmptyESDT(), esdtData)
	})

	t.Run("cannot unmarshall, should return error", func(t *testing.T) {
		errUnmarshall := errors.New("cannot unmarshall")
		marshaller := &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errUnmarshall
			},
		}
		esdtBalance, _ := NewESDTAsBalance(baseTokenID, marshaller)

		esdtData, err := esdtBalance.getESDTData(&trie.DataTrieTrackerStub{})
		require.Nil(t, esdtData)
		require.Equal(t, errUnmarshall, err)
	})

	t.Run("empty value when unmarshalling, should fill mandatory fields", func(t *testing.T) {
		marshaller := &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				expectedObj := obj.(*esdt.ESDigitalToken)
				expectedObj.Properties = []byte("properties")

				return nil
			},
		}
		esdtBalance, _ := NewESDTAsBalance(baseTokenID, marshaller)

		esdtData, err := esdtBalance.getESDTData(&trie.DataTrieTrackerStub{})
		require.Equal(t, &esdt.ESDigitalToken{
			Type:       uint32(core.Fungible),
			Value:      big.NewInt(0),
			Properties: []byte("properties"),
		}, esdtData)
		require.Nil(t, err)
	})
}

func TestEsdtAsBalance_GetBalance(t *testing.T) {
	t.Parallel()

	t.Run("could not load balance, should return 0 value", func(t *testing.T) {
		marshaller := &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errors.New("cannot unmarshall")
			},
		}
		esdtBalance, _ := NewESDTAsBalance(baseTokenID, marshaller)
		balance := esdtBalance.GetBalance(&trie.DataTrieTrackerStub{})
		require.Equal(t, big.NewInt(0), balance)
	})

	t.Run("should work", func(t *testing.T) {
		expectedBalance := big.NewInt(444)
		marshaller := &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				expectedObj := obj.(*esdt.ESDigitalToken)
				expectedObj.Value = expectedBalance

				return nil
			},
		}
		esdtBalance, _ := NewESDTAsBalance(baseTokenID, marshaller)
		balance := esdtBalance.GetBalance(&trie.DataTrieTrackerStub{})
		require.Equal(t, expectedBalance, balance)
	})
}

func TestEsdtAsBalance_AddToBalance(t *testing.T) {
	t.Parallel()

	currentBalance := &esdt.ESDigitalToken{
		Value: big.NewInt(123),
	}
	marshaller := &marshallerMock.MarshalizerMock{}
	esdtBalance, _ := NewESDTAsBalance(baseTokenID, marshaller)

	newValue := big.NewInt(321)
	expectedNewBalance := big.NewInt(0).Add(currentBalance.Value, newValue)
	wasBalanceSaved := false
	accHandler := &trie.DataTrieTrackerStub{
		RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
			require.Equal(t, []byte(baseESDTKeyPrefix+baseTokenID), key)

			storedValue, err := marshaller.Marshal(currentBalance)
			require.Nil(t, err)

			return storedValue, 0, nil
		},
		SaveKeyValueCalled: func(key []byte, value []byte) error {
			expectedBalance := &esdt.ESDigitalToken{
				Value: expectedNewBalance,
			}
			expectedBalanceMarshalledData, err := marshaller.Marshal(expectedBalance)
			require.Nil(t, err)
			require.Equal(t, expectedBalanceMarshalledData, value)
			require.Equal(t, []byte(baseESDTKeyPrefix+baseTokenID), key)

			wasBalanceSaved = true
			return nil
		},
	}

	err := esdtBalance.AddToBalance(accHandler, newValue)
	require.Nil(t, err)
	require.True(t, wasBalanceSaved)

	err = esdtBalance.AddToBalance(accHandler, big.NewInt(-4412))
	require.Equal(t, errorsMx.ErrInsufficientFunds, err)
}

func TestEsdtAsBalance_SubFromBalance(t *testing.T) {
	t.Parallel()

	currentBalance := &esdt.ESDigitalToken{
		Value: big.NewInt(121),
	}
	marshaller := &marshallerMock.MarshalizerMock{}
	esdtBalance, _ := NewESDTAsBalance(baseTokenID, marshaller)

	subBalance := big.NewInt(22)
	expectedNewBalance := big.NewInt(0).Sub(currentBalance.Value, subBalance)
	wasBalanceSaved := false
	accHandler := &trie.DataTrieTrackerStub{
		RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
			require.Equal(t, []byte(baseESDTKeyPrefix+baseTokenID), key)

			storedValue, err := marshaller.Marshal(currentBalance)
			require.Nil(t, err)

			return storedValue, 0, nil
		},
		SaveKeyValueCalled: func(key []byte, value []byte) error {
			expectedBalance := &esdt.ESDigitalToken{
				Value: expectedNewBalance,
			}
			expectedBalanceMarshalledData, err := marshaller.Marshal(expectedBalance)
			require.Nil(t, err)
			require.Equal(t, expectedBalanceMarshalledData, value)
			require.Equal(t, []byte(baseESDTKeyPrefix+baseTokenID), key)

			wasBalanceSaved = true
			return nil
		},
	}

	err := esdtBalance.SubFromBalance(accHandler, subBalance)
	require.Nil(t, err)
	require.True(t, wasBalanceSaved)

	err = esdtBalance.SubFromBalance(accHandler, big.NewInt(122))
	require.Equal(t, errorsMx.ErrInsufficientFunds, err)
}
