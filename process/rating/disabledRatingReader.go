package rating

// DisabledRatingReader represents a nil implementation for the RatingReader interface
type disabledRatingReader struct {
	startRating uint32
}

// NewDisabledRatingReader create a new ratingReader that returns fixed values
func NewDisabledRatingReader(startRating uint32) *disabledRatingReader {
	return &disabledRatingReader{startRating: startRating}
}

//GetRating gets the rating for the public key
func (rr *disabledRatingReader) GetRating(string) uint32 {
	return rr.startRating
}

//UpdateRatingFromTempRating sets the new rating to the value of the tempRating
func (rr *disabledRatingReader) UpdateRatingFromTempRating([]string) error {
	return nil
}

//IsInterfaceNil verifies if the interface is nil
func (rr *disabledRatingReader) IsInterfaceNil() bool {
	return rr == nil
}
