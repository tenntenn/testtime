package greeting_test

import (
	"greeting"
	"testing"
	"time"

	"github.com/tenntenn/testtime"
)

func TestDo(t *testing.T) {

	if !testtime.Overlayed() {
		t.Skip()
	}

	t.Parallel()
	cases := []struct {
		tm   string
		want string
	}{
		{"04:00:00", "おはよう"},
		{"09:00:00", "おはよう"},
		{"10:00:00", "こんにちは"},
		{"16:00:00", "こんにちは"},
		{"17:00:00", "こんばんは"},
		{"03:00:00", "こんばんは"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.tm, func(t *testing.T) {
			t.Parallel()
			testtime.SetTime(t, parseTime(t, tt.tm))
			got := greeting.Do()
			if got != tt.want {
				t.Errorf("want %s but got %s", tt.want, got)
			}
		})
	}
}

func parseTime(t *testing.T, v string) time.Time {
	t.Helper()
	tm, err := time.Parse("2006/01/02 15:04:05", "2006/01/02 "+v)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	return tm
}
