package testscommon

// AddressBlacklistCheckerStub -
type AddressBlacklistCheckerStub struct {
	IsBlacklistedCalled func(addrBytes []byte) bool
}

// IsBlacklisted returns true if the address is blacklisted
func (abcs *AddressBlacklistCheckerStub) IsBlacklisted(addrBytes []byte) bool {
	if abcs.IsBlacklistedCalled != nil {
		return abcs.IsBlacklistedCalled(addrBytes)
	}
	return false
}

// IsInterfaceNil returns true if the receiver is nil
func (abcs *AddressBlacklistCheckerStub) IsInterfaceNil() bool {
	return abcs == nil
}
