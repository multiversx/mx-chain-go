package checking

import "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"

// NodesSetupChecker defines the interface for genesis nodes checker
type NodesSetupChecker interface {
	Check(initialNodes []nodesCoordinator.GenesisNodeInfoHandler) error
	IsInterfaceNil() bool
}

// NodesSetupCheckerFactory defines the interface for a nodes setup checker factory
type NodesSetupCheckerFactory interface {
	CreateNodesSetupChecker(args ArgsNodesSetupChecker) (NodesSetupChecker, error)
	IsInterfaceNil() bool
}
