package trigger

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update"
)

const hardforkTriggerString = "hardfork trigger"
const dataSeparator = "@"
const hardforkGracePeriod = time.Minute * 5
const epochGracePeriod = 4
const minTimeToWaitAfterHardforkInMinutes = 2
const minimumEpochForHarfork = 1
const deltaRoundsForForcedEpoch = uint64(10)
const disabledRoundForForceEpochStart = uint64(math.MaxUint64)

var _ facade.HardforkTrigger = (*trigger)(nil)
var log = logger.GetOrCreate("update/trigger")

// ArgHardforkTrigger contains the
type ArgHardforkTrigger struct {
	Enabled                   bool
	EnabledAuthenticated      bool
	CloseAfterExportInMinutes uint32
	TriggerPubKeyBytes        []byte
	SelfPubKeyBytes           []byte
	ArgumentParser            process.ArgumentsParser
	EpochProvider             update.EpochHandler
	ExportFactoryHandler      update.ExportFactoryHandler
	ChanStopNodeProcess       chan endProcess.ArgEndProcess
	EpochConfirmedNotifier    update.EpochChangeConfirmedNotifier
	ImportStartHandler        update.ImportStartHandler
	RoundHandler              update.RoundHandler
}

// trigger implements a hardfork trigger that is able to notify a set list of handlers if this instance gets triggered
// by external events
type trigger struct {
	enabled                      bool
	enabledAuthenticated         bool
	shouldTriggerFromEpochChange bool
	isTriggerSelf                bool
	triggerReceived              bool
	triggerExecuting             bool
	epoch                        uint32
	round                        uint64
	closeAfterInMinutes          uint32
	triggerPubKey                []byte
	selfPubKey                   []byte
	mutTriggered                 sync.RWMutex
	recordedTriggerMessage       []byte
	getTimestampHandler          func() int64
	argumentParser               process.ArgumentsParser
	epochProvider                update.EpochHandler
	exportFactoryHandler         update.ExportFactoryHandler
	chanStopNodeProcess          chan endProcess.ArgEndProcess
	mutClosers                   sync.RWMutex
	closers                      []update.Closer
	chanTriggerReceived          chan struct{}
	importStartHandler           update.ImportStartHandler
	isWithEarlyEndOfEpoch        bool
	roundHandler                 update.RoundHandler
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
	if check.IfNil(arg.ExportFactoryHandler) {
		return nil, update.ErrNilExportFactoryHandler
	}
	if arg.ChanStopNodeProcess == nil {
		return nil, update.ErrNilChanStopNodeProcess
	}
	if check.IfNil(arg.EpochConfirmedNotifier) {
		return nil, update.ErrNilEpochConfirmedNotifier
	}
	if arg.CloseAfterExportInMinutes < minTimeToWaitAfterHardforkInMinutes {
		return nil, fmt.Errorf("%w, minimum time to wait in minutes: %d",
			update.ErrInvalidTimeToWaitAfterHardfork,
			minTimeToWaitAfterHardforkInMinutes,
		)
	}
	if check.IfNil(arg.ImportStartHandler) {
		return nil, update.ErrNilImportStartHandler
	}
	if check.IfNil(arg.RoundHandler) {
		return nil, fmt.Errorf("%w in update.NewTrigger", update.ErrNilRoundHandler)
	}

	t := &trigger{
		enabled:              arg.Enabled,
		enabledAuthenticated: arg.EnabledAuthenticated,
		selfPubKey:           arg.SelfPubKeyBytes,
		triggerPubKey:        arg.TriggerPubKeyBytes,
		triggerReceived:      false,
		triggerExecuting:     false,
		argumentParser:       arg.ArgumentParser,
		epochProvider:        arg.EpochProvider,
		exportFactoryHandler: arg.ExportFactoryHandler,
		closeAfterInMinutes:  arg.CloseAfterExportInMinutes,
		chanStopNodeProcess:  arg.ChanStopNodeProcess,
		closers:              make([]update.Closer, 0),
		chanTriggerReceived:  make(chan struct{}, 1), //buffer with one value as there might be async calls
		importStartHandler:   arg.ImportStartHandler,
		roundHandler:         arg.RoundHandler,
	}

	t.isTriggerSelf = bytes.Equal(arg.TriggerPubKeyBytes, arg.SelfPubKeyBytes)
	t.getTimestampHandler = t.getCurrentUnixTime
	arg.EpochConfirmedNotifier.RegisterForEpochChangeConfirmed(t.epochConfirmed)

	return t, nil
}

func (t *trigger) getCurrentUnixTime() int64 {
	return time.Now().Unix()
}

