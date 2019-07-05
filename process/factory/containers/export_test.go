package containers

import "github.com/ElrondNetwork/elrond-go/data/block"

func (ic *InterceptorsContainer) Insert(key string, value interface{}) bool {
	return ic.objects.Insert(key, value)
}

func (ppc *PreProcessorsContainer) Insert(key block.Type, value interface{}) bool {
	return ppc.objects.Insert(uint8(key), value)
}

func (ppc *IntermediateTransactionHandlersContainer) Insert(key block.Type, value interface{}) bool {
	return ppc.objects.Insert(uint8(key), value)
}
