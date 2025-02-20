package hooks_test

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/vm"
)

func TestNewSovereignBlockChainHook(t *testing.T) {
	t.Parallel()

	t.Run("nil blockchain hook, should return error", func(t *testing.T) {
		sbh, err := hooks.NewSovereignBlockChainHook(nil)
		require.Nil(t, sbh)
		require.Equal(t, hooks.ErrNilBlockChainHook, err)
	})

	t.Run("should work", func(t *testing.T) {
		args := createMockBlockChainHookArgs()
		bh, _ := hooks.NewBlockChainHookImpl(args)

		sbh, err := hooks.NewSovereignBlockChainHook(bh)
		require.False(t, check.IfNil(sbh))
		require.Nil(t, err)
	})
}

func TestSovereignBlockChainHook_ProcessBuiltInFunction(t *testing.T) {
	t.Parallel()

	funcName := "func"
	builtInFunctionsContainer := vmcommonBuiltInFunctions.NewBuiltInFunctionContainer()
	_ = builtInFunctionsContainer.Add(funcName, &mock.BuiltInFunctionStub{})

	t.Run("normal flow, should work", func(t *testing.T) {
		t.Parallel()

		addrSender := []byte("addr sender")
		addrReceiver := []byte("addr receiver")

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer

		getSenderAccountCalled := &atomic.Flag{}
		getReceiverAccountCalled := &atomic.Flag{}
		ctSaveAccount := &atomic.Counter{}
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrSender, addressContainer)
				getSenderAccountCalled.SetValue(true)
				return stateMock.NewAccountWrapMock(addrSender), nil
			},

			LoadAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrReceiver, addressContainer)
				getReceiverAccountCalled.SetValue(true)
				return stateMock.NewAccountWrapMock(addrReceiver), nil
			},

			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				require.True(t, bytes.Equal(addrSender, account.AddressBytes()) || bytes.Equal(addrReceiver, account.AddressBytes()))
				ctSaveAccount.Increment()
				return nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		sbh, _ := hooks.NewSovereignBlockChainHook(bh)

		input := createContractCallInput(funcName, addrSender, addrReceiver)
		_, err := sbh.ProcessBuiltInFunction(input)
		require.Nil(t, err)

		require.True(t, getSenderAccountCalled.IsSet())
		require.True(t, getReceiverAccountCalled.IsSet())
		require.Equal(t, int64(2), ctSaveAccount.Get())
	})

	t.Run("incoming sovereign scr, sender is ESDTSCAddress, should not load sender", func(t *testing.T) {
		t.Parallel()

		addrSender := core.ESDTSCAddress
		addrReceiver := []byte("addr receiver")

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer

		getSenderAccountCalled := &atomic.Flag{}
		getReceiverAccountCalled := &atomic.Flag{}
		ctSaveAccount := &atomic.Counter{}
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				getSenderAccountCalled.SetValue(true)
				return nil, nil
			},

			LoadAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrReceiver, addressContainer)
				getReceiverAccountCalled.SetValue(true)
				return stateMock.NewAccountWrapMock(addrReceiver), nil
			},

			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				require.Equal(t, addrReceiver, account.AddressBytes())
				ctSaveAccount.Increment()
				return nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		sbh, _ := hooks.NewSovereignBlockChainHook(bh)

		input := createContractCallInput(funcName, addrSender, addrReceiver)
		_, err := sbh.ProcessBuiltInFunction(input)
		require.Nil(t, err)

		require.False(t, getSenderAccountCalled.IsSet())
		require.True(t, getReceiverAccountCalled.IsSet())
		require.Equal(t, int64(1), ctSaveAccount.Get())
	})
}

func TestSovereignBlockChainHook_GetStorageData(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()

	ctProcessTrieReads := 0
	args.Counter = &testscommon.BlockChainHookCounterStub{
		ProcessCrtNumberOfTrieReadsCounterCalled: func() error {
			ctProcessTrieReads++
			return nil
		},
	}

	getAccCt := 0
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			getAccCt++
			return &stateMock.UserAccountStub{
				AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
					return &trie.DataTrieTrackerStub{}
				},
			}, nil
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	sbh, _ := hooks.NewSovereignBlockChainHook(bh)

	for i := 1; i <= 10; i++ {
		_, _, err := sbh.GetStorageData([]byte("addr"), []byte{0x1})
		require.Nil(t, err)
		require.Equal(t, i, getAccCt)
		require.Equal(t, i, ctProcessTrieReads)
	}

	_, _, err := sbh.GetStorageData(vm.StakingSCAddress, []byte{0x1})
	require.Nil(t, err)
	_, _, err = sbh.GetStorageData(vm.ValidatorSCAddress, []byte{0x1})
	require.Nil(t, err)

	require.Equal(t, 12, getAccCt)
	require.Equal(t, 10, ctProcessTrieReads)
}
