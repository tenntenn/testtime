package testtime_test

import (
	"testing"
	"time"
	_ "unsafe"

	"github.com/tenntenn/testtime"
	"github.com/tenntenn/testtime/internal/test"
)

//go:linkname now time.Now
func now() time.Time {
	return testtime.Now()
}

func Test(t *testing.T) {
	func() {
		testtime.Set(time.Time{})
		if !time.Now().IsZero() {
			t.Error("time.Now() must be zero value")
		}

		if !test.Now().IsZero() {
			t.Error("time.Now() must be zero value")
		}

		func() {
			if !time.Now().IsZero() {
				t.Error("time.Now() must be zero value")
			}
		}()

		done := make(chan struct{})
		go func() {
			if time.Now().IsZero() {
				t.Error("time.Now() must not be zero value")
			}
			close(done)
		}()
		<-done
	}()
	if time.Now().IsZero() {
		t.Error("time.Now() must not be zero value")
	}
}
