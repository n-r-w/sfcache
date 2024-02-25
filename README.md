[![Go Reference](https://pkg.go.dev/badge/github.com/n-r-w/sfcache.svg)](https://pkg.go.dev/github.com/n-r-w/sfcache)
[![Go Coverage](https://github.com/n-r-w/sfcache/wiki/coverage.svg)](https://raw.githack.com/wiki/n-r-w/sfcache/coverage.html)
![CI Status](https://github.com/n-r-w/sfcache/actions/workflows/go.yml/badge.svg)
[![Stability](http://badges.github.io/stability-badges/dist/stable.svg)](http://github.com/badges/stability-badges)
[![Go Report](https://goreportcard.com/badge/github.com/n-r-w/sfcache)](https://goreportcard.com/badge/github.com/n-r-w/sfcache)

# sfcache

Singleflight group with internal LRU cache

## Usage

Singleflight is a concurrency method to prevent duplicate work from being executed due to multiple calls for the same resource.
The internal LRU cache is used to store the result of the function call and bypass the singleflight call if the result is already in the cache.

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/n-r-w/sfcache"
)

func main() {
    const (
    // The number of elements in the cache
    cacheSize = 100
    // TTL for the cache
    cacheTTL = 10 * time.Minute
    useCache = true
    )

    // evictCallback can be nil if you don't want to do anything when an item is evicted
    cache := sfcache.New[int, string](cacheSize, evictCallback, cacheTTL)

    // make come call
    key := 1
    val, shared, err := cache.Do(context.Background(), key, useCache, func(ctx context.Context) (string, error) {
        return worker(ctx, key)
    })
    if err != nil {
        panic(err)
    }
    log.Println("Value:", val, "Shared:", shared)

    // make some other call with the same key. This time it should be cached and worker should not be called.
    val, shared, err = cache.Do(context.Background(), key, useCache, func(ctx context.Context) (string, error) {
        return worker(ctx, key)
    })
    if err != nil {
        panic(err)
    }
    log.Println("Value:", val, "Shared:", shared)
}

func evictCallback(key int, value string) {
    log.Println("Evicting", key, value)
}

func worker(_ context.Context, key int) (string, error) {
    log.Println("Doing some work for", key)
    return "value", nil
}
```
