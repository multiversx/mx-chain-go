package factory

import (
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedPeerShardFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil core comp should error", func(t *testing.T) {
		t.Parallel()

		_, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(nil, cryptoComp)

		idcif, err := NewInterceptedPeerShardFactory(*arg)
		assert.Equal(t, process.ErrNilCoreComponentsHolder, err)
		assert.True(t, check.IfNil(idcif))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		coreComp.IntMarsh = nil
		arg := createMockArgument(coreComp, cryptoComp)

		idcif, err := NewInterceptedPeerShardFactory(*arg)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(idcif))
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.ShardCoordinator = nil

		idcif, err := NewInterceptedPeerShardFactory(*arg)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
		assert.True(t, check.IfNil(idcif))
	})
	t.Run("should work and create", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)

		idcif, err := NewInterceptedPeerShardFactory(*arg)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(idcif))

		msg := &p2pFactory.PeerShard{
			ShardId: "5",
		}
		msgBuff, _ := arg.CoreComponents.InternalMarshalizer().Marshal(msg)
		interceptedData, err := idcif.Create(msgBuff)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(interceptedData))
		assert.True(t, strings.Contains(fmt.Sprintf("%T", interceptedData), "*p2p.interceptedPeerShard"))
	})
}
