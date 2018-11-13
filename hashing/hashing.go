package hashing

type Hasher interface {
	Compute(string) []byte
	EmptyHash() []byte
	Size() int
}

var DefHash Hasher

func init() {
	DefHash = &Sha256{}
}
