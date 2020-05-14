package trigger_test

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/trigger"
	"github.com/stretchr/testify/assert"
)

func createMockArgHardforkTrigger() trigger.ArgHardforkTrigger {
	return trigger.ArgHardforkTrigger{
		TriggerPubKeyBytes:   []byte("trigger"),
		SelfPubKeyBytes:      []byte("self"),
		Enabled:              true,
		EnabledAuthenticated: true,
	}
}

func TestNewTrigger_EmptyTriggerPubKeyBytesShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	arg.TriggerPubKeyBytes = nil
	trig, err := trigger.NewTrigger(arg)

	assert.True(t, errors.Is(err, update.ErrInvalidValue))
	assert.True(t, check.IfNil(trig))
}

func TestNewTrigger_EmptySelfPubKeyBytesShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	arg.SelfPubKeyBytes = nil
	trig, err := trigger.NewTrigger(arg)

	assert.True(t, errors.Is(err, update.ErrInvalidValue))
	assert.True(t, check.IfNil(trig))
}

func TestNewTrigger_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	trig, err := trigger.NewTrigger(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(trig))
}

//------- Trigger

func TestTrigger_TriggerNotEnabledShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	arg.Enabled = false
	trig, _ := trigger.NewTrigger(arg)

	err := trig.Trigger()
	assert.Equal(t, update.ErrTriggerNotEnabled, err)

	_, wasTriggered := trig.RecordedTriggerMessage()
	assert.False(t, wasTriggered)
}

func TestTrigger_TriggerEnabledShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	trig, _ := trigger.NewTrigger(arg)
	numTrigCalled := int32(0)
	_ = trig.RegisterHandler(func() {
		atomic.AddInt32(&numTrigCalled, 1)
	})

	payload, wasTriggered := trig.RecordedTriggerMessage()
	assert.Nil(t, payload)
	assert.False(t, wasTriggered)

	err := trig.Trigger()

	// delay as to execute the async calls
	time.Sleep(time.Second)

	payload, wasTriggered = trig.RecordedTriggerMessage()

	assert.Nil(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&numTrigCalled))
	assert.Nil(t, payload)
	assert.True(t, wasTriggered)
}

func TestTrigger_TriggerReceivedNotAHardforkTriggerRetNilButNotCall(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	arg.Enabled = false
	trig, _ := trigger.NewTrigger(arg)

	isHardfork, err := trig.TriggerReceived(nil, make([]byte, 0), nil)
	assert.Nil(t, err)
	assert.False(t, isHardfork)

	_, wasTriggered := trig.RecordedTriggerMessage()
	assert.False(t, wasTriggered)
}

func TestTrigger_TriggerReceivedNotEnabledShouldRetNilButNotCall(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	arg.Enabled = false
	trig, _ := trigger.NewTrigger(arg)
	data := []byte(trigger.HardforkTriggerString + trigger.PayloadSeparator + fmt.Sprintf("%d", 0))

	isHardfork, err := trig.TriggerReceived(nil, data, nil)
	assert.Nil(t, err)
	assert.True(t, isHardfork)

	_, wasTriggered := trig.RecordedTriggerMessage()
	assert.False(t, wasTriggered)
}

func TestTrigger_TriggerReceivedNotEnabledAuthenticatedShouldRetNilButNotCall(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	arg.EnabledAuthenticated = false
	trig, _ := trigger.NewTrigger(arg)
	data := []byte(trigger.HardforkTriggerString + trigger.PayloadSeparator + fmt.Sprintf("%d", 0))

	isHardfork, err := trig.TriggerReceived(nil, data, nil)
	assert.Nil(t, err)
	assert.True(t, isHardfork)

	_, wasTriggered := trig.RecordedTriggerMessage()
	assert.False(t, wasTriggered)
}

func TestTrigger_TriggerReceivedPubkeysMismatchShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	trig, _ := trigger.NewTrigger(arg)
	data := []byte(trigger.HardforkTriggerString + trigger.PayloadSeparator + fmt.Sprintf("%d", 0))

	pubkey := []byte("invalid pubkey")
	isHardfork, err := trig.TriggerReceived(nil, data, pubkey)
	assert.Equal(t, update.ErrTriggerPubKeyMismatch, err)
	assert.True(t, isHardfork)

	_, wasTriggered := trig.RecordedTriggerMessage()
	assert.False(t, wasTriggered)
}

