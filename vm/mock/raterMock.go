package mock

// RaterMock -
type RaterMock struct {
	ComputeRatingCalled  func(string, uint32) uint32
	GetRatingCalled      func(string) uint32
	GetStartRatingCalled func() uint32
	GetChancesCalled     func(uint32) uint32
}

// ComputeRating -
func (rm *RaterMock) ComputeRating(ratingOptionKey string, previousValue uint32) uint32 {
	if rm.ComputeRatingCalled != nil {
		return rm.ComputeRatingCalled(ratingOptionKey, previousValue)
	}
	return 1
}

// GetRating -
func (rm *RaterMock) GetRating(pk string) uint32 {
	if rm.GetRatingCalled != nil {
		return rm.GetRatingCalled(pk)
	}
	return 1
}

// IsInterfaceNil -
func (rm *RaterMock) IsInterfaceNil() bool {
	return rm == nil
}

// GetStartRating -
func (rm *RaterMock) GetStartRating() uint32 {
	if rm.GetStartRatingCalled != nil {
		return rm.GetStartRatingCalled()
	}
	return 5
}

// GetChance -
func (rm *RaterMock) GetChance(rating uint32) uint32 {
	if rm.GetChancesCalled != nil {
		return rm.GetChancesCalled(rating)
	}
	return 5
}
