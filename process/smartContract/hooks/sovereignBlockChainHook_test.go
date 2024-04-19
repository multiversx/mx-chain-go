package hooks_test

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/stretchr/testify/require"
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

func TestSovereignBlockChainHook_IsPayable(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	bh, _ := hooks.NewBlockChainHookImpl(args)
	sbh, _ := hooks.NewSovereignBlockChainHook(bh)

	t.Run("receiver is sys acc address, should not be payable", func(t *testing.T) {
		sender := []byte("addr1")

		payable, err := sbh.IsPayable(sender, core.SystemAccountAddress)
		require.Nil(t, err)
		require.False(t, payable)
	})

	t.Run("receiver is user address, should be payable", func(t *testing.T) {
		sender := []byte("addr1")
		receiver := []byte("addr2")

		payable, err := sbh.IsPayable(sender, receiver)
		require.Nil(t, err)
		require.True(t, payable)
	})

	t.Run("sender is sc, receiver is esdt sc address, should be cross payable", func(t *testing.T) {
		sender := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6, 255, 255}
		receiver := vm.ESDTSCAddress

		payable, err := sbh.IsPayable(sender, receiver)
		require.Nil(t, err)
		require.True(t, payable)
	})

	t.Run("sender is user, receiver is sc, should be payable by code meta data", func(t *testing.T) {
		receiver := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6, 255, 255}
		wasGetAccCalled := false
		argsBlockHook := createMockBlockChainHookArgs()
		argsBlockHook.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
				require.Equal(t, receiver, address)

				acc := &stateMock.AccountWrapMock{}
				acc.SetCodeMetadata([]byte{0, vmcommon.MetadataPayable})
				wasGetAccCalled = true
				return acc, nil
			},
		}

		chainHook, _ := hooks.NewBlockChainHookImpl(argsBlockHook)
		sovChainHook, _ := hooks.NewSovereignBlockChainHook(chainHook)

		sender := []byte("sender")
		payable, err := sovChainHook.IsPayable(sender, receiver)
		require.Nil(t, err)
		require.True(t, payable)
		require.True(t, wasGetAccCalled)
	})
}

func TestSovereignBlockChainHook_isCrossShardForPayableCheck(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	bh, _ := hooks.NewBlockChainHookImpl(args)
	sbh, _ := hooks.NewSovereignBlockChainHook(bh)

	userAddr := make([]byte, 32)

	require.True(t, sbh.IsCrossShardForPayableCheck(userAddr, vm.ESDTSCAddress))
	require.True(t, sbh.IsCrossShardForPayableCheck(vm.StakingSCAddress, userAddr))

	require.False(t, sbh.IsCrossShardForPayableCheck(userAddr, userAddr))
	require.False(t, sbh.IsCrossShardForPayableCheck(vm.StakingSCAddress, vm.ESDTSCAddress))
}
