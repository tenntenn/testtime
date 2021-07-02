# testtime

[![pkg.go.dev][gopkg-badge]][gopkg]

`testtime` provides `time.Now` for testing.

WARNING: This package is an experimental project.

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

```sh
$ go install github.com/tenntenn/testtime/cmd/gotesttime@latest
$ gotesttime
PASS
ok  	main	0.156s
```

<!-- links -->
[gopkg]: https://pkg.go.dev/github.com/tenntenn/testtime
[gopkg-badge]: https://pkg.go.dev/badge/github.com/tenntenn/testtime?status.svg
