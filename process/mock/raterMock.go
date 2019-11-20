package mock

type RaterMock struct {
	ComputeRatingCalled      func(string, string, uint32) uint32
	GetRatingCalled          func(string) uint32
	GetRatingOptionKeyCalled func() []string
	RatingReader             RatingReader
}

func (rm *RaterMock) ComputeRating(pk, ratingOptionKey string, previousValue uint32) uint32 {
	if rm.ComputeRatingCalled != nil {
		return rm.ComputeRatingCalled(pk, ratingOptionKey, previousValue)
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

//SetRatingReader sets the Reader that can read ratings
func (rm *RaterMock) SetRatingReader(reader RatingReader) {

}

type RatingReader interface {
	//GetRating gets the rating for the public key
	GetRating(string) uint32
	//GetRatings gets all the ratings as a map[pk] ratingValue
	GetRatings([]string) map[string]uint32
}
