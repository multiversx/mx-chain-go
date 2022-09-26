package trieIterators

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/keyValStorage"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func createAccountsWrapper() *AccountsWrapper {
	return &AccountsWrapper{
		Mutex:           &sync.Mutex{},
		AccountsAdapter: &stateMock.AccountsStub{},
	}
}

func createMockArgs() ArgTrieIteratorProcessor {
	return ArgTrieIteratorProcessor{
		Accounts:           createAccountsWrapper(),
		QueryService:       &mock.SCQueryServiceStub{},
		PublicKeyConverter: &mock.PubkeyConverterMock{},
	}
}

func TestNewTotalStakedValueProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() ArgTrieIteratorProcessor
		exError  error
	}{
		{
			name: "NilAccounts",
			argsFunc: func() ArgTrieIteratorProcessor {
				arg := createMockArgs()
				arg.Accounts = nil

				return arg
			},
			exError: ErrNilAccountsAdapter,
		},
		{
			name: "ShouldWork",
			argsFunc: func() ArgTrieIteratorProcessor {
				return createMockArgs()
			},
			exError: nil,
		},
		{
			name: "NilQueryService",
			argsFunc: func() ArgTrieIteratorProcessor {
				arg := createMockArgs()
				arg.QueryService = nil

				return arg
			},
			exError: ErrNilQueryService,
		},
		{
			name: "NilPubKeyConverter",
			argsFunc: func() ArgTrieIteratorProcessor {
				arg := createMockArgs()
				arg.PublicKeyConverter = nil

				return arg
			},
			exError: ErrNilPubkeyConverter,
		},
		{
			name: "NilAccountsWrapperMutex",
			argsFunc: func() ArgTrieIteratorProcessor {
				arg := createMockArgs()
				arg.Accounts.Mutex = nil

				return arg
			},
			exError: ErrNilMutex,
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
	arg := createMockArgs()
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return nil, expectedErr
		},
	}
	totalStakedProc, _ := NewTotalStakedValueProcessor(arg)

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue(context.Background())
	require.Nil(t, resTotalStaked)
	require.Equal(t, expectedErr, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValueAccountsAdapterErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	arg := createMockArgs()
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return nil, expectedErr
		},
	}
	totalStakedProc, _ := NewTotalStakedValueProcessor(arg)

	require.False(t, totalStakedProc.IsInterfaceNil())

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue(context.Background())
	require.Nil(t, resTotalStaked)
	require.Equal(t, expectedErr, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_CannotCastAccount(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return nil, nil
		},
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	totalStakedProc, _ := NewTotalStakedValueProcessor(arg)

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue(context.Background())
	require.Nil(t, resTotalStaked)
	require.Equal(t, ErrCannotCastAccountHandlerToUserAccount, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_CannotGetRootHash(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	acc, _ := state.NewUserAccount([]byte("newaddress"))
	acc.SetDataTrie(&trieMock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return nil, expectedErr
		},
	})

	arg := createMockArgs()
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return acc, nil
		},
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	totalStakedProc, _ := NewTotalStakedValueProcessor(arg)

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue(context.Background())
	require.Nil(t, resTotalStaked)
	require.Equal(t, expectedErr, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_ContextShouldTimeout(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewUserAccount([]byte("newaddress"))
	acc.SetDataTrie(&trieMock.TrieStub{
		GetAllLeavesOnChannelCalled: func(chLeaves chan core.KeyValueHolder, _ context.Context, _ []byte, _ common.KeyBuilder) error {
			time.Sleep(time.Second)
			close(chLeaves)
			return nil
		},
		RootCalled: func() ([]byte, error) {
			return nil, nil
		},
	})

	arg := createMockArgs()
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return acc, nil
		},
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	totalStakedProc, _ := NewTotalStakedValueProcessor(arg)

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue(ctxWithTimeout)
	require.Nil(t, resTotalStaked)
	require.Equal(t, ErrTrieOperationsTimeout, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_CannotGetAllLeaves(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	acc, _ := state.NewUserAccount([]byte("newaddress"))
	acc.SetDataTrie(&trieMock.TrieStub{
		GetAllLeavesOnChannelCalled: func(_ chan core.KeyValueHolder, _ context.Context, _ []byte, _ common.KeyBuilder) error {
			return expectedErr
		},
		RootCalled: func() ([]byte, error) {
			return nil, nil
		},
	})

	arg := createMockArgs()
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return acc, nil
		},
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	totalStakedProc, _ := NewTotalStakedValueProcessor(arg)

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue(context.Background())
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

	leafKey2 := "0123456789"
	leafKey3 := "0123456781"
	leafKey4 := "0123456783"
	leafKey5 := "0123456780"
	leafKey6 := "0123456788"
	acc, _ := state.NewUserAccount([]byte("newaddress"))
	acc.SetDataTrie(&trieMock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		GetAllLeavesOnChannelCalled: func(ch chan core.KeyValueHolder, ctx context.Context, rootHash []byte, _ common.KeyBuilder) error {
			go func() {
				leaf1 := keyValStorage.NewKeyValStorage(rootHash, append(marshalledData, suffix...))
				ch <- leaf1

				leaf2 := keyValStorage.NewKeyValStorage([]byte(leafKey2), nil)
				ch <- leaf2

				leaf3 := keyValStorage.NewKeyValStorage([]byte(leafKey3), nil)
				ch <- leaf3

				leaf4 := keyValStorage.NewKeyValStorage([]byte(leafKey4), nil)
				ch <- leaf4

				leaf5 := keyValStorage.NewKeyValStorage([]byte(leafKey5), nil)
				ch <- leaf5

				leaf6 := keyValStorage.NewKeyValStorage([]byte(leafKey6), nil)
				ch <- leaf6

				close(ch)
			}()

			return nil
		},
	})

	expectedErr := errors.New("expected error")
	arg := createMockArgs()
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return acc, nil
		},
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	arg.QueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
			switch string(query.Arguments[0]) {
			case leafKey3:
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.UserError,
				}, nil

			case leafKey4:
				return &vmcommon.VMOutput{}, nil

			case leafKey5:
				return &vmcommon.VMOutput{
					ReturnData: [][]byte{
						big.NewInt(50).Bytes(), big.NewInt(100).Bytes(), big.NewInt(0).Bytes(),
					},
				}, nil

			case leafKey6:
				return &vmcommon.VMOutput{
					ReturnData: [][]byte{
						big.NewInt(60).Bytes(), big.NewInt(500).Bytes(), big.NewInt(0).Bytes(),
					},
				}, nil

			default:
				return nil, expectedErr
			}
		},
	}
	arg.PublicKeyConverter = mock.NewPubkeyConverterMock(10)
	totalStakedProc, _ := NewTotalStakedValueProcessor(arg)

	stakeValues, err := totalStakedProc.GetTotalStakedValue(context.Background())
	require.Equal(t, &api.StakeValues{
		BaseStaked: big.NewInt(490),
		TopUp:      big.NewInt(110),
	}, stakeValues)
	require.Nil(t, err)
}
