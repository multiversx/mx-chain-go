package configs

const (
	minRoundsToKeepUnprocessedData = uint64(1)
	minFloodingRounds              = 2
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
)
