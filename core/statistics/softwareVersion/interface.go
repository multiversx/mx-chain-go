package softwareVersion

// StableTagProviderHandler defines what a stable tag provider should be able to do
type StableTagProviderHandler interface {
	FetchTagVersion() (string, error)
	IsInterfaceNil() bool
}
