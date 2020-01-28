package containers

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/cornelk/hashmap"
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
	return vmc.objects.Insert(key, value)
}

func (ic *interceptorsContainer) Objects() *hashmap.HashMap {
	return ic.objects
}
