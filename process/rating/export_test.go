package rating

func (p *BlockSigningRater) SetRatings(newRatings map[string]uint32) {
	p.ratings = newRatings
}
