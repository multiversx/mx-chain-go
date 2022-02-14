package sender

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

const minTimeBetweenSends = time.Second

// argBaseSender represents the arguments for base sender
type argBaseSender struct {
	messenger                 heartbeat.P2PMessenger
	marshaller                marshal.Marshalizer
	topic                     string
	timeBetweenSends          time.Duration
	timeBetweenSendsWhenError time.Duration
}

type baseSender struct {
	timerHandler
	messenger                 heartbeat.P2PMessenger
	marshaller                marshal.Marshalizer
	topic                     string
	timeBetweenSends          time.Duration
	timeBetweenSendsWhenError time.Duration
}

func createBaseSender(args argBaseSender) baseSender {
	return baseSender{
		timerHandler: &timerWrapper{
			timer: time.NewTimer(args.timeBetweenSends),
		},
		messenger:                 args.messenger,
		marshaller:                args.marshaller,
		topic:                     args.topic,
		timeBetweenSends:          args.timeBetweenSends,
		timeBetweenSendsWhenError: args.timeBetweenSendsWhenError,
	}
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

	return nil
}
