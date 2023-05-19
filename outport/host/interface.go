package host

// SenderHost defines the actions that a host sender should do
type SenderHost interface {
	Send(payload []byte, topic string) error
	Close() error
	IsInterfaceNil() bool
}
