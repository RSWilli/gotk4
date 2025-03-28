package typesystem

import (
	"fmt"
	"log/slog"

	"github.com/diamondburned/gotk4/gir"
)

type Config struct {
	// Primitives contains additional primitives, e.g. GType, which is referenced in the GIRs without a namespace. This allows
	// for redirecting these types to a go package instead of global names.
	Primitives []Type

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

	// ManualTypes contains the gir name to a manual type override that will not be generated. Themanual type
	// must be in the same go package as the generator would place it.
	ManualTypes []Type
}

func (cfg Config) getNamespaceEnv(namespace *Namespace) *env {
	nsCfg := cfg.Namespaces[fmt.Sprintf("%s-%d", namespace.Name, namespace.Version.Major)]

	if nsCfg.Ignored {
		return nil
	}

	var err error
	var v gir.Version

	if nsCfg.MinVersion != "" {
		v, err = gir.ParseVersion(nsCfg.MinVersion)

		if err != nil {
			panic(err)
		}
	}

	return &env{
		cfg:            cfg,
		nsCfg:          nsCfg,
		minVersion:     v,
		ignore:         ignoreOr(nsCfg.IgnoredDefinitions...),
		namespace:      namespace,
		compareParams:  nil,
		compareReturns: nil,
		logger:         slog.Default().With(slog.String("namespace", namespace.v.String())),
	}
}
