package dataThrottle_test

import (
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/dataThrottle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func checkIntegrity(sdt *dataThrottle.SendDataThrottle, name string) error {
	if len(sdt.Names()) != len(sdt.Chans()) {
		return errors.New("len different for names and chans")
	}

	if len(sdt.Names()) != len(sdt.NamesChans()) {
		return errors.New("len different for names and namesChans")
	}

	idxFound := -1
	for i, n := range sdt.Names() {
		if n == name {
			idxFound = i
			break
		}
	}

	if idxFound == -1 && sdt.NamesChans()[name] == nil {
		return errors.New("missing pipe")
	}

	if sdt.NamesChans()[name] != sdt.Chans()[idxFound] {
		return errors.New("pipes mismatch")
	}

	return nil
}

//------- NewSendDataThrottle

func TestNewSendDataThrottle_ShouldNotProduceNil(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	assert.NotNil(t, sdt)
}

func TestNewSendDataThrottle_ShouldAddDefaultPipe(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	assert.Equal(t, 1, len(sdt.Names()))
	assert.Nil(t, checkIntegrity(sdt, dataThrottle.DefaultSendPipe()))
}

//------- AddPipe

func TestSendDataThrottle_AddPipeNewPipeShouldNotProduceErr(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	err := sdt.AddPipe("test")

	assert.Nil(t, err)
}

func TestSendDataThrottle_AddPipeNewPipeShouldAddPipe(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	_ = sdt.AddPipe("test")

	assert.Equal(t, 2, len(sdt.Names()))
	assert.Nil(t, checkIntegrity(sdt, dataThrottle.DefaultSendPipe()))
	assert.Nil(t, checkIntegrity(sdt, "test"))
}

func TestSendDataThrottle_AddPipeDefaultPipeShouldErr(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	err := sdt.AddPipe(dataThrottle.DefaultSendPipe())

	assert.Equal(t, p2p.ErrPipeAlreadyExists, err)
}

func TestSendDataThrottle_AddPipeReAddPipeShouldErr(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	_ = sdt.AddPipe("test")
	err := sdt.AddPipe("test")

	assert.Equal(t, p2p.ErrPipeAlreadyExists, err)
}

//------- RemovePipe

func TestSendDataThrottle_RemovePipeRemoveDefaultShouldErr(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	err := sdt.RemovePipe(dataThrottle.DefaultSendPipe())

	assert.Equal(t, p2p.ErrPipeCanNotBeDeleted, err)
}

func TestSendDataThrottle_RemovePipeRemoveNotFoundPipeShouldErr(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	err := sdt.RemovePipe("test")

	assert.Equal(t, p2p.ErrPipeDoNotExists, err)
}

func TestSendDataThrottle_RemovePipeRemoveLastPipeAddedShouldWork(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	_ = sdt.AddPipe("test1")
	_ = sdt.AddPipe("test2")
	_ = sdt.AddPipe("test3")

	err := sdt.RemovePipe("test3")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(sdt.Names()))
	assert.Nil(t, checkIntegrity(sdt, dataThrottle.DefaultSendPipe()))
	assert.Nil(t, checkIntegrity(sdt, "test1"))
	assert.Nil(t, checkIntegrity(sdt, "test2"))
	assert.Equal(t, "missing pipe", checkIntegrity(sdt, "test3").Error())
}

func TestSendDataThrottle_RemovePipeRemoveFirstPipeAddedShouldWork(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	_ = sdt.AddPipe("test1")
	_ = sdt.AddPipe("test2")
	_ = sdt.AddPipe("test3")

	err := sdt.RemovePipe("test1")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(sdt.Names()))
	assert.Nil(t, checkIntegrity(sdt, dataThrottle.DefaultSendPipe()))
	assert.Equal(t, "missing pipe", checkIntegrity(sdt, "test1").Error())
	assert.Nil(t, checkIntegrity(sdt, "test2"))
	assert.Nil(t, checkIntegrity(sdt, "test3"))
}

func TestSendDataThrottle_RemovePipeRemoveMiddlePipeAddedShouldWork(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	_ = sdt.AddPipe("test1")
	_ = sdt.AddPipe("test2")
	_ = sdt.AddPipe("test3")

	err := sdt.RemovePipe("test2")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(sdt.Names()))
	assert.Nil(t, checkIntegrity(sdt, dataThrottle.DefaultSendPipe()))
	assert.Nil(t, checkIntegrity(sdt, "test1"))
	assert.Equal(t, "missing pipe", checkIntegrity(sdt, "test2").Error())
	assert.Nil(t, checkIntegrity(sdt, "test3"))
}

//------- GetChannelOrDefault

func TestSendDataThrottle_GetChannelOrDefaultNotFoundShouldReturnDefault(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	_ = sdt.AddPipe("test1")

	pipe := sdt.GetChannelOrDefault("missing pipe")

	assert.True(t, sdt.NamesChans()[dataThrottle.DefaultSendPipe()] == pipe)
}

func TestSendDataThrottle_GetChannelOrDefaultFoundShouldReturnChannel(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	_ = sdt.AddPipe("test1")

	pipe := sdt.GetChannelOrDefault("test1")

	assert.True(t, sdt.NamesChans()["test1"] == pipe)
}

//------- CollectFromPipes

func TestSendDataThrottle_CollectFromPipesNoObjectsWaitingShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	objs := sdt.CollectFromPipes()

	assert.Equal(t, 0, len(objs))
}

func TestSendDataThrottle_CollectFromPipesShouldWork(t *testing.T) {
	t.Parallel()

	sdt := dataThrottle.NewSendDataThrottle()

	sdt.AddPipe("test")

	obj1 := &p2p.SendableData{Topic: "test"}
	obj2 := &p2p.SendableData{Topic: "default"}

	chanDone := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(3)

	//send on pipe test
	go func() {
		sdt.GetChannelOrDefault("test") <- obj1
		wg.Done()
	}()

	//send on default pipe
	go func() {
		sdt.GetChannelOrDefault(dataThrottle.DefaultSendPipe()) <- obj2
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
			objs := sdt.CollectFromPipes()

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
