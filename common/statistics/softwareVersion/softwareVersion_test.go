package softwareVersion

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common/mock"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewSoftwareVersionChecker_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	softwareChecker, err := NewSoftwareVersionChecker(nil, &mock.StableTagProviderStub{}, 1)

	assert.Nil(t, softwareChecker)
	assert.Equal(t, core.ErrNilAppStatusHandler, err)
}

func TestNewSoftwareVersionChecker_NilStableTagProviderShouldErr(t *testing.T) {
	t.Parallel()

	softwareChecker, err := NewSoftwareVersionChecker(&statusHandlerMock.AppStatusHandlerStub{}, nil, 1)

	assert.Nil(t, softwareChecker)
	assert.Equal(t, core.ErrNilStatusTagProvider, err)
}

func TestNewSoftwareVersionChecker_InvalidPollingIntervalShouldErr(t *testing.T) {
	t.Parallel()

	statusHandler := &statusHandlerMock.AppStatusHandlerStub{}
	tagProvider := &mock.StableTagProviderStub{}
	softwareChecker, err := NewSoftwareVersionChecker(statusHandler, tagProvider, 0)

	assert.Nil(t, softwareChecker)
	assert.Equal(t, core.ErrInvalidPollingInterval, err)
}

func TestNewSoftwareVersionChecker(t *testing.T) {
	t.Parallel()

	statusHandler := &statusHandlerMock.AppStatusHandlerStub{}
	tagProvider := &mock.StableTagProviderStub{}
	softwareChecker, err := NewSoftwareVersionChecker(statusHandler, tagProvider, 1)

	assert.Nil(t, err)
	assert.NotNil(t, softwareChecker)
}

func TestSoftwareVersionChecker_StartCheckSoftwareVersionShouldWork(t *testing.T) {
	t.Parallel()

	fetchChan := make(chan bool)
	setStringChan := make(chan bool)
	statusHandler := &statusHandlerMock.AppStatusHandlerStub{
		SetStringValueHandler: func(key string, value string) {
			setStringChan <- true
		},
	}
	tagProvider := &mock.StableTagProviderStub{
		FetchTagVersionCalled: func() (string, error) {
			fetchChan <- true
			return "1.0.0", nil
		},
	}

	softwareChecker, _ := NewSoftwareVersionChecker(statusHandler, tagProvider, 1)
	softwareChecker.StartCheckSoftwareVersion()

	select {
	case <-fetchChan:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "fetch was not called")
	}

	select {
	case <-setStringChan:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "fetch was not called")
	}

	assert.Nil(t, softwareChecker.Close())
}

func TestSoftwareVersionChecker_StartCheckSoftwareVersionShouldErrWhenFetchFails(t *testing.T) {
	t.Parallel()

	localErr := errors.New("error")
	fetchChan := make(chan bool)
	setStringChan := make(chan bool)
	statusHandler := &statusHandlerMock.AppStatusHandlerStub{
		SetStringValueHandler: func(key string, value string) {
			setStringChan <- true
		},
	}
	tagProvider := &mock.StableTagProviderStub{
		FetchTagVersionCalled: func() (string, error) {
			fetchChan <- true
			return "", localErr
		},
	}

	softwareChecker, _ := NewSoftwareVersionChecker(statusHandler, tagProvider, 1)
	softwareChecker.StartCheckSoftwareVersion()

	select {
	case <-fetchChan:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "fetch was not called")
	}

	select {
	case <-time.After(100 * time.Millisecond):
	case <-setStringChan:
		assert.Fail(t, "this should have not been called")
	}
}

func TestResourceMonitor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var softwareChecker *softwareVersionChecker
	assert.True(t, softwareChecker.IsInterfaceNil())

	softwareChecker, _ = NewSoftwareVersionChecker(&statusHandlerMock.AppStatusHandlerStub{}, &mock.StableTagProviderStub{}, 1)
	assert.False(t, softwareChecker.IsInterfaceNil())
}
