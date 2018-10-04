package state

//var DefHasher hashing.Hasher
//var DefMarsh marshal.Marshalizer
//
//func init() {
//	DefHasher = &mock.MockHasher{}
//	DefMarsh = &mock.MockMarshalizer{}
//}

type Accounter interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	Commit() error
	Undo() error
	Root() []byte
}
