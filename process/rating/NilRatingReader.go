package rating

type NilRatingReader struct {
}

//GetRating gets the rating for the public key
func (*NilRatingReader) GetRating(string) uint32 {
	return 1
}

//GetRatings gets all the ratings as a map[pk] ratingValue
func (*NilRatingReader) GetRatings(pks []string) map[string]uint32 {
	ratingsMap := make(map[string]uint32)

	for _, val := range pks {
		ratingsMap[val] = 1
	}
	return ratingsMap
}

//IsInterfaceNil verifies if the interface is nil
func (rr *NilRatingReader) IsInterfaceNil() bool {
	return rr == nil
}
