package sender

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/heartbeat"
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
	redundancyHandler         heartbeat.NodeRedundancyHandler
	privKey                   crypto.PrivateKey
}

type baseSender struct {
	timerHandler
	messenger                 heartbeat.P2PMessenger
	marshaller                marshal.Marshalizer
	topic                     string
	timeBetweenSends          time.Duration
	timeBetweenSendsWhenError time.Duration
	thresholdBetweenSends     float64
	redundancy                heartbeat.NodeRedundancyHandler
	privKey                   crypto.PrivateKey
	publicKey                 crypto.PublicKey
	observerPublicKey         crypto.PublicKey
}

func createBaseSender(args argBaseSender) baseSender {
	bs := baseSender{
		messenger:                 args.messenger,
		marshaller:                args.marshaller,
		topic:                     args.topic,
		timeBetweenSends:          args.timeBetweenSends,
		timeBetweenSendsWhenError: args.timeBetweenSendsWhenError,
		thresholdBetweenSends:     args.thresholdBetweenSends,
		redundancy:                args.redundancyHandler,
		privKey:                   args.privKey,
		publicKey:                 args.privKey.GeneratePublic(),
		observerPublicKey:         args.redundancyHandler.ObserverPrivateKey().GeneratePublic(),
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
		return fmt.Errorf("%w for thresholdBetweenSends, received %f, min allowed %f, max allowed %f",
			heartbeat.ErrInvalidThreshold, args.thresholdBetweenSends, minThresholdBetweenSends, maxThresholdBetweenSends)
	}
	if check.IfNil(args.privKey) {
		return heartbeat.ErrNilPrivateKey
	}
	if check.IfNil(args.redundancyHandler) {
		return heartbeat.ErrNilRedundancyHandler
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

func (bs *baseSender) getCurrentPrivateAndPublicKeys() (crypto.PrivateKey, crypto.PublicKey) {
	shouldUseOriginalKeys := !bs.redundancy.IsRedundancyNode() || (bs.redundancy.IsRedundancyNode() && !bs.redundancy.IsMainMachineActive())
	if shouldUseOriginalKeys {
		return bs.privKey, bs.publicKey
	}

	return bs.redundancy.ObserverPrivateKey(), bs.observerPublicKey
}
