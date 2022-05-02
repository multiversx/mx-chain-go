package factory

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedValidatorInfoFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil core comp should error", func(t *testing.T) {
		t.Parallel()

		_, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(nil, cryptoComp)

		isvif, err := NewInterceptedValidatorInfoFactory(*arg)
		assert.Equal(t, process.ErrNilCoreComponentsHolder, err)
		assert.True(t, check.IfNil(isvif))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		coreComp.IntMarsh = nil
		arg := createMockArgument(coreComp, cryptoComp)

		isvif, err := NewInterceptedValidatorInfoFactory(*arg)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(isvif))
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)
		arg.ShardCoordinator = nil

		isvif, err := NewInterceptedValidatorInfoFactory(*arg)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
		assert.True(t, check.IfNil(isvif))
	})
	t.Run("should work and create", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		arg := createMockArgument(coreComp, cryptoComp)

		isvif, err := NewInterceptedValidatorInfoFactory(*arg)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(isvif))

		msg := &message.ShardValidatorInfo{
			ShardId: 5,
		}
		msgBuff, _ := arg.CoreComponents.InternalMarshalizer().Marshal(msg)
		interceptedData, err := isvif.Create(msgBuff)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(interceptedData))
		assert.True(t, strings.Contains(fmt.Sprintf("%T", interceptedData), "*p2p.interceptedValidatorInfo"))
	})
}
