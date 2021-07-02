package testtime_test

import (
	"testing"
	"time"
	_ "unsafe"

	"github.com/tenntenn/testtime"
)

func Test(t *testing.T) {
	func() {
		testtime.Set(time.Time{})
		if !testtime.Now().IsZero() {
			t.Error("testtime.Now() must be zero value")
		}

		if !testtime.Now().IsZero() {
			t.Error("testtime.Now() must be zero value")
		}

		func() {
			if !testtime.Now().IsZero() {
				t.Error("testtime.Now() must be zero value")
			}
		}()

		done := make(chan struct{})
		go func() {
			if testtime.Now().IsZero() {
				t.Error("testtime.Now() must not be zero value")
			}
			close(done)
		}()
		<-done
	}()
	if testtime.Now().IsZero() {
		t.Error("testtime.Now() must not be zero value")
	}
}