func TestTrigger_TriggerReceivedNotEnoughTokensShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	trig, _ := trigger.NewTrigger(arg)
	data := []byte(trigger.HardforkTriggerString)

	isHardfork, err := trig.TriggerReceived(nil, data, arg.TriggerPubKeyBytes)
	assert.True(t, errors.Is(err, update.ErrIncorrectHardforkMessage))
	assert.True(t, isHardfork)

	_, wasTriggered := trig.RecordedTriggerMessage()
	assert.False(t, wasTriggered)
}

func TestTrigger_TriggerReceivedNotAnIntShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	trig, _ := trigger.NewTrigger(arg)
	data := []byte(trigger.HardforkTriggerString + trigger.PayloadSeparator + "not-an-int")

	isHardfork, err := trig.TriggerReceived(nil, data, arg.TriggerPubKeyBytes)
	assert.True(t, errors.Is(err, update.ErrIncorrectHardforkMessage))
	assert.True(t, isHardfork)

	_, wasTriggered := trig.RecordedTriggerMessage()
	assert.False(t, wasTriggered)
}

func TestTrigger_TriggerReceivedOutOfGracePeriodShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	trig, _ := trigger.NewTrigger(arg)
	currentTimeStamp := time.Now().Unix()
	trig.SetTimeHandler(func() int64 {
		return currentTimeStamp
	})
	messageTimeStamp := currentTimeStamp - int64(trigger.HardforkGracePeriod.Seconds()) - 1
	data := []byte(trigger.HardforkTriggerString + trigger.PayloadSeparator + fmt.Sprintf("%d", messageTimeStamp))

	isHardfork, err := trig.TriggerReceived(nil, data, arg.TriggerPubKeyBytes)
	assert.True(t, errors.Is(err, update.ErrIncorrectHardforkMessage))
	assert.True(t, isHardfork)

	_, wasTriggered := trig.RecordedTriggerMessage()
	assert.False(t, wasTriggered)
}

func TestTrigger_TriggerReceivedShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	trig, _ := trigger.NewTrigger(arg)
	numTrigCalled := int32(0)
	payloadReceived := []byte("original message")
	_ = trig.RegisterHandler(func() {
		atomic.AddInt32(&numTrigCalled, 1)
	})
	currentTimeStamp := time.Now().Unix()
	trig.SetTimeHandler(func() int64 {
		return currentTimeStamp
	})
	messageTimeStamp := currentTimeStamp - int64(trigger.HardforkGracePeriod.Seconds())
	data := []byte(trigger.HardforkTriggerString + trigger.PayloadSeparator + fmt.Sprintf("%d", messageTimeStamp))

	payload, wasTriggered := trig.RecordedTriggerMessage()
	assert.Nil(t, payload)
	assert.False(t, wasTriggered)

	isHardfork, err := trig.TriggerReceived(payloadReceived, data, arg.TriggerPubKeyBytes)
	assert.True(t, isHardfork)

	// delay as to execute the async calls
	time.Sleep(time.Second)

	payload, wasTriggered = trig.RecordedTriggerMessage()

	assert.Nil(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&numTrigCalled))
	assert.Equal(t, payloadReceived, payload)
	assert.True(t, wasTriggered)
}

func TestTrigger_TriggerReceivedCreatePayloadShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	trig, _ := trigger.NewTrigger(arg)

	data := trig.CreateData(0)
	numTrigCalled := int32(0)
	payloadReceived := []byte("original message")
	_ = trig.RegisterHandler(func() {
		atomic.AddInt32(&numTrigCalled, 1)
	})

	isHardfork, err := trig.TriggerReceived(payloadReceived, data, arg.TriggerPubKeyBytes)
	assert.True(t, isHardfork)
	assert.Nil(t, err)
}

//------- RegisterHandler

func TestTrigger_RegisterHandlerNilHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	trig, _ := trigger.NewTrigger(arg)

	err := trig.RegisterHandler(nil)

	assert.True(t, errors.Is(err, update.ErrNilHandler))
}

func TestTrigger_RegisterHandlerShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgHardforkTrigger()
	trig, _ := trigger.NewTrigger(arg)

	err := trig.RegisterHandler(func() {})

	assert.Nil(t, err)
	assert.Equal(t, 1, len(trig.RegisteredHandlers()))
}

//------- IsSelfTrigger

func TestTrigger_IsSelfTrigger(t *testing.T) {
	t.Parallel()

	arg1 := createMockArgHardforkTrigger()
	trig1, _ := trigger.NewTrigger(arg1)

	assert.False(t, trig1.IsSelfTrigger())

	arg2 := createMockArgHardforkTrigger()
	arg2.SelfPubKeyBytes = arg2.TriggerPubKeyBytes
	trig2, _ := trigger.NewTrigger(arg2)

	assert.True(t, trig2.IsSelfTrigger())
}
