package common

import (
	"errors"
	"fmt"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/config"
)

var (
	// ErrInvalidHardforkRoundExclusion signals an interval whose start is after its end.
	ErrInvalidHardforkRoundExclusion = errors.New("invalid hardfork round exclusion")
	// ErrOverlappingHardforkRoundExclusions signals overlapping configured intervals.
	ErrOverlappingHardforkRoundExclusions = errors.New("overlapping hardfork round exclusions")
	// ErrRoundExcluded signals that a block or proof belongs to an excluded round.
	ErrRoundExcluded = errors.New("round is excluded")
	// ErrNilRoundExclusionHandler signals that a round exclusion handler is missing.
	ErrNilRoundExclusionHandler = errors.New("nil round exclusion handler")
)

// RoundExclusionHandler reports whether a round is excluded by configuration.
type RoundExclusionHandler interface {
	IsRoundExcluded(round uint64) bool
	IsInterfaceNil() bool
}

type roundExclusionHandler struct {
	intervals []config.HardforkRoundExclusionConfig
}

// NewRoundExclusionHandler creates an immutable round exclusion lookup.
func NewRoundExclusionHandler(intervals []config.HardforkRoundExclusionConfig) (RoundExclusionHandler, error) {
	ordered := append([]config.HardforkRoundExclusionConfig(nil), intervals...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].StartRound < ordered[j].StartRound
	})

	for index, interval := range ordered {
		if interval.StartRound > interval.EndRound {
			return nil, fmt.Errorf("%w: start %d, end %d", ErrInvalidHardforkRoundExclusion, interval.StartRound, interval.EndRound)
		}
		if index > 0 && interval.StartRound <= ordered[index-1].EndRound {
			return nil, fmt.Errorf(
				"%w: [%d, %d] and [%d, %d]",
				ErrOverlappingHardforkRoundExclusions,
				ordered[index-1].StartRound,
				ordered[index-1].EndRound,
				interval.StartRound,
				interval.EndRound,
			)
		}
	}

	return &roundExclusionHandler{intervals: ordered}, nil
}

// ResolveRoundExclusionHandler returns the optional handler or an empty immutable handler.
func ResolveRoundExclusionHandler(handlers ...RoundExclusionHandler) (RoundExclusionHandler, error) {
	if len(handlers) == 0 {
		return NewRoundExclusionHandler(nil)
	}
	if len(handlers) != 1 || check.IfNil(handlers[0]) {
		return nil, ErrNilRoundExclusionHandler
	}

	return handlers[0], nil
}

func (reh *roundExclusionHandler) IsRoundExcluded(round uint64) bool {
	index := sort.Search(len(reh.intervals), func(index int) bool {
		return reh.intervals[index].EndRound >= round
	})

	return index < len(reh.intervals) && reh.intervals[index].StartRound <= round
}

func (reh *roundExclusionHandler) IsInterfaceNil() bool {
	return reh == nil
}
