package guardian

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/guardians"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMocks "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/testscommon/vmcommonMocks"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewGuardedAccount(t *testing.T) {
	marshaller := &marshallerMock.MarshalizerMock{}
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
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1, g2}}
		activeGuardian, err := ga.getActiveGuardian(configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, g2, activeGuardian)
	})
	t.Run("one active and one too old, saved in reverse order", func(t *testing.T) {
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

	t.Run("guardians key not found should return empty", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, expectedErr
			},
		}

		configuredGuardians, err := ga.getConfiguredGuardians(acc)
		require.Nil(t, err)
		require.NotNil(t, configuredGuardians)
		require.True(t, len(configuredGuardians.Slice) == 0)
	})
	t.Run("key found but no guardians, should return empty", func(t *testing.T) {
		t.Parallel()

		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, nil
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
		ga.marshaller = &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				return []byte("wrongly marshalled guardians"), 0, nil
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

		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(expectedConfiguredGuardians)
				return val, 0, err
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
		ga.marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}

		err := ga.saveAccountGuardians(userAccount, nil)
		require.Equal(t, expectedErr, err)
	})
	t.Run("save account guardians OK", func(t *testing.T) {
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
	newGuardian := &guardians.Guardian{
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
		existingGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: 11,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingGuardian}}

		updatedGuardians, err := ga.updateGuardians(newGuardian, configuredGuardians)
		require.Nil(t, updatedGuardians)
		require.True(t, errors.Is(err, process.ErrAccountHasNoActiveGuardian))
	})
	t.Run("updating the existing same active guardian should leave the active guardian unchanged", func(t *testing.T) {
		existingGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: 9,
		}

		newGuardian := newGuardian
		newGuardian.Address = existingGuardian.Address
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingGuardian}}

		updatedGuardians, err := ga.updateGuardians(newGuardian, configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, configuredGuardians, updatedGuardians)
	})
	t.Run("updating the existing same active guardian, when there is also a pending guardian configured, should clean up pending and leave active unchanged", func(t *testing.T) {
		existingActiveGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: 9,
		}
		existingPendingGuardian := &guardians.Guardian{
			Address:         []byte("pending guardian address"),
			ActivationEpoch: 13,
		}

		newGuardian := newGuardian
		newGuardian.Address = existingPendingGuardian.Address
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingActiveGuardian, existingPendingGuardian}}
		expectedUpdatedGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingActiveGuardian, newGuardian}}

		updatedGuardians, err := ga.updateGuardians(newGuardian, configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, expectedUpdatedGuardians, updatedGuardians)
	})
	t.Run("updating the existing same pending guardian while there is an active one should leave the active guardian unchanged but update the pending", func(t *testing.T) {
		existingGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: 9,
		}

		newGuardian := newGuardian
		newGuardian.Address = existingGuardian.Address
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingGuardian}}

		updatedGuardians, err := ga.updateGuardians(newGuardian, configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, configuredGuardians, updatedGuardians)
	})
}

func TestGuardedAccount_setAccountGuardian(t *testing.T) {
	ga := createGuardedAccountWithEpoch(10)
	newGuardian := &guardians.Guardian{
		Address:         []byte("new guardian address"),
		ActivationEpoch: 20,
	}

	t.Run("if updateGuardians returns err, the err should be propagated", func(t *testing.T) {
		existingGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: 11,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{existingGuardian}}
		ua := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		err := ga.setAccountGuardian(ua, newGuardian)
		require.True(t, errors.Is(err, process.ErrAccountHasNoActiveGuardian))
	})
	t.Run("setGuardian same guardian ok, changing existing config", func(t *testing.T) {
		existingGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: 9,
		}
		newGuardian := newGuardian
		newGuardian.Address = existingGuardian.Address
		expectedValue := []byte(nil)
		ua := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				expectedValue, _ = ga.marshaller.Marshal(&guardians.Guardians{Slice: []*guardians.Guardian{existingGuardian, newGuardian}})
				return expectedValue, 0, nil
			},
			AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
				return &trie.DataTrieTrackerStub{
					SaveKeyValueCalled: func(key []byte, value []byte) error {
						require.Equal(t, guardianKey, key)
						require.Equal(t, expectedValue, value)
						return nil
					},
				}
			},
		}

		err := ga.setAccountGuardian(ua, newGuardian)
		require.Nil(t, err)
	})
}

