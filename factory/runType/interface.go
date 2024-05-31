package runType

// runTypeComponentsCreator is the interface for the runTypeComponentsCreator
type runTypeComponentsCreator interface {
	Create() (*runTypeComponents, error)
	IsInterfaceNil() bool
}
