package trieIterators

import (
	"errors"
	"math/big"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/keyValStorage"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/stretchr/testify/require"
)

func createAccountsWrapper() *AccountsWrapper {
	return &AccountsWrapper{
		Mutex:           &sync.Mutex{},
		AccountsAdapter: &mock.AccountsStub{},
	}
}

func TestNewTotalStakedValueProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() (
			accounts *AccountsWrapper,
			blockChain data.ChainHandler,
			queryService process.SCQueryService,
			publicKeyConverter core.PubkeyConverter)
		exError error
	}{
		{
			name: "NilAccounts",
			argsFunc: func() (accounts *AccountsWrapper, blockChain data.ChainHandler, queryService process.SCQueryService, publicKeyConverter core.PubkeyConverter) {
				return nil, &mock.BlockChainMock{}, &mock.SCQueryServiceStub{}, &mock.PubkeyConverterMock{}
			},
			exError: ErrNilAccountsAdapter,
		},
		{
			name: "ShouldWork",
			argsFunc: func() (accounts *AccountsWrapper, blockChain data.ChainHandler, queryService process.SCQueryService, publicKeyConverter core.PubkeyConverter) {
				return createAccountsWrapper(), &mock.BlockChainMock{}, &mock.SCQueryServiceStub{}, &mock.PubkeyConverterMock{}
			},
			exError: nil,
		},
		{
			name: "NilBlockChain",
			argsFunc: func() (accounts *AccountsWrapper, blockChain data.ChainHandler, queryService process.SCQueryService, publicKeyConverter core.PubkeyConverter) {
				return createAccountsWrapper(), nil, &mock.SCQueryServiceStub{}, &mock.PubkeyConverterMock{}
			},
			exError: ErrNilBlockChain,
		},
		{
			name: "NilQueryService",
			argsFunc: func() (accounts *AccountsWrapper, blockChain data.ChainHandler, queryService process.SCQueryService, publicKeyConverter core.PubkeyConverter) {
				return createAccountsWrapper(), &mock.BlockChainMock{}, nil, &mock.PubkeyConverterMock{}
			},
			exError: ErrNilQueryService,
		},
		{
			name: "NilPubKeyConverter",
			argsFunc: func() (accounts *AccountsWrapper, blockChain data.ChainHandler, queryService process.SCQueryService, publicKeyConverter core.PubkeyConverter) {
				return createAccountsWrapper(), &mock.BlockChainMock{}, &mock.SCQueryServiceStub{}, nil
			},
			exError: ErrNilPubkeyConverter,
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
	totalStakedProc, _ := NewTotalStakedValueProcessor(
		&AccountsWrapper{
			Mutex: &sync.Mutex{},
			AccountsAdapter: &mock.AccountsStub{
				RecreateTrieCalled: func(rootHash []byte) error {
					return nil
				},
				GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
					return nil, expectedErr
				},
			},
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{}
			},
		},
		&mock.SCQueryServiceStub{},
		&mock.PubkeyConverterMock{},
	)

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, resTotalStaked)
	require.Equal(t, expectedErr, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValueilHeaderShouldError(t *testing.T) {
	t.Parallel()

	totalStakedProc, _ := NewTotalStakedValueProcessor(
		&AccountsWrapper{
			Mutex:           &sync.Mutex{},
			AccountsAdapter: &mock.AccountsStub{},
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		},
		&mock.SCQueryServiceStub{},
		&mock.PubkeyConverterMock{},
	)

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, resTotalStaked)
	require.Equal(t, ErrNodeNotInitialized, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_CannotRecreateTree(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	totalStakedProc, _ := NewTotalStakedValueProcessor(
		&AccountsWrapper{
			Mutex: &sync.Mutex{},
			AccountsAdapter: &mock.AccountsStub{
				RecreateTrieCalled: func(rootHash []byte) error {
					return expectedErr
				},
			},
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{}
			},
		},
		&mock.SCQueryServiceStub{},
		&mock.PubkeyConverterMock{},
	)

	require.False(t, totalStakedProc.IsInterfaceNil())

	resTotalStaked, err := totalStakedProc.GetTotalStakedValue()
	require.Nil(t, resTotalStaked)
	require.Equal(t, expectedErr, err)
}

func TestTotalStakedValueProcessor_GetTotalStakedValue_CannotCastAccount(t *testing.T) {
	t.Parallel()

	totalStakedProc, _ := NewTotalStakedValueProcessor(
		&AccountsWrapper{
			Mutex: &sync.Mutex{},
			AccountsAdapter: &mock.AccountsStub{
				GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
					return nil, nil
				},
				RecreateTrieCalled: func(rootHash []byte) error {
					return nil
				},
			},
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{}
			},
		},
		&mock.SCQueryServiceStub{},
		&mock.PubkeyConverterMock{},
	)

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

	totalStakedProc, _ := NewTotalStakedValueProcessor(
		&AccountsWrapper{
			Mutex: &sync.Mutex{},
			AccountsAdapter: &mock.AccountsStub{
				GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
					return acc, nil
				},
				RecreateTrieCalled: func(rootHash []byte) error {
					return nil
				},
			},
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{}
			},
		},
		&mock.SCQueryServiceStub{},
		&mock.PubkeyConverterMock{},
	)

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

	totalStakedProc, _ := NewTotalStakedValueProcessor(
		&AccountsWrapper{
			Mutex: &sync.Mutex{},
			AccountsAdapter: &mock.AccountsStub{
				GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
					return acc, nil
				},
				RecreateTrieCalled: func(rootHash []byte) error {
					return nil
				},
			},
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{}
			},
		},
		&mock.SCQueryServiceStub{},
		&mock.PubkeyConverterMock{},
	)

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

	leafKey2 := "0123456789"
	leafKey3 := "0123456781"
	leafKey4 := "0123456783"
	leafKey5 := "0123456780"
	leafKey6 := "0123456788"
	acc, _ := state.NewUserAccount([]byte("newaddress"))
	acc.SetDataTrie(&mock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		GetAllLeavesOnChannelCalled: func(hash []byte) (chan core.KeyValueHolder, error) {
			ch := make(chan core.KeyValueHolder)

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

			return ch, nil
		},
	})

	expectedErr := errors.New("expected error")
	totalStakedProc, _ := NewTotalStakedValueProcessor(
		&AccountsWrapper{
			Mutex: &sync.Mutex{},
			AccountsAdapter: &mock.AccountsStub{
				GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
					return acc, nil
				},
				RecreateTrieCalled: func(rootHash []byte) error {
					return nil
				},
			},
		},
		&mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlock{}
			},
		},
		&mock.SCQueryServiceStub{
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
		}, mock.NewPubkeyConverterMock(10))

	stakeValues, err := totalStakedProc.GetTotalStakedValue()
	require.Equal(t, &api.StakeValues{
		TotalStaked: big.NewInt(600),
		TopUp:       big.NewInt(110),
	}, stakeValues)
	require.Nil(t, err)
}
