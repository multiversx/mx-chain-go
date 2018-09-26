package marshal

type Marshalizer interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(obj interface{}, buff []byte) error
	Version() string
}

var DMarsh Marshalizer

func init() {
	DMarsh = &JsonMarshalizer{}
}

func DefaultMarshalizer() Marshalizer {
	return DMarsh
}
