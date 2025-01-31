package chaosAdapters

type Validator interface {
	PubKey() []byte
	Chances() uint32
	Index() uint32
	Size() int
}
