package cache //white box testing

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetContent(t *testing.T) {
	cc, err := NewContentCache(128, true) //for some reason Set fails if size <= 60
	assert.Empty(t, err)
	assert.Equal(t, uint64(0), cc.cache.Metrics.CostAdded())
	assert.Equal(t, uint64(0), cc.cache.Metrics.CostEvicted())
	assert.Equal(t, uint64(0), cc.cache.Metrics.KeysAdded())
	assert.Equal(t, uint64(0), cc.cache.Metrics.KeysEvicted())
	assert.Equal(t, uint64(0), cc.cache.Metrics.KeysUpdated())

	ccc, err := cc.GetContent("id1", bytes.NewReader([]byte("a")))
	time.Sleep(100 * time.Millisecond) //https://github.com/dgraph-io/ristretto/issues/161
	assert.Empty(t, err)
	assert.Equal(t, ContentCacheCell{id: "id1", Data: bytes.NewBuffer([]byte("a"))}, ccc)
	assert.Equal(t, "hit: 0 miss: 1 keys-added: 1 keys-updated: 0 keys-evicted: 0 cost-added: 60 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 1 hit-ratio: 0.00", cc.cache.Metrics.String())

	ccc, err = cc.GetContent("id1", bytes.NewReader([]byte("a")))
	time.Sleep(100 * time.Millisecond)
	assert.Empty(t, err)
	assert.Equal(t, ContentCacheCell{id: "id1", Data: bytes.NewBuffer([]byte("a"))}, ccc)
	assert.Equal(t, "hit: 1 miss: 1 keys-added: 1 keys-updated: 0 keys-evicted: 0 cost-added: 60 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 2 hit-ratio: 0.50", cc.cache.Metrics.String())

	ccc, err = cc.GetContent("id2", bytes.NewReader([]byte("test")))
	time.Sleep(100 * time.Millisecond)	
	assert.Empty(t, err)
	assert.Equal(t, ContentCacheCell{id: "id2", Data: bytes.NewBuffer([]byte("test"))}, ccc)
	assert.Equal(t, "hit: 1 miss: 2 keys-added: 2 keys-updated: 0 keys-evicted: 0 cost-added: 120 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 3 hit-ratio: 0.33", cc.cache.Metrics.String())

	ccc, err = cc.GetContent("id2", bytes.NewReader([]byte("test")))
	time.Sleep(100 * time.Millisecond)
	assert.Empty(t, err)
	assert.Equal(t, ContentCacheCell{id: "id2", Data: bytes.NewBuffer([]byte("test"))}, ccc)
	assert.Equal(t, "hit: 2 miss: 2 keys-added: 2 keys-updated: 0 keys-evicted: 0 cost-added: 120 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 4 hit-ratio: 0.50", cc.cache.Metrics.String())
}
