package metachain

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// GetEpochToUseEpochStartData returns the epoch to use for epoch start data computation
func GetEpochToUseEpochStartData(header data.HeaderHandler) uint32 {
	epochToUse := header.GetEpoch()
	if header.IsHeaderV3() {
		// for meta headers v3 on the epoch change proposed block, the epoch is not yet updated
		epochToUse = header.GetEpoch() + 1
	}
	return epochToUse
}
