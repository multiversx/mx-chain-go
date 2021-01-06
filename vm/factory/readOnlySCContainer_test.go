package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/require"
)

func TestNewReadOnlySCContainer_NilScContainerShouldErr(t *testing.T) {
	t.Parallel()

	roscc, err := NewReadOnlySCContainer(nil)
	require.Equal(t, vm.ErrNilSystemContractsContainer, err)
	require.True(t, check.IfNil(roscc))
}

func TestNewReadOnlySCContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	scContainer := &mock.SystemSCContainerStub{}
	roscc, err := NewReadOnlySCContainer(scContainer)
	require.NoError(t, err)
	require.NotNil(t, roscc)
}

func TestReadOnlySCContainer_WriteOperationShouldNotChangeAnything(t *testing.T) {
	t.Parallel()

	scContainer := &mock.SystemSCContainerStub{
		AddCalled: func(key []byte, val vm.SystemSmartContract) error {
			require.Fail(t, "Add should have not be called")
			return nil
		},
		RemoveCalled: func(key []byte) {
			require.Fail(t, "Remove should have not be called")
		},
		ReplaceCalled: func(key []byte, val vm.SystemSmartContract) error {
			require.Fail(t, "Replace should have not be called")
			return nil
		},
	}

	key := []byte("abc")

	roscc, _ := NewReadOnlySCContainer(scContainer)

	err := roscc.Add(key, &mock.SystemSCStub{})
	require.NoError(t, err)

	err = roscc.Replace(key, &mock.SystemSCStub{})
	require.NoError(t, err)

	roscc.Remove(append(key, []byte("v2")...))

	require.Zero(t, roscc.Len())
	require.Zero(t, len(roscc.Keys()))
}

func TestReadOnlySCContainer_ReadOperationShouldGetFromOriginal(t *testing.T) {
	t.Parallel()

	expectedSystemSc := &mock.SystemSCStub{
		ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return 4
		},
	}

	expectedKeys := [][]byte{[]byte("key0"), []byte("key1")}
	expectedLen := len(expectedKeys)

	scContainer := &mock.SystemSCContainerStub{
		GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
			return expectedSystemSc, nil
		},
		KeysCalled: func() [][]byte {
			return expectedKeys
		},
		LenCalled: func() int {
			return expectedLen
		},
	}

	key := []byte("abc")

	roscc, _ := NewReadOnlySCContainer(scContainer)

	actualSystemSc, err := roscc.Get(key)
	require.NoError(t, err)
	require.Equal(t, expectedSystemSc, actualSystemSc)

	actualKeys := roscc.Keys()
	require.Equal(t, expectedKeys, actualKeys)

	actualLen := roscc.Len()
	require.Equal(t, expectedLen, actualLen)
}
