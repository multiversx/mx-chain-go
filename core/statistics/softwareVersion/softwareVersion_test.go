package softwareVersion

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
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

	softwareChecker, err := NewSoftwareVersionChecker(&mock.AppStatusHandlerStub{}, nil, 1)

	assert.Nil(t, softwareChecker)
	assert.Equal(t, core.ErrNilStatusTagProvider, err)
}

func TestNewSoftwareVersionChecker_InvalidPollingIntervalShouldErr(t *testing.T) {
	t.Parallel()

	statusHandler := &mock.AppStatusHandlerStub{}
	tagProvider := &mock.StableTagProviderStub{}
	softwareChecker, err := NewSoftwareVersionChecker(statusHandler, tagProvider, 0)

	assert.Nil(t, softwareChecker)
	assert.Equal(t, core.ErrInvalidPollingInterval, err)
}

func TestNewSoftwareVersionChecker(t *testing.T) {
	t.Parallel()

	statusHandler := &mock.AppStatusHandlerStub{}
	tagProvider := &mock.StableTagProviderStub{}
	softwareChecker, err := NewSoftwareVersionChecker(statusHandler, tagProvider, 1)

	assert.Nil(t, err)
	assert.NotNil(t, softwareChecker)
}

func TestSoftwareVersionChecker_StartCheckSoftwareVersionShouldWork(t *testing.T) {
	t.Parallel()

	fetchChan := make(chan bool)
	setStringChan := make(chan bool)
	statusHandler := &mock.AppStatusHandlerStub{
		SetStringValueHandler: func(key string, value string) {
			setStringChan <- true
		},
	}
	tagProvider := &mock.StableTagProviderStub{
		FetchTagVersionCalled: func() (string, error) {
			fetchChan <- true
			return "", nil
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
}

func TestSoftwareVersionChecker_StartCheckSoftwareVersionShouldErrWhenFetchFails(t *testing.T) {
	t.Parallel()

	localErr := errors.New("error")
	fetchChan := make(chan bool)
	setStringChan := make(chan bool)
	statusHandler := &mock.AppStatusHandlerStub{
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
