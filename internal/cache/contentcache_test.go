package cache //white box testing

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetContent(t *testing.T) {
	cc, err := NewContentCache(8, true)
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
	assert.Equal(t, "hit: 0 miss: 1 keys-added: 1 keys-updated: 0 keys-evicted: 0 cost-added: 1 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 1 hit-ratio: 0.00", cc.cache.Metrics.String())

	ccc, err = cc.GetContent("id1", bytes.NewReader([]byte("a")))
	time.Sleep(100 * time.Millisecond)
	assert.Empty(t, err)
	assert.Equal(t, ContentCacheCell{id: "id1", Data: bytes.NewBuffer([]byte("a"))}, ccc)
	assert.Equal(t, "hit: 1 miss: 1 keys-added: 1 keys-updated: 0 keys-evicted: 0 cost-added: 1 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 2 hit-ratio: 0.50", cc.cache.Metrics.String())

	ccc, err = cc.GetContent("id2", bytes.NewReader([]byte("test")))
	time.Sleep(100 * time.Millisecond)
	assert.Empty(t, err)
	assert.Equal(t, ContentCacheCell{id: "id2", Data: bytes.NewBuffer([]byte("test"))}, ccc)
	assert.Equal(t, "hit: 1 miss: 2 keys-added: 2 keys-updated: 0 keys-evicted: 0 cost-added: 5 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 3 hit-ratio: 0.33", cc.cache.Metrics.String())

	//below should not happen as hash is used as id
	ccc, err = cc.GetContent("id2", bytes.NewReader([]byte("update")))
	time.Sleep(100 * time.Millisecond)
	assert.Empty(t, err)
	assert.Equal(t, ContentCacheCell{id: "id2", Data: bytes.NewBuffer([]byte("test"))}, ccc)
	assert.Equal(t, "hit: 2 miss: 2 keys-added: 2 keys-updated: 0 keys-evicted: 0 cost-added: 5 cost-evicted: 0 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 4 hit-ratio: 0.50", cc.cache.Metrics.String())

	ccc, err = cc.GetContent("id3", bytes.NewReader([]byte("overflow")))
	time.Sleep(100 * time.Millisecond)
	assert.Empty(t, err)
	assert.Equal(t, ContentCacheCell{id: "id3", Data: bytes.NewBuffer([]byte("overflow"))}, ccc)
	assert.Equal(t, "hit: 2 miss: 3 keys-added: 3 keys-updated: 0 keys-evicted: 2 cost-added: 13 cost-evicted: 5 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 5 hit-ratio: 0.40", cc.cache.Metrics.String())

	//below should not happen as content size is checked before putting in cache
	ccc, err = cc.GetContent("id4", bytes.NewReader([]byte("content too big")))
	time.Sleep(100 * time.Millisecond)
	assert.Empty(t, err)
	assert.Equal(t, ContentCacheCell{id: "id4", Data: bytes.NewBuffer([]byte("content "))}, ccc)
	assert.Equal(t, "hit: 2 miss: 4 keys-added: 4 keys-updated: 0 keys-evicted: 3 cost-added: 21 cost-evicted: 13 sets-dropped: 0 sets-rejected: 0 gets-dropped: 0 gets-kept: 0 gets-total: 6 hit-ratio: 0.33", cc.cache.Metrics.String())
}
