package cache //white box testing

import (
	"bytes"
	"testing"
	"sync"
	"time"
	"fmt"
	"io"

	"github.com/stretchr/testify/assert"
)

type CacheKeyMock string

func TestGetContent(t *testing.T) {
	fmt.Println("testing GetContent...")
	load := func(r io.ReadSeeker) (*bytes.Buffer, int64, error) {
		b, err := io.ReadAll(r)
		nb := bytes.NewBuffer(b)
		if err != nil {
			return nb, 0, err
		}
		return nb, int64(nb.Len()), nil 
	}
	cc, err := NewContentCache[CacheKeyMock](8, true, load)
	assert.Empty(t, err)

	k1 := CacheKeyMock("id1")
	k2 := CacheKeyMock("id2")
	k3 := CacheKeyMock("id3")

	ccc, found := cc.GetContent(k1)
	assert.Equal(t, false, found)
	assert.Equal(t, "hit: 0 miss: 1 keys-added: 0 keys-updated: 0 keys-evicted: 0 cost-added: 0 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 1 hit-ratio: 0.00", cc.cache.Metrics.String())

	ccc, err = cc.SetContent(k1, bytes.NewReader([]byte("a")))
	time.Sleep(100 * time.Millisecond) //https://github.com/dgraph-io/ristretto/issues/161
	assert.Empty(t, err)
	assert.Equal(t, bytes.NewBuffer([]byte("a")), ccc)
	assert.Equal(t, "hit: 0 miss: 1 keys-added: 1 keys-updated: 0 keys-evicted: 0 cost-added: 1 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 1 hit-ratio: 0.00", cc.cache.Metrics.String())

	ccc, found = cc.GetContent(k1)
	assert.Equal(t, true, found)
	assert.Equal(t, bytes.NewBuffer([]byte("a")), ccc)
	assert.Equal(t, "hit: 1 miss: 1 keys-added: 1 keys-updated: 0 keys-evicted: 0 cost-added: 1 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 2 hit-ratio: 0.50", cc.cache.Metrics.String())

	ccc, err = cc.SetContent(k1, bytes.NewReader([]byte("test")))
	time.Sleep(100 * time.Millisecond)
	assert.Empty(t, err)
	assert.Equal(t, bytes.NewBuffer([]byte("test")), ccc)
	assert.Equal(t, "hit: 1 miss: 1 keys-added: 1 keys-updated: 1 keys-evicted: 0 cost-added: 4 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 2 hit-ratio: 0.50", cc.cache.Metrics.String())

	ccc, found = cc.GetContent(k1)
	assert.Equal(t, true, found)
	assert.Equal(t, bytes.NewBuffer([]byte("test")), ccc)
	assert.Equal(t, "hit: 2 miss: 1 keys-added: 1 keys-updated: 1 keys-evicted: 0 cost-added: 4 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 3 hit-ratio: 0.67", cc.cache.Metrics.String())

	ccc, err = cc.SetContent(k2, bytes.NewReader([]byte("overflow")))
	time.Sleep(100 * time.Millisecond)
	assert.Empty(t, err)
	assert.Equal(t, bytes.NewBuffer([]byte("overflow")), ccc)
	assert.Equal(t, "hit: 2 miss: 1 keys-added: 2 keys-updated: 1 keys-evicted: 1 cost-added: 12 cost-evicted: 4 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 3 hit-ratio: 0.67", cc.cache.Metrics.String())

	ccc, found = cc.GetContent(k2)
	assert.Equal(t, true, found)
	assert.Equal(t, bytes.NewBuffer([]byte("overflow")), ccc)
	assert.Equal(t, "hit: 3 miss: 1 keys-added: 2 keys-updated: 1 keys-evicted: 1 cost-added: 12 cost-evicted: 4 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 4 hit-ratio: 0.75", cc.cache.Metrics.String())

	ccc, err = cc.SetContent(k3, bytes.NewReader([]byte("content too big")))
	time.Sleep(100 * time.Millisecond)
	assert.Empty(t, err)
	assert.Equal(t, bytes.NewBuffer([]byte("content too big")), ccc)
	assert.Equal(t, "hit: 3 miss: 1 keys-added: 2 keys-updated: 1 keys-evicted: 1 cost-added: 12 cost-evicted: 4 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 4 hit-ratio: 0.75", cc.cache.Metrics.String())

	ccc, found = cc.GetContent(k3)
	assert.Equal(t, false, found)
	assert.Equal(t, "hit: 3 miss: 2 keys-added: 2 keys-updated: 1 keys-evicted: 1 cost-added: 12 cost-evicted: 4 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 5 hit-ratio: 0.60", cc.cache.Metrics.String())
}

func TestDataRace(t *testing.T) {
	fmt.Println("testing data race...")
	var wg sync.WaitGroup
	load := func(r io.ReadSeeker) (*bytes.Buffer, int64, error) {
		b, err := io.ReadAll(r)
		nb := bytes.NewBuffer(b)
		if err != nil {
			return nb, 0, err
		}
		return nb, int64(nb.Len()), nil 
	}
	cc, err := NewContentCache[CacheKeyMock](16, true, load)
	assert.Empty(t, err)
	k1 := CacheKeyMock("id1")
	k2 := CacheKeyMock("id2")
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := cc.SetContent(k1, bytes.NewReader([]byte("data")))
		time.Sleep(100 * time.Millisecond) 
		assert.Empty(t, err)
		_, err = cc.SetContent(k2, bytes.NewReader([]byte("race")))
		time.Sleep(100 * time.Millisecond) 
		assert.Empty(t, err)
		_, found := cc.GetContent(k1)
		assert.Equal(t, true, found)
		_, found = cc.GetContent(k2)
		assert.Equal(t, true, found)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := cc.SetContent(k1, bytes.NewReader([]byte("race")))
		time.Sleep(100 * time.Millisecond) 
		assert.Empty(t, err)
		_, err = cc.SetContent(k2, bytes.NewReader([]byte("data")))
		time.Sleep(100 * time.Millisecond) 
		assert.Empty(t, err)
		_, found := cc.GetContent(k1)
		assert.Equal(t, true, found)
		_, found = cc.GetContent(k2)
		assert.Equal(t, true, found)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := cc.SetContent(k2, bytes.NewReader([]byte("data")))
		time.Sleep(100 * time.Millisecond) 
		assert.Empty(t, err)
		_, err = cc.SetContent(k1, bytes.NewReader([]byte("race")))
		time.Sleep(100 * time.Millisecond) 
		assert.Empty(t, err)
		_, found := cc.GetContent(k1)
		assert.Equal(t, true, found)
		_, found = cc.GetContent(k2)
		assert.Equal(t, true, found)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := cc.SetContent(k2, bytes.NewReader([]byte("race")))
		time.Sleep(100 * time.Millisecond) 
		assert.Empty(t, err)
		_, err = cc.SetContent(k1, bytes.NewReader([]byte("data")))
		time.Sleep(100 * time.Millisecond) 
		assert.Empty(t, err)
		_, found := cc.GetContent(k1)
		assert.Equal(t, true, found)
		_, found = cc.GetContent(k2)
		assert.Equal(t, true, found)
	}()
	wg.Wait()
}
