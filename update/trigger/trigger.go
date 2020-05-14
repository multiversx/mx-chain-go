package trigger

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update"
)

const hardforkTriggerString = "hardfork trigger"
const dataSeparator = "@"
const hardforkGracePeriod = time.Minute * 5
const epochGracePeriod = 4

var _ facade.HardforkTrigger = (*trigger)(nil)

// ArgHardforkTrigger contains the
type ArgHardforkTrigger struct {
	TriggerPubKeyBytes   []byte
	SelfPubKeyBytes      []byte
	Enabled              bool
	EnabledAuthenticated bool
	ArgumentParser       process.ArgumentsParser
	EpochProvider        update.EpochHandler
}

// trigger implements a hardfork trigger that is able to notify a set list of handlers if this instance gets triggered
// by external events
type trigger struct {
	mutTriggerHandlers     sync.RWMutex
	triggerHandlers        []func()
	triggerPubKey          []byte
	selfPubKey             []byte
	enabled                bool
	enabledAuthenticated   bool
	isTriggerSelf          bool
	mutTriggered           sync.RWMutex
	triggered              bool
	recordedTriggerMessage []byte
	getTimestampHandler    func() int64
	argumentParser         process.ArgumentsParser
	epochProvider          update.EpochHandler
}

// NewTrigger returns the trigger instance
func NewTrigger(arg ArgHardforkTrigger) (*trigger, error) {
	if len(arg.TriggerPubKeyBytes) == 0 {
		return nil, fmt.Errorf("%w hardfork trigger public key bytes length is 0", update.ErrInvalidValue)
	}
	if len(arg.SelfPubKeyBytes) == 0 {
		return nil, fmt.Errorf("%w self public key bytes length is 0", update.ErrInvalidValue)
	}
	if check.IfNil(arg.ArgumentParser) {
		return nil, update.ErrNilArgumentParser
	}
	if check.IfNil(arg.EpochProvider) {
		return nil, update.ErrNilEpochHandler
	}

	t := &trigger{
		triggerHandlers:      make([]func(), 0),
		enabled:              arg.Enabled,
		enabledAuthenticated: arg.EnabledAuthenticated,
		selfPubKey:           arg.SelfPubKeyBytes,
		triggerPubKey:        arg.TriggerPubKeyBytes,
		triggered:            false,
		argumentParser:       arg.ArgumentParser,
		epochProvider:        arg.EpochProvider,
	}
	t.isTriggerSelf = bytes.Equal(arg.TriggerPubKeyBytes, arg.SelfPubKeyBytes)

	t.getTimestampHandler = t.getCurrentUnixTime

	return t, nil
}

func (t *trigger) getCurrentUnixTime() int64 {
	return time.Now().Unix()
}

// Trigger will start of the hardfork process
func (t *trigger) Trigger() error {
	if !t.enabled {
		return update.ErrTriggerNotEnabled
	}

	return t.doTrigger(nil)
}

func (t *trigger) doTrigger(payload []byte) error {
	t.mutTriggered.Lock()
	t.triggered = true
	t.recordedTriggerMessage = payload
	t.mutTriggered.Unlock()

	t.callAddedDataHandlers()

	return nil
}

// TriggerReceived is called whenever a trigger is received from the p2p side
func (t *trigger) TriggerReceived(originalPayload []byte, data []byte, pkBytes []byte) (bool, error) {
	err := t.argumentParser.ParseData(string(data))
	if err != nil {
		return false, nil
	}

	receivedFunction, err := t.argumentParser.GetFunction()
	if err != nil {
		return false, nil
	}

	if receivedFunction != hardforkTriggerString {
		return false, nil
	}

	isTriggerEnabled := t.enabled && t.enabledAuthenticated
	if !isTriggerEnabled {
		//should not return error as to allow the message to get to other peers
		return true, nil
	}

	if !bytes.Equal(pkBytes, t.triggerPubKey) {
		return true, update.ErrTriggerPubKeyMismatch
	}

	arguments, err := t.argumentParser.GetFunctionArguments()
	if err != nil {
		return true, err
	}

	if len(arguments) != 2 {
		return true, update.ErrNotEnoughArgumentsForHardForkTrigger
	}

	timestamp, err := t.getIntFromArgument(string(arguments[0]))
	if err != nil {
		return true, err
	}

	currentTimeStamp := t.getTimestampHandler()
	if timestamp+int64(hardforkGracePeriod.Seconds()) < currentTimeStamp {
		return true, fmt.Errorf("%w timestamp out of grace period message", update.ErrIncorrectHardforkMessage)
	}

	epoch, err := t.getIntFromArgument(string(arguments[1]))
	if err != nil {
		return true, err
	}

	currentEpoch := int64(t.epochProvider.MetaEpoch())
	if currentEpoch-epoch > epochGracePeriod {
		return true, fmt.Errorf("%w epoch out of grace perdiod", update.ErrIncorrectHardforkMessage)
	}

	return true, t.doTrigger(originalPayload)
}

func (t *trigger) getIntFromArgument(value string) (int64, error) {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w, convert error, `%s` is not a valid int",
			update.ErrIncorrectHardforkMessage,
			value,
		)
	}

	return n, nil
}

func (t *trigger) callAddedDataHandlers() {
	t.mutTriggerHandlers.RLock()
	for _, handler := range t.triggerHandlers {
		go handler()
	}
	t.mutTriggerHandlers.RUnlock()
}

// RegisterHandler will add a trigger event handler to the list
func (t *trigger) RegisterHandler(handler func()) error {
	if handler == nil {
		return fmt.Errorf("%w when setting a hardfork trigger", update.ErrNilHandler)
	}

	t.mutTriggerHandlers.Lock()
	t.triggerHandlers = append(t.triggerHandlers, handler)
	t.mutTriggerHandlers.Unlock()

	return nil
}

// IsSelfTrigger returns true if self public key is the trigger public key set in the configs
func (t *trigger) IsSelfTrigger() bool {
	return t.isTriggerSelf
}

// RecordedTriggerMessage returns the trigger message that set the trigger
func (t *trigger) RecordedTriggerMessage() ([]byte, bool) {
	t.mutTriggered.RLock()
	defer t.mutTriggered.RUnlock()

	return t.recordedTriggerMessage, t.triggered
}

// CreateData creates a correct hardfork trigger message based on the identifier and the additional information
func (t *trigger) CreateData(epoch uint32) []byte {
	payload := hardforkTriggerString +
		dataSeparator + hex.EncodeToString([]byte(fmt.Sprintf("%d", t.getTimestampHandler()))) +
		dataSeparator + hex.EncodeToString([]byte(fmt.Sprintf("%d", epoch)))

	return []byte(payload)
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
