// +build overlaytesttime
//go:build overlaytesttime

package testtime

import (
	"sync"
	_ "unsafe" // for go:linkname
)

//go:linkname timeMap time.timeMap
var timeMap sync.Map
