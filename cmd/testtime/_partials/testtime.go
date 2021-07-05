// It will be added to GOROOT/src/time/time.go.

var timeMap sync.Map

func Now() Time {
	pcs := make([]uintptr, 10)
	n := runtime.Callers(1, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, hasNext := frames.Next()
		v, ok := timeMap.Load(goroutineID() + ":" + frame.Function)
		if ok {
			return v.(func() Time)()
		}

		if !hasNext {
			break
		}
	}
	return _Now()
}

func funcName(skip int) (string, bool) {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", false
	}
	fnc := runtime.FuncForPC(pc)

	return goroutineID() + ":" + fnc.Name(), true
}

func goroutineID() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	for i := 10; i < n; i++ {
		if buf[i] == ' ' {
			return string(buf[10:i])
		}
	}
	return ""
}

// End of testtime's code
