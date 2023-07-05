package disabled

type stateStatistics struct{}

func NewStateStatistics() *stateStatistics {
	return &stateStatistics{}
}

func (s *stateStatistics) Reset() {
}

func (s *stateStatistics) AddNumCache(value int) {
}

func (s *stateStatistics) NumCache() int {
	return 0
}

func (s *stateStatistics) AddNumPersister(value int) {
}

func (s *stateStatistics) NumPersister() int {
	return 0
}

func (s *stateStatistics) AddNumTrie(value int) {
}

func (s *stateStatistics) NumTrie() int {
	return 0
}

func (s *stateStatistics) IsInterfaceNil() bool {
	return s == nil
}
