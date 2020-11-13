package containers

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/container"
	vmcommon "github.com/ElrondNetwork/elrond-go/core/vm-common"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.VirtualMachinesContainer = (*virtualMachinesContainer)(nil)

var logVMContainer = logger.GetOrCreate("factory/containers/vmContainer")

// virtualMachinesContainer is an VM holder organized by type
type virtualMachinesContainer struct {
	objects *container.MutexMap
}

// NewVirtualMachinesContainer will create a new instance of a container
func NewVirtualMachinesContainer() *virtualMachinesContainer {
	return &virtualMachinesContainer{
		objects: container.NewMutexMap(),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (vmc *virtualMachinesContainer) Get(key []byte) (vmcommon.VMExecutionHandler, error) {
	value, ok := vmc.objects.Get(string(key))
	if !ok {
		return nil, fmt.Errorf("%w in vm container for key %v", process.ErrInvalidContainerKey, key)
	}

	vm, ok := value.(vmcommon.VMExecutionHandler)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return vm, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (vmc *virtualMachinesContainer) Add(key []byte, vm vmcommon.VMExecutionHandler) error {
	if check.IfNilReflect(vm) {
		return process.ErrNilContainerElement
	}

	ok := vmc.objects.Insert(string(key), vm)

	if !ok {
		return process.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (vmc *virtualMachinesContainer) AddMultiple(keys [][]byte, vms []vmcommon.VMExecutionHandler) error {
	if len(keys) != len(vms) {
		return process.ErrLenMismatch
	}

	for idx, key := range keys {
		err := vmc.Add(key, vms[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (vmc *virtualMachinesContainer) Replace(key []byte, vm vmcommon.VMExecutionHandler) error {
	if check.IfNilReflect(vm) {
		return process.ErrNilContainerElement
	}

	vmc.objects.Set(string(key), vm)
	return nil
}

// Remove will remove an object at a given key
func (vmc *virtualMachinesContainer) Remove(key []byte) {
	vmc.objects.Remove(string(key))
}

// Len returns the length of the added objects
func (vmc *virtualMachinesContainer) Len() int {
	return vmc.objects.Len()
}

// Keys returns all the keys from the container
func (vmc *virtualMachinesContainer) Keys() [][]byte {
	keys := vmc.objects.Keys()
	keysBytes := make([][]byte, 0, len(keys))
	for _, k := range keys {
		key, ok := k.(string)
		if !ok {
			continue
		}
		keysBytes = append(keysBytes, []byte(key))
	}

	return keysBytes
}

// Close closes the items in the container (meaningful for Arwen out-of-process)
func (vmc *virtualMachinesContainer) Close() error {
	var withError bool

	for _, item := range vmc.objects.Values() {
		asCloser, ok := item.(interface{ Close() error })
		if !ok {
			continue
		}

		err := asCloser.Close()
		if err != nil {
			logVMContainer.Error("Cannot close item in container", "err", err)
			withError = true
		}
	}

	if withError {
		return ErrCloseVMContainer
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmc *virtualMachinesContainer) IsInterfaceNil() bool {
	return vmc == nil
}
