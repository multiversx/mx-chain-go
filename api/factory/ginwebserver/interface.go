package ginwebserver

type resetHandler interface {
	Reset()
	IsInterfaceNil() bool
}
