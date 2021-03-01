package gin

type resetHandler interface {
	Reset()
	IsInterfaceNil() bool
}
