package runType

// runTypeComponentsCreator is the interface for the runTypeComponentsCreator
type runTypeComponentsCreator interface {
	Create() (*runTypeComponents, error)
	IsInterfaceNil() bool
}

type runTypeCoreComponentsCreator interface {
	Create() *runTypeCoreComponents
	IsInterfaceNil() bool
}
