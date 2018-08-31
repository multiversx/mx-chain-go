package copy

import "github.com/satori/go.uuid"

func PeersToPeerMap(dest *[]uuid.UUID, src []uuid.UUID) {
	for _, value := range src {
		*dest = append(*dest, value)
	}
}

func PeerPath(dest *[]int, src []int, senderId int) {
	for _, value := range src {
		*dest = append(*dest, value)
	}
	*dest = append(*dest, senderId)

}
