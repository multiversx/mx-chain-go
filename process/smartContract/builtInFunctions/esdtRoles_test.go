package builtInFunctions

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/require"
)

func TestNewESDTRolesFunc_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	esdtRolesF, err := NewESDTRolesFunc(nil, false)

	require.Equal(t, process.ErrNilMarshalizer, err)
	require.Nil(t, esdtRolesF)
}

func TestEsdtRoles_ProcessBuiltinFunction_NilVMInputShouldErr(t *testing.T) {
	t.Parallel()

	esdtRolesF, _ := NewESDTRolesFunc(nil, false)

	_, err := esdtRolesF.ProcessBuiltinFunction(nil, &mock.UserAccountStub{}, nil)
	require.Equal(t, process.ErrNilVmInput, err)
}

func TestEsdtRoles_ProcessBuiltinFunction_WrongCalledShouldErr(t *testing.T) {
	t.Parallel()

	esdtRolesF, _ := NewESDTRolesFunc(nil, false)

	_, err := esdtRolesF.ProcessBuiltinFunction(nil, &mock.UserAccountStub{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			CallerAddr: []byte{},
			Arguments:  [][]byte{[]byte("1"), []byte("2")},
		},
	})
	require.Equal(t, process.ErrAddressIsNotESDTSystemSC, err)
}

func TestEsdtRoles_ProcessBuiltinFunction_NilAccountDestShouldErr(t *testing.T) {
	t.Parallel()

	esdtRolesF, _ := NewESDTRolesFunc(nil, false)

	_, err := esdtRolesF.ProcessBuiltinFunction(nil, nil, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			CallerAddr: vm.ESDTSCAddress,
			Arguments:  [][]byte{[]byte("1"), []byte("2")},
		},
	})
	require.Equal(t, process.ErrNilUserAccount, err)
}

func TestEsdtRoles_ProcessBuiltinFunction_GetRolesFailShouldErr(t *testing.T) {
	t.Parallel()

	esdtRolesF, _ := NewESDTRolesFunc(&mock.MarshalizerMock{}, false)

	localErr := errors.New("local err")
	_, err := esdtRolesF.ProcessBuiltinFunction(nil, &mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					return nil, localErr
				},
			}
		},
	}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			CallerAddr: vm.ESDTSCAddress,
			Arguments:  [][]byte{[]byte("1"), []byte("2")},
		},
	})
	require.Equal(t, localErr, err)
}

func TestEsdtRoles_ProcessBuiltinFunction_SetRolesShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtRolesF, _ := NewESDTRolesFunc(marshalizer, true)

	acc := &mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					roles := &esdt.ESDTRoles{}
					return marshalizer.Marshal(roles)
				},
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					roles := &esdt.ESDTRoles{}
					_ = marshalizer.Unmarshal(roles, value)
					require.Equal(t, roles.Roles, [][]byte{[]byte(core.ESDTRoleLocalMint)})
					return nil
				},
			}
		},
	}
	_, err := esdtRolesF.ProcessBuiltinFunction(nil, acc, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			CallerAddr: vm.ESDTSCAddress,
			Arguments:  [][]byte{[]byte("1"), []byte(core.ESDTRoleLocalMint)},
		},
	})
	require.Nil(t, err)
}

func TestEsdtRoles_ProcessBuiltinFunction_SaveFailedShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtRolesF, _ := NewESDTRolesFunc(marshalizer, true)

	localErr := errors.New("local err")
	acc := &mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					roles := &esdt.ESDTRoles{}
					return marshalizer.Marshal(roles)
				},
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					return localErr
				},
			}
		},
	}
	_, err := esdtRolesF.ProcessBuiltinFunction(nil, acc, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			CallerAddr: vm.ESDTSCAddress,
			Arguments:  [][]byte{[]byte("1"), []byte(core.ESDTRoleLocalMint)},
		},
	})
	require.Equal(t, localErr, err)
}

