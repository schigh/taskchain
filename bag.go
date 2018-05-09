package taskgraph

import "sync"

type bag struct {
	sync.RWMutex
	data map[string]interface{}
}

func (b *bag) get(key string) interface{} {
	b.Lock()
	defer b.Unlock()

	return b.data[key]
}

func (b *bag) set(key string, value interface{}) {
	b.Lock()
	if b.data == nil {
		b.data = make(map[string]interface{}, 1)
	}
	b.data[key] = value
	b.Unlock()
}

func (b *bag) remove(key string) {
	b.Lock()
	if b.data != nil {
		delete(b.data, key)
	}
	b.Unlock()
}

func (b *bag) absorb(other *bag) {
	b.Lock()
	other.Lock()
	defer func(b1, b2 *bag) {
		b1.Unlock()
		b2.Unlock()
	}(b, other)

	for k, v := range other.data {
		if b.data[k] == nil && v != nil {
			b.data[k] = v
		}
	}
}
