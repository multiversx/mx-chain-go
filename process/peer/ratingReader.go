package peer

import "sync"

type RatingReader struct {
	mutReader  sync.Mutex
	getRating  func(string) uint32
	getRatings func([]string) map[string]uint32
}

//GetRating returns the Rating for the specified public key
func (bsr *RatingReader) GetRating(pk string) uint32 {
	bsr.mutReader.Lock()
	rating := bsr.getRating(pk)
	bsr.mutReader.Unlock()
	return rating
}

//GetRatings gets all the ratings that the current rater has
func (bsr *RatingReader) GetRatings(addresses []string) map[string]uint32 {
	bsr.mutReader.Lock()
	ratings := bsr.getRatings(addresses)
	bsr.mutReader.Unlock()
	return ratings
}

//IsInterfaceNil checks if the underlying object is nil
func (bsr *RatingReader) IsInterfaceNil() bool {
	return bsr == nil
}
