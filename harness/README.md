# Test Harness

This test harness will try a number of commands in sequence testing put, get,
del, cache misses, invalid arguments and client hangups. It is by no means
exhaustive and runs on rails meaning it will not cope well with malformed 
server responses. It is deliberately basic with the idea that you will replace
it with something far superior.

## Usage

```
go run main.go <address>
```

Example:

```
go run main.go localhost:1234
```
