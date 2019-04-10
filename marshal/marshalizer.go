package marshal

// Marshalizer defines the 2 basic operations: serialize (marshal) and deserialize (unmarshal)
type Marshalizer interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(obj interface{}, buff []byte) error
}
