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
var errMissingPipe = errors.New("missing pipe")
var errPipesMismatch = errors.New("pipes mismatch")

func checkIntegrity(oplb *loadBalancer.OutgoingPipeLoadBalancer, name string) error {
	if len(oplb.Names()) != len(oplb.Chans()) {
		return errLenDifferent
	}

	if len(oplb.Names()) != len(oplb.NamesChans()) {
		return errLenDifferentNamesChans
	}

	idxFound := -1
	for i, n := range oplb.Names() {
		if n == name {
			idxFound = i
			break
		}
	}

	if idxFound == -1 && oplb.NamesChans()[name] == nil {
		return errMissingPipe
	}

	if oplb.NamesChans()[name] != oplb.Chans()[idxFound] {
		return errPipesMismatch
	}

	return nil
}

//------- NewOutgoingPipeLoadBalancer

func TestNewOutgoingPipeLoadBalancer_ShouldNotProduceNil(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	assert.NotNil(t, oplb)
}

func TestNewOutgoingPipeLoadBalancer_ShouldAddDefaultPipe(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	assert.Equal(t, 1, len(oplb.Names()))
	assert.Nil(t, checkIntegrity(oplb, loadBalancer.DefaultSendPipe()))
}

//------- AddPipe

func TestOutgoingPipeLoadBalancer_AddPipeNewPipeShouldNotErrAndAddNewPipe(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	err := oplb.AddPipe("test")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(oplb.Names()))
	assert.Nil(t, checkIntegrity(oplb, loadBalancer.DefaultSendPipe()))
	assert.Nil(t, checkIntegrity(oplb, "test"))
}

func TestOutgoingPipeLoadBalancer_AddPipeDefaultPipeShouldErr(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	err := oplb.AddPipe(loadBalancer.DefaultSendPipe())

	assert.Equal(t, p2p.ErrPipeAlreadyExists, err)
}

func TestOutgoingPipeLoadBalancer_AddPipeReAddPipeShouldErr(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	_ = oplb.AddPipe("test")
	err := oplb.AddPipe("test")

	assert.Equal(t, p2p.ErrPipeAlreadyExists, err)
}

//------- RemovePipe

func TestOutgoingPipeLoadBalancer_RemovePipeRemoveDefaultShouldErr(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	err := oplb.RemovePipe(loadBalancer.DefaultSendPipe())

	assert.Equal(t, p2p.ErrPipeCanNotBeDeleted, err)
}

func TestOutgoingPipeLoadBalancer_RemovePipeRemoveNotFoundPipeShouldErr(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	err := oplb.RemovePipe("test")

	assert.Equal(t, p2p.ErrPipeDoNotExists, err)
}

func TestOutgoingPipeLoadBalancer_RemovePipeRemoveLastPipeAddedShouldWork(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	_ = oplb.AddPipe("test1")
	_ = oplb.AddPipe("test2")
	_ = oplb.AddPipe("test3")

	err := oplb.RemovePipe("test3")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(oplb.Names()))
	assert.Nil(t, checkIntegrity(oplb, loadBalancer.DefaultSendPipe()))
	assert.Nil(t, checkIntegrity(oplb, "test1"))
	assert.Nil(t, checkIntegrity(oplb, "test2"))
	assert.Equal(t, errMissingPipe, checkIntegrity(oplb, "test3"))
}

func TestOutgoingPipeLoadBalancer_RemovePipeRemoveFirstPipeAddedShouldWork(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	_ = oplb.AddPipe("test1")
	_ = oplb.AddPipe("test2")
	_ = oplb.AddPipe("test3")

	err := oplb.RemovePipe("test1")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(oplb.Names()))
	assert.Nil(t, checkIntegrity(oplb, loadBalancer.DefaultSendPipe()))
	assert.Equal(t, errMissingPipe, checkIntegrity(oplb, "test1"))
	assert.Nil(t, checkIntegrity(oplb, "test2"))
	assert.Nil(t, checkIntegrity(oplb, "test3"))
}

func TestOutgoingPipeLoadBalancer_RemovePipeRemoveMiddlePipeAddedShouldWork(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	_ = oplb.AddPipe("test1")
	_ = oplb.AddPipe("test2")
	_ = oplb.AddPipe("test3")

	err := oplb.RemovePipe("test2")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(oplb.Names()))
	assert.Nil(t, checkIntegrity(oplb, loadBalancer.DefaultSendPipe()))
	assert.Nil(t, checkIntegrity(oplb, "test1"))
	assert.Equal(t, errMissingPipe, checkIntegrity(oplb, "test2"))
	assert.Nil(t, checkIntegrity(oplb, "test3"))
}

//------- GetChannelOrDefault

func TestOutgoingPipeLoadBalancer_GetChannelOrDefaultNotFoundShouldReturnDefault(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	_ = oplb.AddPipe("test1")

	pipe := oplb.GetChannelOrDefault("missing pipe")

	assert.True(t, oplb.NamesChans()[loadBalancer.DefaultSendPipe()] == pipe)
}

func TestOutgoingPipeLoadBalancer_GetChannelOrDefaultFoundShouldReturnChannel(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	_ = oplb.AddPipe("test1")

	pipe := oplb.GetChannelOrDefault("test1")

	assert.True(t, oplb.NamesChans()["test1"] == pipe)
}

//------- CollectFromPipes

func TestOutgoingPipeLoadBalancer_CollectFromPipesNoObjectsWaitingShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	objs := oplb.CollectFromPipes()

	assert.Equal(t, 0, len(objs))
}

func TestOutgoingPipeLoadBalancer_CollectFromPipesShouldWork(t *testing.T) {
	t.Parallel()

	oplb := loadBalancer.NewOutgoingPipeLoadBalancer()

	oplb.AddPipe("test")

	obj1 := &p2p.SendableData{Topic: "test"}
	obj2 := &p2p.SendableData{Topic: "default"}

	chanDone := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(3)

	//send on pipe test
	go func() {
		oplb.GetChannelOrDefault("test") <- obj1
		wg.Done()
	}()

	//send on default pipe
	go func() {
		oplb.GetChannelOrDefault(loadBalancer.DefaultSendPipe()) <- obj2
		wg.Done()
	}()

	//func to wait finishing sending and receiving
	go func() {
		wg.Wait()
		chanDone <- true
	}()

	//func to periodically consume from pipes
	go func() {
		foundObj1 := false
		foundObj2 := false

		for {
			objs := oplb.CollectFromPipes()

			for idx := range objs {
				if !foundObj1 {
					if objs[idx] == obj1 {
						foundObj1 = true
					}
				}

				if !foundObj2 {
					if objs[idx] == obj2 {
						foundObj2 = true
					}
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
	case <-time.After(time.Second * 2):
		assert.Fail(t, "timeout")
		return
	}
}
