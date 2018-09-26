package hashing

import "github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"

var services map[interface{}]interface{}

func init() {

	if services == nil {
		services = make(map[interface{}]interface{})
	}

	InjectDefaultServices()
}

func InjectDefaultServices() {
	PutService("Hasher", hashing.Sha256Impl{})
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

func GetHasherService() Hasher {
	return GetService("Hasher").(Hasher)
}
