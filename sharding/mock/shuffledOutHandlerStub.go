package mock

// ShuffledOutHandlerStub -
type ShuffledOutHandlerStub struct {
	ProcessCalled         func(newShardID uint32) error
	RegisterHandlerCalled func(handler func(newShardID uint32))
	CurrentShardIDCalled  func() uint32
}

// Process -
func (s *ShuffledOutHandlerStub) Process(newShardID uint32) error {
	if s.ProcessCalled != nil {
		return s.ProcessCalled(newShardID)
	}

	return nil
}

// RegisterHandler -
func (s *ShuffledOutHandlerStub) RegisterHandler(handler func(newShardID uint32)) {
	if s.RegisterHandlerCalled != nil {
		s.RegisterHandlerCalled(handler)
	}
}

// CurrentShardID -
func (s *ShuffledOutHandlerStub) CurrentShardID() uint32 {
	if s.CurrentShardIDCalled != nil {
		return s.CurrentShardIDCalled()
	}

	return 0
}

// IsInterfaceNil -
func (s *ShuffledOutHandlerStub) IsInterfaceNil() bool {
	return s == nil
}
