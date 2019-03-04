package loadBalancer_test

import (
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var errLenDifferent = errors.New("len different for names and chans")
var errLenDifferentNamesChans = errors.New("len different for names and chans")
var errMissingChannel = errors.New("missing channel")
var errChannelsMismatch = errors.New("channels mismatch")
var durationWait = time.Second * 2

func checkIntegrity(oclb *loadBalancer.OutgoingChannelLoadBalancer, name string) error {
	if len(oclb.Names()) != len(oclb.Chans()) {
		return errLenDifferent
	}

	if len(oclb.Names()) != len(oclb.NamesChans()) {
		return errLenDifferentNamesChans
	}

	idxFound := -1
	for i, n := range oclb.Names() {
		if n == name {
			idxFound = i
			break
		}
	}

	if idxFound == -1 && oclb.NamesChans()[name] == nil {
		return errMissingChannel
	}

	if oclb.NamesChans()[name] != oclb.Chans()[idxFound] {
		return errChannelsMismatch
	}

	return nil
}

//------- NewOutgoingChannelLoadBalancer

func TestNewOutgoingChannelLoadBalancer_ShouldNotProduceNil(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	assert.NotNil(t, oclb)
}

func TestNewOutgoingChannelLoadBalancer_ShouldAddDefaultChannel(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	assert.Equal(t, 1, len(oclb.Names()))
	assert.Nil(t, checkIntegrity(oclb, loadBalancer.DefaultSendChannel()))
}

//------- AddChannel

func TestOutgoingChannelLoadBalancer_AddChannelNewChannelShouldNotErrAndAddNewChannel(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	err := oclb.AddChannel("test")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(oclb.Names()))
	assert.Nil(t, checkIntegrity(oclb, loadBalancer.DefaultSendChannel()))
	assert.Nil(t, checkIntegrity(oclb, "test"))
}

func TestOutgoingChannelLoadBalancer_AddChannelDefaultChannelShouldErr(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	err := oclb.AddChannel(loadBalancer.DefaultSendChannel())

	assert.Equal(t, p2p.ErrChannelAlreadyExists, err)
}

func TestOutgoingChannelLoadBalancer_AddChannelReAddChannelShouldErr(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	_ = oclb.AddChannel("test")
	err := oclb.AddChannel("test")

	assert.Equal(t, p2p.ErrChannelAlreadyExists, err)
}

//------- RemoveChannel

func TestOutgoingChannelLoadBalancer_RemoveChannelRemoveDefaultShouldErr(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	err := oclb.RemoveChannel(loadBalancer.DefaultSendChannel())

	assert.Equal(t, p2p.ErrChannelCanNotBeDeleted, err)
}

func TestOutgoingChannelLoadBalancer_RemoveChannelRemoveNotFoundChannelShouldErr(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	err := oclb.RemoveChannel("test")

	assert.Equal(t, p2p.ErrChannelDoesNotExist, err)
}

func TestOutgoingChannelLoadBalancer_RemoveChannelRemoveLastChannelAddedShouldWork(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	_ = oclb.AddChannel("test1")
	_ = oclb.AddChannel("test2")
	_ = oclb.AddChannel("test3")

	err := oclb.RemoveChannel("test3")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(oclb.Names()))
	assert.Nil(t, checkIntegrity(oclb, loadBalancer.DefaultSendChannel()))
	assert.Nil(t, checkIntegrity(oclb, "test1"))
	assert.Nil(t, checkIntegrity(oclb, "test2"))
	assert.Equal(t, errMissingChannel, checkIntegrity(oclb, "test3"))
}

func TestOutgoingChannelLoadBalancer_RemoveChannelRemoveFirstChannelAddedShouldWork(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	_ = oclb.AddChannel("test1")
	_ = oclb.AddChannel("test2")
	_ = oclb.AddChannel("test3")

	err := oclb.RemoveChannel("test1")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(oclb.Names()))
	assert.Nil(t, checkIntegrity(oclb, loadBalancer.DefaultSendChannel()))
	assert.Equal(t, errMissingChannel, checkIntegrity(oclb, "test1"))
	assert.Nil(t, checkIntegrity(oclb, "test2"))
	assert.Nil(t, checkIntegrity(oclb, "test3"))
}

func TestOutgoingChannelLoadBalancer_RemoveChannelRemoveMiddleChannelAddedShouldWork(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	_ = oclb.AddChannel("test1")
	_ = oclb.AddChannel("test2")
	_ = oclb.AddChannel("test3")

	err := oclb.RemoveChannel("test2")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(oclb.Names()))
	assert.Nil(t, checkIntegrity(oclb, loadBalancer.DefaultSendChannel()))
	assert.Nil(t, checkIntegrity(oclb, "test1"))
	assert.Equal(t, errMissingChannel, checkIntegrity(oclb, "test2"))
	assert.Nil(t, checkIntegrity(oclb, "test3"))
}

//------- GetChannelOrDefault

func TestOutgoingChannelLoadBalancer_GetChannelOrDefaultNotFoundShouldReturnDefault(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	_ = oclb.AddChannel("test1")

	channel := oclb.GetChannelOrDefault("missing channel")

	assert.True(t, oclb.NamesChans()[loadBalancer.DefaultSendChannel()] == channel)
}

func TestOutgoingChannelLoadBalancer_GetChannelOrDefaultFoundShouldReturnChannel(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	_ = oclb.AddChannel("test1")

	channel := oclb.GetChannelOrDefault("test1")

	assert.True(t, oclb.NamesChans()["test1"] == channel)
}

//------- CollectOneElementFromChannels

func TestOutgoingChannelLoadBalancer_CollectFromChannelsNoObjectsShouldWaitBlocking(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	chanDone := make(chan struct{})

	go func() {
		_ = oclb.CollectOneElementFromChannels()

		chanDone <- struct{}{}
	}()

	select {
	case <-chanDone:
		assert.Fail(t, "should have not received object")
	case <-time.After(durationWait):
	}
}

func TestOutgoingChannelLoadBalancer_CollectOneElementFromChannelsShouldWork(t *testing.T) {
	t.Parallel()

	oclb := loadBalancer.NewOutgoingChannelLoadBalancer()

	oclb.AddChannel("test")

	obj1 := &p2p.SendableData{Topic: "test"}
	obj2 := &p2p.SendableData{Topic: "default"}

	chanDone := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(3)

	//send on channel test
	go func() {
		oclb.GetChannelOrDefault("test") <- obj1
		wg.Done()
	}()

	//send on default channel
	go func() {
		oclb.GetChannelOrDefault(loadBalancer.DefaultSendChannel()) <- obj2
		wg.Done()
	}()

	//func to wait finishing sending and receiving
	go func() {
		wg.Wait()
		chanDone <- true
	}()

	//func to periodically consume from channels
	go func() {
		foundObj1 := false
		foundObj2 := false

		for {
			obj := oclb.CollectOneElementFromChannels()

			if !foundObj1 {
				if obj == obj1 {
					foundObj1 = true
				}
			}

			if !foundObj2 {
				if obj == obj2 {
					foundObj2 = true
				}
			}

			if foundObj1 && foundObj2 {
				break
			}
		}

		wg.Done()
	}()

	select {
	case <-chanDone:
		return
	case <-time.After(durationWait):
		assert.Fail(t, "timeout")
		return
	}
}
