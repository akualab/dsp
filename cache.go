package dsp

// Primitive cache. TODO: optimize.
type cache struct {
	cap        int
	start, len int
	idx        uint32
	store      map[uint32]Value
}

func newCache(cap int) *cache {

	return &cache{
		cap:   cap,
		store: map[uint32]Value{},
	}
}

func (c *cache) set(idx uint32, vec Value) {
	c.store[idx] = vec
}

func (c *cache) get(idx uint32) (Value, bool) {
	v, ok := c.store[idx]
	return v, ok
}

func (c *cache) clear() {
	c.store = map[uint32]Value{}
}
