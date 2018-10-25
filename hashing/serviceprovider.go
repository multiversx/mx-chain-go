package hashing

import "github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"

var services map[string]Hasher

const Hash = "hasher"

var DefHasher Hasher

func init() {

	if services == nil {
		services = make(map[string]Hasher)
	}

	DefHasher = sha256.Sha256{}

	InjectDefaultServices()
}

func InjectDefaultServices() {
	PutService(Hash, DefHasher)
}

func GetService(key string) Hasher {
	return services[key]
}

func PutService(key string, value Hasher) {

	if key == "" || value == nil {
		return
	}

	services[key] = value
}

func GetHasherService() Hasher {
	return GetService(Hash).(Hasher)
}
