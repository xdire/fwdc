package fwdc

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

const testFuncDelay = time.Duration(100 * time.Millisecond)

type testFetcher struct {
	Delay time.Duration
}

func (f *testFetcher) Fetch(key string) (string, error) {
	res, err := fetch(key)
	if err != nil {
		return "", err
	}
	return res.(string), nil
}

func fetch(key any) (any, error) {
	time.Sleep(testFuncDelay)
	return fmt.Sprintf("for key value is, %s!", key), nil
}

func TestCacheConcurrent(t *testing.T) {
	// Create a new cache with int as value type
	cache := New[string, string](&testFetcher{testFuncDelay})
	// Use a wait group to manage concurrent access
	wg := sync.WaitGroup{}
	// Simulate multiple concurrent requests
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Send duplicate request for each key
			if i > 5 {
				i = i % 5
			}
			// Use simple Get functions
			if i < 10 {
				key := fmt.Sprintf("test_key_%d", i)
				_, err := cache.Get(key)
				if err != nil {
					t.Fatalf("Failed to get key %v: %v", key, err)
				}
			}
			// Use GetFn
			if i < 10 {
				key := fmt.Sprintf("test_key_%d", i)
				_, err := cache.GetFn(key, fetch)
				if err != nil {
					t.Fatalf("Failed to get key %v: %v", key, err)
				}
			}
		}(i)
	}
	// Wait for all requests to complete
	wg.Wait()

	// Just to ensure the cache is not empty
	_, errNonEmpty := cache.Get("test_key")
	assert.NoError(t, errNonEmpty)

	// Test keys consistency
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("test_key_%d", i)
		val, err := cache.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("for key value is, %s!", key), val)
	}

}
