package detector

import "github.com/ElrondNetwork/elrond-go/sharding"

func (ssd *SigningSlashingDetector) DoubleSigners(
	group1 []sharding.Validator,
	group2 []sharding.Validator,
	bitmap1 []byte,
	bitmap2 []byte,
) []sharding.Validator {
	return ssd.doubleSigners(0, group1, group2, bitmap1, bitmap2)
}