func TestEsdtRoles_ProcessBuiltinFunction_UnsetRolesDoesNotExistsShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtRolesF, _ := NewESDTRolesFunc(marshalizer, false)

	acc := &mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					roles := &esdt.ESDTRoles{}
					return marshalizer.Marshal(roles)
				},
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					roles := &esdt.ESDTRoles{}
					_ = marshalizer.Unmarshal(roles, value)
					require.Len(t, roles.Roles, 0)
					return nil
				},
			}
		},
	}
	_, err := esdtRolesF.ProcessBuiltinFunction(nil, acc, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			CallerAddr: vm.ESDTSCAddress,
			Arguments:  [][]byte{[]byte("1"), []byte(core.ESDTRoleLocalMint)},
		},
	})
	require.Nil(t, err)
}

func TestEsdtRoles_ProcessBuiltinFunction_UnsetRolesShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtRolesF, _ := NewESDTRolesFunc(marshalizer, false)

	acc := &mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					roles := &esdt.ESDTRoles{
						Roles: [][]byte{[]byte(core.ESDTRoleLocalMint)},
					}
					return marshalizer.Marshal(roles)
				},
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					roles := &esdt.ESDTRoles{}
					_ = marshalizer.Unmarshal(roles, value)
					require.Len(t, roles.Roles, 0)
					return nil
				},
			}
		},
	}
	_, err := esdtRolesF.ProcessBuiltinFunction(nil, acc, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			CallerAddr: vm.ESDTSCAddress,
			Arguments:  [][]byte{[]byte("1"), []byte(core.ESDTRoleLocalMint)},
		},
	})
	require.Nil(t, err)
}

func TestEsdtRoles_CheckAllowedToExecuteNilAccountShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtRolesF, _ := NewESDTRolesFunc(marshalizer, false)

	err := esdtRolesF.CheckAllowedToExecute(nil, []byte("ID"), []byte(core.ESDTRoleLocalBurn))
	require.Equal(t, process.ErrNilUserAccount, err)
}

func TestEsdtRoles_CheckAllowedToExecuteCannotGetESDTRole(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtRolesF, _ := NewESDTRolesFunc(marshalizer, false)

	localErr := errors.New("local err")
	err := esdtRolesF.CheckAllowedToExecute(&mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					return nil, localErr
				},
			}
		},
	}, []byte("ID"), []byte(core.ESDTRoleLocalBurn))
	require.Equal(t, localErr, err)
}

func TestEsdtRoles_CheckAllowedToExecuteIsNewNotAllowed(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtRolesF, _ := NewESDTRolesFunc(marshalizer, false)

	err := esdtRolesF.CheckAllowedToExecute(&mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					return nil, nil
				},
			}
		},
	}, []byte("ID"), []byte(core.ESDTRoleLocalBurn))
	require.Equal(t, process.ErrActionNotAllowed, err)
}

func TestEsdtRoles_CheckAllowed_ShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtRolesF, _ := NewESDTRolesFunc(marshalizer, false)

	err := esdtRolesF.CheckAllowedToExecute(&mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					roles := &esdt.ESDTRoles{
						Roles: [][]byte{[]byte(core.ESDTRoleLocalMint)},
					}
					return marshalizer.Marshal(roles)
				},
			}
		},
	}, []byte("ID"), []byte(core.ESDTRoleLocalMint))
	require.Nil(t, err)
}

func TestEsdtRoles_CheckAllowedToExecuteRoleNotFind(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtRolesF, _ := NewESDTRolesFunc(marshalizer, false)

	err := esdtRolesF.CheckAllowedToExecute(&mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					roles := &esdt.ESDTRoles{
						Roles: [][]byte{[]byte(core.ESDTRoleLocalBurn)},
					}
					return marshalizer.Marshal(roles)
				},
			}
		},
	}, []byte("ID"), []byte(core.ESDTRoleLocalMint))
	require.Equal(t, process.ErrActionNotAllowed, err)
}
