package enablers

type roundFlagsHolder struct {
	disableAsyncCallV1   *roundFlag
	supernovaEnableRound *roundFlag
}

// IsDisableAsyncCallV1Enabled returns true for the round where async calls will be disabled in processor v1
func (holder *roundFlagsHolder) IsDisableAsyncCallV1Enabled() bool {
	return holder.disableAsyncCallV1.IsSet()
}

// SupernovaEnableRoundEnabled returns true for the round where supernova will be enabled
func (holder *roundFlagsHolder) SupernovaEnableRoundEnabled() bool {
	return holder.supernovaEnableRound.IsSet()
}

// SupernovaActivationRound returns the round when supernova will be enabled
func (holder *roundFlagsHolder) SupernovaActivationRound() uint64 {
	return holder.supernovaEnableRound.round
}