func TestGuardedAccount_instantSetGuardian(t *testing.T) {
	currentEpoch := uint32(10)
	ga := createGuardedAccountWithEpoch(currentEpoch)
	newGuardian := &guardians.Guardian{
		Address:         []byte("new guardian address"),
		ActivationEpoch: 20,
	}
	txGuardianAddress := []byte("guardian address")
	guardianServiceUID := []byte("testID")

	t.Run("getActiveGuardianErr with err (no active guardian) should error", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{}}

		ua := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		err := ga.instantSetGuardian(ua, newGuardian.Address, txGuardianAddress, guardianServiceUID)
		require.Equal(t, process.ErrAccountHasNoActiveGuardian, err)
	})
	t.Run("tx signed by different than active guardian should err", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("active guardian address"),
			ActivationEpoch: 1,
		}

		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian}}

		ua := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		err := ga.instantSetGuardian(ua, newGuardian.Address, txGuardianAddress, guardianServiceUID)
		require.Equal(t, process.ErrTransactionAndAccountGuardianMismatch, err)
	})
	t.Run("immediately set the guardian if setGuardian tx is signed by active guardian", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         txGuardianAddress,
			ActivationEpoch: 1,
			ServiceUID:      guardianServiceUID,
		}
		newGuardian := &guardians.Guardian{
			Address:         []byte("new guardian address"),
			ActivationEpoch: currentEpoch,
			ServiceUID:      []byte("testServiceID2"),
		}

		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian}}
		expectedValue, _ := ga.marshaller.Marshal(&guardians.Guardians{Slice: []*guardians.Guardian{newGuardian}})

		ua := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
			AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
				return &trie.DataTrieTrackerStub{
					SaveKeyValueCalled: func(key []byte, value []byte) error {
						require.Equal(t, guardianKey, key)
						require.Equal(t, expectedValue, value)
						return nil
					},
				}
			}}

		err := ga.instantSetGuardian(ua, newGuardian.Address, txGuardianAddress, newGuardian.ServiceUID)
		require.Nil(t, err)
	})
}

func TestGuardedAccount_GetActiveGuardian(t *testing.T) {
	currentEpoch := uint32(10)
	ga := createGuardedAccountWithEpoch(currentEpoch)

	t.Run("wrong account type should err", func(t *testing.T) {
		var uah *vmcommonMocks.UserAccountStub
		activeGuardian, err := ga.GetActiveGuardian(uah)
		require.Nil(t, activeGuardian)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("getConfiguredGuardians with err should err - no active", func(t *testing.T) {
		dataTrieErr := errors.New("expected error")
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, dataTrieErr
			},
		}
		activeGuardian, err := ga.GetActiveGuardian(uah)
		require.Nil(t, activeGuardian)
		require.Equal(t, process.ErrAccountHasNoGuardianSet, err)
	})
	t.Run("no guardian should return err", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		activeGuardian, err := ga.GetActiveGuardian(uah)
		require.Nil(t, activeGuardian)
		require.Equal(t, process.ErrAccountHasNoGuardianSet, err)
	})
	t.Run("one pending guardian should return err", func(t *testing.T) {
		pendingGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch + 1,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{pendingGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		activeGuardian, err := ga.GetActiveGuardian(uah)
		require.Nil(t, activeGuardian)
		require.Equal(t, process.ErrAccountHasNoActiveGuardian, err)
	})
	t.Run("one active guardian should return the active", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		guardian, err := ga.GetActiveGuardian(uah)
		require.Equal(t, activeGuardian.Address, guardian)
		require.Nil(t, err)
	})
	t.Run("one active guardian and one pending new guardian", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		pendingGuardian := &guardians.Guardian{
			Address:         []byte("pending guardian address"),
			ActivationEpoch: currentEpoch + 1,
		}

		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian, pendingGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		guardian, err := ga.GetActiveGuardian(uah)
		require.Equal(t, activeGuardian.Address, guardian)
		require.Nil(t, err)
	})
	t.Run("one active guardian and one disabled (old) guardian", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		oldGuardian := &guardians.Guardian{
			Address:         []byte("old guardian address"),
			ActivationEpoch: currentEpoch - 5,
		}

		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian, oldGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		guardian, err := ga.GetActiveGuardian(uah)
		require.Equal(t, activeGuardian.Address, guardian)
		require.Nil(t, err)
	})
}

