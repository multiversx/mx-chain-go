package resolver

type RequestDataType string

const (
	HashType  RequestDataType = "HashType"
	NonceType RequestDataType = "NonceType"
)

type RequestData struct {
	Type  RequestDataType
	Value []byte
}
