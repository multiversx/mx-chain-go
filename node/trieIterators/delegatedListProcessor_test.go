package trieIterators

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDelegatedListProcessor(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dlp, err := NewDelegatedListProcessor(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
			require.Nil(t, dlp)
		})
	}

	dlp, err := NewDelegatedListProcessor(createMockArgs())
	require.NotNil(t, dlp)
	require.Nil(t, err)
}

func TestDelegatedListProc_GetDelegatorsListGetAllContractAddressesFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	arg := createMockArgs()
	arg.QueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
			return nil, nil, expectedErr
		},
	}
	dlp, _ := NewDelegatedListProcessor(arg)

	list, err := dlp.GetDelegatorsList(context.Background())
	assert.Nil(t, list)
	assert.Equal(t, expectedErr, err)

	arg = createMockArgs()
	arg.QueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.UserError,
			}, nil, nil
		},
	}
	dlp, _ = NewDelegatedListProcessor(arg)

	list, err = dlp.GetDelegatorsList(context.Background())
	assert.Nil(t, list)
	assert.True(t, errors.Is(err, epochStart.ErrExecutingSystemScCode))
}

func TestDelegatedListProc_GetDelegatorsListContextShouldTimeout(t *testing.T) {
	t.Parallel()

	delegators := [][]byte{[]byte("delegator1"), []byte("delegator2")}

	arg := createMockArgs()
	arg.PublicKeyConverter = testscommon.NewPubkeyConverterMock(10)
	delegationSc := [][]byte{[]byte("delegationSc1"), []byte("delegationSc2")}
	arg.QueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
			switch query.FuncName {
			case "getAllContractAddresses":
				return &vmcommon.VMOutput{
					ReturnData: delegationSc,
				}, nil, nil
			case "getUserActiveStake":
				for index, delegator := range delegators {
					if bytes.Equal(delegator, query.Arguments[0]) {
						value := big.NewInt(int64(index + 1))
						return &vmcommon.VMOutput{
							ReturnData: [][]byte{value.Bytes()},
						}, nil, nil
					}
				}
			}

			return nil, nil, fmt.Errorf("not an expected call")
		},
	}
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return createScAccount(addressContainer, delegators, addressContainer, time.Second), nil
		},
		RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
			return nil
		},
	}
	dlp, _ := NewDelegatedListProcessor(arg)

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	delegatorsValues, err := dlp.GetDelegatorsList(ctxWithTimeout)
	require.Nil(t, delegatorsValues)
	require.Equal(t, ErrTrieOperationsTimeout, err)
}

func TestDelegatedListProc_GetDelegatorsListShouldWork(t *testing.T) {
	t.Parallel()

	delegators := [][]byte{[]byte("delegator1"), []byte("delegator2")}

	arg := createMockArgs()
	arg.PublicKeyConverter = testscommon.NewPubkeyConverterMock(10)
	delegationSc := [][]byte{[]byte("delegationSc1"), []byte("delegationSc2")}
	arg.QueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
			switch query.FuncName {
			case "getAllContractAddresses":
				return &vmcommon.VMOutput{
					ReturnData: delegationSc,
				}, nil, nil
			case "getUserActiveStake":
				for index, delegator := range delegators {
					if bytes.Equal(delegator, query.Arguments[0]) {
						value := big.NewInt(int64(index + 1))
						return &vmcommon.VMOutput{
							ReturnData: [][]byte{value.Bytes()},
						}, nil, nil
					}
				}
			}

			return nil, nil, fmt.Errorf("not an expected call")
		},
	}
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return createScAccount(addressContainer, delegators, addressContainer, 0), nil
		},
		RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
			return nil
		},
	}
	dlp, _ := NewDelegatedListProcessor(arg)

	delegatorsValues, err := dlp.GetDelegatorsList(context.Background())
	require.Nil(t, err)
	require.Equal(t, 2, len(delegatorsValues))

	encodedDelegator1PubKey, err := arg.PublicKeyConverter.Encode(delegators[0])
	require.Nil(t, err)

	encodedDelegationSc1, err := arg.PublicKeyConverter.Encode(delegationSc[0])
	require.Nil(t, err)

	encodedDelegationSc2, err := arg.PublicKeyConverter.Encode(delegationSc[1])
	require.Nil(t, err)

	expectedDelegator1 := api.Delegator{
		DelegatorAddress: encodedDelegator1PubKey,
		DelegatedTo: []*api.DelegatedValue{
			{
				DelegationScAddress: encodedDelegationSc1,
				Value:               "1",
			},
			{
				DelegationScAddress: encodedDelegationSc2,
				Value:               "1",
			},
		},
		Total:         "2",
		TotalAsBigInt: big.NewInt(2),
	}

	encodedDelegator2PubKey, err := arg.PublicKeyConverter.Encode(delegators[1])
	require.Nil(t, err)

	expectedDelegator2 := api.Delegator{
		DelegatorAddress: encodedDelegator2PubKey,
		DelegatedTo: []*api.DelegatedValue{
			{
				DelegationScAddress: encodedDelegationSc1,
				Value:               "2",
			},
			{
				DelegationScAddress: encodedDelegationSc2,
				Value:               "2",
			},
		},
		Total:         "4",
		TotalAsBigInt: big.NewInt(4),
	}

	assert.Equal(t, []*api.Delegator{&expectedDelegator1, &expectedDelegator2}, delegatorsValues)
}

func TestDelegatedListProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var dlp *delegatedListProcessor
	require.True(t, dlp.IsInterfaceNil())

	dlp, _ = NewDelegatedListProcessor(createMockArgs())
	require.False(t, dlp.IsInterfaceNil())
}
