package goroutines

// NewGoCounterWithHandler -
func NewGoCounterWithHandler(goRoutinesFetcher func() string, filter func(string) bool) *goCounter {
	return &goCounter{
		filter:            filter,
		snapshots:         make([]*snapshot, 0),
		goRoutinesFetcher: goRoutinesFetcher,
	}
}
