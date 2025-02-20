package latestData

type sovereignLatestDataProvider struct {
	*latestDataProvider
}

// NewSovereignLatestDataProvider creates a latest storage data provider for sovereign chain
func NewSovereignLatestDataProvider(args ArgsLatestDataProvider) (*sovereignLatestDataProvider, error) {
	ldp, err := newLatestDataProvider(args, newSovereignEpochStartRoundLoader())
	if err != nil {
		return nil, err
	}
	return &sovereignLatestDataProvider{
		ldp,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sldp *sovereignLatestDataProvider) IsInterfaceNil() bool {
	return sldp == nil
}
