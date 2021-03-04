package stakeValuesProcessor

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/keyValStorage"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/stretchr/testify/require"
)

func TestNewTotalStakedValueProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() (nodePrice string, marshalizer marshal.Marshalizer, accounts state.AccountsAdapter)
		exError  error
	}{
		{
			name: "NilMarshalizer",
			argsFunc: func() (nodePrice string, marshalizer marshal.Marshalizer, accounts state.AccountsAdapter) {
				return "100", nil, &mock.AccountsStub{}
			},
			exError: ErrNilMarshalizer,
		},
		{
			name: "InvalidNodePrice",
			argsFunc: func() (nodePrice string, marshalizer marshal.Marshalizer, accounts state.AccountsAdapter) {
				return "AAA", &mock.MarshalizerMock{}, &mock.AccountsStub{}
			},
			exError: ErrInvalidNodePrice,
		},
		{
			name: "NilAccounts",
			argsFunc: func() (nodePrice string, marshalizer marshal.Marshalizer, accounts state.AccountsAdapter) {
				return "100", &mock.MarshalizerMock{}, nil
			},
			exError: ErrNilAccountsAdapter,
		},
		{
			name: "ShouldWork",
			argsFunc: func() (nodePrice string, marshalizer marshal.Marshalizer, accounts state.AccountsAdapter) {
				return "100", &mock.MarshalizerMock{}, &mock.AccountsStub{}
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTotalStakedValueProcessor(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
		})
	}
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_CannotGetAccount(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	totalStakedProc, _ := NewTotalStakedValueProcessor("100", &mock.MarshalizerMock{}, &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
			return nil, expectedErr
		},
	})

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, resTotalStaked)
	require.Equal(t, expectedErr, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_CannotCastAccount(t *testing.T) {
	t.Parallel()

	totalStakedProc, _ := NewTotalStakedValueProcessor("100", &mock.MarshalizerMock{}, &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
			return nil, nil
		},
	})

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, resTotalStaked)
	require.Equal(t, ErrCannotCastAccountHandlerToUserAccount, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_CannotGetRootHash(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	acc, _ := state.NewUserAccount([]byte("newaddress"))
	acc.SetDataTrie(&mock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return nil, expectedErr
		},
	})

	totalStakedProc, _ := NewTotalStakedValueProcessor("100", &mock.MarshalizerMock{}, &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
			return acc, nil
		},
	})

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, resTotalStaked)
	require.Equal(t, expectedErr, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_CannotGetAllLeaves(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	acc, _ := state.NewUserAccount([]byte("newaddress"))
	acc.SetDataTrie(&mock.TrieStub{
		GetAllLeavesOnChannelCalled: func(rootHash []byte) (chan core.KeyValueHolder, error) {
			return nil, expectedErr
		},
	})

	totalStakedProc, _ := NewTotalStakedValueProcessor("100", &mock.MarshalizerMock{}, &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
			return acc, nil
		},
	})

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, resTotalStaked)
	require.Equal(t, expectedErr, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue(t *testing.T) {
	t.Parallel()

	rootHash := []byte("hash")
	totalStaked := big.NewInt(100000000000)
	marshalizer := &mock.MarshalizerFake{}
	validatorData := &systemSmartContracts.ValidatorDataV2{
		TotalStakeValue: totalStaked,
		NumRegistered:   2,
	}
	marshalledData, _ := marshalizer.Marshal(validatorData)

	suffix := append(rootHash, vm.ValidatorSCAddress...)

	acc, _ := state.NewUserAccount([]byte("newaddress"))
	acc.SetDataTrie(&mock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		GetAllLeavesOnChannelCalled: func(hash []byte) (chan core.KeyValueHolder, error) {
			ch := make(chan core.KeyValueHolder)

			go func() {
				trieLeaf := keyValStorage.NewKeyValStorage(rootHash, append(marshalledData, suffix...))
				ch <- trieLeaf
				close(ch)
			}()

			return ch, nil
		},
	})

	totalStakedProc, _ := NewTotalStakedValueProcessor("750", marshalizer, &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
			return acc, nil
		},
	})

	expectedTopUp := big.NewInt(99999998500)
	stakeValues, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, err)
	require.Equal(t, totalStaked, stakeValues.TotalStaked)
	require.Equal(t, expectedTopUp, stakeValues.TopUp)
}
