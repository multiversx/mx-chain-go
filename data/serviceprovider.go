package data

var services map[interface{}]interface{}

func init() {

	if services == nil {
		services = make(map[interface{}]interface{})
	}

	InjectDefaultServices()
}

func InjectDefaultServices() {
	//PutService("Blocker", block.BlockImpl1{})
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

//func GetBlockerService() Blocker {
//	return GetService("Blocker").(Blocker)
//}
