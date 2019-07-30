package core

// AppStatusHandler interface will handle different implementations of monitoring tools, such as Prometheus of term-ui
type AppStatusHandler interface {
	Increment(key string)
	Decrement(key string)
	SetInt64Value(key string, value int64)
	SetUInt64Value(key string, value uint64)
	Close()
}

// ConnectedAddressesHandler interface will be used for passing the network component to AppStatusPolling
type ConnectedAddressesHandler interface {
	ConnectedAddresses() []string
}
