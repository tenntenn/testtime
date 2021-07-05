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

// SetTime sets a fixed time with its caller.
func SetTime(t *testing.T, tm time.Time) bool {
	t.Helper()
	name, ok := funcName(1)
	if !ok {
		return false
	}
	timeMap.Store(name, func() time.Time {
		return tm
	})

	t.Cleanup(func() {
		timeMap.Delete(name)
	})

	return true
}

// SetFunc sets a function which returns time.Time.
func SetFunc(t *testing.T, f func() time.Time) bool {
	t.Helper()
	name, ok := funcName(1)
	if !ok {
		return false
	}
	timeMap.Store(name, f)

	t.Cleanup(func() {
		timeMap.Delete(name)
	})

	return true
}

// Now returns a fixed time which is related with the caller function by Set.
// If the caller is not related with any fixed time, Now calls time.Now and returns its returned value.
func Now() time.Time {
	pcs := make([]uintptr, 10)
	n := runtime.Callers(1, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, hasNext := frames.Next()
		v, ok := timeMap.Load(frame.Function)
		if ok {
			return v.(func() time.Time)()
		}

		if !hasNext {
			break
		}
	}
	return time.Now()
}

func funcName(skip int) (string, bool) {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", false
	}
	fnc := runtime.FuncForPC(pc)
	return fnc.Name(), true
}
