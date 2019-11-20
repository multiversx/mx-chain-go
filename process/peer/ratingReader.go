package peer

type RatingReader struct {
	getRating  func(string) uint32
	getRatings func([]string) map[string]uint32
}

//GetRating returns the Rating for the specified public key
func (bsr *RatingReader) GetRating(pk string) uint32 {
	return bsr.getRating(pk)
}

//GetRatings gets all the ratings that the current rater has
func (bsr *RatingReader) GetRatings(addresses []string) map[string]uint32 {
	return bsr.getRatings(addresses)
}
