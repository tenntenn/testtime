package testtime

import (
	_ "unsafe" // for go:linkname
)

//go:linkname overlayed time.overlayed
var overlayed bool

// Overlayed returns whether time.go in time package was overlayed by testtime or not.
func Overlayed() bool {
	return overlayed
}
