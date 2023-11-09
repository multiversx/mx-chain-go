package testscommon

import "github.com/multiversx/mx-chain-core-go/data"

// VersionedHeaderFactoryStub -
type VersionedHeaderFactoryStub struct {
	CreateCalled func(epoch uint32) data.HeaderHandler
}

// Create -
func (vhfs *VersionedHeaderFactoryStub) Create(epoch uint32) data.HeaderHandler {
	if vhfs.CreateCalled != nil {
		return vhfs.CreateCalled(epoch)
	}
	return nil
}

// IsInterfaceNil -
func (vhfs *VersionedHeaderFactoryStub) IsInterfaceNil() bool {
	return vhfs == nil
}
