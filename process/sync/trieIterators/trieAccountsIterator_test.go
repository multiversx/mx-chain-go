package trieIterators

import (
	"context"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func getTrieAccountsIteratorArgs() ArgsTrieAccountsIterator {
	return ArgsTrieAccountsIterator{
		Marshaller: &marshallerMock.MarshalizerMock{},
		Accounts:   &stateMock.AccountsStub{},
	}
}

func dummyIterator(_ common.UserAccountHandler) error {
	return nil
}

func TestNewTrieAccountsIterator(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		args.Marshaller = nil

		tai, err := NewTrieAccountsIterator(args)
		require.Nil(t, tai)
		require.Equal(t, errNilMarshaller, err)
	})

	t.Run("nil accounts", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		args.Accounts = nil

		tai, err := NewTrieAccountsIterator(args)
		require.Nil(t, tai)
		require.Equal(t, errNilAccountsAdapter, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		tai, err := NewTrieAccountsIterator(args)
		require.NotNil(t, tai)
		require.NoError(t, err)
	})
}

func TestTrieAccountsIterator_Process(t *testing.T) {
	t.Parallel()

	var expectedErr = errors.New("expected error")

	t.Run("skip processing if no handler", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		args.Accounts = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return nil, errors.New("error that should not be returned")
			},
		}
		tai, _ := NewTrieAccountsIterator(args)
		err := tai.Process()
		require.NoError(t, err)
	})

	t.Run("cannot get root hash", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		args.Accounts = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}
		tai, _ := NewTrieAccountsIterator(args)

		err := tai.Process(dummyIterator)
		require.Equal(t, expectedErr, err)
	})

	t.Run("cannot get all leaves", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		args.Accounts = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
			GetAllLeavesCalled: func(_ *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
				return expectedErr
			},
		}
		tai, _ := NewTrieAccountsIterator(args)

		err := tai.Process(dummyIterator)
		require.Equal(t, expectedErr, err)
	})

	t.Run("cannot get existing account", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		args.Accounts = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
			GetAllLeavesCalled: func(iter *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
				userAcc := &stateMock.AccountWrapMock{
					RootHash: []byte("rootHash"),
				}
				userAccBytes, _ := args.Marshaller.Marshal(userAcc)
				iter.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("addr"), userAccBytes)
				close(iter.LeavesChan)
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return nil, expectedErr
			},
		}
		tai, _ := NewTrieAccountsIterator(args)

		err := tai.Process(dummyIterator)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should ignore non-accounts leaves", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		args.Accounts = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
			GetAllLeavesCalled: func(iter *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
				userAcc := &stateMock.AccountWrapMock{
					RootHash: []byte("rootHash"),
				}
				userAccBytes, _ := args.Marshaller.Marshal(userAcc)
				iter.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("addr"), userAccBytes)
				iter.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("non-addr"), []byte("not an account"))
				close(iter.LeavesChan)
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.AccountWrapMock{}, nil
			},
		}
		tai, _ := NewTrieAccountsIterator(args)

		err := tai.Process(dummyIterator)
		require.NoError(t, err)
	})

	t.Run("should ignore user account without root hash", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		args.Accounts = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
			GetAllLeavesCalled: func(iter *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
				userAcc := &stateMock.AccountWrapMock{
					RootHash: nil,
				}
				userAccBytes, _ := args.Marshaller.Marshal(userAcc)
				iter.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("addr"), userAccBytes)
				close(iter.LeavesChan)
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.AccountWrapMock{}, nil
			},
		}
		tai, _ := NewTrieAccountsIterator(args)

		err := tai.Process(dummyIterator)
		require.NoError(t, err)
	})

	t.Run("should ignore accounts that cannot be casted", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		args.Accounts = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
			GetAllLeavesCalled: func(iter *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
				userAcc := stateMock.AccountWrapMock{
					RootHash: []byte("root"),
				}
				userAccBytes, _ := args.Marshaller.Marshal(userAcc)
				iter.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("addr"), userAccBytes)
				close(iter.LeavesChan)
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return accounts.NewEmptyPeerAccount(), nil
			},
		}
		tai, _ := NewTrieAccountsIterator(args)

		err := tai.Process(dummyIterator)
		require.NoError(t, err)
	})

	t.Run("should work with dummy handler", func(t *testing.T) {
		t.Parallel()

		args := getTrieAccountsIteratorArgs()
		args.Accounts = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
			GetAllLeavesCalled: func(iter *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
				userAcc := &stateMock.AccountWrapMock{
					RootHash: []byte("rootHash"),
				}
				userAccBytes, _ := args.Marshaller.Marshal(userAcc)
				iter.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("addr"), userAccBytes)
				close(iter.LeavesChan)
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.AccountWrapMock{}, nil
			},
		}
		tai, _ := NewTrieAccountsIterator(args)

		err := tai.Process(dummyIterator)
		require.NoError(t, err)
	})

	t.Run("one handler returns error, should error", func(t *testing.T) {
		t.Parallel()

		handler1 := func(account common.UserAccountHandler) error {
			return nil
		}
		handler2 := func(account common.UserAccountHandler) error {
			return expectedErr
		}
		args := getTrieAccountsIteratorArgs()
		args.Accounts = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
			GetAllLeavesCalled: func(iter *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
				userAcc := &stateMock.AccountWrapMock{
					RootHash: []byte("rootHash"),
				}
				userAccBytes, _ := args.Marshaller.Marshal(userAcc)
				iter.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("addr"), userAccBytes)
				close(iter.LeavesChan)
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.AccountWrapMock{RootHash: []byte("rootHash")}, nil
			},
		}
		tai, _ := NewTrieAccountsIterator(args)

		err := tai.Process(handler1, handler2)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work with handlers", func(t *testing.T) {
		t.Parallel()

		handlersReceived := make(map[int]struct{})
		handler1 := func(account common.UserAccountHandler) error {
			handlersReceived[0] = struct{}{}
			return nil
		}
		handler2 := func(account common.UserAccountHandler) error {
			handlersReceived[1] = struct{}{}
			return nil
		}
		args := getTrieAccountsIteratorArgs()
		args.Accounts = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
			GetAllLeavesCalled: func(iter *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
				userAcc := &stateMock.AccountWrapMock{
					RootHash: []byte("rootHash"),
				}
				userAccBytes, _ := args.Marshaller.Marshal(userAcc)
				iter.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("addr"), userAccBytes)
				close(iter.LeavesChan)
				return nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.AccountWrapMock{RootHash: []byte("rootHash")}, nil
			},
		}
		tai, _ := NewTrieAccountsIterator(args)

		err := tai.Process(handler1, handler2)
		require.NoError(t, err)
		require.Len(t, handlersReceived, 2)
	})
}
