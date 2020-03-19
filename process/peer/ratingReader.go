package peer

// RatingReader will handle the fetching of the ratings
type RatingReader struct {
	getRating func(string) uint32
}

//GetRating returns the Rating for the specified public key
func (bsr *RatingReader) GetRating(pk string) uint32 {
	rating := bsr.getRating(pk)
	return rating
}

//IsInterfaceNil checks if the underlying object is nil
func (bsr *RatingReader) IsInterfaceNil() bool {
	return bsr == nil
}
