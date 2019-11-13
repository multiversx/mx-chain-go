package rating

func (p *BlockSigningRater) SetRatings(newRatings map[string]int64) {
	p.ratings = newRatings
}
