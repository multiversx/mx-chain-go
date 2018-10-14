package marshal

type Marshalizer interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(obj interface{}, buff []byte) error
}

var DefMarsh Marshalizer

func init() {
	DefMarsh = &JsonMarshalizer{}
}
