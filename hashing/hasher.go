package hashing

type Hasher interface {
	CalculateHash(interface{}) interface{}
}
