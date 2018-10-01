package hashing

type Hasher interface {
	Compute(string) []byte
	EmptyHash() []byte
	Size() int
}
