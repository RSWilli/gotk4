package typesystem

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/gir"
)

type Config struct {
	// MinVersions maps the versioned package name (e.g. GLib-2) to the minimal version that should be supported in
	// type resolution. Everything that is deprecated longer than this version will be ignored.
	MinVersions map[string]string
}

// getSkipFuncs returns a map that contains every skip func for each of the given namespaces
func (c *Config) getSkipFuncs(repos []*repoWithIncludes) map[versionedNamespace]skipFunc {
	m := make(map[versionedNamespace]skipFunc)

	for _, repo := range repos {
		for _, ns := range repo.namespaces {
			if minVersion, ok := c.MinVersions[fmt.Sprintf("%s-%d", ns.Name, ns.Version.Major)]; ok {
				v, err := gir.ParseVersion(minVersion)

				if err != nil {
					log.Panicf("received invalid minVersion %s: %v", minVersion, err)
				}

				m[ns.versionedName] = skipDeprecatedLongerThan(ns.versionedName.String(), v)
			} else {
				m[ns.versionedName] = defaultSkipFunc
			}
		}
	}

	return m
}
