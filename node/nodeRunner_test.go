//go:build !race

package node

import (
	"os"
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/api"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const originalConfigsPath = "../cmd/node/config"

func TestNewNodeRunner(t *testing.T) {
	t.Parallel()

	t.Run("nil configs should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "nil configs provided"
		runner, err := NewNodeRunner(nil)
		assert.Nil(t, runner)
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("with valid configs should work", func(t *testing.T) {
		t.Parallel()

		configs := testscommon.CreateTestConfigs(t, originalConfigsPath)
		runner, err := NewNodeRunner(configs)
		assert.NotNil(t, runner)
		assert.Nil(t, err)
	})
}

func TestNodeRunner_StartAndCloseNodeUsingSIGINT(t *testing.T) {
	t.Parallel()

	configs := testscommon.CreateTestConfigs(t, originalConfigsPath)
	runner, _ := NewNodeRunner(configs)

	trigger := mock.NewApplicationRunningTrigger()
	err := logger.AddLogObserver(trigger, &logger.PlainFormatter{})
	require.Nil(t, err)

	// start a go routine that will send the SIGINT message after 1 second after the node has started
	go func() {
		timeout := time.Minute * 5
		select {
		case <-trigger.ChanClose():
		case <-time.After(timeout):
			require.Fail(t, "timeout waiting for application to start")
		}
		time.Sleep(time.Second)

		log.Info("sending SIGINT to self")
		errKill := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		assert.Nil(t, errKill)
	}()

	err = runner.Start()
	assert.Nil(t, err)
}

func TestCopyDirectory(t *testing.T) {
	t.Parallel()

	file1Name := "file1.toml"
	file1Contents := []byte("file1")
	file2Name := "file2.toml"
	file2Contents := []byte("file2")
	file3Name := "file3.toml"
	file3Contents := []byte("file3")
	file4Name := "file4.toml"
	file4Contents := []byte("file4")

	tempDir := t.TempDir()

	// generating dummy structure like
	// file1
	// src
	//   +- file2
	//   +- dir1
	//         +- file3
	//   +- dir2
	//         +- file4

	err := os.WriteFile(path.Join(tempDir, file1Name), file1Contents, os.ModePerm)
	require.Nil(t, err)
	err = os.MkdirAll(path.Join(tempDir, "src", "dir1"), os.ModePerm)
	require.Nil(t, err)
	err = os.MkdirAll(path.Join(tempDir, "src", "dir2"), os.ModePerm)
	require.Nil(t, err)
	err = os.WriteFile(path.Join(tempDir, "src", file2Name), file2Contents, os.ModePerm)
	require.Nil(t, err)
	err = os.WriteFile(path.Join(tempDir, "src", "dir1", file3Name), file3Contents, os.ModePerm)
	require.Nil(t, err)
	err = os.WriteFile(path.Join(tempDir, "src", "dir2", file4Name), file4Contents, os.ModePerm)
	require.Nil(t, err)

	err = copyDirectory(path.Join(tempDir, "src"), path.Join(tempDir, "dst"))
	require.Nil(t, err)
	copySingleFile(path.Join(tempDir, "dst"), path.Join(tempDir, file1Name))

	// after copy, check that the files are the same
	buff, err := os.ReadFile(path.Join(tempDir, "dst", file1Name))
	require.Nil(t, err)
	assert.Equal(t, file1Contents, buff)

	buff, err = os.ReadFile(path.Join(tempDir, "dst", file2Name))
	require.Nil(t, err)
	assert.Equal(t, file2Contents, buff)

	buff, err = os.ReadFile(path.Join(tempDir, "dst", "dir1", file3Name))
	require.Nil(t, err)
	assert.Equal(t, file3Contents, buff)

	buff, err = os.ReadFile(path.Join(tempDir, "dst", "dir2", file4Name))
	require.Nil(t, err)
	assert.Equal(t, file4Contents, buff)
}

func TestWaitForSignal(t *testing.T) {
	t.Parallel()

	closedCalled := make(map[string]struct{})
	healthServiceClosableComponent := &mock.CloserStub{
		CloseCalled: func() error {
			closedCalled["healthService"] = struct{}{}
			return nil
		},
	}
	facadeClosableComponent := &mock.CloserStub{
		CloseCalled: func() error {
			closedCalled["facade"] = struct{}{}
			return nil
		},
	}
	httpClosableComponent := &api.UpgradeableHttpServerHandlerStub{
		CloseCalled: func() error {
			closedCalled["http"] = struct{}{}
			return nil
		},
	}
	internalNodeClosableComponent1 := &components.ComponentHandlerStub{
		Name: factory.CoreComponentsName,
		CloseCalled: func() error {
			closedCalled[factory.CoreComponentsName] = struct{}{}
			return nil
		},
	}
	internalNodeClosableComponent2 := &components.ComponentHandlerStub{
		Name: factory.StatusComponentsName,
		CloseCalled: func() error {
			closedCalled[factory.StatusComponentsName] = struct{}{}
			return nil
		},
	}
	n, _ := NewNode()
	err := n.AddClosableComponents(internalNodeClosableComponent1)
	assert.Nil(t, err)
	err = n.AddClosableComponents(internalNodeClosableComponent2)
	assert.Nil(t, err)

	// do not run these tests in parallel as they are using the same map
	t.Run("should return nextOperationShouldStop if SIGINT is received", func(t *testing.T) {
		closedCalled = make(map[string]struct{})
		stopChan := make(chan endProcess.ArgEndProcess)
		sigs := make(chan os.Signal, 1)

		go func() {
			time.Sleep(time.Millisecond * 100) // wait for the waitForSignal to start
			sigs <- syscall.SIGINT
		}()

		nextOperation := waitForSignal(
			sigs,
			stopChan,
			healthServiceClosableComponent,
			facadeClosableComponent,
			httpClosableComponent,
			n,
			1,
		)

		assert.Equal(t, nextOperationShouldStop, nextOperation)
		assert.Equal(t, createExpectedComponentsCalledMap(), closedCalled)
	})
	t.Run("should return nextOperationShouldRestart if shuffled out is received", func(t *testing.T) {
		closedCalled = make(map[string]struct{})
		stopChan := make(chan endProcess.ArgEndProcess, 1)
		sigs := make(chan os.Signal)

		go func() {
			time.Sleep(time.Millisecond * 100) // wait for the waitForSignal to start
			stopChan <- endProcess.ArgEndProcess{
				Reason:      common.ShuffledOut,
				Description: "test",
			}
		}()

		nextOperation := waitForSignal(
			sigs,
			stopChan,
			healthServiceClosableComponent,
			facadeClosableComponent,
			httpClosableComponent,
			n,
			1,
		)

		assert.Equal(t, nextOperationShouldRestart, nextOperation)
		assert.Equal(t, createExpectedComponentsCalledMap(), closedCalled)
	})
	t.Run("wrong configuration should not stop the node", func(t *testing.T) {
		closedCalled = make(map[string]struct{})
		stopChan := make(chan endProcess.ArgEndProcess, 1)
		sigs := make(chan os.Signal)

		go func() {
			time.Sleep(time.Millisecond * 100) // wait for the waitForSignal to start
			stopChan <- endProcess.ArgEndProcess{
				Reason:      common.WrongConfiguration,
				Description: "test",
			}
		}()

		functionFinished := make(chan struct{})
		go func() {
			_ = waitForSignal(
				sigs,
				stopChan,
				healthServiceClosableComponent,
				facadeClosableComponent,
				httpClosableComponent,
				n,
				1,
			)
			close(functionFinished)
		}()

		select {
		case <-functionFinished:
			assert.Fail(t, "function should not have finished")
		case <-time.After(maxTimeToClose + time.Second*2):
			// ok, timeout reached, function did not finish
		}

		assert.Equal(t, createExpectedComponentsCalledMap(), closedCalled)
	})

	delayedComponent := &components.ComponentHandlerStub{
		Name: factory.ConsensusComponentsName,
		CloseCalled: func() error {
			time.Sleep(time.Minute)
			return nil
		},
	}
	err = n.AddClosableComponents(delayedComponent)
	assert.Nil(t, err)

	t.Run("force closing the node when SIGINT is received", func(t *testing.T) {
		closedCalled = make(map[string]struct{})
		stopChan := make(chan endProcess.ArgEndProcess)
		sigs := make(chan os.Signal, 1)

		go func() {
			time.Sleep(time.Millisecond * 100) // wait for the waitForSignal to start
			sigs <- syscall.SIGINT
		}()

		nextOperation := waitForSignal(
			sigs,
			stopChan,
			healthServiceClosableComponent,
			facadeClosableComponent,
			httpClosableComponent,
			n,
			1,
		)

		assert.Equal(t, nextOperationShouldStop, nextOperation)

		expectedMap := createExpectedComponentsCalledMap()
		// these exceptions appear because the delayedComponent prevented the call of the first 2 components
		// as the closable components are called in reversed order
		delete(expectedMap, factory.CoreComponentsName)
		delete(expectedMap, factory.StatusComponentsName)

		assert.Equal(t, expectedMap, closedCalled)
	})
	t.Run("force closing the node when shuffle out is received", func(t *testing.T) {
		closedCalled = make(map[string]struct{})
		stopChan := make(chan endProcess.ArgEndProcess, 1)
		sigs := make(chan os.Signal)

		go func() {
			time.Sleep(time.Millisecond * 100) // wait for the waitForSignal to start
			stopChan <- endProcess.ArgEndProcess{
				Reason:      common.ShuffledOut,
				Description: "test",
			}
		}()

		nextOperation := waitForSignal(
			sigs,
			stopChan,
			healthServiceClosableComponent,
			facadeClosableComponent,
			httpClosableComponent,
			n,
			1,
		)

		expectedMap := createExpectedComponentsCalledMap()
		// these exceptions appear because the delayedComponent prevented the call of the first 2 components
		// as the closable components are called in reversed order
		delete(expectedMap, factory.CoreComponentsName)
		delete(expectedMap, factory.StatusComponentsName)
		// in this case, even if the node is shuffled out, it should stop as some components were not closed
		assert.Equal(t, nextOperationShouldStop, nextOperation)
		assert.Equal(t, expectedMap, closedCalled)
	})
}

func createExpectedComponentsCalledMap() map[string]struct{} {
	return map[string]struct{}{
		"healthService":              {},
		"facade":                     {},
		"http":                       {},
		factory.CoreComponentsName:   {},
		factory.StatusComponentsName: {},
	}
}
