package systemSmartContracts

import "github.com/multiversx/mx-chain-go/vm"

// VMContextCreatorHandler defines a handler able to create vm context
type VMContextCreatorHandler interface {
	CreateVmContext(args VMContextArgs) (vm.ContextHandler, error)
	IsInterfaceNil() bool
}
