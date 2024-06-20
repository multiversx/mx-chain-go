package latestData

type sovereignLatestDataProvider struct {
	*latestDataProvider
}

func NewSovereignLatestDataProvider(args ArgsLatestDataProvider) (*sovereignLatestDataProvider, error) {
	ldp, err := newLatestDataProvider(args, newSovereignEpochStartRoundLoader())
	if err != nil {
		return nil, err
	}

	return &sovereignLatestDataProvider{
		ldp,
	}, nil
}
