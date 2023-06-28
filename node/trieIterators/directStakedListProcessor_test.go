package trieIterators

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dslp, err := NewDirectStakedListProcessor(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
			require.Nil(t, dslp)
		})
	}

	dslp, err := NewDirectStakedListProcessor(createMockArgs())
	require.NotNil(t, dslp)
	require.Nil(t, err)
}

func TestDirectStakedListProc_GetDelegatorsListContextShouldTimeout(t *testing.T) {
	t.Parallel()

	validators := [][]byte{[]byte("validator1"), []byte("validator2")}

	arg := createMockArgs()
	arg.PublicKeyConverter = testscommon.NewPubkeyConverterMock(10)
	arg.QueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
			return nil, fmt.Errorf("not an expected call")
		},
	}
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return createScAccount(addressContainer, validators, addressContainer, time.Second), nil
		},
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	dslp, _ := NewDirectStakedListProcessor(arg)

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	directStakedList, err := dslp.GetDirectStakedList(ctxWithTimeout)
	require.Equal(t, ErrTrieOperationsTimeout, err)
	require.Nil(t, directStakedList)
}

func TestDirectStakedListProc_GetDelegatorsListShouldWork(t *testing.T) {
	t.Parallel()

	validators := [][]byte{[]byte("validator1"), []byte("validator2")}

	arg := createMockArgs()
	arg.PublicKeyConverter = testscommon.NewPubkeyConverterMock(10)
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
	arg.Accounts.AccountsAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return createScAccount(addressContainer, validators, addressContainer, 0), nil
		},
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	dslp, _ := NewDirectStakedListProcessor(arg)

	directStakedList, err := dslp.GetDirectStakedList(context.Background())
	require.Nil(t, err)
	require.Equal(t, 2, len(directStakedList))
	encodedValidator0PubKey, err := arg.PublicKeyConverter.Encode(validators[0])
	require.Nil(t, err)
	encodedValidator1PubKey, err := arg.PublicKeyConverter.Encode(validators[1])
	require.Nil(t, err)

	expectedDirectStake1 := api.DirectStakedValue{
		Address:    encodedValidator0PubKey,
		BaseStaked: "9",
		TopUp:      "1",
		Total:      "10",
	}
	expectedDirectStake2 := api.DirectStakedValue{
		Address:    encodedValidator1PubKey,
		BaseStaked: "18",
		TopUp:      "2",
		Total:      "20",
	}

	assert.Equal(t, []*api.DirectStakedValue{&expectedDirectStake1, &expectedDirectStake2}, directStakedList)
}

func createScAccount(address []byte, leaves [][]byte, rootHash []byte, timeSleep time.Duration) state.UserAccountHandler {
	argsAccCreation := state.ArgsAccountCreation{
		Hasher:              &hashingMocks.HasherMock{},
		Marshaller:          &marshallerMock.MarshalizerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	acc, _ := state.NewUserAccount(address, argsAccCreation)
	acc.SetDataTrie(&trieMock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
			go func() {
				time.Sleep(timeSleep)
				for _, leafBuff := range leaves {
					leaf := keyValStorage.NewKeyValStorage(leafBuff, nil)
					leavesChannels.LeavesChan <- leaf
				}

				close(leavesChannels.LeavesChan)
				leavesChannels.ErrChan.Close()
			}()

			return nil
		},
	})

	return acc
}

func TestDirectStakedListProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var dslp *directStakedListProcessor
	require.True(t, dslp.IsInterfaceNil())

	dslp, _ = NewDirectStakedListProcessor(createMockArgs())
	require.False(t, dslp.IsInterfaceNil())
}
