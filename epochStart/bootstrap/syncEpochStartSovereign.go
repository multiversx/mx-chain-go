package bootstrap

type epochStartSovereignSyncer struct {
	*epochStartMetaSyncer
}

// internal constructor
func newEpochStartSovereignSyncer(args ArgsNewEpochStartMetaSyncer) (*epochStartSovereignSyncer, error) {
	baseSyncer, err := newEpochStartMetaSyncer(args)
	if err != nil {
		return nil, err
	}

	topicProvider := &sovereignTopicProvider{}
	baseSyncer.epochStartTopicProviderHandler = topicProvider
	baseSyncer.singleDataInterceptor, err = createSingleDataInterceptor(args, topicProvider.getTopic())
	if err != nil {
		return nil, err
	}

	return &epochStartSovereignSyncer{
		epochStartMetaSyncer: baseSyncer,
	}, nil
}
