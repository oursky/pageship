package cache //white box testing

import (
	"fmt"
	"sync"
	"testing"
	"time"
	"net/http"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	fmt.Println("testing Get...")
	cc, err := NewContentCache(8, true)
	assert.Empty(t, err)

	k1 := "id1"
	k2 := "id2"
	k3 := "id3"

	ccc, found := cc.Get(k1)
	assert.Equal(t, false, found)
	assert.Equal(t, "hit: 0 miss: 1 keys-added: 0 keys-updated: 0 keys-evicted: 0 cost-added: 0 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 1 hit-ratio: 0.00", cc.cache.Metrics.String())

	resp1 := Response{Header: make(http.Header), StatusCode: 200, Body: []byte("a")}
	cc.Set(k1, &resp1)
	time.Sleep(100 * time.Millisecond) //https://github.com/dgraph-io/ristretto/issues/161
	assert.Equal(t, "hit: 0 miss: 1 keys-added: 1 keys-updated: 0 keys-evicted: 0 cost-added: 1 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 1 hit-ratio: 0.00", cc.cache.Metrics.String())

	ccc, found = cc.Get(k1)
	assert.Equal(t, true, found)
	assert.Equal(t, &resp1, ccc)
	assert.Equal(t, "hit: 1 miss: 1 keys-added: 1 keys-updated: 0 keys-evicted: 0 cost-added: 1 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 2 hit-ratio: 0.50", cc.cache.Metrics.String())

	resp2 := Response{Header: make(http.Header), StatusCode: 200, Body: []byte("test")}
	cc.Set(k1, &resp2)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "hit: 1 miss: 1 keys-added: 1 keys-updated: 1 keys-evicted: 0 cost-added: 4 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 2 hit-ratio: 0.50", cc.cache.Metrics.String())

	ccc, found = cc.Get(k1)
	assert.Equal(t, true, found)
	assert.Equal(t, &resp2, ccc)
	assert.Equal(t, "hit: 2 miss: 1 keys-added: 1 keys-updated: 1 keys-evicted: 0 cost-added: 4 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 3 hit-ratio: 0.67", cc.cache.Metrics.String())

	resp3 := Response{Header: make(http.Header), StatusCode: 200, Body: []byte("overflow")}
	cc.Set(k2, &resp3)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "hit: 2 miss: 1 keys-added: 2 keys-updated: 1 keys-evicted: 1 cost-added: 12 cost-evicted: 4 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 3 hit-ratio: 0.67", cc.cache.Metrics.String())

	ccc, found = cc.Get(k2)
	assert.Equal(t, true, found)
	assert.Equal(t, &resp3, ccc)
	assert.Equal(t, "hit: 3 miss: 1 keys-added: 2 keys-updated: 1 keys-evicted: 1 cost-added: 12 cost-evicted: 4 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 4 hit-ratio: 0.75", cc.cache.Metrics.String())

	resp4 := Response{Header: make(http.Header), StatusCode: 200, Body: []byte("content too big")}
	cc.Set(k3, &resp4)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "hit: 3 miss: 1 keys-added: 2 keys-updated: 1 keys-evicted: 1 cost-added: 12 cost-evicted: 4 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 4 hit-ratio: 0.75", cc.cache.Metrics.String())

	ccc, found = cc.Get(k3)
	assert.Equal(t, false, found)
	assert.Equal(t, "hit: 3 miss: 2 keys-added: 2 keys-updated: 1 keys-evicted: 1 cost-added: 12 cost-evicted: 4 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 5 hit-ratio: 0.60", cc.cache.Metrics.String())
}

func TestDataRace(t *testing.T) {
	fmt.Println("testing data race...")
	var wg sync.WaitGroup
	cc, err := NewContentCache(16, true)
	assert.Empty(t, err)
	k1 := "id1"
	k2 := "id2"
	wg.Add(1)
	go func() {
		defer wg.Done()
		cc.Set(k1, &Response{Header: make(http.Header), StatusCode: 200, Body: []byte("data")})
		time.Sleep(100 * time.Millisecond)
		assert.Empty(t, err)
		cc.Set(k2, &Response{Header: make(http.Header), StatusCode: 200, Body: []byte("race")})
		time.Sleep(100 * time.Millisecond)
		assert.Empty(t, err)
		_, found := cc.Get(k1)
		assert.Equal(t, true, found)
		_, found = cc.Get(k2)
		assert.Equal(t, true, found)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		cc.Set(k1, &Response{Header: make(http.Header), StatusCode: 200, Body: []byte("race")})
		time.Sleep(100 * time.Millisecond)
		assert.Empty(t, err)
		cc.Set(k2, &Response{Header: make(http.Header), StatusCode: 200, Body: []byte("data")})
		time.Sleep(100 * time.Millisecond)
		assert.Empty(t, err)
		_, found := cc.Get(k1)
		assert.Equal(t, true, found)
		_, found = cc.Get(k2)
		assert.Equal(t, true, found)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		cc.Set(k2, &Response{Header: make(http.Header), StatusCode: 200, Body: []byte("data")})
		time.Sleep(100 * time.Millisecond)
		assert.Empty(t, err)
		cc.Set(k1, &Response{Header: make(http.Header), StatusCode: 200, Body: []byte("race")})
		time.Sleep(100 * time.Millisecond)
		assert.Empty(t, err)
		_, found := cc.Get(k1)
		assert.Equal(t, true, found)
		_, found = cc.Get(k2)
		assert.Equal(t, true, found)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		cc.Set(k2, &Response{Header: make(http.Header), StatusCode: 200, Body: []byte("race")})
		time.Sleep(100 * time.Millisecond)
		assert.Empty(t, err)
		cc.Set(k1, &Response{Header: make(http.Header), StatusCode: 200, Body: []byte("data")})
		time.Sleep(100 * time.Millisecond)
		assert.Empty(t, err)
		_, found := cc.Get(k1)
		assert.Equal(t, true, found)
		_, found = cc.Get(k2)
		assert.Equal(t, true, found)
	}()
	wg.Wait()
}
