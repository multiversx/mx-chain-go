package service

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/hasher"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

var services map[interface{}]interface{}

func init() {

	if services == nil {
		services = make(map[interface{}]interface{})
	}

	InjectDefaultServices()
}

func InjectDefaultServices() {
	PutService("IBlockService", data.BlockServiceImpl{})
	PutService("IBlockChainService", data.BlockChainServiceImpl{})
	PutService("IChronologyService", chronology.ChronologyServiceImpl{})
	PutService("IConsensusService", consensus.ConsensusServiceImpl{})
	PutService("IHasherService", hasher.Sha256Impl{})
	PutService("Marshalizer", &marshal.JsonMarshalizer{})
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

func GetBlockService() data.IBlockService {
	return GetService("IBlockService").(data.IBlockService)
}

func GetBlockChainService() data.IBlockChainService {
	return GetService("IBlockChainService").(data.IBlockChainService)
}

func GetChronologyService() chronology.IChronologyService {
	return GetService("IChronologyService").(chronology.IChronologyService)
}

func GetConsensusService() consensus.IConsensusService {
	return GetService("IConsensusService").(consensus.IConsensusService)
}

func GetHasherService() hasher.IHasherService {
	return GetService("IHasherService").(hasher.IHasherService)
}

func GetMarshalizerService() marshal.Marshalizer {
	return GetService("Marshalizer").(marshal.Marshalizer)
}