func TestGuardedAccount_getPendingGuardian(t *testing.T) {
	currentEpoch := uint32(10)
	ga := createGuardedAccountWithEpoch(currentEpoch)

	t.Run("nil guardians/empty guardians should err", func(t *testing.T) {
		pendingGuardian, err := ga.getPendingGuardian(nil)
		require.Nil(t, pendingGuardian)
		require.Equal(t, process.ErrAccountHasNoPendingGuardian, err)

		configuredGuardians := &guardians.Guardians{}
		pendingGuardian, err = ga.getPendingGuardian(configuredGuardians)
		require.Nil(t, pendingGuardian)
		require.Equal(t, process.ErrAccountHasNoPendingGuardian, err)
	})
	t.Run("one pending guardian should return it", func(t *testing.T) {
		pendingGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch + 1,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{pendingGuardian}}
		pGuardian, err := ga.getPendingGuardian(configuredGuardians)
		require.Nil(t, err)
		require.Equal(t, pendingGuardian, pGuardian)
	})
	t.Run("one active guardian should err", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian}}
		guardian, err := ga.getPendingGuardian(configuredGuardians)
		require.Nil(t, guardian)
		require.Equal(t, process.ErrAccountHasNoPendingGuardian, err)
	})
	t.Run("one active guardian and one pending new guardian", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		pendingGuardian := &guardians.Guardian{
			Address:         []byte("pending guardian address"),
			ActivationEpoch: currentEpoch + 1,
		}

		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian, pendingGuardian}}
		guardian, err := ga.getPendingGuardian(configuredGuardians)
		require.Equal(t, pendingGuardian, guardian)
		require.Nil(t, err)
	})
	t.Run("one active guardian and one disabled (old) guardian should err", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		oldGuardian := &guardians.Guardian{
			Address:         []byte("old guardian address"),
			ActivationEpoch: currentEpoch - 5,
		}

		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian, oldGuardian}}
		guardian, err := ga.getPendingGuardian(configuredGuardians)
		require.Nil(t, guardian)
		require.Equal(t, process.ErrAccountHasNoPendingGuardian, err)
	})
}

func TestGuardedAccount_SetGuardian(t *testing.T) {
	currentEpoch := uint32(10)
	ga := createGuardedAccountWithEpoch(currentEpoch)
	guardianServiceUID := []byte("testID")
	initialServiceUID := []byte("test2ID")
	g1 := &guardians.Guardian{
		Address:         []byte("guardian address 1"),
		ActivationEpoch: currentEpoch - 2,
		ServiceUID:      initialServiceUID,
	}
	g2 := &guardians.Guardian{
		Address:         []byte("guardian address 2"),
		ActivationEpoch: currentEpoch - 1,
		ServiceUID:      initialServiceUID,
	}
	newGuardianAddress := []byte("new guardian address")

	t.Run("invalid user account handler should err", func(t *testing.T) {
		err := ga.SetGuardian(nil, newGuardianAddress, g1.Address, guardianServiceUID)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("transaction signed by current active guardian but instantSetGuardian returns error", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}
		err := ga.SetGuardian(uah, newGuardianAddress, g2.Address, guardianServiceUID)
		require.Equal(t, process.ErrTransactionAndAccountGuardianMismatch, err)
	})
	t.Run("instantly set guardian if tx signed by current active guardian", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1}}
		newGuardian := &guardians.Guardian{
			Address:         newGuardianAddress,
			ActivationEpoch: currentEpoch,
			ServiceUID:      guardianServiceUID,
		}
		expectedNewGuardians, _ := ga.marshaller.Marshal(&guardians.Guardians{Slice: []*guardians.Guardian{newGuardian}})

		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
			AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
				return &trie.DataTrieTrackerStub{
					SaveKeyValueCalled: func(key []byte, value []byte) error {
						require.Equal(t, guardianKey, key)
						require.Equal(t, expectedNewGuardians, value)
						return nil
					},
				}
			},
		}
		err := ga.SetGuardian(uah, newGuardianAddress, g1.Address, guardianServiceUID)
		require.Nil(t, err)
	})
	t.Run("nil guardian serviceUID should err", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1}}
		saveKeyValueCalled := false
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
			AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
				return &trie.DataTrieTrackerStub{
					SaveKeyValueCalled: func(_ []byte, _ []byte) error {
						saveKeyValueCalled = true
						return nil
					},
				}
			},
		}
		err := ga.SetGuardian(uah, newGuardianAddress, g1.Address, nil)
		require.False(t, saveKeyValueCalled)
		require.Equal(t, process.ErrNilGuardianServiceUID, err)
	})
	t.Run("tx not signed by active guardian sets guardian with delay", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1}}
		newGuardian := &guardians.Guardian{
			Address:         newGuardianAddress,
			ActivationEpoch: currentEpoch + ga.guardianActivationEpochsDelay,
			ServiceUID:      guardianServiceUID,
		}
		expectedNewGuardians, _ := ga.marshaller.Marshal(&guardians.Guardians{Slice: []*guardians.Guardian{g1, newGuardian}})

		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
			AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
				return &trie.DataTrieTrackerStub{
					SaveKeyValueCalled: func(key []byte, value []byte) error {
						require.Equal(t, guardianKey, key)
						require.Equal(t, expectedNewGuardians, value)
						return nil
					},
				}
			},
		}
		err := ga.SetGuardian(uah, newGuardianAddress, nil, guardianServiceUID)
		require.Nil(t, err)
	})
}

