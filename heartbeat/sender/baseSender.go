package sender

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

var randomizer = &random.ConcurrentSafeIntRandomizer{}

const minTimeBetweenSends = time.Second
const minThresholdBetweenSends = 0.05 // 5%
const maxThresholdBetweenSends = 1.00 // 100%

// argBaseSender represents the arguments for base sender
type argBaseSender struct {
	messenger                 heartbeat.P2PMessenger
	marshaller                marshal.Marshalizer
	topic                     string
	timeBetweenSends          time.Duration
	timeBetweenSendsWhenError time.Duration
	thresholdBetweenSends     float64
}

type baseSender struct {
	timerHandler
	messenger                 heartbeat.P2PMessenger
	marshaller                marshal.Marshalizer
	topic                     string
	timeBetweenSends          time.Duration
	timeBetweenSendsWhenError time.Duration
	thresholdBetweenSends     float64
}

func createBaseSender(args argBaseSender) baseSender {
	bs := baseSender{
		messenger:                 args.messenger,
		marshaller:                args.marshaller,
		topic:                     args.topic,
		timeBetweenSends:          args.timeBetweenSends,
		timeBetweenSendsWhenError: args.timeBetweenSendsWhenError,
		thresholdBetweenSends:     args.thresholdBetweenSends,
	}
	bs.timerHandler = &timerWrapper{
		timer: time.NewTimer(bs.computeRandomDuration(bs.timeBetweenSends)),
	}

	return bs
}

func checkBaseSenderArgs(args argBaseSender) error {
	if check.IfNil(args.messenger) {
		return heartbeat.ErrNilMessenger
	}
	if check.IfNil(args.marshaller) {
		return heartbeat.ErrNilMarshaller
	}
	if len(args.topic) == 0 {
		return heartbeat.ErrEmptySendTopic
	}
	if args.timeBetweenSends < minTimeBetweenSends {
		return fmt.Errorf("%w for timeBetweenSends", heartbeat.ErrInvalidTimeDuration)
	}
	if args.timeBetweenSendsWhenError < minTimeBetweenSends {
		return fmt.Errorf("%w for timeBetweenSendsWhenError", heartbeat.ErrInvalidTimeDuration)
	}
	if args.thresholdBetweenSends < minThresholdBetweenSends || args.thresholdBetweenSends > maxThresholdBetweenSends {
		return fmt.Errorf("%w for thresholdBetweenSends, receieved %f, min allowed %f, max allowed %f",
			heartbeat.ErrInvalidThreshold, args.thresholdBetweenSends, minThresholdBetweenSends, maxThresholdBetweenSends)
	}

	return nil
}

func (bs *baseSender) computeRandomDuration(baseDuration time.Duration) time.Duration {
	timeBetweenSendsInNano := baseDuration.Nanoseconds()
	maxThreshold := float64(timeBetweenSendsInNano) * bs.thresholdBetweenSends
	randThreshold := randomizer.Intn(int(maxThreshold))

	ret := time.Duration(timeBetweenSendsInNano + int64(randThreshold))
	return ret
}
