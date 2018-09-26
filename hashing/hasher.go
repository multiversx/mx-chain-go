package hashing

type Hasher interface {
	CalculateHash(interface{}) []byte
	EmptyHash() []byte
	HashSize() int
}