func TestGuardedAccount_HasActiveGuardian(t *testing.T) {
	t.Parallel()

	currentEpoch := uint32(10)
	ga := createGuardedAccountWithEpoch(currentEpoch)

	t.Run("nil account type should return false", func(t *testing.T) {
		var uah *stateMocks.UserAccountStub
		require.False(t, ga.HasActiveGuardian(uah))
	})
	t.Run("getConfiguredGuardians with err should return false", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, expectedErr
			},
		}
		require.False(t, ga.HasActiveGuardian(uah))
	})
	t.Run("no guardian should return false", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		require.False(t, ga.HasActiveGuardian(uah))
	})
	t.Run("one pending guardian should return false", func(t *testing.T) {
		pendingGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch + 1,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{pendingGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		require.False(t, ga.HasActiveGuardian(uah))
	})
	t.Run("one active guardian should return true", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		require.True(t, ga.HasActiveGuardian(uah))
	})
	t.Run("one active guardian and one pending new guardian should return true", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		pendingGuardian := &guardians.Guardian{
			Address:         []byte("pending guardian address"),
			ActivationEpoch: currentEpoch + 1,
		}

		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian, pendingGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		require.True(t, ga.HasActiveGuardian(uah))
	})
	t.Run("one active guardian and one disabled (old) guardian should return true", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		oldGuardian := &guardians.Guardian{
			Address:         []byte("old guardian address"),
			ActivationEpoch: currentEpoch - 5,
		}

		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian, oldGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		require.True(t, ga.HasActiveGuardian(uah))
	})
}

func TestGuardedAccount_HasPendingGuardian(t *testing.T) {
	t.Parallel()

	currentEpoch := uint32(10)
	ga := createGuardedAccountWithEpoch(currentEpoch)

	t.Run("nil account type should return false", func(t *testing.T) {
		var uah *stateMocks.UserAccountStub
		require.False(t, ga.HasPendingGuardian(uah))
	})
	t.Run("getConfiguredGuardians with err should return false", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, expectedErr
			},
		}
		require.False(t, ga.HasPendingGuardian(uah))
	})
	t.Run("no guardian should return false", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		require.False(t, ga.HasPendingGuardian(uah))
	})
	t.Run("one pending guardian should return true", func(t *testing.T) {
		pendingGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch + 1,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{pendingGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		require.True(t, ga.HasPendingGuardian(uah))
	})
	t.Run("one active guardian should return false", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		require.False(t, ga.HasPendingGuardian(uah))
	})
	t.Run("one active guardian and one pending new guardian should return true", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		pendingGuardian := &guardians.Guardian{
			Address:         []byte("pending guardian address"),
			ActivationEpoch: currentEpoch + 1,
		}

		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian, pendingGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		require.True(t, ga.HasPendingGuardian(uah))
	})
	t.Run("one active guardian and one disabled (old) guardian should return false", func(t *testing.T) {
		activeGuardian := &guardians.Guardian{
			Address:         []byte("guardian address"),
			ActivationEpoch: currentEpoch - 1,
		}
		oldGuardian := &guardians.Guardian{
			Address:         []byte("old guardian address"),
			ActivationEpoch: currentEpoch - 5,
		}

		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{activeGuardian, oldGuardian}}
		uah := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}

		require.False(t, ga.HasPendingGuardian(uah))
	})
}

