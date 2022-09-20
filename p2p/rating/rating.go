package rating

import (
	"github.com/ElrondNetwork/elrond-go-p2p/rating"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// ArgPeersRatingHandler is the DTO used to create a new peers rating handler
type ArgPeersRatingHandler = rating.ArgPeersRatingHandler

// NewPeersRatingHandler returns a new peers rating handler
func NewPeersRatingHandler(args ArgPeersRatingHandler) (p2p.PeersRatingHandler, error) {
	return rating.NewPeersRatingHandler(args)
}
