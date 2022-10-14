// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/cortexproject/cortex/blob/master/pkg/storegateway/chunk_bytes_pool.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package storegateway

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/thanos-io/thanos/pkg/pool"
	"github.com/thanos-io/thanos/pkg/store/storepb"
)

type chunkBytesPool struct {
	pool *pool.BucketedBytes

	// Metrics.
	requestedBytes prometheus.Counter
	returnedBytes  prometheus.Counter
}

func newChunkBytesPool(minBucketSize, maxBucketSize int, maxChunkPoolBytes uint64, reg prometheus.Registerer) (*chunkBytesPool, error) {
	upstream, err := pool.NewBucketedBytes(minBucketSize, maxBucketSize, 2, maxChunkPoolBytes)
	if err != nil {
		return nil, err
	}

	return &chunkBytesPool{
		pool: upstream,
		requestedBytes: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "cortex_bucket_store_chunk_pool_requested_bytes_total",
			Help: "Total bytes requested to chunk bytes pool.",
		}),
		returnedBytes: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "cortex_bucket_store_chunk_pool_returned_bytes_total",
			Help: "Total bytes returned by the chunk bytes pool.",
		}),
	}, nil
}

func (p *chunkBytesPool) Get(sz int) (*[]byte, error) {
	buffer, err := p.pool.Get(sz)
	if err != nil {
		return buffer, err
	}

	p.requestedBytes.Add(float64(sz))
	p.returnedBytes.Add(float64(cap(*buffer)))

	return buffer, err
}

func (p *chunkBytesPool) Put(b *[]byte) {
	p.pool.Put(b)
}

// chunksPool is a memory pool of chunk objects.
type chunksPool struct {
	pool sync.Pool
}

// newChunksPool creates a new chunks pool.
func newChunksPool() *chunksPool {
	return &chunksPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &storepb.Chunk{}
			},
		},
	}
}

// Get returns a new chunk from the pool.
func (p *chunksPool) get() *storepb.Chunk {
	return p.pool.Get().(*storepb.Chunk)
}

// put returns a chunk to the pool.
func (p *chunksPool) put(chk *storepb.Chunk) {
	chk.Data = nil
	p.pool.Put(chk)
}
