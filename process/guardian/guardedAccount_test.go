package guardian

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/guardians"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/testscommon/vmcommonMocks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestNewAccountGuardianChecker(t *testing.T) {
	marshaller := &testscommon.MarshalizerMock{}
	en := &epochNotifier.EpochNotifierStub{}
	ga, err := NewGuardedAccount(marshaller, en, 10)
	require.Nil(t, err)
	require.NotNil(t, ga)

	ga, err = NewGuardedAccount(nil, en, 10)
	require.Equal(t, process.ErrNilMarshalizer, err)
	require.Nil(t, ga)

	ga, err = NewGuardedAccount(marshaller, nil, 10)
	require.Equal(t, process.ErrNilEpochNotifier, err)
	require.Nil(t, ga)

	ga, err = NewGuardedAccount(marshaller, en, 0)
	require.Equal(t, process.ErrInvalidSetGuardianEpochsDelay, err)
	require.Nil(t, ga)
}

func TestGuardedAccount_getActiveGuardian(t *testing.T) {
	ga := createGuardedAccountWithEpoch(9)

	t.Run("no guardians", func(t *testing.T) {
		t.Parallel()

		configuredGuardians := &guardians.Guardians{}
		activeGuardian, err := ga.getActiveGuardian(configuredGuardians)
		require.Nil(t, activeGuardian)
		require.Equal(t, process.ErrAccountHasNoActiveGuardian, err)
	})
	t.Run("one pending guardian", func(t *testing.T) {
		t.Parallel()

		g1 := &guardians.Guardian{Address: []byte("addr1"), ActivationEpoch: 11}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1}}
		activeGuardian, err := ga.getActiveGuardian(configuredGuardians)
		require.Nil(t, activeGuardian)
		require.Equal(t, process.ErrAccountHasNoActiveGuardian, err)
	})
	t.Run("one active guardian", func(t *testing.T) {
		t.Parallel()

		g1 := &guardians.Guardian{Address: []byte("addr1"), ActivationEpoch: 9}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1}}
		activeGuardian, err := ga.getActiveGuardian(configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, g1, activeGuardian)
	})
	t.Run("one active and one pending", func(t *testing.T) {
		t.Parallel()

		g1 := &guardians.Guardian{Address: []byte("addr1"), ActivationEpoch: 9}
		g2 := &guardians.Guardian{Address: []byte("addr2"), ActivationEpoch: 30}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1, g2}}
		activeGuardian, err := ga.getActiveGuardian(configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, g1, activeGuardian)
	})
	t.Run("one active and one too old", func(t *testing.T) {
		t.Parallel()

		g1 := &guardians.Guardian{Address: []byte("addr1"), ActivationEpoch: 8}
		g2 := &guardians.Guardian{Address: []byte("addr2"), ActivationEpoch: 9}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g2, g1}}
		activeGuardian, err := ga.getActiveGuardian(configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, g2, activeGuardian)
	})
}

func TestGuardedAccount_getConfiguredGuardians(t *testing.T) {
	ga := createGuardedAccountWithEpoch(10)

	t.Run("guardians key not found should err", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		acc := &state.UserAccountStub{
			RetrieveValueFromDataTrieTrackerCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}

		configuredGuardians, err := ga.getConfiguredGuardians(acc)
		require.Nil(t, configuredGuardians)
		require.Equal(t, expectedErr, err)
	})
	t.Run("key found but no guardians, should return empty", func(t *testing.T) {
		t.Parallel()

		acc := &state.UserAccountStub{
			RetrieveValueFromDataTrieTrackerCalled: func(key []byte) ([]byte, error) {
				return nil, nil
			},
		}

		configuredGuardians, err := ga.getConfiguredGuardians(acc)
		require.Nil(t, err)
		require.NotNil(t, configuredGuardians)
		require.True(t, len(configuredGuardians.Slice) == 0)
	})
	t.Run("unmarshal guardians error should return error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		ga := createGuardedAccountWithEpoch(10)
		ga.marshaller = &testscommon.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		acc := &state.UserAccountStub{
			RetrieveValueFromDataTrieTrackerCalled: func(key []byte) ([]byte, error) {
				return []byte("wrongly marshalled guardians"), nil
			},
		}

		configuredGuardians, err := ga.getConfiguredGuardians(acc)
		require.Nil(t, configuredGuardians)
		require.Equal(t, expectedErr, err)
	})
	t.Run("unmarshal guardians error should return error", func(t *testing.T) {
		t.Parallel()

		g1 := &guardians.Guardian{Address: []byte("addr1"), ActivationEpoch: 9}
		expectedConfiguredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1}}

		acc := &state.UserAccountStub{
			RetrieveValueFromDataTrieTrackerCalled: func(key []byte) ([]byte, error) {
				return ga.marshaller.Marshal(expectedConfiguredGuardians)
			},
		}

		configuredGuardians, err := ga.getConfiguredGuardians(acc)
		require.Nil(t, err)
		require.Equal(t, expectedConfiguredGuardians, configuredGuardians)
	})
}

