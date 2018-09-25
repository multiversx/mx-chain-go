package data

import (
	block "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	blockchain "github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

var services map[interface{}]interface{}

func init() {

	if services == nil {
		services = make(map[interface{}]interface{})
	}

	InjectDefaultServices()
}

func InjectDefaultServices() {
	PutService("Blocker", block.BlockImpl1{})
	PutService("BlockChainer", blockchain.BlockChainImpl{})
}

func GetService(key interface{}) interface{} {
	return services[key]
}

func PutService(key interface{}, value interface{}) {

	if key == nil || value == nil {
		return
	}

	services[key] = value
}

func GetBlockerService() Blocker {
	return GetService("Blocker").(Blocker)
}

func GetBlockChainerService() BlockChainer {
	return GetService("BlockChainer").(BlockChainer)
}
