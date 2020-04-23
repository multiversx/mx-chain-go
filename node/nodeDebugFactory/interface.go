package nodeDebugFactory

import "github.com/ElrondNetwork/elrond-go/debug"

// NodeWrapper is the interface that defines the behavior of a Node that can work with debug handlers
type NodeWrapper interface {
	AddQueryHandler(name string, handler debug.QueryHandler) error
	IsInterfaceNil() bool
}
