package chaos

// func TestChaosConfig_verify(t *testing.T) {
// 	t.Run("with valid configuration", func(t *testing.T) {
// 		config := &chaosConfig{
// 			SelectedProfileName: "dummy",
// 			Profiles: []chaosProfile{
// 				{
// 					Name: "dummy",
// 				},
// 			},
// 		}

// 		err := config.verify()
// 		require.NoError(t, err)
// 	})

// 	t.Run("with no selected profile", func(t *testing.T) {
// 		config := &chaosConfig{
// 			SelectedProfileName: "",
// 		}

// 		err := config.verify()
// 		require.ErrorContains(t, err, "no selected profile")
// 	})

// 	t.Run("with undefined selected profile", func(t *testing.T) {
// 		config := &chaosConfig{
// 			SelectedProfileName: "dummy",
// 		}

// 		err := config.verify()
// 		require.ErrorContains(t, err, "selected profile does not exist: dummy")
// 	})
// }

// func TestChaosProfile_verify(t *testing.T) {
// 	t.Run("with valid configuration", func(t *testing.T) {
// 		config := &chaosProfile{
// 			Failures: []failureDefinition{
// 				{
// 					Name:     string(failureConsensusV2SkipSendingBlock),
// 					Triggers: []string{"true"},
// 				},
// 			},
// 		}

// 		err := config.verify()
// 		require.NoError(t, err)
// 	})

// 	t.Run("with unknown failure entries", func(t *testing.T) {
// 		config := &chaosProfile{
// 			Failures: []failureDefinition{
// 				{
// 					Name: "unknown",
// 				},
// 			},
// 		}

// 		err := config.verify()
// 		require.ErrorContains(t, err, "unknown failure: unknown")
// 	})

// 	t.Run("with failure entries without triggers", func(t *testing.T) {
// 		config := &chaosProfile{
// 			Failures: []failureDefinition{
// 				{
// 					Name: string(failureConsensusV2SkipSendingBlock),
// 				},
// 			},
// 		}

// 		err := config.verify()
// 		require.ErrorContains(t, err, "failure consensusV2SkipSendingBlock has no triggers")
// 	})

// 	t.Run("with failure entries that require parameters", func(t *testing.T) {
// 		config := &chaosProfile{
// 			Failures: []failureDefinition{
// 				{
// 					Name:     string(failureConsensusV1DelayBroadcastingFinalBlockAsLeader),
// 					Triggers: []string{"true"},
// 				},
// 			},
// 		}

// 		err := config.verify()
// 		require.ErrorContains(t, err, "failure consensusV1DelayBroadcastingFinalBlockAsLeader requires the parameter 'duration'")

// 		config = &chaosProfile{
// 			Failures: []failureDefinition{
// 				{
// 					Name:     string(failureConsensusV2DelayLeaderSignature),
// 					Triggers: []string{"true"},
// 				},
// 			},
// 		}

// 		err = config.verify()
// 		require.ErrorContains(t, err, "failure consensusV2DelayLeaderSignature requires the parameter 'duration'")
// 	})
// }
