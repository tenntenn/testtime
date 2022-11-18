package testtime

import (
	"runtime"
	"sync"
	"testing"
	"time"
	_ "unsafe" // for go:linkname
)

//go:linkname timeMap time.timeMap
var timeMap sync.Map

// SetTime stores a fixed time which can be identified by the caller's goroutine.
// The fixed time will be deleted at the end of test.
func SetTime(t *testing.T, tm time.Time) {
	t.Helper()

	key := goroutineID()
	timeMap.Store(key, func() time.Time {
		return tm
	})

	t.Cleanup(func() {
		timeMap.Delete(key)
	})
}

// SetFunc stores a function which returns time.Time which can be identified by the caller's goroutine.
// The function will be deleted at the end of test.
func SetFunc(t *testing.T, f func() time.Time) {
	t.Helper()

	key := goroutineID()
	timeMap.Store(key, f)

	t.Cleanup(func() {
		timeMap.Delete(key)
	})
}

// Now returns a fixed time which is related with the goroutine by SetTime or SetFunc.
// If the current goroutine is not related with any fixed time or function, Now calls time.Now and returns its returned value.
func Now() time.Time {
	v, ok := timeMap.Load(goroutineID())
	if ok {
		return v.(func() time.Time)()
	}
	return time.Now()
}

func goroutineID() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// 10: len("goroutine ")
	for i := 10; i < n; i++ {
		if buf[i] == ' ' {
			return string(buf[10:i])
		}
	}
	return ""
}
