package transaction

// Encoder represents a byte slice to string encoder
type Encoder interface {
	Encode(buff []byte) string
	IsInterfaceNil() bool
}

// Marshalizer is able to encode an object to its byte slice representation
type Marshalizer interface {
	Marshal(obj interface{}) ([]byte, error)
	IsInterfaceNil() bool
}
