package cache

import (
	"fmt"
	"sync"

	"github.com/dgraph-io/ristretto"
)

type ContentCache[K any, V any, R any] struct {
	m     *mm
	size  int64
	cache *ristretto.Cache
	load  func(r R) (V, int64, error)
}

func NewContentCache[K any, V any, R any](contentCacheSize int64, metrics bool, load func(r R) (V, int64, error)) (*ContentCache[K, V, R], error) {
	m := New()
	size := contentCacheSize
	nc := size / 1000
	if (nc < 100) { //mainly for testing
		nc = 100
	}
	cache, err := ristretto.NewCache(&ristretto.Config{
		//NumCounters is 10 times estimated max number of items in cache, as suggested in https://pkg.go.dev/github.com/dgraph-io/ristretto@v0.1.1#Config
		NumCounters: nc, //limit / 10 KB small files * 10
		MaxCost:     size,
		BufferItems: 64,
		Metrics:     metrics,
		IgnoreInternalCost: true,
	})
	if err != nil {
		return nil, err
	}

	return &ContentCache[K, V, R]{m: m, size: size, cache: cache, load: load}, nil
}

func (c *ContentCache[K, V, R]) keyToString(key K) string {
	return fmt.Sprintf("%v", key)
}

func (c *ContentCache[K, V, R]) GetContent(key K) (V, bool) {
	l := c.m.RLock(key)
	defer l.RUnlock()

	v, b := c.cache.Get(c.keyToString(key))
	if v == nil {
		var nv V
		return nv, b
	}
	return v.(V), b
}

func (c *ContentCache[K, V, R]) SetContent(key K, r R) (V, error) {
	l := c.m.Lock(key)
	defer l.Unlock()

	b, n, err := c.load(r)
	if err != nil {
		return b, err
	}

	c.cache.Set(c.keyToString(key), b, n)
	return b, nil
}

//modified from https://stackoverflow.com/questions/40931373/how-to-gc-a-map-of-mutexes-in-go
type mm struct {
    ml sync.Mutex              
    ma map[interface{}]*mentry 
}

type mentry struct {
    m   *mm         
    el  sync.RWMutex 
    cnt int        
    key interface{}
}

type Unlocker interface {
    Unlock()
}

type RUnlocker interface {
    RUnlock()
}

func New() *mm {
    return &mm{ma: make(map[interface{}]*mentry)}
}

func (m *mm) Lock(key interface{}) Unlocker {
    m.ml.Lock()
    e, ok := m.ma[key]
    if !ok {
        e = &mentry{m: m, key: key}
        m.ma[key] = e
    }
    e.cnt++
    m.ml.Unlock()

    e.el.Lock()

    return e
}

func (me *mentry) Unlock() {
    m := me.m

    m.ml.Lock()
    e, ok := m.ma[me.key]
    if !ok {
        m.ml.Unlock()
        panic(fmt.Errorf("Unlock requested for key=%v but no entry found", me.key))
    }
    e.cnt--        
    if e.cnt < 1 { 
        delete(m.ma, me.key)
    }
    m.ml.Unlock()

    e.el.Unlock()
}


func (m *mm) RLock(key interface{}) RUnlocker {
    m.ml.Lock()
    e, ok := m.ma[key]
    if !ok {
        e = &mentry{m: m, key: key}
        m.ma[key] = e
    }
    e.cnt++
    m.ml.Unlock()

    e.el.RLock()

    return e
}

func (me *mentry) RUnlock() {
    m := me.m

    m.ml.Lock()
    e, ok := m.ma[me.key]
    if !ok {
        m.ml.Unlock()
        panic(fmt.Errorf("Unlock requested for key=%v but no entry found", me.key))
    }
    e.cnt--        
    if e.cnt < 1 { 
        delete(m.ma, me.key)
    }
    m.ml.Unlock()

    e.el.RUnlock()
}
