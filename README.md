# testtime

[![pkg.go.dev][gopkg-badge]][gopkg]

## How to use

`testtime` package provides `testtime.Now()` and `testtime.SetTime()`.
`testtime.SetTime()` stores a fixed time to a map with goroutine ID  of `testtime.SetTime()` as a key.
When goroutine ID of `testtime.Now()` is related to a fixed time by `testtime.SetTime()`, it returns the fixed time otherwise it returns current time which is returned by `time.Now()`.

```go
package main

import (
	"fmt"
	"time"
	"testing"

	"github.com/tenntenn/testtime"
)

func Test(t *testing.T) {

	t.Run("A", func(t *testing.T) {
		// set zero value
		testtime.SetTime(t, time.Time{})
		// true
		if time.Now().IsZero {
			t.Error("error")
		}
	})

	t.Run("B", func(t *testing.T) {
		// set func which return zero value
		f := func() time.Time {
			return time.Time{}
		}
		testtime.SetFunc(t, f)
		// true
		if time.Now().IsZero {
			t.Error("error")
		}
	})

	// false
	if !time.Now().IsZero {
		t.Error("error")
	}
}
```

The `testtime` command replace `time.Now` to `testtime.Now`.
It prints a file path of overlay JSON which can be given to `-overlay` flag of `go test` command.

```sh
$ go install github.com/tenntenn/testtime/cmd/testtime@latest
$ go test -overlay=`testtime`
PASS
ok  	main	0.156s
```

The `testtime` command creates an overlay JSON file and `time.go` which is replaced `time.Now` in `$GOPATH/pkg/testtime` directory. The `testtime` command does not update these files without `-u` flag.

```sh
$ cat `testtime` | jq
{
  "Replace": {
    "/usr/local/go/src/time/time.go": "/Users/tenntenn/go/pkg/testtime/time_go1.23.1.go"
  }
}

$ diff /usr/local/go/src/time/time.go /Users/tenntenn/go/pkg/testtime/time_go1.23.1.go
94a95,96
> 	"runtime"
> 	"sync"
1159c1161
< func Now() Time {
---
> func _Now() Time {
1695a1698,1726
> 
> // It will be added to GOROOT/src/time/time.go.
> 
> //go:linkname timeMap
> var timeMap sync.Map
> 
> // Now returns a fixed time which is related with the goroutine by SetTime or SetFunc.
> // If the current goroutine is not related with any fixed time or function, Now calls time.Now and returns its returned value.
> func Now() Time {
> 	v, ok := timeMap.Load(goroutineID())
> 	if ok {
> 		return v.(func() Time)()
> 	}
> 	return _Now()
> }
> 
> func goroutineID() string {
> 	var buf [64]byte
> 	n := runtime.Stack(buf[:], false)
> 	// 10: len("goroutine ")
> 	for i := 10; i < n; i++ {
> 		if buf[i] == ' ' {
> 			return string(buf[10:i])
> 		}
> 	}
> 	return ""
> }
> 
> // End of testtime's code
```

## Examples

See [_examples](./_examples) directory.

<!-- links -->
[gopkg]: https://pkg.go.dev/github.com/tenntenn/testtime
[gopkg-badge]: https://pkg.go.dev/badge/github.com/tenntenn/testtime?status.svg
