package configs

import "time"

const (
	minRoundsToKeepUnprocessedData = uint64(1)
	minBlockProcessingTimeMs       = uint32(1)
	// the fallback prefetch patience must always be set explicitly, a zero would request on sight
	minExtraDelayForRequestBlockInfoMs = uint32(1)
)

const (
	defaultMaxMetaNoncesBehind                    = 15
	defaultMaxMetaNoncesBehindForGlobalStuck      = 30
	defaultMaxShardNoncesBehind                   = 15
	defaultMaxRoundsWithoutNewBlockReceived       = 10
	defaultMaxRoundsWithoutCommittedBlock         = 10
	defaultRoundModulusTriggerWhenSyncIsStuck     = 20
	defaultMaxSyncWithErrorsAllowed               = 20
	defaultMaxRoundsToKeepUnprocessedMiniBlocks   = 3000
	defaultMaxRoundsToKeepUnprocessedTransactions = 3000
	defaultMaxConsecutiveRoundsOfRatingDecrease   = 600
	defaultMaxRoundsOfInactivityAccepted          = 3
	defaultMaxBlockProcessingTimeMs               = 100
	defaultNumHeadersToRequestInAdvance           = 10
)

// defaults for the block data propagation delays, expressed as durations so that they can be
// returned directly by the corresponding getters
const (
	defaultExtraDelayForBroadcastBlockInfo     = 1000 * time.Millisecond
	defaultExtraDelayBetweenBroadcastMbsAndTxs = 1000 * time.Millisecond
	defaultExtraDelayForRequestBlockInfo       = 3000 * time.Millisecond
)