func TestGuardedAccount_saveAccountGuardians(t *testing.T) {
	userAccount := &vmcommonMocks.UserAccountStub{
		AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
			return &trie.DataTrieTrackerStub{
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					return nil
				},
			}
		},
	}

	t.Run("marshaling error should return err", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		ga := createGuardedAccountWithEpoch(10)
		ga.marshaller = &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}

		err := ga.saveAccountGuardians(userAccount, nil)
		require.Equal(t, expectedErr, err)
	})
	t.Run("", func(t *testing.T) {
		t.Parallel()

		SaveKeyValueCalled := false
		userAccount.AccountDataHandlerCalled = func() vmcommon.AccountDataHandler {
			return &trie.DataTrieTrackerStub{
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					SaveKeyValueCalled = true
					return nil
				},
			}
		}

		ga := createGuardedAccountWithEpoch(10)
		err := ga.saveAccountGuardians(userAccount, nil)
		require.Nil(t, err)
		require.True(t, SaveKeyValueCalled)
	})
}

func TestGuardedAccount_updateGuardians(t *testing.T) {
	ga := createGuardedAccountWithEpoch(10)
	newGuardian:= &guardians.Guardian{
		Address:         []byte("new guardian address"),
		ActivationEpoch: 20,
	}

	t.Run("update empty guardian list with new guardian", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{}}
		expectedGuardians := append(configuredGuardians.Slice, newGuardian)
		updatedGuardians, err := ga.updateGuardians(newGuardian, configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, expectedGuardians, updatedGuardians.Slice)
	})
	t.Run("updating when there is an existing pending guardian and no active should error", func(t *testing.T) {
		existingGuardian:= &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: 11,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingGuardian}}

		updatedGuardians, err := ga.updateGuardians(newGuardian, configuredGuardians)
		require.Nil(t, updatedGuardians)
		require.True(t, errors.Is(err, process.ErrAccountHasNoActiveGuardian))
	})
	t.Run("updating the existing same active guardian should leave the active guardian unchanged", func(t *testing.T) {
		existingGuardian:= &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: 9,
		}

		newGuardian:= newGuardian
		newGuardian.Address = existingGuardian.Address
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingGuardian}}

		updatedGuardians, err := ga.updateGuardians(newGuardian, configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, configuredGuardians, updatedGuardians)
	})
	t.Run("updating the existing same active guardian, when there is also a pending guardian configured, should clean up pending and leave active unchanged", func(t *testing.T) {
		existingActiveGuardian:= &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: 9,
		}
		existingPendingGuardian := &guardians.Guardian{
			Address: []byte("pending guardian address"),
			ActivationEpoch: 13,
		}

		newGuardian:= newGuardian
		newGuardian.Address = existingPendingGuardian.Address
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingActiveGuardian, existingPendingGuardian}}
		expectedUpdatedGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingActiveGuardian, newGuardian}}

		updatedGuardians, err := ga.updateGuardians(newGuardian, configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, expectedUpdatedGuardians, updatedGuardians)
	})
	// todo: if todo in guarded account is implemented, this will need to be changed
	t.Run("updating the existing same pending guardian while there is an active one should leave the active guardian unchanged but update the pending", func(t *testing.T) {
		existingGuardian:= &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: 9,
		}

		newGuardian:= newGuardian
		newGuardian.Address = existingGuardian.Address
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingGuardian}}

		updatedGuardians, err := ga.updateGuardians(newGuardian, configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, configuredGuardians, updatedGuardians)
	})
}

func TestGuardedAccount_setAccountGuardian(t *testing.T) {

}

func TestGuardedAccount_instantSetGuardian(t *testing.T) {

}

func TestGuardedAccount_EpochConfirmed(t *testing.T) {

}

func TestGuardedAccount_GetActiveGuardian(t *testing.T) {

}

func TestGuardedAccount_SetGuardian(t *testing.T) {

}

func TestAccountGuardianChecker_IsInterfaceNil(t *testing.T) {

}

func createGuardedAccountWithEpoch(epoch uint32) *guardedAccount {
	marshaller := &testscommon.MarshalizerMock{}
	en := &epochNotifier.EpochNotifierStub{
		RegisterNotifyHandlerCalled: func(handler vmcommon.EpochSubscriberHandler) {
			handler.EpochConfirmed(epoch, 0)
		},
	}

	ga, _ := NewGuardedAccount(marshaller, en, 10)
	return ga
}
