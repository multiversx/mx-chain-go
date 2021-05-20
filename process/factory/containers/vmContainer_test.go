package containers_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewVirtualMachinesContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	assert.NotNil(t, c)
	assert.False(t, c.IsInterfaceNil())
}

//------- Add

func TestVirtualMachinesContainer_AddAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	_ = c.Add([]byte("0001"), &mock.VMExecutionHandlerStub{})
	err := c.Add([]byte("0001"), &mock.VMExecutionHandlerStub{})

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestVirtualMachinesContainer_AddNilShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	err := c.Add([]byte("0001"), nil)

	assert.Equal(t, process.ErrNilContainerElement, err)
}

func TestVirtualMachinesContainer_AddShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	err := c.Add([]byte("0001"), &mock.VMExecutionHandlerStub{})

	assert.Nil(t, err)
	assert.Equal(t, 1, c.Len())
}

//------- AddMultiple

func TestVirtualMachinesContainer_AddMultipleAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	keys := [][]byte{[]byte("0001"), []byte("0001")}
	vms := []vmcommon.VMExecutionHandler{&mock.VMExecutionHandlerStub{}, &mock.VMExecutionHandlerStub{}}

	err := c.AddMultiple(keys, vms)

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestVirtualMachinesContainer_AddMultipleLenMismatchShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	keys := [][]byte{[]byte("0001")}
	vms := []vmcommon.VMExecutionHandler{&mock.VMExecutionHandlerStub{}, &mock.VMExecutionHandlerStub{}}

	err := c.AddMultiple(keys, vms)

	assert.Equal(t, process.ErrLenMismatch, err)
}

func TestVirtualMachinesContainer_AddMultipleShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	keys := [][]byte{[]byte("0001"), []byte("0002")}
	vms := []vmcommon.VMExecutionHandler{&mock.VMExecutionHandlerStub{}, &mock.VMExecutionHandlerStub{}}

	err := c.AddMultiple(keys, vms)

	assert.Nil(t, err)
	assert.Equal(t, 2, c.Len())
}

//------- Get

func TestVirtualMachinesContainer_GetNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	key := []byte("0001")
	keyNotFound := []byte("0002")
	val := &mock.VMExecutionHandlerStub{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(keyNotFound)

	assert.Nil(t, valRecovered)
	assert.True(t, errors.Is(err, process.ErrInvalidContainerKey))
}

func TestVirtualMachinesContainer_GetWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	key := []byte("0001")

	_ = c.Insert(key, "string value")
	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrWrongTypeInContainer, err)
}

func TestVirtualMachinesContainer_GetShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	key := []byte("0001")
	val := &mock.VMExecutionHandlerStub{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(key)

	assert.True(t, val == valRecovered)
	assert.Nil(t, err)
}

//------- Replace

func TestVirtualMachinesContainer_ReplaceNilValueShouldErrAndNotModify(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	key := []byte("0001")
	val := &mock.VMExecutionHandlerStub{}

	_ = c.Add(key, val)
	err := c.Replace(key, nil)

	valRecovered, _ := c.Get(key)

	assert.Equal(t, process.ErrNilContainerElement, err)
	assert.Equal(t, val, valRecovered)
}

func TestVirtualMachinesContainer_ReplaceShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	key := []byte("0001")
	val := &mock.VMExecutionHandlerStub{}
	val2 := &mock.VMExecutionHandlerStub{}

	_ = c.Add(key, val)
	err := c.Replace(key, val2)

	valRecovered, _ := c.Get(key)

	assert.True(t, val2 == valRecovered)
	assert.Nil(t, err)
}

//------- Remove

func TestVirtualMachinesContainer_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	key := []byte("0001")
	val := &mock.VMExecutionHandlerStub{}

	_ = c.Add(key, val)
	c.Remove(key)

	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.True(t, errors.Is(err, process.ErrInvalidContainerKey))
}

//------- Len

func TestVirtualMachinesContainer_LenShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	_ = c.Add([]byte("0001"), &mock.VMExecutionHandlerStub{})
	assert.Equal(t, 1, c.Len())

	keys := c.Keys()
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, []byte("0001"), keys[0])

	_ = c.Add([]byte("0002"), &mock.VMExecutionHandlerStub{})
	assert.Equal(t, 2, c.Len())

	keys = c.Keys()
	assert.Equal(t, 2, len(keys))
	assert.Contains(t, keys, []byte("0001"))
	assert.Contains(t, keys, []byte("0002"))

	c.Remove([]byte("0002"))
	assert.Equal(t, 1, c.Len())

	keys = c.Keys()
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, []byte("0001"), keys[0])
}

func TestVirtualMachinesContainer_Close(t *testing.T) {
	t.Skip("fix this test whe the TODO on vmContainer.go L179 will be solved")
	t.Parallel()

	c := containers.NewVirtualMachinesContainer()

	closeCalledItem1 := false
	item1 := struct {
		*mock.CloserStub
		vmcommon.VMExecutionHandler
	}{
		CloserStub: &mock.CloserStub{
			CloseCalled: func() error {
				closeCalledItem1 = true

				return nil
			},
		},
	}

	closeCalledItem2 := false
	expectedErr := errors.New("expected Error")
	item2 := struct {
		*mock.CloserStub
		vmcommon.VMExecutionHandler
	}{
		CloserStub: &mock.CloserStub{
			CloseCalled: func() error {
				closeCalledItem2 = true

				return expectedErr
			},
		},
	}

	cleanCalledItem3 := false
	item3 := struct {
		*mock.CleanerStub
		vmcommon.VMExecutionHandler
	}{
		CleanerStub: &mock.CleanerStub{
			CleanCalled: func() {
				cleanCalledItem3 = true
			},
		},
	}
	closeCalledItem4 := false
	cleanCalledItem4 := false
	item4 := struct {
		*mock.CloserStub
		*mock.CleanerStub
		vmcommon.VMExecutionHandler
	}{
		CleanerStub: &mock.CleanerStub{
			CleanCalled: func() {
				cleanCalledItem4 = true
			},
		},
		CloserStub: &mock.CloserStub{
			CloseCalled: func() error {
				closeCalledItem4 = true

				return nil
			},
		},
	}

	_ = c.Add([]byte("key1"), item1)
	_ = c.Add([]byte("key2"), item2)
	_ = c.Add([]byte("key3"), item3)
	_ = c.Add([]byte("key4"), item4)

	err := c.Close()
	assert.Equal(t, containers.ErrCloseVMContainer, err)
	assert.True(t, closeCalledItem1)
	assert.True(t, closeCalledItem2)
	assert.True(t, cleanCalledItem3)
	assert.True(t, closeCalledItem4)
	assert.False(t, cleanCalledItem4) //not calling Close and Clean on the same object
}
