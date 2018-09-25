package chronology

import (
	epoch "github.com/ElrondNetwork/elrond-go-sandbox/chronology/epoch"
	round "github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"
)

var services map[interface{}]interface{}

func init() {

	if services == nil {
		services = make(map[interface{}]interface{})
	}

	InjectDefaultServices()
}

func InjectDefaultServices() {
	PutService("Epocher", epoch.EpochImpl{})
	PutService("Rounder", round.RoundImpl{})
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

func GetEpocherService() Epocher {
	return GetService("Epocher").(Epocher)
}

func GetRounderService() Rounder {
	return GetService("Rounder").(Rounder)
}
