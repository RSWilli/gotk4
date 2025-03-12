package typesystem

import (
	"log"

	"github.com/diamondburned/gotk4/gir"
)

type girType interface {
	GetInfoAttrs() gir.InfoAttrs
}

type context interface {
	goName() string
	majorVersion() int
	findType(t *gir.Type) Type
	findAnyType(t gir.AnyType) Type

	skipType(t girType) bool
}

type namespaceContext struct {
	skipTypeFunc func(t girType) bool
	*Namespace
}

func (c namespaceContext) goName() string {
	return c.GoName
}

func (c namespaceContext) majorVersion() int {
	return c.Version.Major
}

func (c namespaceContext) skipType(t girType) bool {
	return c.skipTypeFunc(t)
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
