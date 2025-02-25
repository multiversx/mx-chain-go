package chaosImpl

import mainFactory "github.com/multiversx/mx-chain-go/factory"

// NodeHandler -
type NodeHandler interface {
	GetCoreComponents() mainFactory.CoreComponentsHolder
}
