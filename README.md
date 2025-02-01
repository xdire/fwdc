# Forwardcache
Forward cache helper library

## What it does
Thread-safe abstraction to fetch and store some data as a memory
cache, allows multiple requesters to get the key value that might be
present somewhere else and can be fetched with an abstract function call

## Usage
Refer the test in [Cache test](cache_test.go)

Create simple fetch function
```go
func fetch(key any) (any, error) {
	time.Sleep(testFuncDelay)
	return fmt.Sprintf("for key value is, %s!", key), nil
}
```
Or create or struct with `Fetch` interface
```go
func (f *testFetcher) Fetch(key string) (string, error) {
	res, err := fetch(key)
	if err != nil {
		return "", err
	}
	return res.(string), nil
}
```
Create cache instance
```go
cache := NewWithConfig[string, string](Configuration[string, string]{
		&testFetcher{testFuncDelay},
	})
```
Get the key with function initialized with constructor
```go
val, err := cache.Get(key)
if err != nil {
    t.Fatalf("Failed to get key %v: %v", key, err)
}
```
Or pass some function along
```go
val, err := cache.GetFn(key, fetch)
if err != nil {
    t.Fatalf("Failed to get key %v: %v", key, err)
}
```
