package detector

import "github.com/ElrondNetwork/elrond-go/sharding"

func DoubleSigners(
	group1 []sharding.Validator,
	group2 []sharding.Validator,
	bitmap1 []byte,
	bitmap2 []byte,
) []sharding.Validator {
	return doubleSigners(group1, group2, bitmap1, bitmap2)
}