func TestGuardedAccount_CleanOtherThanActive(t *testing.T) {
	t.Parallel()

	currentEpoch := uint32(10)
	g0 := &guardians.Guardian{
		Address:         []byte("old guardian"),
		ActivationEpoch: currentEpoch - 4,
	}
	g1 := &guardians.Guardian{
		Address:         []byte("active guardian"),
		ActivationEpoch: currentEpoch - 2,
	}
	g2 := &guardians.Guardian{
		Address:         []byte("pending guardian"),
		ActivationEpoch: currentEpoch + 2,
	}

	t.Run("no configured guardians does not change the guardians", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{}}
		ga := createGuardedAccountWithEpoch(currentEpoch)

		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
			AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
				return &trie.DataTrieTrackerStub{
					SaveKeyValueCalled: func(key []byte, value []byte) error {
						require.Fail(t, "should not save anything")
						return nil
					},
				}
			},
		}

		ga.CleanOtherThanActive(acc)
	})
	t.Run("one pending guardian should clean the pending", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g2}}
		ga := createGuardedAccountWithEpoch(currentEpoch)
		expectedConfig := &guardians.Guardians{Slice: []*guardians.Guardian{}}
		expectedValue, _ := ga.marshaller.Marshal(expectedConfig)

		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
			AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
				return &trie.DataTrieTrackerStub{
					SaveKeyValueCalled: func(key []byte, value []byte) error {
						require.Equal(t, guardianKey, key)
						require.Equal(t, value, expectedValue)
						return nil
					},
				}
			},
		}

		ga.CleanOtherThanActive(acc)
	})
	t.Run("one active guardian should set again the active", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1}}
		ga := createGuardedAccountWithEpoch(currentEpoch)

		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
			AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
				return &trie.DataTrieTrackerStub{
					SaveKeyValueCalled: func(key []byte, value []byte) error {
						require.Equal(t, guardianKey, key)
						expectedMarshalledGuardians, _ := ga.marshaller.Marshal(configuredGuardians)
						require.Equal(t, expectedMarshalledGuardians, value)
						return nil
					},
				}
			},
		}

		ga.CleanOtherThanActive(acc)
	})
	t.Run("one active and one pending should set again the active (effect is cleaning the pending)", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1, g2}}
		ga := createGuardedAccountWithEpoch(currentEpoch)

		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
			AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
				return &trie.DataTrieTrackerStub{
					SaveKeyValueCalled: func(key []byte, value []byte) error {
						require.Equal(t, guardianKey, key)
						expectedMarshalledGuardians, _ := ga.marshaller.Marshal(&guardians.Guardians{Slice: []*guardians.Guardian{g1}})
						require.Equal(t, expectedMarshalledGuardians, value)
						return nil
					},
				}
			},
		}

		ga.CleanOtherThanActive(acc)
	})
	t.Run("one active and one disabled should set again the active (effect is cleaning the disabled)", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g0, g1}}
		ga := createGuardedAccountWithEpoch(currentEpoch)

		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
			AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
				return &trie.DataTrieTrackerStub{
					SaveKeyValueCalled: func(key []byte, value []byte) error {
						require.Equal(t, guardianKey, key)
						expectedMarshalledGuardians, _ := ga.marshaller.Marshal(&guardians.Guardians{Slice: []*guardians.Guardian{g1}})
						require.Equal(t, expectedMarshalledGuardians, value)
						return nil
					},
				}
			},
		}

		ga.CleanOtherThanActive(acc)
	})
}

