package chaos

import mainFactory "github.com/multiversx/mx-chain-go/factory"

type NodeHandler interface {
	GetCoreComponents() mainFactory.CoreComponentsHolder
}
