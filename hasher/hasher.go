package hasher

type Hasher interface {
	CalculateHash(interface{}) interface{}
}
