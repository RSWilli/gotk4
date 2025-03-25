package typesystem

import "github.com/diamondburned/gotk4/gir"

type Config struct {
	// Namespaces contains the configuration for the given versioned namespace (e.g. GLib-2). If a key is missing in this map,
	// then the typesystem will resolve everything in that namespace
	Namespaces map[string]NamespaceConfig
}

type NamespaceConfig struct {
	// Ignored signifies that the whole namespace should be treated as not existing
	Ignored bool

	// MinVersion declares the minimal version that should be supported in
	// type resolution. Everything that is deprecated longer than this version will be ignored.
	MinVersion string

	IgnoredDefinitions []IgnoreFunc

	// ManualTypes contains the gir name to a manual type override that will be imported instead of generated
	ManualTypes []Type
}

func (c NamespaceConfig) getEnv(namespace *Namespace) *env {
	var err error
	var v gir.Version

	if c.MinVersion != "" {
		v, err = gir.ParseVersion(c.MinVersion)

		if err != nil {
			panic(err)
		}
	}

	// TODO: do we need to add the manual types to the ignored functions?

	return &env{
		minVersion:     v,
		ignore:         ignoreOr(c.IgnoredDefinitions...),
		namespace:      namespace,
		compareParams:  nil,
		compareReturns: nil,
	}
}
