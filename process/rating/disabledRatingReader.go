package rating

// DisabledRatingReader represents a nil implementation for the RatingReader interface
type DisabledRatingReader struct {
}

// GetRating gets the rating for the public key
func (*DisabledRatingReader) GetRating(string) uint32 {
	return 1
}

// GetRatings gets all the ratings as a map[pk] ratingValue
func (*DisabledRatingReader) GetRatings(pks []string) map[string]uint32 {
	ratingsMap := make(map[string]uint32)

	for _, val := range pks {
		ratingsMap[val] = 1
	}
	return ratingsMap
}

// IsInterfaceNil verifies if the interface is nil
func (rr *DisabledRatingReader) IsInterfaceNil() bool {
	return rr == nil
}
