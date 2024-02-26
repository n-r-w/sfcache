package sfcache

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/n-r-w/singleflight/v2"
)

// Group singleflight group with LRU cache
type Group[K comparable, V any] struct {
	muClear sync.RWMutex

	cache   *expirable.LRU[K, V]
	size    int
	ttl     time.Duration
	onEvict expirable.EvictCallback[K, V]

	sfgroup singleflight.Group[K, V]
}

// New creates a new singleflight Group with the specified LRU cache size and TTL.
// Size parameter set to 0 makes cache of unlimited size, e.g. turns LRU mechanism off.
// Providing 0 TTL turns expiring off.
func New[K comparable, V any](size int, onEvict expirable.EvictCallback[K, V], ttl time.Duration) *Group[K, V] {
	return &Group[K, V]{
		size:    size,
		ttl:     ttl,
		onEvict: onEvict,
		cache:   expirable.NewLRU[K, V](size, onEvict, ttl),
	}
}

// Do executes and returns the results of the given function, making sure that
// only one execution is in-flight for a given key at a time. If a duplicate
// comes in, the duplicate caller waits for the original to complete and
// receives the same results.
//
// The context passed to the fn function is a context that preserves all values
// from the passed context but is cancelled by the singleflight only when all
// awaiting caller's contexts are cancelled (no caller is awaiting the result).
// If there are multiple callers, context passed to one caller does not affect
// the execution and returned values of others except if the function result is
// dependent on the context values.
//
// If useCache is false, it will work as a simple singleflight.
//
// The return value shared indicates whether v was given to multiple callers.
func (g *Group[K, V]) Do(ctx context.Context,
	key K,
	useCache bool,
	fn func(ctx context.Context) (V, error),
) (val V, shared bool, err error) {
	g.muClear.RLock()
	defer g.muClear.RUnlock()

	if useCache {
		if v, ok := g.cache.Get(key); ok {
			return v, true, nil
		}
	}

	val, shared, err = g.sfgroup.Do(ctx, key, func(ctxFunc context.Context) (V, error) {
		v, err1 := fn(ctxFunc)
		if err1 != nil {
			var zeroV V
			return zeroV, err1
		}

		return v, nil
	})

	if err != nil {
		return val, shared, err
	}

	if useCache {
		_ = g.cache.Add(key, val)
	}

	return val, shared, err
}

// Clear clears the cache
func (g *Group[K, V]) Clear() {
	g.muClear.Lock()
	defer g.muClear.Unlock()

	g.cache = expirable.NewLRU[K, V](g.size, g.onEvict, g.ttl)
}