func (t *trigger) epochConfirmed(epoch uint32) {
	if !t.enabled {
		return
	}

	shouldStartHardfork := t.computeTriggerStartOfEpoch(epoch)
	if !shouldStartHardfork {
		return
	}

	t.doTrigger()
}

func (t *trigger) computeTriggerStartOfEpoch(receivedTrigger uint32) bool {
	t.mutTriggered.Lock()
	defer t.mutTriggered.Unlock()

	if !t.triggerReceived {
		return false
	}
	if t.triggerExecuting {
		return false
	}
	if receivedTrigger < t.epoch {
		return false
	}

	t.triggerExecuting = true
	return true
}

// Trigger will start the hardfork process
func (t *trigger) Trigger(epoch uint32, withEarlyEndOfEpoch bool) error {
	if !t.enabled {
		return update.ErrTriggerNotEnabled
	}

	round := t.computeHardforkRound(withEarlyEndOfEpoch)
	logInfo := []interface{}{
		"epoch", epoch, "withEarlyEndOfEpoch", withEarlyEndOfEpoch,
	}
	if withEarlyEndOfEpoch {
		logInfo = append(logInfo, []interface{}{
			"round",
			round,
		}...)
	}
	log.Info("hardfork trigger", logInfo...)
	t.isWithEarlyEndOfEpoch = withEarlyEndOfEpoch
	t.round = round

	if epoch < minimumEpochForHarfork {
		return fmt.Errorf("%w, minimum epoch accepted is %d", update.ErrInvalidEpoch, minimumEpochForHarfork)
	}

	shouldTrigger, err := t.computeAndSetTrigger(epoch, nil, withEarlyEndOfEpoch, round) //original payload is nil because this node is the originator
	if err != nil {
		return err
	}
	if !shouldTrigger {
		log.Debug("hardfork won't trigger now, will wait for epoch change")

		return nil
	}

	t.doTrigger()

	return nil
}

func (t *trigger) computeHardforkRound(withEarlyEndOfEpoch bool) uint64 {
	if !withEarlyEndOfEpoch {
		return disabledRoundForForceEpochStart
	}

	currentRound := t.roundHandler.Index()
	if currentRound < 0 {
		//do not overflow on uint64 when current round is negative
		return deltaRoundsForForcedEpoch
	}

	return uint64(currentRound) + deltaRoundsForForcedEpoch
}

// computeAndSetTrigger needs to do 2 things atomically: set the original payload and epoch and determine if the trigger
// can be called
func (t *trigger) computeAndSetTrigger(epoch uint32, originalPayload []byte, withEarlyEndOfEpoch bool, round uint64) (bool, error) {
	t.mutTriggered.Lock()
	defer t.mutTriggered.Unlock()

	t.triggerReceived = true
	if t.triggerExecuting {
		return false, update.ErrTriggerAlreadyInAction
	}
	t.epoch = epoch
	if len(originalPayload) > 0 {
		t.recordedTriggerMessage = originalPayload
	}

	if withEarlyEndOfEpoch {
		t.epochProvider.ForceEpochStart(round)
	}

	if len(originalPayload) == 0 {
		t.writeOnNotifyChan()
	}

	shouldSetTriggerFromEpochChange := epoch > t.epochProvider.MetaEpoch()
	if shouldSetTriggerFromEpochChange {
		t.shouldTriggerFromEpochChange = true
		return false, nil
	}

	if t.shouldTriggerFromEpochChange {
		return false, nil
	}

	t.triggerExecuting = true

	return true, nil
}

func (t *trigger) writeOnNotifyChan() {
	//writing on the notification chan should not be blocking as to allow self to initiate the hardfork process
	select {
	case t.chanTriggerReceived <- struct{}{}:
	default:
	}
}

func (t *trigger) doTrigger() {
	t.callClose()
	t.exportAll()
}

func (t *trigger) exportAll() {
	t.mutTriggered.Lock()
	defer t.mutTriggered.Unlock()

	log.Debug("hardfork trigger exportAll called")

	epoch := t.epoch

	go func() {
		exportHandler, err := t.exportFactoryHandler.Create()
		if err != nil {
			log.Error("error while creating export handler", "error", err)
			return
		}

		log.Info("started hardFork export process")
		err = exportHandler.ExportAll(epoch)
		if err != nil {
			log.Error("error while exporting data", "error", err)
			return
		}
		log.Info("finished hardFork export process")
		errNotCritical := t.importStartHandler.SetStartImport()
		if errNotCritical != nil {
			log.Error("error setting the node to start the import after the restart",
				"error", errNotCritical)
		}

		wait := time.Duration(t.closeAfterInMinutes) * time.Minute
		log.Info("node will still be active for", "time duration", wait)

		time.Sleep(wait)
		argument := endProcess.ArgEndProcess{
			Reason:      "HardForkExport",
			Description: "Node finished the export process with success",
		}
		t.chanStopNodeProcess <- argument
	}()
}

