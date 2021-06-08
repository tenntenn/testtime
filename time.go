package testtime

import (
	"runtime"
	"sync"
	"time"
)

var timeMap sync.Map

// Set sets a fixed time with its caller.
func Set(tm time.Time) bool {
	name, ok := funcName(1)
	if !ok {
		return false
	}
	timeMap.Store(name, tm)
	return true
}

// Now returns a fixed time which is related with the caller function by Set.
// If the caller is not related with  any fixed time, Now calls time.Now and returns its returned value.
// Now can replaces time.Now by go:linkname as a following.
//
//	//go:linkname now time.Now
//	func now() time.Time {
//		return testtime.Now()
//	}
//
//	func f() {
//		func() {
//			// set zero value
//			testtime.Set(time.Time{})
//			// true
//			fmt.Println(time.Now().IsZero())
//		}()
//		// false
//		fmt.Println(time.Now().IsZero())
//	}
func Now() time.Time {
	pcs := make([]uintptr, 10)
	n := runtime.Callers(1, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, hasNext := frames.Next()
		tm, ok := timeMap.Load(frame.Function)
		if ok {
			return tm.(time.Time)
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
