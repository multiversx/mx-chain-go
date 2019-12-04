package mock

type RaterMock struct {
	ComputeRatingCalled      func(string, uint32) uint32
	GetRatingCalled          func(string) uint32
	GetRatingOptionKeyCalled func() []string
	RatingReader             RatingReader
}

func GetNewMockRater() *RaterMock {
	return &RaterMock{
		ComputeRatingCalled: func(s string, u uint32) uint32 {
			return 1
		},
		GetRatingCalled: func(s string) uint32 {
			return 1
		},
		GetRatingOptionKeyCalled: func() []string {
			options := make([]string, 0)
			options = append(options, "a")
			options = append(options, "b")
			options = append(options, "c")
			options = append(options, "d")
			return options
		},
		RatingReader: nil,
	}
}

func (rm *RaterMock) ComputeRating(ratingOptionKey string, previousValue uint32) uint32 {
	if rm.ComputeRatingCalled != nil {
		return rm.ComputeRatingCalled(ratingOptionKey, previousValue)
	}
	return 1
}

func (rm *RaterMock) GetRating(pk string) uint32 {
	return rm.GetRatingCalled(pk)
}

func (rm *RaterMock) GetRatings([]string) map[string]uint32 {
	return make(map[string]uint32)
}

func (rm *RaterMock) GetRatingOptionKeys() []string {
	if rm.GetRatingOptionKeyCalled != nil {
		return rm.GetRatingOptionKeyCalled()
	}
	return make([]string, 0)
}

func (rm *RaterMock) GetStartRating() uint32 {
	return 5
}

//SetRatingReader sets the Reader that can read ratings
func (rm *RaterMock) SetRatingReader(reader RatingReader) {

}

//SetRatingReader sets the Reader that can read ratings
func (rm *RaterMock) IsInterfaceNil() bool {
	return false
}

type RatingReader interface {
	//GetRating gets the rating for the public key
	GetRating(string) uint32
	//GetRatings gets all the ratings as a map[pk] ratingValue
	GetRatings([]string) map[string]uint32
}
