package typesystem

import (
	"log"

	"github.com/diamondburned/gotk4/gir"
)

type girType interface {
	GetInfoAttrs() gir.InfoAttrs
}

type skipFunc func(gt girType) bool

func defaultSkipFunc(gt girType) bool {
	attrs := gt.GetInfoAttrs()

	return !attrs.IsIntrospectable()
}

func skipDeprecatedLongerThan(namespace string, minVersion gir.Version) func(girType) bool {
	// minVersion := gir.Version{
	// 	Major: namespaceV.Major,
	// 	Minor: max(0, namespaceV.Minor-10),
	// 	Patch: 0,
	// }

	return func(gt girType) bool {
		attrs := gt.GetInfoAttrs()

		if attrs.Deprecated && attrs.DeprecatedVersion.Lte(minVersion) {
			log.Printf("skipping type in %s that is deprecated since %s, min allowed version: %s", namespace, attrs.DeprecatedVersion, minVersion)
			return true
		}

		return defaultSkipFunc(gt)
	}
}
