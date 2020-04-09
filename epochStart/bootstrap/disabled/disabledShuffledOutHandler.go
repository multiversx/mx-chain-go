package disabled

type shuffledOutHandler struct {
}

// NewShuffledOutHandler will return a new instance of shuffledOutHandler
func NewShuffledOutHandler() *shuffledOutHandler {
	return &shuffledOutHandler{}
}

// Process won't do anything and will return nil
func (s *shuffledOutHandler) Process(_ uint32) error {
	return nil
}

// RegisterHandler won't do anything
func (s *shuffledOutHandler) RegisterHandler(_ func(newShardID uint32)) {
}

// CurrentShardID won't return a real shard ID
func (s *shuffledOutHandler) CurrentShardID() uint32 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *shuffledOutHandler) IsInterfaceNil() bool {
	return s == nil
}
