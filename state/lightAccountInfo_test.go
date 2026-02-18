package state

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLightAccountInfo_ReadOnlyFields(t *testing.T) {
	addr := []byte("test-address-32-bytes-long-12345")
	balance := big.NewInt(999)
	codeMeta := []byte{0x00, 0x00}
	rootHash := []byte("some-root-hash")

	acct := NewLightAccountInfo(addr, 42, balance, codeMeta, rootHash)

	assert.Equal(t, uint64(42), acct.GetNonce())
	assert.Equal(t, big.NewInt(999), acct.GetBalance())
	assert.Equal(t, addr, acct.AddressBytes())
	assert.Equal(t, codeMeta, acct.GetCodeMetadata())
	assert.Equal(t, rootHash, acct.GetRootHash())
	assert.False(t, acct.IsInterfaceNil())
}

func TestLightAccountInfo_GetBalance_ReturnsCopy(t *testing.T) {
	acct := NewLightAccountInfo([]byte("addr"), 0, big.NewInt(100), nil, nil)

	b1 := acct.GetBalance()
	b1.SetInt64(999)

	// Original should be unaffected
	assert.Equal(t, big.NewInt(100), acct.GetBalance())
}

func TestLightAccountInfo_GetBalance_NilReturnsZero(t *testing.T) {
	acct := NewLightAccountInfo([]byte("addr"), 0, nil, nil, nil)
	assert.Equal(t, big.NewInt(0), acct.GetBalance())
}

func TestLightAccountInfo_IsGuarded_NotGuarded(t *testing.T) {
	// Empty code metadata = not guarded
	acct := NewLightAccountInfo([]byte("addr"), 0, big.NewInt(0), nil, nil)
	assert.False(t, acct.IsGuarded())

	acct2 := NewLightAccountInfo([]byte("addr"), 0, big.NewInt(0), []byte{0x00, 0x00}, nil)
	assert.False(t, acct2.IsGuarded())
}

func TestLightAccountInfo_IsGuarded_Guarded(t *testing.T) {
	// MetadataGuarded = 0x08 in byte[0] of code metadata
	guardedCodeMeta := []byte{0x08, 0x00}
	acct := NewLightAccountInfo([]byte("addr"), 0, big.NewInt(0), guardedCodeMeta, nil)
	assert.True(t, acct.IsGuarded())
}

func TestLightAccountInfo_IsGuarded_WithOtherBitsSet(t *testing.T) {
	// Guarded bit (0x08) combined with other flags
	codeMeta := []byte{0x0C, 0x00} // 0x08 | 0x04
	acct := NewLightAccountInfo([]byte("addr"), 0, big.NewInt(0), codeMeta, nil)
	assert.True(t, acct.IsGuarded())
}

func TestLightAccountInfo_ZeroValueGetters(t *testing.T) {
	acct := NewLightAccountInfo([]byte("addr"), 0, big.NewInt(0), nil, nil)

	assert.Nil(t, acct.GetCodeHash())
	assert.Nil(t, acct.GetRootHash())
	assert.Nil(t, acct.DataTrie())
	assert.Nil(t, acct.GetOwnerAddress())
	assert.Nil(t, acct.GetUserName())
	assert.Equal(t, big.NewInt(0), acct.GetDeveloperReward())
}

func TestLightAccountInfo_GetRootHash_ReturnsStoredHash(t *testing.T) {
	rootHash := []byte("data-trie-root-hash-32-bytes-xx")
	acct := NewLightAccountInfo([]byte("addr"), 0, big.NewInt(0), nil, rootHash)
	assert.Equal(t, rootHash, acct.GetRootHash())
}

func TestLightAccountInfo_MutatingMethodsPanic(t *testing.T) {
	acct := NewLightAccountInfo([]byte("addr"), 0, big.NewInt(0), nil, nil)

	panicCases := []struct {
		name string
		fn   func()
	}{
		{"IncreaseNonce", func() { acct.IncreaseNonce(1) }},
		{"SetCode", func() { acct.SetCode([]byte{}) }},
		{"SetCodeMetadata", func() { acct.SetCodeMetadata([]byte{}) }},
		{"SetCodeHash", func() { acct.SetCodeHash([]byte{}) }},
		{"SetRootHash", func() { acct.SetRootHash([]byte{}) }},
		{"SetDataTrie", func() { acct.SetDataTrie(nil) }},
		{"RetrieveValue", func() { _, _, _ = acct.RetrieveValue([]byte{}) }},
		{"SaveKeyValue", func() { _ = acct.SaveKeyValue([]byte{}, []byte{}) }},
		{"AddToBalance", func() { _ = acct.AddToBalance(big.NewInt(1)) }},
		{"SubFromBalance", func() { _ = acct.SubFromBalance(big.NewInt(1)) }},
		{"ClaimDeveloperRewards", func() { _, _ = acct.ClaimDeveloperRewards([]byte{}) }},
		{"AddToDeveloperReward", func() { acct.AddToDeveloperReward(big.NewInt(1)) }},
		{"ChangeOwnerAddress", func() { _ = acct.ChangeOwnerAddress([]byte{}, []byte{}) }},
		{"SetOwnerAddress", func() { acct.SetOwnerAddress([]byte{}) }},
		{"SetUserName", func() { acct.SetUserName([]byte{}) }},
		{"GetAllLeaves", func() { _ = acct.GetAllLeaves(nil, nil) }},
	}

	for _, tc := range panicCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Panics(t, tc.fn, "expected %s to panic", tc.name)
		})
	}
}

func TestLightAccountInfo_NilIsInterfaceNil(t *testing.T) {
	var acct *lightAccountInfo
	assert.True(t, acct.IsInterfaceNil())
}