func TestGuardedAccount_GetConfiguredGuardians(t *testing.T) {
	currentEpoch := uint32(10)
	g0 := &guardians.Guardian{
		Address:         []byte("old guardian"),
		ActivationEpoch: currentEpoch - 4,
	}
	g1 := &guardians.Guardian{
		Address:         []byte("active guardian"),
		ActivationEpoch: currentEpoch - 2,
	}
	g2 := &guardians.Guardian{
		Address:         []byte("pending guardian"),
		ActivationEpoch: currentEpoch + 2,
	}
	ga := createGuardedAccountWithEpoch(currentEpoch)

	t.Run("unmarshall error", func(t *testing.T) {
		t.Parallel()

		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				return []byte("wrong data"), 0, nil
			},
		}
		active, pending, err := ga.GetConfiguredGuardians(acc)
		require.Nil(t, active)
		require.Nil(t, pending)
		require.NotNil(t, err)
	})
	t.Run("empty configured guardians", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				return nil, 0, expectedErr
			},
		}
		active, pending, err := ga.GetConfiguredGuardians(acc)
		require.Nil(t, active)
		require.Nil(t, pending)
		require.Nil(t, err)
	})
	t.Run("one pending guardian", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g2}}
		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}
		active, pending, err := ga.GetConfiguredGuardians(acc)
		require.Nil(t, active)
		require.Equal(t, g2, pending)
		require.Nil(t, err)
	})
	t.Run("one active guardian", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1}}
		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}
		active, pending, err := ga.GetConfiguredGuardians(acc)
		require.Equal(t, g1, active)
		require.Nil(t, pending)
		require.Nil(t, err)
	})
	t.Run("one active and one pending", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g1, g2}}
		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}
		active, pending, err := ga.GetConfiguredGuardians(acc)
		require.Equal(t, g1, active)
		require.Equal(t, g2, pending)
		require.Nil(t, err)
	})
	t.Run("one old and one active", func(t *testing.T) {
		configuredGuardians := &guardians.Guardians{Slice: []*guardians.Guardian{g0, g1}}
		acc := &stateMocks.UserAccountStub{
			RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
				val, err := ga.marshaller.Marshal(configuredGuardians)
				return val, 0, err
			},
		}
		active, pending, err := ga.GetConfiguredGuardians(acc)
		require.Equal(t, g1, active)
		require.Nil(t, pending)
		require.Nil(t, err)
	})
}

func TestGuardedAccount_EpochConfirmed(t *testing.T) {
	ga := createGuardedAccountWithEpoch(0)
	ga.EpochConfirmed(1, 0)
	require.Equal(t, uint32(1), ga.currentEpoch)

	ga.EpochConfirmed(111, 0)
	require.Equal(t, uint32(111), ga.currentEpoch)
}

func TestGuardedAccount_IsInterfaceNil(t *testing.T) {
	var gah process.GuardedAccountHandler
	require.True(t, check.IfNil(gah))

	var ga *guardedAccount
	require.True(t, check.IfNil(ga))

	ga, _ = NewGuardedAccount(&marshallerMock.MarshalizerMock{}, &epochNotifier.EpochNotifierStub{}, 10)
	require.False(t, check.IfNil(ga))
}

func TestGuardedAccount_EpochConcurrency(t *testing.T) {
	t.Parallel()

	marshaller := &marshallerMock.MarshalizerMock{}
	currentEpoch := uint32(0)
	en := forking.NewGenericEpochNotifier()
	ga, _ := NewGuardedAccount(marshaller, en, 2)
	ctx := context.Background()
	go func() {
		epochTime := time.Millisecond
		timer := time.NewTimer(epochTime)
		defer timer.Stop()

		for {
			timer.Reset(epochTime)
			select {
			case <-timer.C:
				hdr := &block.Header{
					Epoch: currentEpoch,
				}
				en.CheckEpoch(hdr)
				currentEpoch++
			case <-ctx.Done():
				return
			}
		}
	}()

	uah := &stateMocks.UserAccountStub{
		AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
			return &trie.DataTrieTrackerStub{
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					return nil
				},
			}
		},
	}
	err := ga.SetGuardian(uah, []byte("guardian address"), nil, []byte("uuid"))
	require.Nil(t, err)
}

func createGuardedAccountWithEpoch(epoch uint32) *guardedAccount {
	marshaller := &marshallerMock.MarshalizerMock{}
	en := &epochNotifier.EpochNotifierStub{
		RegisterNotifyHandlerCalled: func(handler vmcommon.EpochSubscriberHandler) {
			handler.EpochConfirmed(epoch, 0)
		},
	}

	ga, _ := NewGuardedAccount(marshaller, en, 10)
	return ga
}
