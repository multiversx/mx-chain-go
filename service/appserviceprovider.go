package service

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
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
	PutService("IConsensusService", consensus.ConsensusServiceImpl{})
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

func GetConsensusService() consensus.IConsensusService {
	return GetService("IConsensusService").(consensus.IConsensusService)
}

func GetMarshalizerService() marshal.Marshalizer {
	return GetService("Marshalizer").(marshal.Marshalizer)
}
