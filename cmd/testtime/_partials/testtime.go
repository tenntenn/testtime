// It will be added to GOROOT/src/time/time.go.

//go:linkname timeMap
var timeMap sync.Map

var overlayed = true

// Now returns a fixed time which is related with the goroutine by SetTime or SetFunc.
// If the current goroutine is not related with any fixed time or function, Now calls time.Now and returns its returned value.
func Now() Time {
	v, ok := timeMap.Load(goroutineID())
	if ok {
		return v.(func() Time)()
	}
	return _Now()
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

// End of testtime's code
