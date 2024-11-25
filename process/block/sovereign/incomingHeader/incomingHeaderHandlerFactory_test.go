package incomingHeader

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/stretchr/testify/require"
)

func createWSCfg() config.WebSocketConfig {
	return config.WebSocketConfig{
		MarshallerType: "json",
		HasherType:     "keccak",
	}
}

func TestCreateIncomingHeaderProcessor(t *testing.T) {
	t.Parallel()

	runTypeComps := mock.NewRunTypeComponentsStub()
	headersPool := &dataRetriever.PoolsHolderStub{
		HeadersCalled: func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{}
		},
	}

	t.Run("nil run type comps, should not work", func(t *testing.T) {
		headerProc, err := CreateIncomingHeaderProcessor(
			createWSCfg(),
			headersPool,
			11,
			nil,
		)
		require.Equal(t, errorsMx.ErrNilRunTypeComponents, err)
		require.Nil(t, headerProc)
	})

	t.Run("invalid marshaller, should not work", func(t *testing.T) {
		cfg := createWSCfg()
		cfg.MarshallerType = ""

		headerProc, err := CreateIncomingHeaderProcessor(
			cfg,
			headersPool,
			11,
			runTypeComps,
		)
		require.NotNil(t, err)
		require.Nil(t, headerProc)
	})

	t.Run("invalid hasher, should not work", func(t *testing.T) {
		cfg := createWSCfg()
		cfg.HasherType = ""

		headerProc, err := CreateIncomingHeaderProcessor(
			cfg,
			headersPool,
			11,
			runTypeComps,
		)
		require.NotNil(t, err)
		require.Nil(t, headerProc)
	})

	t.Run("should work", func(t *testing.T) {
		headerProc, err := CreateIncomingHeaderProcessor(
			createWSCfg(),
			headersPool,
			11,
			runTypeComps,
		)
		require.Nil(t, err)
		require.False(t, headerProc.IsInterfaceNil())
	})
}
