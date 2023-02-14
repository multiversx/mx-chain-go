package notifier

type httpClientHandler interface {
	Post(route string, payload interface{}) error
	IsInterfaceNil() bool
}
