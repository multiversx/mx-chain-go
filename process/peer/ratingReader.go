package peer

type RatingReader struct {
	getRating  func(string) uint32
	getRatings func([]string) map[string]uint32
}

//GetRating returns the Rating for the specified public key
func (bsr *RatingReader) GetRating(pk string) uint32 {
	rating := bsr.getRating(pk)
	return rating
}

//GetRatings gets all the ratings that the current rater has
func (bsr *RatingReader) GetRatings(addresses []string) map[string]uint32 {
	ratings := bsr.getRatings(addresses)
	return ratings
}

//IsInterfaceNil checks if the underlying object is nil
func (bsr *RatingReader) IsInterfaceNil() bool {
	return bsr == nil
}
