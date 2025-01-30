package fwdc

import (
	"fmt"
	"sync"
)

const fetchFail = "failed to fetch key"
const cacheTypeFail = "cache value type error"

type FetcherFn func(key any) (any, error)

// Cache Code should reflect the implementation of forward memory cache implemented in Golang with generics
type Cache[K any, T any] interface {
	Get(key K) (T, error)
	GetFn(key K, fetcher FetcherFn) (T, error)
}

type Fetcher[K any, T any] interface {
	Fetch(key K) (T, error)
}

type fetchOp[T any] struct {
	signal chan struct{}
	res    T
}

type Manager[K comparable, T any] struct {
	cache        map[K]T
	cacheInFetch map[K]*fetchOp[T]
	mu           sync.RWMutex
	fetcher      Fetcher[K, T]
}

func New[K comparable, T any](f Fetcher[K, T]) *Manager[K, T] {
	return &Manager[K, T]{
		cache:        make(map[K]T),
		cacheInFetch: make(map[K]*fetchOp[T]),
		fetcher:      f,
	}
}

// Get or add value for the key with trying to utilize the function in the constructor
func (m *Manager[K, T]) Get(key K) (T, error) {
	return m.getFn(key, func(key any) (any, error) {
		return m.fetcher.Fetch(key.(K))
	})
}

// GetFn or add value using dedicated function
func (m *Manager[K, T]) GetFn(key K, fn FetcherFn) (T, error) {
	return m.getFn(key, fn)
}

func (m *Manager[K, T]) getFn(key K, fn FetcherFn) (T, error) {

	// Try to fetch from the memory first
	m.mu.RLock()
	if v, ok := m.cache[key]; ok {
		m.mu.RUnlock()
		return v, nil
	}

	if v, ok := m.cacheInFetch[key]; ok {
		m.mu.RUnlock()
		<-v.signal
		return v.res, nil
	}
	m.mu.RUnlock()

	// Now do the lock for fetch init
	m.mu.Lock()
	// Recheck that at the moment of locking we still doesn't have value in cache
	if v, ok := m.cache[key]; ok {
		m.mu.Unlock()
		return v, nil
	}
	// Nor in transit
	if v, ok := m.cacheInFetch[key]; ok {
		m.mu.Unlock()
		<-v.signal
		return v.res, nil
	}

	// Initiate fetch operation
	op := &fetchOp[T]{
		signal: make(chan struct{}, 1),
	}
	m.cacheInFetch[key] = op

	// Release lock and proceed with fetch for this thread
	m.mu.Unlock()
	data, err := fn(key)
	if err != nil {
		// Empty assert for err return
		v, _ := data.(T)
		return v, fmt.Errorf("%s %v: %w", fetchFail, key, err)
	}

	// Fail if function returned value that cannot be cast
	if v, ok := data.(T); !ok {
		return v, fmt.Errorf("%s %v: %w", cacheTypeFail, key, err)
	}

	// Write the result into the cache and clear the fetch for it
	m.mu.Lock()
	m.cache[key] = data.(T)
	delete(m.cacheInFetch, key)
	m.mu.Unlock()

	// Free the
	op.res = data.(T)
	op.signal <- struct{}{}

	close(op.signal)

	return data.(T), nil
}
