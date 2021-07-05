# testtime

[![pkg.go.dev][gopkg-badge]][gopkg]

`testtime` package provides `timetime.Now()` and `testtime.Set()`.
`testtime.Set()` stores a fixed time to a map with a caller of `testtime.Set()` as a key.
When a caller or its ancestor caller of `timetime.Now()` is related to a fixed time by `testtime.Set()`, it returns the fixed time otherwise it returns current time which is returned by `time.Now()`.

```go
package main

import (
	"fmt"
	"time"
	"testing"

	"github.com/tenntenn/testtime"
)

func Test(t *testing.T) {
	func() {
		// set zero value
		testtime.Set(time.Time{})
		// true
		if time.Now().IsZero {
			t.Error("error")
		}
	}()
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
    "/usr/local/go/src/time/time.go": "/Users/tenntenn/go/pkg/testtime/time_go1.16.go"
  }
}
$ diff /usr/local/go/src/time/time.go /Users/tenntenn/go/pkg/testtime/time_go1.16.go
79a80,81
> 	"runtime"
> 	"sync"
1066c1068
< func Now() Time {
---
> func _Now() Time {
1521a1524,1556
> 
> // It will be added to GOROOT/src/time/time.go.
> 
> var timeMap testtime_sync.Map
> 
> func Now() Time {
> 	pcs := make([]uintptr, 10)
> 	n := testtime_runtime.Callers(1, pcs)
> 	frames := testtime_runtime.CallersFrames(pcs[:n])
> 	for {
> 		frame, hasNext := frames.Next()
> 		tm, ok := timeMap.Load(frame.Function)
> 		if ok {
> 			return tm.(Time)
> 		}
> 
> 		if !hasNext {
> 			break
> 		}
> 	}
> 	return _Now()
> }
> 
> func funcName(skip int) (string, bool) {
> 	pc, _, _, ok := testtime_runtime.Caller(skip + 1)
> 	if !ok {
> 		return "", false
> 	}
> 	fnc := testtime_runtime.FuncForPC(pc)
> 	return fnc.Name(), true
> }
> 
> // End of testtime's code
```

<!-- links -->
[gopkg]: https://pkg.go.dev/github.com/tenntenn/testtime
[gopkg-badge]: https://pkg.go.dev/badge/github.com/tenntenn/testtime?status.svg
