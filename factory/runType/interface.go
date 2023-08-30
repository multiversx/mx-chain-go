package runType

// RunTypeComponentsCreator defines the run type components creator
type RunTypeComponentsCreator interface {
	Create() (*runTypeComponents, error)
	IsInterfaceNil() bool
}
