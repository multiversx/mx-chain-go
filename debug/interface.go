package debug

// QueryHandler defines the behavior of a queryable debug handler
type QueryHandler interface {
	Query(search string) []string
	IsInterfaceNil() bool
}
