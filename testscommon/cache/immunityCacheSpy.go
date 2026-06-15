package cache

// ImmunityCacheSpy is a spy for the ImmunityCache
type ImmunityCacheSpy struct {
	*CacherStub
	SetOldestImmuneNonceCalled func(uint64)
}

// ImmunizeKeys is a spy for the ImmunizeKeys method of the ImmunityCache
func (c *ImmunityCacheSpy) ImmunizeKeys(_ [][]byte, _ uint64) (int, int) { return 0, 0 }

// SetOldestImmuneNonce is a spy for the SetOldestImmuneNonce method of the ImmunityCache
func (c *ImmunityCacheSpy) SetOldestImmuneNonce(nonce uint64) {
	if c.SetOldestImmuneNonceCalled != nil {
		c.SetOldestImmuneNonceCalled(nonce)
	}
}

func (c *ImmunityCacheSpy) RemoveWithResult(_ []byte) bool { return false }
func (c *ImmunityCacheSpy) NumBytes() int                  { return 0 }
func (c *ImmunityCacheSpy) Diagnose(_ bool)                {}
