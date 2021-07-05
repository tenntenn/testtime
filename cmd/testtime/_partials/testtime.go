// It will be added to GOROOT/src/time/time.go.

var timeMap testtime_sync.Map

func Now() Time {
	pcs := make([]uintptr, 10)
	n := testtime_runtime.Callers(1, pcs)
	frames := testtime_runtime.CallersFrames(pcs[:n])
	for {
		frame, hasNext := frames.Next()
		tm, ok := timeMap.Load(frame.Function)
		if ok {
			return tm.(Time)
		}

		if !hasNext {
			break
		}
	}
	return _Now()
}

func funcName(skip int) (string, bool) {
	pc, _, _, ok := testtime_runtime.Caller(skip + 1)
	if !ok {
		return "", false
	}
	fnc := testtime_runtime.FuncForPC(pc)
	return fnc.Name(), true
}

// End of testtime's code
