package closure

import "runtime/pprof"

var profile *pprof.Profile

func init() {
	profile = pprof.NewProfile("gotk4/closures")
}
