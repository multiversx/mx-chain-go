package processor_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieNodesInterceptorProcessor_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	tnip, err := processor.NewTrieNodesInterceptorProcessor(nil)
	assert.Nil(t, tnip)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestNewTrieNodesInterceptorProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	tnip, err := processor.NewTrieNodesInterceptorProcessor(&mock.CacherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, tnip)
}

//------- Validate

func TestTrieNodesInterceptorProcessor_ValidateShouldWork(t *testing.T) {
	t.Parallel()

	tnip, _ := processor.NewTrieNodesInterceptorProcessor(&mock.CacherMock{})

	assert.Nil(t, tnip.Validate(nil))
}

//------- Save

func TestTrieNodesInterceptorProcessor_SaveWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	tnip, _ := processor.NewTrieNodesInterceptorProcessor(&mock.CacherMock{})

	err := tnip.Save(nil)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestTrieNodesInterceptorProcessor_SaveShouldPutInCacher(t *testing.T) {
	t.Parallel()

	putCalled := false
	cacher := &mock.CacherStub{
		PutCalled: func(key []byte, value interface{}) (evicted bool) {
			putCalled = true
			return false
		},
	}
	tnip, _ := processor.NewTrieNodesInterceptorProcessor(cacher)

	err := tnip.Save(&trie.InterceptedTrieNode{})
	assert.Nil(t, err)
	assert.True(t, putCalled)
}

//------- IsInterfaceNil

func TestTrieNodesInterceptorProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var tnip *processor.TrieNodeInterceptorProcessor
	assert.True(t, check.IfNil(tnip))
}
