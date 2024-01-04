package state

// SignRate -
type SignRate struct {
	NumSuccess uint32
	NumFailure uint32
}

// GetNumSuccess -
func (s *SignRate) GetNumSuccess() uint32 {
	return s.NumSuccess

}

// GetNumFailure -
func (s *SignRate) GetNumFailure() uint32 {
	return s.NumFailure
}
