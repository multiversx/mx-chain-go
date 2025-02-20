package checking

type nodesSetupCheckerFactory struct {
}

// NewNodesSetupCheckerFactory creates a nodes setup checker factory
func NewNodesSetupCheckerFactory() *nodesSetupCheckerFactory {
	return &nodesSetupCheckerFactory{}
}

// CreateNodesSetupChecker creates a new nodes setup checker
func (f *nodesSetupCheckerFactory) CreateNodesSetupChecker(args ArgsNodesSetupChecker) (NodesSetupChecker, error) {
	return NewNodesSetupChecker(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *nodesSetupCheckerFactory) IsInterfaceNil() bool {
	return f == nil
}
