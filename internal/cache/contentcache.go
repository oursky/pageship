package cache

import (
	"fmt"
	"sync"
	"time"
	"http"

	"github.com/dgraph-io/ristretto"
)

type ContentCache struct {
	m     *mm
	size  int64
	cache *ristretto.Cache
	debug bool
}

type Response struct {
	Header     http.Header
	Body       []byte
	StatusCode int
}

func NewContentCache(contentCacheSize int64, debug bool) (*ContentCache, error) {
	m := New()
	size := contentCacheSize
	nc := size / 1000
	if nc < 100 { //mainly for testing
		nc = 100
	}
	cache, err := ristretto.NewCache(&ristretto.Config{
		//NumCounters is 10 times estimated max number of items in cache, as suggested in https://pkg.go.dev/github.com/dgraph-io/ristretto@v0.1.1#Config
		NumCounters:        nc, //limit / 10 KB small files * 10
		MaxCost:            size,
		BufferItems:        64,
		Metrics:            debug,
		IgnoreInternalCost: true,
	})
	if err != nil {
		return nil, err
	}

	return &ContentCache{m: m, size: size, cache: cache, load: load, debug: debug}, nil
}

func (c *ContentCache) Get(key string) (*Response, bool) {
	l := c.m.RLock(key)
	defer l.RUnlock()

	v, b := c.cache.Get(c.keyToString(key))
	if c.debug {
		fmt.Println("getted: ", c.cache.Metrics.String())
	}
	if v == nil {
		var nv V
		return nv, b
	}
	return v.(V), b
}

func (c *ContentCache) Set(key string, value *Response) {
	l := c.m.Lock(key)
	defer l.Unlock()

	b, n, err := c.load(r)
	if err != nil {
		return b, err
	}

	c.cache.Set(c.keyToString(key), b, n)
	if c.debug {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("setted: ", c.cache.Metrics.String())
	}
}

// modified from https://stackoverflow.com/questions/40931373/how-to-gc-a-map-of-mutexes-in-go
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
