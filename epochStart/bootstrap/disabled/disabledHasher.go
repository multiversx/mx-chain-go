package disabled

type hasher struct {
}

func NewHasher() *hasher {
	return &hasher{}
}

func (h *hasher) Compute(_ string) []byte {
	return nil
}

func (h *hasher) Size() int {
	return 0
}

func (h *hasher) IsInterfaceNil() bool {
	return h == nil
}
