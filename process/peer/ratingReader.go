package peer

// RatingReader will handle the fetching of the ratings
type RatingReader struct {
	getRating                  func(string) uint32
	updateRatingFromTempRating func([]string) error
}

//GetRating returns the Rating for the specified public key
func (bsr *RatingReader) GetRating(pk string) uint32 {
	rating := bsr.getRating(pk)
	return rating
}

//UpdateRatingFromTempRating returns the TempRating for the specified public key
func (bsr *RatingReader) UpdateRatingFromTempRating(pks []string) error {
	return bsr.updateRatingFromTempRating(pks)
}

//IsInterfaceNil checks if the underlying object is nil
func (bsr *RatingReader) IsInterfaceNil() bool {
	return bsr == nil
}
