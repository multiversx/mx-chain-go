package trieIterators

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/keyValStorage"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
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
		{
			name: "ShouldWork",
			argsFunc: func() ArgTrieIteratorProcessor {
				return createMockArgs()
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDelegatedListProcessor(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
		})
	}

	dlp, _ := NewDelegatedListProcessor(createMockArgs())
	assert.False(t, check.IfNil(dlp))
}

func TestDelegatedListProc_GetDelegatorsListGetAllContractAddressesFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	arg := createMockArgs()
	arg.QueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
			return nil, expectedErr
		},
	}
	dlp, _ := NewDelegatedListProcessor(arg)

	list, err := dlp.GetDelegatorsList()
	assert.Nil(t, list)
	assert.Equal(t, expectedErr, err)

	arg = createMockArgs()
	arg.QueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.UserError,
			}, nil
		},
	}
	dlp, _ = NewDelegatedListProcessor(arg)

	list, err = dlp.GetDelegatorsList()
	assert.Nil(t, list)
	assert.True(t, errors.Is(err, epochStart.ErrExecutingSystemScCode))
}

func TestDelegatedListProc_GetDelegatorsListShouldWork(t *testing.T) {
	t.Parallel()

	delegators := [][]byte{[]byte("delegator1"), []byte("delegator2")}

	arg := createMockArgs()
	arg.PublicKeyConverter = mock.NewPubkeyConverterMock(10)
	delegationSc := [][]byte{[]byte("delegationSc1"), []byte("delegationSc2")}
	arg.QueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
			switch query.FuncName {
			case "getAllContractAddresses":
				return &vmcommon.VMOutput{
					ReturnData: delegationSc,
				}, nil
			case "getUserActiveStake":
				for index, delegator := range delegators {
					if bytes.Equal(delegator, query.Arguments[0]) {
						value := big.NewInt(int64(index + 1))
						return &vmcommon.VMOutput{
							ReturnData: [][]byte{value.Bytes()},
						}, nil
					}
				}
			}

			return nil, fmt.Errorf("not an expected call")
		},
	}
	arg.BlockChain = &mock.BlockChainMock{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.MetaBlock{}
		},
	}
	arg.Accounts.AccountsAdapter = &testscommon.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
			return createDelegationScAccount(addressContainer, delegators, addressContainer), nil
		},
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	dlp, _ := NewDelegatedListProcessor(arg)

	delegatorsValues, err := dlp.GetDelegatorsList()
	require.Nil(t, err)
	require.Equal(t, 2, len(delegatorsValues))
	expectedDelegator1 := api.Delegator{
		DelegatorAddress: arg.PublicKeyConverter.Encode(delegators[0]),
		DelegatedTo: []*api.DelegatedValue{
			{
				DelegationScAddress: arg.PublicKeyConverter.Encode(delegationSc[0]),
				Value:               "1",
			},
			{
				DelegationScAddress: arg.PublicKeyConverter.Encode(delegationSc[1]),
				Value:               "1",
			},
		},
		Total:         "2",
		TotalAsBigInt: big.NewInt(2),
	}

	expectedDelegator2 := api.Delegator{
		DelegatorAddress: arg.PublicKeyConverter.Encode(delegators[1]),
		DelegatedTo: []*api.DelegatedValue{
			{
				DelegationScAddress: arg.PublicKeyConverter.Encode(delegationSc[0]),
				Value:               "2",
			},
			{
				DelegationScAddress: arg.PublicKeyConverter.Encode(delegationSc[1]),
				Value:               "2",
			},
		},
		Total:         "4",
		TotalAsBigInt: big.NewInt(4),
	}

	assert.Equal(t, []*api.Delegator{&expectedDelegator1, &expectedDelegator2}, delegatorsValues)
}

func createDelegationScAccount(address []byte, leaves [][]byte, rootHash []byte) state.UserAccountHandler {
	acc, _ := state.NewUserAccount(address)
	acc.SetDataTrie(&testscommon.TrieStub{
		RootCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		GetAllLeavesOnChannelCalled: func(hash []byte) (chan core.KeyValueHolder, error) {
			ch := make(chan core.KeyValueHolder)

			go func() {
				for _, leafBuff := range leaves {
					leaf := keyValStorage.NewKeyValStorage(leafBuff, nil)
					ch <- leaf
				}

				close(ch)
			}()

			return ch, nil
		},
	})

	return acc
}
