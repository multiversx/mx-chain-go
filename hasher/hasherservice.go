package hasher

type IHasherService interface {
	CalculateHash(interface{}) interface{}
}
