package gin

import "context"

type resetHandler interface {
	Reset()
	IsInterfaceNil() bool
}

type server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}
