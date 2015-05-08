package dsp

// Primitive cache. TODO: optimize.
type cache struct {
	cap        int
	start, len int
	idx        int
	store      map[int]Value
}

func newCache(cap int) *cache {

	return &cache{
		cap:   cap,
		store: map[int]Value{},
	}
}

func (c *cache) set(idx int, vec Value) {
	c.store[idx] = vec
}

func (c *cache) get(idx int) (Value, bool) {
	v, ok := c.store[idx]
	return v, ok
}

func (c *cache) clear() {
	c.store = map[int]Value{}
}
