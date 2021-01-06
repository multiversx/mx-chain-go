package totalStakedAPI

import (
	"errors"
	"math/big"
	"testing"
	"time"

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
		argsFunc func() (marshalizer marshal.Marshalizer, cacheDuration time.Duration, accounts state.AccountsAdapter)
		exError  error
	}{
		{
			name: "NilMarshalizer",
			argsFunc: func() (marshalizer marshal.Marshalizer, cacheDuration time.Duration, accounts state.AccountsAdapter) {
				return nil, time.Duration(100), &mock.AccountsStub{}
			},
			exError: ErrNilMarshalizer,
		},
		{
			name: "InvalidCacheDuration",
			argsFunc: func() (marshalizer marshal.Marshalizer, cacheDuration time.Duration, accounts state.AccountsAdapter) {
				return &mock.MarshalizerMock{}, time.Duration(0), &mock.AccountsStub{}
			},
			exError: ErrInvalidTotalStakedValueCacheDuration,
		},
		{
			name: "NilAccounts",
			argsFunc: func() (marshalizer marshal.Marshalizer, cacheDuration time.Duration, accounts state.AccountsAdapter) {
				return &mock.MarshalizerMock{}, time.Duration(1), nil
			},
			exError: ErrNilAccountsAdapter,
		},
		{
			name: "ShouldWork",
			argsFunc: func() (marshalizer marshal.Marshalizer, cacheDuration time.Duration, accounts state.AccountsAdapter) {
				return &mock.MarshalizerMock{}, time.Duration(100), &mock.AccountsStub{}
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

func TestTotalStakedValueProcessor_GetTotalStakedValue_ValueFromCache(t *testing.T) {
	t.Parallel()

	totalStaked := big.NewInt(10000)
	totalStakedProc, _ := NewTotalStakedValueProcessor(&mock.MarshalizerMock{}, 5*time.Second, &mock.AccountsStub{})
	require.False(t, totalStakedProc.IsInterfaceNil())

	totalStakedProc.totalStakedValue = totalStaked
	totalStakedProc.lastComputeTime = time.Now()

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, err)
	require.Equal(t, totalStaked, resTotalStaked)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_CannotGetAccount(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	totalStakedProc, _ := NewTotalStakedValueProcessor(&mock.MarshalizerMock{}, 5*time.Second, &mock.AccountsStub{
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

	totalStakedProc, _ := NewTotalStakedValueProcessor(&mock.MarshalizerMock{}, 5*time.Second, &mock.AccountsStub{
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

	totalStakedProc, _ := NewTotalStakedValueProcessor(&mock.MarshalizerMock{}, 5*time.Second, &mock.AccountsStub{
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

	totalStakedProc, _ := NewTotalStakedValueProcessor(&mock.MarshalizerMock{}, 5*time.Second, &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
			return acc, nil
		},
	})

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, resTotalStaked)
	require.Equal(t, expectedErr, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_ShouldWork(t *testing.T) {
	t.Parallel()

	rootHash := []byte("hash")
	totalStaked := big.NewInt(100000000000)
	marshalizer := &mock.MarshalizerFake{}
	validatorData := &systemSmartContracts.ValidatorDataV2{
		TotalStakeValue: totalStaked,
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
				trieLeaf := keyValStorage.NewKeyValStorage(append(rootHash), append(marshalledData, suffix...))
				ch <- trieLeaf
				close(ch)
			}()

			return ch, nil
		},
	})

	totalStakedProc, _ := NewTotalStakedValueProcessor(marshalizer, 5*time.Second, &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
			return acc, nil
		},
	})

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, err)
	require.Equal(t, totalStaked, resTotalStaked)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_ValueChange(t *testing.T) {
	t.Parallel()

	rootHash := []byte("hash")
	totalStaked := big.NewInt(100000000000)
	marshalizer := &mock.MarshalizerFake{}
	validatorData := &systemSmartContracts.ValidatorDataV2{
		TotalStakeValue: totalStaked,
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
				trieLeaf := keyValStorage.NewKeyValStorage(append(rootHash), append(marshalledData, suffix...))
				ch <- trieLeaf
				close(ch)
			}()

			return ch, nil
		},
	})

	totalStakedProc, _ := NewTotalStakedValueProcessor(marshalizer, 1*time.Second, &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
			return acc, nil
		},
	})

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, err)
	require.Equal(t, totalStaked, resTotalStaked)

	time.Sleep(time.Second)

	validatorData = &systemSmartContracts.ValidatorDataV2{
		TotalStakeValue: big.NewInt(200000000000),
	}
	marshalledData, _ = marshalizer.Marshal(validatorData)
	acc.SetDataTrie(&mock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		GetAllLeavesOnChannelCalled: func(hash []byte) (chan core.KeyValueHolder, error) {
			ch := make(chan core.KeyValueHolder)

			go func() {
				trieLeaf := keyValStorage.NewKeyValStorage(append(rootHash), append(marshalledData, suffix...))
				ch <- trieLeaf
				close(ch)
			}()

			return ch, nil
		},
	})

	resTotalStaked, err = totalStakedProc.GetTotalStakedValue()
	require.Nil(t, err)
	require.Equal(t, big.NewInt(200000000000), resTotalStaked)
}
