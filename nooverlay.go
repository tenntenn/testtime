// +build !overlaytesttime
//go:build !overlaytesttime

package testtime

import "sync"

var timeMap sync.Map
