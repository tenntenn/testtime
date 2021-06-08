# testtime

[![pkg.go.dev][gopkg-badge]][gopkg]

`testtime` provides `time.Now` for testing.

https://play.golang.org/p/ML5nhtXLOWA

```go
package main

import (
	"fmt"
	"time"
	_ "unsafe" // for go:linkname

	"github.com/tenntenn/testtime"
)

// replace time.Now
//go:linkname now time.Now
func now() time.Time {
	return testtime.Now()
}

func main() {
	func() {
		// set zero value
		testtime.Set(time.Time{})
		// true
		fmt.Println(time.Now().IsZero())
	}()
	// false
	fmt.Println(time.Now().IsZero())
}
```

<!-- links -->
[gopkg]: https://pkg.go.dev/github.com/tenntenn/testtime
[gopkg-badge]: https://pkg.go.dev/badge/github.com/tenntenn/testtime?status.svg
