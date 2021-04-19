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
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDirectStakedListProcessor(t *testing.T) {
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
			_, err := NewDirectStakedListProcessor(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
		})
	}

	dslp, _ := NewDirectStakedListProcessor(createMockArgs())
	assert.False(t, check.IfNil(dslp))
}

func TestDirectStakedListProc_GetDelegatorsListShouldWork(t *testing.T) {
	t.Parallel()

	validators := [][]byte{[]byte("validator1"), []byte("validator2")}

	arg := createMockArgs()
	arg.PublicKeyConverter = mock.NewPubkeyConverterMock(10)
	arg.QueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
			switch query.FuncName {
			case "getTotalStakedTopUpStakedBlsKeys":
				for index, validator := range validators {
					if bytes.Equal(validator, query.Arguments[0]) {
						topUpValue := big.NewInt(int64(index + 1))
						totalStakedValue := big.NewInt(int64(index+1) * 10)

						return &vmcommon.VMOutput{
							ReturnData: [][]byte{topUpValue.Bytes(), totalStakedValue.Bytes(), make([]byte, 0)},
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
	arg.Accounts.AccountsAdapter = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (state.AccountHandler, error) {
			return createValidatorScAccount(addressContainer, validators, addressContainer), nil
		},
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	dslp, _ := NewDirectStakedListProcessor(arg)

	directStakedList, err := dslp.GetDirectStakedList()
	require.Nil(t, err)
	require.Equal(t, 2, len(directStakedList))

	expectedDirectStake1 := api.DirectStakedValue{
		Address: arg.PublicKeyConverter.Encode(validators[0]),
		Staked:  "9",
		TopUp:   "1",
		Total:   "10",
	}
	expectedDirectStake2 := api.DirectStakedValue{
		Address: arg.PublicKeyConverter.Encode(validators[1]),
		Staked:  "18",
		TopUp:   "2",
		Total:   "20",
	}

	assert.Equal(t, []*api.DirectStakedValue{&expectedDirectStake1, &expectedDirectStake2}, directStakedList)
}

//
func createValidatorScAccount(address []byte, leaves [][]byte, rootHash []byte) state.UserAccountHandler {
	acc, _ := state.NewUserAccount(address)
	acc.SetDataTrie(&mock.TrieStub{
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
