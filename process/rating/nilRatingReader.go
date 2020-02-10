package rating

// NilRatingReader represents a nil implementation for the RatingReader interface
type nilRatingReader struct {
	startRating uint32
}

// NewNilRatingReader create a new ratingReader that returns fixed values
func NewNilRatingReader(startRating uint32) *nilRatingReader {
	return &nilRatingReader{startRating: startRating}
}

//GetRating gets the rating for the public key
func (nrr *nilRatingReader) GetRating(string) uint32 {
	return nrr.startRating
}

//UpdateRatingFromTempRating sets the new rating to the value of the tempRating
func (nrr *nilRatingReader) UpdateRatingFromTempRating([]string) {
}

//IsInterfaceNil verifies if the interface is nil
func (rr *nilRatingReader) IsInterfaceNil() bool {
	return rr == nil
}
