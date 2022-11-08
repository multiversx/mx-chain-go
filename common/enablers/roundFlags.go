package enablers

type roundFlagsHolder struct {
	disableAsyncCallV1 *roundFlag
}

// IsDisableAsyncCallV1Enabled returns true for the round where async calls will be disabled in processor v1
func (holder *roundFlagsHolder) IsDisableAsyncCallV1Enabled() bool {
	return holder.disableAsyncCallV1.IsSet()
}