// TriggerReceived is called whenever a trigger is received from the p2p side
func (t *trigger) TriggerReceived(originalPayload []byte, data []byte, pkBytes []byte) (bool, error) {
	receivedFunction, arguments, err := t.argumentParser.ParseCallData(string(data))
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

	if len(arguments) < 2 {
		return true, update.ErrIncorrectHardforkMessage
	}

	timestamp, err := t.getIntFromArgument(string(arguments[0]))
	if err != nil {
		return true, err
	}

	currentTimeStamp := t.getTimestampHandler()
	if timestamp+int64(hardforkGracePeriod.Seconds()) < currentTimeStamp {
		return true, fmt.Errorf("%w message timestamp out of grace period message", update.ErrIncorrectHardforkMessage)
	}

	epoch, err := t.getIntFromArgument(string(arguments[1]))
	if err != nil {
		return true, err
	}
	if epoch < minimumEpochForHarfork {
		return true, fmt.Errorf("%w, minimum epoch accepted is %d", update.ErrInvalidEpoch, minimumEpochForHarfork)
	}

	earlyEndOfEpochRound := disabledRoundForForceEpochStart
	withEarlyEndOfEpoch := false
	if len(arguments) == 4 {
		withEarlyEndOfEpoch = t.getBoolFromArgument(string(arguments[2]))
		earlyEndOfEpochRound, err = t.getUintFromArgument(string(arguments[3]))
		if err != nil {
			return true, err
		}
	}

	currentEpoch := int64(t.epochProvider.MetaEpoch())
	if currentEpoch-epoch > epochGracePeriod {
		return true, fmt.Errorf("%w epoch out of grace period", update.ErrIncorrectHardforkMessage)
	}

	shouldTrigger, err := t.computeAndSetTrigger(uint32(epoch), originalPayload, withEarlyEndOfEpoch, earlyEndOfEpochRound)
	if err != nil {
		log.Debug("received trigger", "status", err)
		return true, nil
	}
	if !shouldTrigger {
		return true, nil
	}

	t.doTrigger()

	return true, nil
}

func (t *trigger) callClose() {
	log.Debug("calling close on all registered instances")

	t.mutClosers.RLock()
	for _, c := range t.closers {
		err := c.Close()
		if err != nil {
			log.Warn("error closing registered instance", "error", err)
		}
	}
	t.mutClosers.RUnlock()
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

func (t *trigger) getUintFromArgument(value string) (uint64, error) {
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w, convert error, `%s` is not a valid int",
			update.ErrIncorrectHardforkMessage,
			value,
		)
	}

	return n, nil
}

func (t *trigger) getBoolFromArgument(value string) bool {
	upperTrue := strings.ToUpper(fmt.Sprintf("%v", true))
	return strings.ToUpper(value) == upperTrue
}

// IsSelfTrigger returns true if self public key is the trigger public key set in the configs
func (t *trigger) IsSelfTrigger() bool {
	return t.isTriggerSelf
}

// RecordedTriggerMessage returns the trigger message that set the trigger
func (t *trigger) RecordedTriggerMessage() ([]byte, bool) {
	t.mutTriggered.RLock()
	defer t.mutTriggered.RUnlock()

	return t.recordedTriggerMessage, t.triggerReceived
}

// CreateData creates a correct hardfork trigger message based on the identifier and the additional information
func (t *trigger) CreateData() []byte {
	t.mutTriggered.RLock()
	payload := hardforkTriggerString +
		dataSeparator + hex.EncodeToString([]byte(fmt.Sprintf("%d", t.getTimestampHandler()))) +
		dataSeparator + hex.EncodeToString([]byte(fmt.Sprintf("%d", t.epoch))) +
		dataSeparator + hex.EncodeToString([]byte(fmt.Sprintf("%v", t.isWithEarlyEndOfEpoch))) +
		dataSeparator + hex.EncodeToString([]byte(fmt.Sprintf("%d", t.round)))
	t.mutTriggered.RUnlock()

	return []byte(payload)
}

// AddCloser will add a closer interface on the existing list
func (t *trigger) AddCloser(closer update.Closer) error {
	if check.IfNil(closer) {
		return update.ErrNilCloser
	}

	t.mutClosers.Lock()
	t.closers = append(t.closers, closer)
	t.mutClosers.Unlock()

	return nil
}

// NotifyTriggerReceived will write a struct{}{} on the provided channel as soon as a trigger is received
// this is done to decrease the latency of the heartbeat sending system
func (t *trigger) NotifyTriggerReceived() <-chan struct{} {
	return t.chanTriggerReceived
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
