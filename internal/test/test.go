package test

import "time"

// Now is used by test of testtime.
// see ../../time_test.go
func Now() time.Time {
	return time.Now()
}
