package testtime_test

import (
	"testing"
	"time"
	_ "unsafe"

	"github.com/tenntenn/testtime"
)

func Test(t *testing.T) {

	t.Run("SetTime", func(t *testing.T) {
		tm := parseTime(t, "2022/11/17 17:21:00")
		testtime.SetTime(t, tm)
		if !testtime.Now().Equal(tm) {
			t.Error("testtime.Now() must be", tm)
		}
	})

	t.Run("SetFunc", func(t *testing.T) {
		tm := parseTime(t, "2022/11/17 17:23:00")
		testtime.SetFunc(t, func() time.Time { return tm })

		if !testtime.Now().Equal(tm) {
			t.Error("testtime.Now() must be", tm)
		}
	})

	testtime.SetTime(t, time.Time{})
	if !testtime.Now().IsZero() {
		t.Error("testtime.Now() must be zero value")
	}
}

func parseTime(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.Parse("2006/01/02 15:04:05", s)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	return tm
}
