package containers

import (
	"github.com/multiversx/mx-chain-core-go/core/container"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

func (ic *interceptorsContainer) Insert(key string, value interface{}) bool {
	return ic.objects.Insert(key, value)
}

func (ppc *preProcessorsContainer) Insert(key block.Type, value interface{}) bool {
	return ppc.objects.Insert(uint8(key), value)
}

func (ppc *intermediateTransactionHandlersContainer) Insert(key block.Type, value interface{}) bool {
	return ppc.objects.Insert(uint8(key), value)
}

func (vmc *virtualMachinesContainer) Insert(key []byte, value interface{}) bool {
	return vmc.objects.Insert(string(key), value)
}

func (ic *interceptorsContainer) Objects() *container.MutexMap {
	return ic.objects
}
