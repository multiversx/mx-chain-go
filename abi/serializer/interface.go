package serializer

type valuesCodec interface {
	EncodeNested(value any) ([]byte, error)
	EncodeTopLevel(value any) ([]byte, error)
	DecodeNested(data []byte, value any) error
	DecodeTopLevel(data []byte, value any) error
}
