package configs_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common/configs"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func defaultConsensusConfigsByRound() []config.ConsensusConfigByRound {
	return []config.ConsensusConfigByRound{
		{
			EnableRound: 0,
			SubroundsTiming: []config.SubroundTiming{
				{StartTime: 0.0, EndTime: 0.05},
				{StartTime: 0.05, EndTime: 0.25},
				{StartTime: 0.25, EndTime: 0.85},
				{StartTime: 0.85, EndTime: 0.95},
			},
			ProcessingThresholdPercent: 85,
		},
	}
}

func TestNewCommonConfigsHandler(t *testing.T) {
	t.Parallel()

	t.Run("should return error for empty config by epoch", func(t *testing.T) {
		t.Parallel()

		pce, err := configs.NewCommonConfigsHandler(nil, []config.EpochStartConfigByRound{}, []config.ConsensusConfigByEpoch{}, defaultConsensusConfigsByRound())
		require.Nil(t, pce)
		require.Equal(t, configs.ErrEmptyCommonConfigsByEpoch, err)
	})

	t.Run("should return error for empty config by round", func(t *testing.T) {
		t.Parallel()

		pce, err := configs.NewCommonConfigsHandler([]config.EpochStartConfigByEpoch{{EnableEpoch: 0}}, nil, []config.ConsensusConfigByEpoch{{EnableEpoch: 0}}, defaultConsensusConfigsByRound())
		require.Nil(t, pce)
		require.Equal(t, configs.ErrEmptyCommonConfigsByRound, err)
	})

	t.Run("should return error for duplicated epoch configs", func(t *testing.T) {
		t.Parallel()

		conf := []config.EpochStartConfigByEpoch{
			{EnableEpoch: 0, GracePeriodRounds: 1},
			{EnableEpoch: 0, GracePeriodRounds: 2},
		}
		pce, err := configs.NewCommonConfigsHandler(conf, []config.EpochStartConfigByRound{}, []config.ConsensusConfigByEpoch{}, defaultConsensusConfigsByRound())
		require.Nil(t, pce)
		require.Equal(t, configs.ErrDuplicatedEpochConfig, err)
	})

	t.Run("should return error for missing epoch 0 config", func(t *testing.T) {
		t.Parallel()

		conf := []config.EpochStartConfigByEpoch{
			{EnableEpoch: 1, GracePeriodRounds: 1},
			{EnableEpoch: 2, GracePeriodRounds: 2},
		}
		pce, err := configs.NewCommonConfigsHandler(conf, []config.EpochStartConfigByRound{}, []config.ConsensusConfigByEpoch{}, defaultConsensusConfigsByRound())
		require.Nil(t, pce)
		require.Equal(t, configs.ErrMissingEpochZeroConfig, err)
	})

	t.Run("should return error for empty consensus configs by round", func(t *testing.T) {
		t.Parallel()

		pce, err := configs.NewCommonConfigsHandler(
			[]config.EpochStartConfigByEpoch{{EnableEpoch: 0}},
			[]config.EpochStartConfigByRound{{EnableRound: 0}},
			[]config.ConsensusConfigByEpoch{{EnableEpoch: 0}},
			nil,
		)
		require.Nil(t, pce)
		require.Equal(t, configs.ErrEmptyConsensusConfigsByRound, err)
	})

	t.Run("should return error for missing round 0 consensus config", func(t *testing.T) {
		t.Parallel()

		pce, err := configs.NewCommonConfigsHandler(
			[]config.EpochStartConfigByEpoch{{EnableEpoch: 0}},
			[]config.EpochStartConfigByRound{{EnableRound: 0}},
			[]config.ConsensusConfigByEpoch{{EnableEpoch: 0}},
			[]config.ConsensusConfigByRound{
				{
					EnableRound: 1,
					SubroundsTiming: []config.SubroundTiming{
						{StartTime: 0.0, EndTime: 0.05},
						{StartTime: 0.05, EndTime: 0.25},
						{StartTime: 0.25, EndTime: 0.85},
						{StartTime: 0.85, EndTime: 0.95},
					},
					ProcessingThresholdPercent: 85,
				},
			},
		)
		require.Nil(t, pce)
		require.Equal(t, configs.ErrMissingRoundZeroConfig, err)
	})

	t.Run("should return error for duplicated round consensus configs", func(t *testing.T) {
		t.Parallel()

		validTiming := []config.SubroundTiming{
			{StartTime: 0.0, EndTime: 0.05},
			{StartTime: 0.05, EndTime: 0.25},
			{StartTime: 0.25, EndTime: 0.85},
			{StartTime: 0.85, EndTime: 0.95},
		}
		pce, err := configs.NewCommonConfigsHandler(
			[]config.EpochStartConfigByEpoch{{EnableEpoch: 0}},
			[]config.EpochStartConfigByRound{{EnableRound: 0}},
			[]config.ConsensusConfigByEpoch{{EnableEpoch: 0}},
			[]config.ConsensusConfigByRound{
				{EnableRound: 0, SubroundsTiming: validTiming, ProcessingThresholdPercent: 85},
				{EnableRound: 0, SubroundsTiming: validTiming, ProcessingThresholdPercent: 85},
			},
		)
		require.Nil(t, pce)
		require.Equal(t, configs.ErrDuplicatedRoundConfig, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		conf := []config.EpochStartConfigByEpoch{
			{EnableEpoch: 0, GracePeriodRounds: 0},
			{EnableEpoch: 2, GracePeriodRounds: 2},
			{EnableEpoch: 1, GracePeriodRounds: 1},
		}
		confByRound := []config.EpochStartConfigByRound{
			{EnableRound: 0, MaxRoundsWithoutCommittedStartInEpochBlock: 10},
			{EnableRound: 1, MaxRoundsWithoutCommittedStartInEpochBlock: 11},
		}
		consensusConf := []config.ConsensusConfigByEpoch{
			{EnableEpoch: 0, NumRoundsToWaitBeforeSignalingChronologyStuck: 10},
			{EnableEpoch: 1, NumRoundsToWaitBeforeSignalingChronologyStuck: 11},
		}

		pce, err := configs.NewCommonConfigsHandler(conf, confByRound, consensusConf, defaultConsensusConfigsByRound())
		require.NotNil(t, pce)
		require.NoError(t, err)
		require.False(t, pce.IsInterfaceNil())

		require.Equal(t, uint32(0), pce.GetOrderedEpochStartConfigByEpoch(0).GracePeriodRounds)
		require.Equal(t, uint32(1), pce.GetOrderedEpochStartConfigByEpoch(1).GracePeriodRounds)
		require.Equal(t, uint32(2), pce.GetOrderedEpochStartConfigByEpoch(2).GracePeriodRounds)
	})
}

func TestCommonConfigsByEpoch_Getters(t *testing.T) {
	t.Parallel()

	conf := []config.EpochStartConfigByEpoch{
		{EnableEpoch: 0, GracePeriodRounds: 10, ExtraDelayForRequestBlockInfoInMilliseconds: 20},
		{EnableEpoch: 1, GracePeriodRounds: 11, ExtraDelayForRequestBlockInfoInMilliseconds: 21},
		{EnableEpoch: 2, GracePeriodRounds: 12, ExtraDelayForRequestBlockInfoInMilliseconds: 22},
	}

	confByRound := []config.EpochStartConfigByRound{
		{EnableRound: 0, MaxRoundsWithoutCommittedStartInEpochBlock: 30},
		{EnableRound: 1, MaxRoundsWithoutCommittedStartInEpochBlock: 31},
	}

	consensusConf := []config.ConsensusConfigByEpoch{
		{EnableEpoch: 0, NumRoundsToWaitBeforeSignalingChronologyStuck: 10},
		{EnableEpoch: 1, NumRoundsToWaitBeforeSignalingChronologyStuck: 11},
	}

	consensusConfByRound := []config.ConsensusConfigByRound{
		{
			EnableRound: 0,
			SubroundsTiming: []config.SubroundTiming{
				{StartTime: 0.0, EndTime: 0.05},
				{StartTime: 0.05, EndTime: 0.25},
				{StartTime: 0.25, EndTime: 0.85},
				{StartTime: 0.85, EndTime: 0.95},
			},
			ProcessingThresholdPercent: 85,
		},
		{
			EnableRound: 10,
			SubroundsTiming: []config.SubroundTiming{
				{StartTime: 0.0, EndTime: 0.05},
				{StartTime: 0.05, EndTime: 0.35},
				{StartTime: 0.35, EndTime: 0.55},
				{StartTime: 0.55, EndTime: 0.95},
			},
			ProcessingThresholdPercent: 85,
		},
	}

	t.Run("get grace period rounds by epoch", func(t *testing.T) {
		t.Parallel()

		cc, _ := configs.NewCommonConfigsHandler(conf, confByRound, consensusConf, consensusConfByRound)

		gracePeriodRounds := cc.GetGracePeriodRoundsByEpoch(0)
		require.Equal(t, uint32(10), gracePeriodRounds)

		gracePeriodRounds = cc.GetGracePeriodRoundsByEpoch(1)
		require.Equal(t, uint32(11), gracePeriodRounds)
	})

	t.Run("get extra delay for request block info", func(t *testing.T) {
		t.Parallel()

		cc, _ := configs.NewCommonConfigsHandler(conf, confByRound, consensusConf, consensusConfByRound)

		extraDelayForRequests := cc.GetExtraDelayForRequestBlockInfoInMs(0)
		require.Equal(t, uint32(20), extraDelayForRequests)

		extraDelayForRequests = cc.GetExtraDelayForRequestBlockInfoInMs(1)
		require.Equal(t, uint32(21), extraDelayForRequests)
	})

	t.Run("get max rounds without committed start in epoch block by round", func(t *testing.T) {
		t.Parallel()

		cc, _ := configs.NewCommonConfigsHandler(conf, confByRound, consensusConf, consensusConfByRound)

		maxRoundsWithoutCommitedStartInEpochBlock := cc.GetMaxRoundsWithoutCommittedStartInEpochBlockInRound(0)
		require.Equal(t, uint32(30), maxRoundsWithoutCommitedStartInEpochBlock)

		maxRoundsWithoutCommitedStartInEpochBlock = cc.GetMaxRoundsWithoutCommittedStartInEpochBlockInRound(1)
		require.Equal(t, uint32(31), maxRoundsWithoutCommitedStartInEpochBlock)
	})

	t.Run("get subrounds timing by round", func(t *testing.T) {
		t.Parallel()

		// subround index constants mirroring bls.SrStartRound=0, SrBlock=1, SrSignature=2, SrEndRound=3
		const srStartRound, srBlock, srSignature = 0, 1, 2

		cc, _ := configs.NewCommonConfigsHandler(conf, confByRound, consensusConf, consensusConfByRound)

		timing := cc.GetSubroundsTimingByRound(0)
		require.Equal(t, consensusConfByRound[0].SubroundsTiming[srStartRound].EndTime, timing.SubroundsTiming[srStartRound].EndTime)
		require.Equal(t, consensusConfByRound[0].SubroundsTiming[srSignature].EndTime, timing.SubroundsTiming[srSignature].EndTime)
		require.Equal(t, consensusConfByRound[0].ProcessingThresholdPercent, timing.ProcessingThresholdPercent)

		timing = cc.GetSubroundsTimingByRound(5)
		require.Equal(t, consensusConfByRound[0].SubroundsTiming[srBlock].EndTime, timing.SubroundsTiming[srBlock].EndTime)
		require.Equal(t, consensusConfByRound[0].ProcessingThresholdPercent, timing.ProcessingThresholdPercent)

		timing = cc.GetSubroundsTimingByRound(10)
		require.Equal(t, consensusConfByRound[1].SubroundsTiming[srBlock].EndTime, timing.SubroundsTiming[srBlock].EndTime)
		require.Equal(t, consensusConfByRound[1].SubroundsTiming[srSignature].EndTime, timing.SubroundsTiming[srSignature].EndTime)
		require.Equal(t, consensusConfByRound[1].ProcessingThresholdPercent, timing.ProcessingThresholdPercent)

		timing = cc.GetSubroundsTimingByRound(999)
		require.Equal(t, consensusConfByRound[1].SubroundsTiming[srBlock].EndTime, timing.SubroundsTiming[srBlock].EndTime)
	})

	t.Run("get active timing boundary round", func(t *testing.T) {
		t.Parallel()

		cc, _ := configs.NewCommonConfigsHandler(conf, confByRound, consensusConf, consensusConfByRound)

		require.Equal(t, uint64(0), cc.GetActiveTimingBoundaryRound(0))
		require.Equal(t, uint64(0), cc.GetActiveTimingBoundaryRound(9))
		require.Equal(t, uint64(10), cc.GetActiveTimingBoundaryRound(10))
		require.Equal(t, uint64(10), cc.GetActiveTimingBoundaryRound(999))
	})
}

func TestCheckConsensusConfigsByRound(t *testing.T) {
	t.Parallel()

	baseEpochConf := []config.EpochStartConfigByEpoch{{EnableEpoch: 0}}
	baseRoundConf := []config.EpochStartConfigByRound{{EnableRound: 0}}
	baseConsensusEpoch := []config.ConsensusConfigByEpoch{{EnableEpoch: 0}}

	validConfig := config.ConsensusConfigByRound{
		EnableRound: 0,
		SubroundsTiming: []config.SubroundTiming{
			{StartTime: 0.0, EndTime: 0.05},
			{StartTime: 0.05, EndTime: 0.25},
			{StartTime: 0.25, EndTime: 0.85},
			{StartTime: 0.85, EndTime: 0.95},
		},
		ProcessingThresholdPercent: 85,
	}

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cc, err := configs.NewCommonConfigsHandler(baseEpochConf, baseRoundConf, baseConsensusEpoch,
			[]config.ConsensusConfigByRound{validConfig})
		require.NoError(t, err)
		require.NotNil(t, cc)
	})

	t.Run("wrong tuple count", func(t *testing.T) {
		t.Parallel()

		bad := config.ConsensusConfigByRound{
			EnableRound: 0,
			SubroundsTiming: []config.SubroundTiming{
				{StartTime: 0.0, EndTime: 0.05},
				{StartTime: 0.05, EndTime: 0.25},
				{StartTime: 0.25, EndTime: 0.85},
				// missing end-round entry
			},
			ProcessingThresholdPercent: 85,
		}
		_, err := configs.NewCommonConfigsHandler(baseEpochConf, baseRoundConf, baseConsensusEpoch,
			[]config.ConsensusConfigByRound{bad})
		require.Equal(t, configs.ErrInvalidSubroundsTimingCount, err)
	})

	t.Run("negative value", func(t *testing.T) {
		t.Parallel()

		bad := config.ConsensusConfigByRound{
			EnableRound: 0,
			SubroundsTiming: []config.SubroundTiming{
				{StartTime: -0.1, EndTime: 0.05}, // negative start
				{StartTime: 0.05, EndTime: 0.25},
				{StartTime: 0.25, EndTime: 0.85},
				{StartTime: 0.85, EndTime: 0.95},
			},
			ProcessingThresholdPercent: 85,
		}
		_, err := configs.NewCommonConfigsHandler(baseEpochConf, baseRoundConf, baseConsensusEpoch,
			[]config.ConsensusConfigByRound{bad})
		require.Equal(t, configs.ErrNegativeSubroundTiming, err)
	})

	t.Run("subround start >= end", func(t *testing.T) {
		t.Parallel()

		bad := config.ConsensusConfigByRound{
			EnableRound: 0,
			SubroundsTiming: []config.SubroundTiming{
				{StartTime: 0.0, EndTime: 0.05},
				{StartTime: 0.25, EndTime: 0.25}, // start == end
				{StartTime: 0.25, EndTime: 0.85},
				{StartTime: 0.85, EndTime: 0.95},
			},
			ProcessingThresholdPercent: 85,
		}
		_, err := configs.NewCommonConfigsHandler(baseEpochConf, baseRoundConf, baseConsensusEpoch,
			[]config.ConsensusConfigByRound{bad})
		require.Equal(t, configs.ErrInvalidSubroundTimingRange, err)
	})

	t.Run("non-contiguous subrounds due to overlap", func(t *testing.T) {
		t.Parallel()

		bad := config.ConsensusConfigByRound{
			EnableRound: 0,
			SubroundsTiming: []config.SubroundTiming{
				{StartTime: 0.0, EndTime: 0.05},
				{StartTime: 0.03, EndTime: 0.25}, // start(1) < end(0): overlaps
				{StartTime: 0.25, EndTime: 0.85},
				{StartTime: 0.85, EndTime: 0.95},
			},
			ProcessingThresholdPercent: 85,
		}
		_, err := configs.NewCommonConfigsHandler(baseEpochConf, baseRoundConf, baseConsensusEpoch,
			[]config.ConsensusConfigByRound{bad})
		require.Equal(t, configs.ErrOverlappingSubroundTiming, err)
	})

	t.Run("non-contiguous subrounds due to gap", func(t *testing.T) {
		t.Parallel()

		bad := config.ConsensusConfigByRound{
			EnableRound: 0,
			SubroundsTiming: []config.SubroundTiming{
				{StartTime: 0.0, EndTime: 0.05},
				{StartTime: 0.10, EndTime: 0.25}, // gap
				{StartTime: 0.25, EndTime: 0.85},
				{StartTime: 0.85, EndTime: 0.95},
			},
			ProcessingThresholdPercent: 85,
		}
		_, err := configs.NewCommonConfigsHandler(baseEpochConf, baseRoundConf, baseConsensusEpoch,
			[]config.ConsensusConfigByRound{bad})
		require.Equal(t, configs.ErrOverlappingSubroundTiming, err)
	})

	t.Run("value >= 1.0", func(t *testing.T) {
		t.Parallel()

		bad := config.ConsensusConfigByRound{
			EnableRound: 0,
			SubroundsTiming: []config.SubroundTiming{
				{StartTime: 0.0, EndTime: 0.05},
				{StartTime: 0.05, EndTime: 0.25},
				{StartTime: 0.25, EndTime: 0.85},
				{StartTime: 0.85, EndTime: 1.0}, // end == 1.0
			},
			ProcessingThresholdPercent: 85,
		}
		_, err := configs.NewCommonConfigsHandler(baseEpochConf, baseRoundConf, baseConsensusEpoch,
			[]config.ConsensusConfigByRound{bad})
		require.Equal(t, configs.ErrSubroundTimingExceedsRound, err)
	})

	t.Run("processing threshold 0", func(t *testing.T) {
		t.Parallel()

		bad := validConfig
		bad.ProcessingThresholdPercent = 0
		_, err := configs.NewCommonConfigsHandler(baseEpochConf, baseRoundConf, baseConsensusEpoch,
			[]config.ConsensusConfigByRound{bad})
		require.Equal(t, configs.ErrInvalidProcessingThreshold, err)
	})

	t.Run("processing threshold > 100", func(t *testing.T) {
		t.Parallel()

		bad := validConfig
		bad.ProcessingThresholdPercent = 101
		_, err := configs.NewCommonConfigsHandler(baseEpochConf, baseRoundConf, baseConsensusEpoch,
			[]config.ConsensusConfigByRound{bad})
		require.Equal(t, configs.ErrInvalidProcessingThreshold, err)
	})
}
