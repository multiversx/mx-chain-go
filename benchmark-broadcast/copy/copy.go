package copy

import "github.com/satori/go.uuid"

// PeersToPeerMap will copy the values from source to destination
func PeersToPeerMap(dest *[]uuid.UUID, src []uuid.UUID) {
	for _, value := range src {
		*dest = append(*dest, value)
	}
}

// PeerPath will construct a peer path
func PeerPath(dest *[]int, src []int, senderId int) {
	for _, value := range src {
		*dest = append(*dest, value)
	}
	*dest = append(*dest, senderId)

}
