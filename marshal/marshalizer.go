package marshal

type Marshalizer interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(obj interface{}, buff []byte) error
	Version() string
}

//var jsn = &JsonMarshalizer{}
//
//func Json() Marshalizer {
//	return jsn
//}
